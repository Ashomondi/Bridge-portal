package academics

import (
	"context"
	"net/http"
	"strings"
)

// SubjectAssignmentInput is one row of the Dean's assignment request.
type SubjectAssignmentInput struct {
	SubjectID           string `json:"subject_id"`
	TeacherMembershipID string `json:"teacher_membership_id"`
}

// AssignmentRequest assigns (or replaces) the class teacher and any number of
// subject teachers for one class (C59).
type AssignmentRequest struct {
	ClassID                  string                   `json:"class_id"`
	AcademicYearID           string                   `json:"academic_year_id,omitempty"`
	ClassTeacherMembershipID string                   `json:"class_teacher_membership_id,omitempty"`
	SubjectAssignments       []SubjectAssignmentInput `json:"subject_assignments,omitempty"`
}

// AssignmentResult is returned to the Dean of Studies workspace.
type AssignmentResult struct {
	Assigned              bool   `json:"assigned"`
	ClassID               string `json:"class_id"`
	AcademicYearID        string `json:"academic_year_id"`
	ClassTeacherAssigned  bool   `json:"class_teacher_assigned"`
	SubjectAssignmentsSet int    `json:"subject_assignments_set"`
}

// AssignTeachers implements POST /api/v1/academics/assignments.
func (s *Service) AssignTeachers(ctx context.Context, req AssignmentRequest) (AssignmentResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return AssignmentResult{}, err
	}
	if !s.canManage(actor, PermAssignmentsManage) {
		return AssignmentResult{}, errForbidden(CodeForbidden, "Only the Dean of Studies may assign teachers to classes and subjects.")
	}

	req.ClassID = strings.TrimSpace(req.ClassID)
	if req.ClassID == "" {
		return AssignmentResult{}, errValidation("class_id is required.")
	}
	if req.ClassTeacherMembershipID == "" && len(req.SubjectAssignments) == 0 {
		return AssignmentResult{}, errValidation("Provide a class_teacher_membership_id or at least one subject assignment.")
	}

	result := AssignmentResult{ClassID: req.ClassID}
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		class, err := tx.ClassByID(ctx, actor.SchoolID, req.ClassID)
		if err != nil {
			return err
		}
		yearID := strings.TrimSpace(req.AcademicYearID)
		if yearID == "" {
			yearID, err = tx.CurrentAcademicYearID(ctx, actor.SchoolID)
			if err != nil {
				return errValidation("An academic year is required before assigning teachers.")
			}
		}
		result.AcademicYearID = yearID
		yearExists, err := tx.AcademicYearExists(ctx, actor.SchoolID, yearID)
		if err != nil {
			return err
		}
		if !yearExists {
			return errValidation("The academic year does not belong to this school.")
		}

		if req.ClassTeacherMembershipID != "" {
			if err := validateMembership(ctx, tx, actor.SchoolID, req.ClassTeacherMembershipID); err != nil {
				return err
			}
			if err := tx.AssignClassTeacher(ctx, actor.SchoolID, req.ClassID, req.ClassTeacherMembershipID, actor.UserID); err != nil {
				return err
			}
			result.ClassTeacherAssigned = true
		}

		for _, entry := range req.SubjectAssignments {
			entry.SubjectID = strings.TrimSpace(entry.SubjectID)
			entry.TeacherMembershipID = strings.TrimSpace(entry.TeacherMembershipID)
			if entry.SubjectID == "" || entry.TeacherMembershipID == "" {
				return errValidation("Every subject assignment needs subject_id and teacher_membership_id.")
			}
			subject, err := tx.SubjectByID(ctx, actor.SchoolID, entry.SubjectID)
			if err != nil {
				return err
			}
			if subject.GradeLevelID != class.GradeLevelID {
				return errValidation("Subject does not belong to this class's grade level.")
			}
			if err := validateMembership(ctx, tx, actor.SchoolID, entry.TeacherMembershipID); err != nil {
				return err
			}
			if err := tx.UpsertSubjectAssignment(ctx, actor.SchoolID, SubjectTeacherAssignment{
				AcademicYearID:      yearID,
				ClassID:             req.ClassID,
				SubjectID:           entry.SubjectID,
				TeacherMembershipID: entry.TeacherMembershipID,
				AssignedBy:          actor.UserID,
			}); err != nil {
				return err
			}
			result.SubjectAssignmentsSet++
		}

		return s.audit(ctx, tx, actor, "teacher_assignments_updated", "class", req.ClassID, map[string]any{
			"academic_year_id":        yearID,
			"class_teacher_assigned":  result.ClassTeacherAssigned,
			"subject_assignments_set": result.SubjectAssignmentsSet,
		})
	})
	if err != nil {
		return AssignmentResult{}, err
	}
	result.Assigned = true
	return result, nil
}

// ListAssignments returns the current teacher assignments for a class.
func (s *Service) ListAssignments(ctx context.Context, classID string) (map[string]any, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}
	classID = strings.TrimSpace(classID)
	if classID == "" {
		return nil, errValidation("class_id is required.")
	}
	var result map[string]any
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		isClassTeacher, _ := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if !s.canManage(actor, PermAssignmentsManage) && !isClassTeacher {
			return errForbidden(CodeForbidden, "You may only view assignments for your own class.")
		}
		assignments, err := tx.ListSubjectAssignments(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		classTeacher, _ := tx.ClassTeacherMembership(ctx, actor.SchoolID, classID)
		result = map[string]any{
			"class_id":                    classID,
			"class_teacher_membership_id": classTeacher,
			"subject_assignments":         assignments,
		}
		return nil
	})
	return result, err
}

func validateMembership(ctx context.Context, tx Store, schoolID, membershipID string) error {
	mem, err := tx.MembershipByID(ctx, schoolID, membershipID)
	if err != nil {
		return errValidation("The selected teacher membership does not exist in this school.")
	}
	if mem.Status != "" && strings.ToLower(mem.Status) != "active" {
		return errValidation("The selected teacher membership is not active.")
	}
	if mem.Role == RoleParent {
		return errValidation("A parent membership cannot be assigned as a teacher.")
	}
	return nil
}

func (s *Service) handleAssignTeachers(w http.ResponseWriter, r *http.Request) {
	var req AssignmentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.AssignTeachers(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleListAssignments(w http.ResponseWriter, r *http.Request) {
	result, err := s.ListAssignments(r.Context(), r.URL.Query().Get("class_id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodPost, "/api/v1/academics/assignments", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleAssignTeachers(w, r)
	})
	registerRoute(http.MethodGet, "/api/v1/academics/assignments", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleListAssignments(w, r)
	})
}
