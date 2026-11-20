package academics

import (
	"context"
	"net/http"
	"strings"
)

// MarkInput is one raw score supplied by a subject teacher.
type MarkInput struct {
	StudentID string  `json:"student_id"`
	Score     float64 `json:"score"`
}

// MarksRequest saves one subject's marks for a class and term.
type MarksRequest struct {
	TermID string      `json:"term_id"`
	Marks  []MarkInput `json:"marks"`
}

// MarksResult is returned after saving marks.
type MarksResult struct {
	ClassID    string `json:"class_id"`
	SubjectID  string `json:"subject_id"`
	TermID     string `json:"term_id"`
	SavedCount int    `json:"saved_count"`
}

// SaveMarks implements PUT /api/v1/classes/{id}/subjects/{sid}/marks.
// Only the currently assigned subject teacher may write, and writing is locked
// once the class results are released (AC-15, RESULTS_LOCKED_RELEASED).
func (s *Service) SaveMarks(ctx context.Context, classID, subjectID string, req MarksRequest) (MarksResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return MarksResult{}, err
	}
	classID = strings.TrimSpace(classID)
	subjectID = strings.TrimSpace(subjectID)
	req.TermID = strings.TrimSpace(req.TermID)
	if classID == "" || subjectID == "" || req.TermID == "" {
		return MarksResult{}, errValidation("class_id, subject_id and term_id are required.")
	}
	if len(req.Marks) == 0 {
		return MarksResult{}, errValidation("At least one mark is required.")
	}

	var result MarksResult
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		class, err := tx.ClassByID(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		subject, err := tx.SubjectByID(ctx, actor.SchoolID, subjectID)
		if err != nil {
			return err
		}
		if subject.GradeLevelID != class.GradeLevelID {
			return errValidation("This subject does not belong to the class's grade level.")
		}
		term, err := tx.TermByID(ctx, actor.SchoolID, req.TermID)
		if err != nil {
			return err
		}

		release, err := tx.ResultRelease(ctx, actor.SchoolID, classID, req.TermID)
		if err == nil && (release.Status == "released" || release.Status == "rereleased") {
			return errConflict(CodeResultsLocked, "Results for this class and term are released; reopen the results before editing marks.")
		}

		assignment, err := tx.SubjectAssignment(ctx, actor.SchoolID, classID, subjectID, term.AcademicYearID)
		if err != nil || assignment.TeacherMembershipID != actor.MembershipID {
			return errForbidden(CodeSubjectTeacherScope, "You are not the assigned subject teacher for this class and subject.")
		}

		students, err := tx.EnrolledStudents(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		enrolled := map[string]bool{}
		for _, student := range students {
			enrolled[student.StudentID] = true
		}

		seen := map[string]bool{}
		rows := make([]Mark, 0, len(req.Marks))
		for _, in := range req.Marks {
			in.StudentID = strings.TrimSpace(in.StudentID)
			if in.StudentID == "" || !enrolled[in.StudentID] {
				return errValidation("Marks include a student who is not enrolled in this class.")
			}
			if seen[in.StudentID] {
				return errValidation("A student may only appear once per marks submission.")
			}
			seen[in.StudentID] = true
			if in.Score < 0 || in.Score > 100 {
				return errValidation("Marks must be between 0 and 100.")
			}
			rows = append(rows, Mark{
				StudentID: in.StudentID,
				ClassID:   classID,
				TermID:    req.TermID,
				SubjectID: subjectID,
				Score:     in.Score,
			})
		}
		count, err := tx.UpsertMarks(ctx, actor.SchoolID, rows, actor.UserID)
		if err != nil {
			return err
		}
		result = MarksResult{ClassID: classID, SubjectID: subjectID, TermID: req.TermID, SavedCount: count}
		return s.audit(ctx, tx, actor, "marks_entered", "class", classID, map[string]any{
			"subject_id":  subjectID,
			"term_id":     req.TermID,
			"saved_count": count,
		})
	})
	return result, err
}

// ReassignMarksRequest reassigns a class/subject to another teacher.
type ReassignMarksRequest struct {
	SubjectID              string `json:"subject_id"`
	NewTeacherMembershipID string `json:"new_teacher_membership_id"`
	AcademicYearID         string `json:"academic_year_id,omitempty"`
	Reason                 string `json:"reason"`
}

// ReassignMarksResult is returned after a reassignment.
type ReassignMarksResult struct {
	Reassigned                  bool   `json:"reassigned"`
	ClassID                     string `json:"class_id"`
	SubjectID                   string `json:"subject_id"`
	NewTeacherMembershipID      string `json:"new_teacher_membership_id"`
	PreviousTeacherMembershipID string `json:"previous_teacher_membership_id,omitempty"`
}

// ReassignMarks implements POST /api/v1/classes/{id}/marks-assignments. The
// new assignment takes effect immediately; the previous teacher's next write is
// rejected (AC-17).
func (s *Service) ReassignMarks(ctx context.Context, classID string, req ReassignMarksRequest) (ReassignMarksResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return ReassignMarksResult{}, err
	}
	if !s.canReassignMarks(actor) {
		return ReassignMarksResult{}, errForbidden(CodeForbidden, "Only the Dean of Studies, Principal or Deputy may reassign marks entry.")
	}
	classID = strings.TrimSpace(classID)
	req.SubjectID = strings.TrimSpace(req.SubjectID)
	req.NewTeacherMembershipID = strings.TrimSpace(req.NewTeacherMembershipID)
	req.Reason = strings.TrimSpace(req.Reason)
	if classID == "" || req.SubjectID == "" || req.NewTeacherMembershipID == "" {
		return ReassignMarksResult{}, errValidation("class_id, subject_id and new_teacher_membership_id are required.")
	}
	if req.Reason == "" {
		return ReassignMarksResult{}, errValidation("A reason is required when reassigning marks entry.")
	}

	var result ReassignMarksResult
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		class, err := tx.ClassByID(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		subject, err := tx.SubjectByID(ctx, actor.SchoolID, req.SubjectID)
		if err != nil {
			return err
		}
		if subject.GradeLevelID != class.GradeLevelID {
			return errValidation("This subject does not belong to the class's grade level.")
		}
		if err := validateMembership(ctx, tx, actor.SchoolID, req.NewTeacherMembershipID); err != nil {
			return err
		}
		yearID := strings.TrimSpace(req.AcademicYearID)
		if yearID == "" {
			yearID, err = tx.CurrentAcademicYearID(ctx, actor.SchoolID)
			if err != nil {
				return errValidation("An academic year is required before reassigning marks entry.")
			}
		}
		yearExists, err := tx.AcademicYearExists(ctx, actor.SchoolID, yearID)
		if err != nil {
			return err
		}
		if !yearExists {
			return errValidation("The academic year does not belong to this school.")
		}

		previous := ""
		if old, err := tx.SubjectAssignment(ctx, actor.SchoolID, classID, req.SubjectID, yearID); err == nil {
			previous = old.TeacherMembershipID
		}
		if err := tx.UpsertSubjectAssignment(ctx, actor.SchoolID, SubjectTeacherAssignment{
			AcademicYearID:      yearID,
			ClassID:             classID,
			SubjectID:           req.SubjectID,
			TeacherMembershipID: req.NewTeacherMembershipID,
			AssignedBy:          actor.UserID,
		}); err != nil {
			return err
		}
		result = ReassignMarksResult{
			Reassigned:                  true,
			ClassID:                     classID,
			SubjectID:                   req.SubjectID,
			NewTeacherMembershipID:      req.NewTeacherMembershipID,
			PreviousTeacherMembershipID: previous,
		}
		return s.audit(ctx, tx, actor, "marks_entry_reassigned", "class", classID, map[string]any{
			"subject_id":                     req.SubjectID,
			"previous_teacher_membership_id": previous,
			"new_teacher_membership_id":      req.NewTeacherMembershipID,
			"reason":                         req.Reason,
			"academic_year_id":               yearID,
		})
	})
	return result, err
}

func (s *Service) canReassignMarks(actor Identity) bool {
	switch actor.Role {
	case RoleDean, RolePrincipal, RoleDeputy:
		return true
	}
	return contains(actor.Permissions, PermMarksReassign)
}

func (s *Service) handlePutMarks(w http.ResponseWriter, r *http.Request) {
	var req MarksRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.SaveMarks(r.Context(), r.PathValue("id"), r.PathValue("sid"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleReassignMarks(w http.ResponseWriter, r *http.Request) {
	var req ReassignMarksRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.ReassignMarks(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodPut, "/api/v1/classes/{id}/subjects/{sid}/marks", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handlePutMarks(w, r)
	})
	registerRoute(http.MethodPost, "/api/v1/classes/{id}/marks-assignments", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleReassignMarks(w, r)
	})
}
