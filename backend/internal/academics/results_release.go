package academics

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ReleaseRequest releases one class's results for a term.
type ReleaseRequest struct {
	TermID string `json:"term_id"`
}

// ReleaseResult is returned after a successful hard-gated release.
type ReleaseResult struct {
	ClassID               string `json:"class_id"`
	TermID                string `json:"term_id"`
	Status                string `json:"status"`
	GradingScaleVersionID string `json:"grading_scale_version_id"`
	SnapshotCount         int    `json:"snapshot_count"`
	GuardiansNotified     int    `json:"guardians_notified"`
}

// ReopenRequest asks to reopen a released class for corrections.
type ReopenRequest struct {
	TermID string `json:"term_id"`
	Reason string `json:"reason"`
}

// ReopenResult reports the new state.
type ReopenResult struct {
	ClassID string `json:"class_id"`
	TermID  string `json:"term_id"`
	Status  string `json:"status"`
}

// CorrectionNoteInput is the teacher's mandatory explanation for one changed mark.
type CorrectionNoteInput struct {
	StudentID   string `json:"student_id"`
	SubjectID   string `json:"subject_id"`
	TeacherNote string `json:"teacher_note"`
}

// RereleaseRequest re-publishes corrected marks.
type RereleaseRequest struct {
	TermID          string                `json:"term_id"`
	CorrectionNotes []CorrectionNoteInput `json:"correction_notes"`
}

// RereleaseResult reports the affected scope of a re-release.
type RereleaseResult struct {
	ClassID                  string `json:"class_id"`
	TermID                   string `json:"term_id"`
	Status                   string `json:"status"`
	AmendedMarks             int    `json:"amended_marks"`
	AffectedStudentsNotified int    `json:"affected_students_notified"`
}

// ReleaseResults implements POST /api/v1/classes/{id}/results/release.
// The class teacher cannot release until every subject has a mark for every
// enrolled learner. There is no override in Release 1 (AC-16).
func (s *Service) ReleaseResults(ctx context.Context, classID string, req ReleaseRequest) (ReleaseResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return ReleaseResult{}, err
	}
	classID = strings.TrimSpace(classID)
	req.TermID = strings.TrimSpace(req.TermID)
	if classID == "" || req.TermID == "" {
		return ReleaseResult{}, errValidation("class_id and term_id are required.")
	}

	var result ReleaseResult
	var guardians []string
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		allowed, err := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeClassTeacherScope, "Only the class teacher for this class may release results.")
		}
		if _, err := tx.ClassByID(ctx, actor.SchoolID, classID); err != nil {
			return err
		}
		term, err := tx.TermByID(ctx, actor.SchoolID, req.TermID)
		if err != nil {
			return err
		}
		currentYear, err := tx.CurrentAcademicYearID(ctx, actor.SchoolID)
		if err != nil {
			return errValidation("An active academic year is required before releasing results.")
		}
		if term.AcademicYearID != currentYear {
			return errValidation("Results can only be released for a term in the current academic year.")
		}

		previousRelease, previousErr := tx.ResultRelease(ctx, actor.SchoolID, classID, req.TermID)
		if previousErr == nil {
			switch previousRelease.Status {
			case "released", "rereleased":
				return errConflict(CodeResultsLocked, "Results for this class and term have already been released.")
			case "reopened":
				return errConflict(CodeResultsReopened, "Results are reopened; correct the marks and re-release them with a teacher's note.")
			}
		}

		students, err := tx.EnrolledStudents(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		subjects, err := tx.SubjectsForClass(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		marks, err := tx.MarksForClassTerm(ctx, actor.SchoolID, classID, req.TermID)
		if err != nil {
			return err
		}

		markByKey := map[string]Mark{}
		for _, mark := range marks {
			markByKey[key(mark.StudentID, mark.SubjectID)] = mark
		}

		type missingSubject struct {
			SubjectID            string `json:"subject_id"`
			SubjectName          string `json:"subject_name"`
			MissingStudentsCount int    `json:"missing_students_count"`
		}
		missingCount := 0
		missingSubjects := []missingSubject{}
		for _, subject := range subjects {
			missing := 0
			for _, student := range students {
				if _, ok := markByKey[key(student.StudentID, subject.ID)]; !ok {
					missing++
				}
			}
			if missing > 0 {
				missingCount += missing
				missingSubjects = append(missingSubjects, missingSubject{
					SubjectID:            subject.ID,
					SubjectName:          subject.Name,
					MissingStudentsCount: missing,
				})
			}
		}
		if missingCount > 0 {
			return &APIError{
				Status:  http.StatusUnprocessableEntity,
				Code:    CodeResultsIncomplete,
				Message: "Cannot release results until every subject has marks for every student.",
				Details: map[string]any{
					"missing_count":    missingCount,
					"missing_subjects": missingSubjects,
				},
			}
		}

		scale, err := tx.ActiveGradingScale(ctx, actor.SchoolID)
		if err != nil {
			return errValidation("Configure a CBC grading scale before releasing results.")
		}
		comments, _ := tx.ReportCommentsForClassTerm(ctx, actor.SchoolID, classID, req.TermID)
		commentByStudent := map[string]string{}
		for _, c := range comments {
			commentByStudent[c.StudentID] = c.CommentText
		}

		release := previousRelease
		if previousErr != nil {
			release = ResultRelease{ID: newUUID(), ClassID: classID, TermID: req.TermID}
		}
		now := time.Now()
		release.Status = "released"
		release.GradingScaleVersionID = scale.ID
		release.ReleasedBy = actor.UserID
		release.ReleasedAt = &now
		release.ReopenedBy = ""
		release.ReopenedAt = nil
		release.RereleasedBy = ""
		release.RereleasedAt = nil

		snapshots := make([]Snapshot, 0, len(students)*len(subjects))
		for _, student := range students {
			for _, subject := range subjects {
				mark := markByKey[key(student.StudentID, subject.ID)]
				snapshots = append(snapshots, Snapshot{
					StudentID:       student.StudentID,
					TermID:          req.TermID,
					SubjectID:       subject.ID,
					Score:           mark.Score,
					CBCLevel:        resolveCBCLevel(scale, mark.Score),
					SnapshotVersion: 1,
					ReportComment:   commentByStudent[student.StudentID],
					PublishedAt:     now,
				})
			}
		}
		if err := tx.SaveResultRelease(ctx, actor.SchoolID, release); err != nil {
			return err
		}
		if err := tx.ReplaceSnapshots(ctx, actor.SchoolID, release.ID, snapshots); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, actor, "results_released", "class", classID, map[string]any{
			"term_id":                  req.TermID,
			"grading_scale_version_id": scale.ID,
			"snapshot_count":           len(snapshots),
			"student_count":            len(students),
		}); err != nil {
			return err
		}
		guardians, _ = tx.GuardianIDsForClass(ctx, actor.SchoolID, classID)
		result = ReleaseResult{
			ClassID:               classID,
			TermID:                req.TermID,
			Status:                "released",
			GradingScaleVersionID: scale.ID,
			SnapshotCount:         len(snapshots),
		}
		return nil
	})
	if err != nil {
		return ReleaseResult{}, err
	}
	notified, _ := s.notifyMany(ctx, actor.SchoolID, guardians, "academics.results",
		"Results released",
		"Your child's results for this term are now available in the app.",
		map[string]any{"class_id": classID, "term_id": req.TermID})
	result.GuardiansNotified = notified
	return result, nil
}

// ReopenResults implements POST /api/v1/classes/{id}/results/reopen. The last
// released snapshot stays visible to parents until re-release (AC-18).
func (s *Service) ReopenResults(ctx context.Context, classID string, req ReopenRequest) (ReopenResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return ReopenResult{}, err
	}
	classID = strings.TrimSpace(classID)
	req.TermID = strings.TrimSpace(req.TermID)
	if classID == "" || req.TermID == "" {
		return ReopenResult{}, errValidation("class_id and term_id are required.")
	}
	var out ReopenResult
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		allowed, err := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeClassTeacherScope, "Only the class teacher for this class may reopen results.")
		}
		release, err := tx.ResultRelease(ctx, actor.SchoolID, classID, req.TermID)
		if err != nil {
			return errConflict(CodeResultsNotReleased, "Results have not been released for this class and term.")
		}
		if release.Status != "released" && release.Status != "rereleased" {
			return errConflict(CodeResultsNotReleased, "Only released results can be reopened.")
		}
		now := time.Now()
		release.Status = "reopened"
		release.ReopenedBy = actor.UserID
		release.ReopenedAt = &now
		if err := tx.SaveResultRelease(ctx, actor.SchoolID, release); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, actor, "results_reopened", "class", classID, map[string]any{
			"term_id": req.TermID,
			"reason":  strings.TrimSpace(req.Reason),
		}); err != nil {
			return err
		}
		out = ReopenResult{ClassID: classID, TermID: req.TermID, Status: "reopened"}
		return nil
	})
	return out, err
}

// RereleaseResults implements POST /api/v1/classes/{id}/results/re-release.
// Every changed mark needs a non-empty teacher note; only guardians of changed
// students are notified; unaffected parents keep the old snapshot (AC-18).
func (s *Service) RereleaseResults(ctx context.Context, classID string, req RereleaseRequest) (RereleaseResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return RereleaseResult{}, err
	}
	classID = strings.TrimSpace(classID)
	req.TermID = strings.TrimSpace(req.TermID)
	if classID == "" || req.TermID == "" {
		return RereleaseResult{}, errValidation("class_id and term_id are required.")
	}

	var out RereleaseResult
	var affectedGuardians []string
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		allowed, err := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeClassTeacherScope, "Only the class teacher for this class may re-release results.")
		}
		release, err := tx.ResultRelease(ctx, actor.SchoolID, classID, req.TermID)
		if err != nil {
			return errConflict(CodeResultsNotReleased, "Results have not been released for this class and term.")
		}
		if release.Status != "reopened" {
			return errConflict(CodeResultsNotReopened, "Results must be reopened before they can be re-released.")
		}
		scale, err := gradingScaleForRelease(ctx, tx, actor.SchoolID, release.GradingScaleVersionID)
		if err != nil {
			return errValidation("The locked grading scale for this release is unavailable.")
		}

		currentMarks, err := tx.MarksForClassTerm(ctx, actor.SchoolID, classID, req.TermID)
		if err != nil {
			return err
		}
		previous, err := tx.SnapshotsForRelease(ctx, actor.SchoolID, release.ID)
		if err != nil {
			return err
		}
		previousByKey := map[string]Snapshot{}
		for _, snapshot := range previous {
			previousByKey[key(snapshot.StudentID, snapshot.SubjectID)] = snapshot
		}

		type changedMark struct {
			StudentID string
			SubjectID string
		}
		changed := map[string]changedMark{}
		for _, mark := range currentMarks {
			old, exists := previousByKey[key(mark.StudentID, mark.SubjectID)]
			level := resolveCBCLevel(scale, mark.Score)
			if !exists || math.Abs(old.Score-mark.Score) > 0.0001 || old.CBCLevel != level {
				changed[key(mark.StudentID, mark.SubjectID)] = changedMark{StudentID: mark.StudentID, SubjectID: mark.SubjectID}
			}
		}

		notes := map[string]string{}
		for _, note := range req.CorrectionNotes {
			studentID := strings.TrimSpace(note.StudentID)
			subjectID := strings.TrimSpace(note.SubjectID)
			if studentID == "" || subjectID == "" {
				return errValidation("Every correction note needs student_id and subject_id.")
			}
			notes[key(studentID, subjectID)] = strings.TrimSpace(note.TeacherNote)
		}

		missing := []map[string]string{}
		for k := range changed {
			if notes[k] == "" {
				parts := strings.SplitN(k, "|", 2)
				missing = append(missing, map[string]string{"student_id": parts[0], "subject_id": parts[1]})
			}
		}
		if len(missing) > 0 {
			return &APIError{
				Status:  http.StatusUnprocessableEntity,
				Code:    CodeRereleaseNote,
				Message: "Every changed mark requires a teacher_note explaining the correction.",
				Details: map[string]any{
					"missing_notes": missing,
					"changed_count": len(changed),
				},
			}
		}

		comments, _ := tx.ReportCommentsForClassTerm(ctx, actor.SchoolID, classID, req.TermID)
		commentByStudent := map[string]string{}
		for _, c := range comments {
			commentByStudent[c.StudentID] = c.CommentText
		}
		subjectNames := map[string]string{}
		subjects, _ := tx.SubjectsForClass(ctx, actor.SchoolID, classID)
		for _, subject := range subjects {
			subjectNames[subject.ID] = subject.Name
		}

		now := time.Now()
		snapshots := make([]Snapshot, 0, len(currentMarks))
		amended := 0
		for _, mark := range currentMarks {
			k := key(mark.StudentID, mark.SubjectID)
			old, exists := previousByKey[k]
			version := old.SnapshotVersion
			isAmended := old.IsAmended
			note := old.AmendmentNote
			if _, ok := changed[k]; ok {
				version++
				isAmended = true
				note = notes[k]
				amended++
			}
			if version == 0 {
				version = 1
			}
			snapshots = append(snapshots, Snapshot{
				StudentID:       mark.StudentID,
				TermID:          req.TermID,
				SubjectID:       mark.SubjectID,
				Score:           mark.Score,
				CBCLevel:        resolveCBCLevel(scale, mark.Score),
				IsAmended:       isAmended,
				AmendmentNote:   note,
				SnapshotVersion: version,
				ReportComment:   commentByStudent[mark.StudentID],
				PublishedAt:     now,
			})
			_ = exists
		}
		if err := tx.ReplaceSnapshots(ctx, actor.SchoolID, release.ID, snapshots); err != nil {
			return err
		}

		release.Status = "rereleased"
		release.RereleasedBy = actor.UserID
		release.RereleasedAt = &now
		if err := tx.SaveResultRelease(ctx, actor.SchoolID, release); err != nil {
			return err
		}

		affectedStudents := map[string]bool{}
		for _, item := range changed {
			affectedStudents[item.StudentID] = true
		}
		studentIDs := make([]string, 0, len(affectedStudents))
		for studentID := range affectedStudents {
			studentIDs = append(studentIDs, studentID)
		}
		sort.Strings(studentIDs)
		for _, studentID := range studentIDs {
			users, _ := tx.GuardianIDsForStudent(ctx, actor.SchoolID, studentID)
			affectedGuardians = append(affectedGuardians, users...)
		}

		amendedDetails := make([]map[string]string, 0, len(changed))
		for k := range changed {
			parts := strings.SplitN(k, "|", 2)
			amendedDetails = append(amendedDetails, map[string]string{
				"student_id":   parts[0],
				"subject_id":   parts[1],
				"subject_name": subjectNames[parts[1]],
				"teacher_note": notes[k],
			})
		}
		if err := s.audit(ctx, tx, actor, "results_rereleased", "class", classID, map[string]any{
			"term_id":           req.TermID,
			"amended_marks":     amended,
			"affected_students": len(affectedStudents),
			"amendments":        amendedDetails,
		}); err != nil {
			return err
		}
		out = RereleaseResult{
			ClassID:                  classID,
			TermID:                   req.TermID,
			Status:                   "rereleased",
			AmendedMarks:             amended,
			AffectedStudentsNotified: len(affectedStudents),
		}
		return nil
	})
	if err != nil {
		return RereleaseResult{}, err
	}
	// Best-effort fan-out; the count above is the number of affected students,
	// not the number of guardian notifications.
	_, _ = s.notifyMany(ctx, actor.SchoolID, affectedGuardians, "academics.results.amended",
		"Corrected results published",
		"A corrected mark has been published for your child. The teacher's explanation is shown with the result.",
		map[string]any{"class_id": classID, "term_id": req.TermID})
	return out, nil
}

func (s *Service) handleReleaseResults(w http.ResponseWriter, r *http.Request) {
	var req ReleaseRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.ReleaseResults(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleReopenResults(w http.ResponseWriter, r *http.Request) {
	var req ReopenRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.ReopenResults(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleRereleaseResults(w http.ResponseWriter, r *http.Request) {
	var req RereleaseRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.RereleaseResults(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodPost, "/api/v1/classes/{id}/results/release", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleReleaseResults(w, r)
	})
	registerRoute(http.MethodPost, "/api/v1/classes/{id}/results/reopen", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleReopenResults(w, r)
	})
	registerRoute(http.MethodPost, "/api/v1/classes/{id}/results/re-release", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleRereleaseResults(w, r)
	})
}
