package academics

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

// RosterResponse is the full roster visible only to authorised staff of a class.
type RosterResponse struct {
	ClassID  string            `json:"class_id"`
	Students []EnrolledStudent `json:"students"`
}

// GetRoster implements GET /api/v1/classes/{id}/roster. The caller must be the
// class teacher, a subject teacher assigned to the class, the Dean, Principal
// or Deputy (Docs/Ikonex.md section 4).
func (s *Service) GetRoster(ctx context.Context, classID string) (RosterResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return RosterResponse{}, err
	}
	classID = strings.TrimSpace(classID)
	if classID == "" {
		return RosterResponse{}, errValidation("class_id is required.")
	}
	var out RosterResponse
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		if _, err := tx.ClassByID(ctx, actor.SchoolID, classID); err != nil {
			return err
		}
		allowed := false
		switch actor.Role {
		case RoleDean, RolePrincipal, RoleDeputy:
			allowed = true
		}
		if !allowed {
			allowed, _ = tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		}
		if !allowed {
			allowed, _ = tx.HasAnySubjectAssignment(ctx, actor.SchoolID, actor.MembershipID, classID)
		}
		if !allowed {
			return errForbidden(CodeForbidden, "You do not have permission to view this class's full roster.")
		}
		students, err := tx.EnrolledStudents(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		out = RosterResponse{ClassID: classID, Students: students}
		return nil
	})
	return out, err
}

// SubjectMeanStat is an anonymised per-subject aggregate. It deliberately has
// no student identifiers (Docs/Ikonex.md section 4).
type SubjectMeanStat struct {
	SubjectID           string         `json:"subject_id"`
	SubjectName         string         `json:"subject_name"`
	MarkedStudents      int            `json:"marked_students"`
	MeanScore           float64        `json:"mean_score"`
	PredominantCBCLevel string         `json:"predominant_cbc_level"`
	LevelDistribution   map[string]int `json:"level_distribution"`
}

// ClassStats is the anonymised class aggregate returned to any teacher.
type ClassStats struct {
	ClassID        string            `json:"class_id"`
	TermID         string            `json:"term_id"`
	TotalStudents  int               `json:"total_students"`
	ClassMean      float64           `json:"class_mean"`
	ClassMeanScore float64           `json:"class_mean_score"`
	SubjectMeans   []SubjectMeanStat `json:"subject_means"`
}

// GetClassStats implements GET /api/v1/classes/{id}/stats. Any teacher in the
// school may call it, but the response contains only class/subject means and
// level distributions — never a student name, admission number or mark row.
func (s *Service) GetClassStats(ctx context.Context, classID, termID string) (ClassStats, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return ClassStats{}, err
	}
	classID = strings.TrimSpace(classID)
	if classID == "" {
		return ClassStats{}, errValidation("class_id is required.")
	}
	switch actor.Role {
	case RoleTeacher, RoleDean, RolePrincipal, RoleDeputy:
	default:
		if !contains(actor.Permissions, PermClassStatsView) {
			return ClassStats{}, errForbidden(CodeForbidden, "Only teachers may view class statistics.")
		}
	}

	var out ClassStats
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		if _, err := tx.ClassByID(ctx, actor.SchoolID, classID); err != nil {
			return err
		}
		var term Term
		if strings.TrimSpace(termID) != "" {
			term, err = tx.TermByID(ctx, actor.SchoolID, strings.TrimSpace(termID))
		} else {
			term, err = tx.TermForDate(ctx, actor.SchoolID, timeNow())
		}
		if err != nil {
			return errValidation("A term_id is required to view class statistics.")
		}

		students, err := tx.EnrolledStudents(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		subjects, err := tx.SubjectsForClass(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		marks, err := tx.MarksForClassTerm(ctx, actor.SchoolID, classID, term.ID)
		if err != nil {
			return err
		}

		scale, err := gradingScaleForRelease(ctx, tx, actor.SchoolID, releaseScaleID(ctx, tx, actor.SchoolID, classID, term.ID))
		if err != nil {
			return err
		}

		enrolled := map[string]bool{}
		for _, st := range students {
			enrolled[st.StudentID] = true
		}
		bySubject := map[string][]Mark{}
		var all []Mark
		for _, m := range marks {
			if !enrolled[m.StudentID] {
				continue
			}
			bySubject[m.SubjectID] = append(bySubject[m.SubjectID], m)
			all = append(all, m)
		}

		out = ClassStats{
			ClassID:       classID,
			TermID:        term.ID,
			TotalStudents: len(students),
			SubjectMeans:  []SubjectMeanStat{},
		}
		if len(all) > 0 {
			var sum float64
			for _, m := range all {
				sum += m.Score
			}
			out.ClassMeanScore = round2(sum / float64(len(all)))
			out.ClassMean = out.ClassMeanScore
		}

		sort.Slice(subjects, func(i, j int) bool { return subjects[i].Name < subjects[j].Name })
		for _, subject := range subjects {
			subjectMarks := bySubject[subject.ID]
			stat := SubjectMeanStat{
				SubjectID:         subject.ID,
				SubjectName:       subject.Name,
				MarkedStudents:    len(subjectMarks),
				LevelDistribution: map[string]int{LevelEE: 0, LevelME: 0, LevelAE: 0, LevelBE: 0},
			}
			if len(subjectMarks) > 0 {
				var sum float64
				for _, m := range subjectMarks {
					sum += m.Score
					stat.LevelDistribution[resolveCBCLevel(scale, m.Score)]++
				}
				stat.MeanScore = round2(sum / float64(len(subjectMarks)))
			}
			stat.PredominantCBCLevel = predominantLevel(stat.LevelDistribution)
			out.SubjectMeans = append(out.SubjectMeans, stat)
		}
		return nil
	})
	return out, err
}

func releaseScaleID(ctx context.Context, tx Store, schoolID, classID, termID string) string {
	release, err := tx.ResultRelease(ctx, schoolID, classID, termID)
	if err != nil {
		return ""
	}
	// Released, reopened and re-released terms all keep the scale locked at the
	// first release (AC-19).
	return release.GradingScaleVersionID
}

func predominantLevel(distribution map[string]int) string {
	best := ""
	bestCount := 0
	for _, level := range []string{LevelEE, LevelME, LevelAE, LevelBE} {
		if distribution[level] > bestCount {
			best = level
			bestCount = distribution[level]
		}
	}
	return best
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// timeNow is a variable so tests can be deterministic if needed.
var timeNow = func() time.Time { return time.Now() }

func (s *Service) handleRoster(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetRoster(r.Context(), r.PathValue("id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleClassStats(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetClassStats(r.Context(), r.PathValue("id"), r.URL.Query().Get("term_id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodGet, "/api/v1/classes/{id}/roster", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleRoster(w, r)
	})
	registerRoute(http.MethodGet, "/api/v1/classes/{id}/stats", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleClassStats(w, r)
	})
}
