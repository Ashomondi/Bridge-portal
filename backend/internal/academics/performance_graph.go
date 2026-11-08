package academics

import (
	"context"
	"net/http"
	"sort"
	"strings"
)

// StudentSubjectResult is one parent-facing subject result. It contains no
// class rank or position (AC-19).
type StudentSubjectResult struct {
	SubjectID     string  `json:"subject_id"`
	SubjectName   string  `json:"subject_name"`
	Score         float64 `json:"score"`
	CBCLevel      string  `json:"cbc_level"`
	LevelValue    int     `json:"level_value"`
	IsAmended     bool    `json:"is_amended"`
	AmendmentNote string  `json:"amendment_note,omitempty"`
}

// StudentResultsResponse is the last released snapshot for one term.
type StudentResultsResponse struct {
	StudentID     string                 `json:"student_id"`
	TermID        string                 `json:"term_id"`
	TermName      string                 `json:"term_name"`
	Status        string                 `json:"status"`
	ReportComment string                 `json:"report_comment,omitempty"`
	Subjects      []StudentSubjectResult `json:"subjects"`
}

// PerformanceSubject is one subject's CBC level in the multi-term graph.
type PerformanceSubject struct {
	SubjectID     string `json:"subject_id"`
	SubjectName   string `json:"subject_name"`
	CBCLevel      string `json:"cbc_level"`
	LevelValue    int    `json:"level_value"`
	IsAmended     bool   `json:"is_amended"`
	AmendmentNote string `json:"amendment_note,omitempty"`
}

// TermPerformance groups one term's subject levels.
type TermPerformance struct {
	TermID     string               `json:"term_id"`
	TermName   string               `json:"term_name"`
	Year       int                  `json:"year"`
	TermNumber int                  `json:"term_number"`
	Subjects   []PerformanceSubject `json:"subjects"`
}

// PerformanceHistoryResponse is the non-ranking CBC graph data (C42, C43).
type PerformanceHistoryResponse struct {
	StudentID string            `json:"student_id"`
	Terms     []TermPerformance `json:"terms"`
}

// GetStudentResults implements GET /api/v1/students/{id}/results. While a
// class is reopened this returns the pre-reopen snapshot because the snapshot
// table is only replaced on re-release (AC-18).
func (s *Service) GetStudentResults(ctx context.Context, studentID, termID string) (StudentResultsResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return StudentResultsResponse{}, err
	}
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return StudentResultsResponse{}, errValidation("student_id is required.")
	}
	var out StudentResultsResponse
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		if err := s.authorizeStudentAccess(ctx, tx, actor, studentID); err != nil {
			return err
		}
		termID = strings.TrimSpace(termID)
		if termID == "" {
			latest, err := latestTermWithSnapshot(ctx, tx, actor.SchoolID, studentID)
			if err != nil {
				return err
			}
			termID = latest
		}
		term, err := tx.TermByID(ctx, actor.SchoolID, termID)
		if err != nil {
			return err
		}
		snapshots, err := tx.SnapshotsForStudentTerm(ctx, actor.SchoolID, studentID, termID)
		if err != nil {
			return err
		}
		if len(snapshots) == 0 {
			return errNotFound("Results have not been released for this student and term.")
		}
		out = StudentResultsResponse{
			StudentID: studentID,
			TermID:    termID,
			TermName:  termName(term),
			Status:    "released",
			Subjects:  make([]StudentSubjectResult, 0, len(snapshots)),
		}
		for _, snapshot := range snapshots {
			name := snapshot.SubjectID
			if subject, err := tx.SubjectByID(ctx, actor.SchoolID, snapshot.SubjectID); err == nil {
				name = subject.Name
			}
			if out.ReportComment == "" {
				out.ReportComment = snapshot.ReportComment
			}
			out.Subjects = append(out.Subjects, StudentSubjectResult{
				SubjectID:     snapshot.SubjectID,
				SubjectName:   name,
				Score:         snapshot.Score,
				CBCLevel:      snapshot.CBCLevel,
				LevelValue:    levelValue(snapshot.CBCLevel),
				IsAmended:     snapshot.IsAmended,
				AmendmentNote: snapshot.AmendmentNote,
			})
		}
		sort.Slice(out.Subjects, func(i, j int) bool { return out.Subjects[i].SubjectName < out.Subjects[j].SubjectName })
		return nil
	})
	return out, err
}

// GetPerformanceHistory implements GET /api/v1/students/{id}/performance-history.
// It plots CBC levels per subject per term and never computes or returns class
// rank, position or a leaderboard (AC-19).
func (s *Service) GetPerformanceHistory(ctx context.Context, studentID string) (PerformanceHistoryResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return PerformanceHistoryResponse{}, err
	}
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return PerformanceHistoryResponse{}, errValidation("student_id is required.")
	}
	var out PerformanceHistoryResponse
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		if err := s.authorizeStudentAccess(ctx, tx, actor, studentID); err != nil {
			return err
		}
		snapshots, err := tx.SnapshotsForStudent(ctx, actor.SchoolID, studentID)
		if err != nil {
			return err
		}
		byTerm := map[string][]Snapshot{}
		for _, snapshot := range snapshots {
			byTerm[snapshot.TermID] = append(byTerm[snapshot.TermID], snapshot)
		}
		out = PerformanceHistoryResponse{StudentID: studentID, Terms: []TermPerformance{}}
		for termID, termSnapshots := range byTerm {
			term, err := tx.TermByID(ctx, actor.SchoolID, termID)
			if err != nil {
				continue
			}
			performance := TermPerformance{
				TermID:     termID,
				TermName:   termName(term),
				Year:       term.Year,
				TermNumber: term.TermNumber,
				Subjects:   make([]PerformanceSubject, 0, len(termSnapshots)),
			}
			for _, snapshot := range termSnapshots {
				name := snapshot.SubjectID
				if subject, err := tx.SubjectByID(ctx, actor.SchoolID, snapshot.SubjectID); err == nil {
					name = subject.Name
				}
				performance.Subjects = append(performance.Subjects, PerformanceSubject{
					SubjectID:     snapshot.SubjectID,
					SubjectName:   name,
					CBCLevel:      snapshot.CBCLevel,
					LevelValue:    levelValue(snapshot.CBCLevel),
					IsAmended:     snapshot.IsAmended,
					AmendmentNote: snapshot.AmendmentNote,
				})
			}
			sort.Slice(performance.Subjects, func(i, j int) bool {
				return performance.Subjects[i].SubjectName < performance.Subjects[j].SubjectName
			})
			out.Terms = append(out.Terms, performance)
		}
		sort.Slice(out.Terms, func(i, j int) bool {
			if out.Terms[i].Year != out.Terms[j].Year {
				return out.Terms[i].Year < out.Terms[j].Year
			}
			return out.Terms[i].TermNumber < out.Terms[j].TermNumber
		})
		return nil
	})
	return out, err
}

// StudentMarksSummary is the internal-api fact pack for Sos's AI assistant.
type StudentMarksSummary struct {
	StudentID string               `json:"student_id"`
	TermID    string               `json:"term_id"`
	Marks     []SubjectMarkSummary `json:"marks"`
}

// SubjectMarkSummary is one subject mark and derived CBC level.
type SubjectMarkSummary struct {
	SubjectID   string  `json:"subject_id"`
	SubjectName string  `json:"subject_name"`
	Score       float64 `json:"score"`
	CBCLevel    string  `json:"cbc_level"`
}

// GetStudentAttendanceSummary exposes the attendance aggregate to Ian's
// internalapi (no identity is needed; the caller has already resolved the
// tenant and verified guardian ownership).
func (s *Service) GetStudentAttendanceSummary(ctx context.Context, schoolID, studentID string) (AttendanceSummary, error) {
	var out AttendanceSummary
	err := s.tx(ctx, schoolID, func(tx Store) error {
		if _, err := tx.StudentByID(ctx, schoolID, studentID); err != nil {
			return err
		}
		var err error
		out, err = tx.AttendanceForStudent(ctx, schoolID, studentID)
		return err
	})
	return out, err
}

// GetStudentMarksSummary exposes the marks aggregate to Ian's internalapi.
func (s *Service) GetStudentMarksSummary(ctx context.Context, schoolID, studentID, termID string) (StudentMarksSummary, error) {
	var out StudentMarksSummary
	err := s.tx(ctx, schoolID, func(tx Store) error {
		schoolStudent, err := tx.StudentByID(ctx, schoolID, studentID)
		if err != nil {
			return err
		}
		termID = strings.TrimSpace(termID)
		if termID == "" {
			term, err := tx.TermForDate(ctx, schoolID, timeNow())
			if err != nil {
				return errValidation("A term_id is required.")
			}
			termID = term.ID
		}
		marks, err := tx.MarksForStudentTerm(ctx, schoolID, studentID, termID)
		if err != nil {
			return err
		}
		scale, err := gradingScaleForRelease(ctx, tx, schoolID, releaseScaleID(ctx, tx, schoolID, schoolStudent.CurrentClassID, termID))
		if err != nil {
			return err
		}
		out = StudentMarksSummary{StudentID: studentID, TermID: termID, Marks: make([]SubjectMarkSummary, 0, len(marks))}
		for _, mark := range marks {
			name := mark.SubjectID
			if subject, err := tx.SubjectByID(ctx, schoolID, mark.SubjectID); err == nil {
				name = subject.Name
			}
			out.Marks = append(out.Marks, SubjectMarkSummary{
				SubjectID:   mark.SubjectID,
				SubjectName: name,
				Score:       mark.Score,
				CBCLevel:    resolveCBCLevel(scale, mark.Score),
			})
		}
		return nil
	})
	return out, err
}

func (s *Service) authorizeStudentAccess(ctx context.Context, tx Store, actor Identity, studentID string) error {
	if _, err := tx.StudentByID(ctx, actor.SchoolID, studentID); err != nil {
		return err
	}
	if guardian, _ := tx.IsGuardian(ctx, actor.SchoolID, actor.UserID, studentID); guardian {
		return nil
	}
	switch actor.Role {
	case RoleTeacher, RoleDean, RolePrincipal, RoleDeputy:
		if allowed, _ := tx.TeacherCanViewStudent(ctx, actor.SchoolID, actor.MembershipID, studentID); allowed {
			return nil
		}
	}
	return errForbidden(CodeForbidden, "You may only view results for a child you are a guardian of or a learner in your class.")
}

func latestTermWithSnapshot(ctx context.Context, tx Store, schoolID, studentID string) (string, error) {
	snapshots, err := tx.SnapshotsForStudent(ctx, schoolID, studentID)
	if err != nil {
		return "", err
	}
	if len(snapshots) == 0 {
		return "", errNotFound("No released results exist for this student.")
	}
	terms := map[string]Term{}
	for _, snapshot := range snapshots {
		if _, ok := terms[snapshot.TermID]; ok {
			continue
		}
		term, err := tx.TermByID(ctx, schoolID, snapshot.TermID)
		if err == nil {
			terms[snapshot.TermID] = term
		}
	}
	var latest Term
	for _, term := range terms {
		if latest.ID == "" || term.Year > latest.Year || (term.Year == latest.Year && term.TermNumber > latest.TermNumber) {
			latest = term
		}
	}
	if latest.ID == "" {
		return "", errNotFound("No released results exist for this student.")
	}
	return latest.ID, nil
}

func termName(term Term) string {
	return "Term " + itoa(term.TermNumber) + " " + itoa(term.Year)
}

func (s *Service) handleStudentResults(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetStudentResults(r.Context(), r.PathValue("id"), r.URL.Query().Get("term_id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handlePerformanceHistory(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetPerformanceHistory(r.Context(), r.PathValue("id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodGet, "/api/v1/students/{id}/results", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleStudentResults(w, r)
	})
	registerRoute(http.MethodGet, "/api/v1/students/{id}/performance-history", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handlePerformanceHistory(w, r)
	})
}
