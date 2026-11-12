package onboarding

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var (
	errDormitoryNotFound   = errors.New("onboarding: dormitory not found for this school")
	errAcademicYearMissing = errors.New("onboarding: academic year not found")
)

type admitStudentRequest struct {
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	DOB             string `json:"dob"`
	Gender          string `json:"gender"`
	ClassID         string `json:"class_id"`
	DormitoryID     string `json:"dormitory_id"`
	GuardianName    string `json:"guardian_name"`
	GuardianPhone   string `json:"guardian_phone"`
	Relationship    string `json:"relationship"`
}

func (s *Service) handleAdmitStudent(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req admitStudentRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.AdmissionNumber = strings.TrimSpace(req.AdmissionNumber)
	req.FullName = strings.TrimSpace(req.FullName)
	req.DOB = strings.TrimSpace(req.DOB)
	req.Gender = strings.TrimSpace(req.Gender)
	req.ClassID = strings.TrimSpace(req.ClassID)
	req.DormitoryID = strings.TrimSpace(req.DormitoryID)
	req.GuardianName = strings.TrimSpace(req.GuardianName)
	req.GuardianPhone = strings.TrimSpace(req.GuardianPhone)
	req.Relationship = strings.TrimSpace(req.Relationship)

	if err := validateAdmitRequest(req); err != nil {
		middleware.RespondErrorDetails(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	dob, _ := time.Parse(dateLayout, req.DOB)

	ctx := r.Context()
	var studentID, enrolmentID, gradeLevelID, academicYearID string
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`SELECT grade_level_id::text FROM classes WHERE id = $1 AND school_id = $2`,
			req.ClassID, identity.SchoolID).Scan(&gradeLevelID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errClassNotFound
			}
			return err
		}
		if req.DormitoryID != "" {
			var exists bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS (SELECT 1 FROM dormitories WHERE id = $1 AND school_id = $2)`,
				req.DormitoryID, identity.SchoolID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errDormitoryNotFound
			}
		}

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO students
			    (school_id, admission_number, full_name, dob, gender, current_class_id, current_dormitory_id)
			 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid)
			 RETURNING id::text`,
			identity.SchoolID, req.AdmissionNumber, req.FullName, dob, req.Gender,
			req.ClassID, req.DormitoryID).Scan(&studentID); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT id::text FROM academic_years
			  WHERE school_id = $1 AND is_current = TRUE
			  ORDER BY year DESC LIMIT 1`,
			identity.SchoolID).Scan(&academicYearID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errNoCurrentAcademicYear
			}
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO enrolments (school_id, student_id, academic_year_id, grade_level_id, class_id, status)
			 VALUES ($1, $2, $3, $4, $5, 'active')
			 ON CONFLICT (school_id, student_id, academic_year_id) DO UPDATE
			   SET grade_level_id = EXCLUDED.grade_level_id, class_id = EXCLUDED.class_id, status = 'active'
			 RETURNING id::text`,
			identity.SchoolID, studentID, academicYearID, gradeLevelID, req.ClassID).Scan(&enrolmentID); err != nil {
			return err
		}

		var guardianUserID string
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO users (phone_number, full_name, is_active)
			 VALUES ($1, $2, TRUE)
			 ON CONFLICT (phone_number) DO UPDATE
			   SET full_name = EXCLUDED.full_name, is_active = TRUE, updated_at = NOW()
			 RETURNING id::text`,
			req.GuardianPhone, req.GuardianName).Scan(&guardianUserID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO memberships (school_id, user_id, role, status)
			 VALUES ($1, $2, 'parent', 'active')
			 ON CONFLICT (school_id, user_id, role) DO NOTHING`,
			identity.SchoolID, guardianUserID); err != nil {
			return err
		}
		relationship := req.Relationship
		if relationship == "" {
			relationship = "guardian"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO student_guardians (school_id, student_id, user_id, relationship, is_primary)
			 VALUES ($1, $2, $3, $4, TRUE)
			 ON CONFLICT (school_id, student_id, user_id) DO NOTHING`,
			identity.SchoolID, studentID, guardianUserID, relationship); err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "student.admitted",
				ResourceType: "student",
				ResourceID:   studentID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"admission_number": req.AdmissionNumber, "class_id": req.ClassID},
			})
		}
		return nil
	})

	switch {
	case errors.Is(err, errClassNotFound):
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_CLASS", "The class does not belong to this school.")
		return
	case errors.Is(err, errDormitoryNotFound):
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_DORMITORY", "The dormitory does not belong to this school.")
		return
	case errors.Is(err, errNoCurrentAcademicYear):
		middleware.RespondError(w, http.StatusConflict, "SETUP_REQUIRED", "Complete the school setup wizard before admitting students.")
		return
	}
	if isUniqueViolation(err) {
		middleware.RespondError(w, http.StatusConflict, "DUPLICATE_STUDENT", "A student with this admission number already exists.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to admit the student.")
		return
	}

	billingChargeCreated := s.chargeStudent(ctx, identity.SchoolID, studentID, gradeLevelID, academicYearID)
	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"student_id":             studentID,
		"enrolment_id":           enrolmentID,
		"billing_charge_created": billingChargeCreated,
	})
}

type reEnrolRequest struct {
	AcademicYear int `json:"academic_year"`
}

func (s *Service) handleReEnrolStudents(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req reEnrolRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	if req.AcademicYear < 2000 || req.AcademicYear > 2100 {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", "academic_year must be between 2000 and 2100.")
		return
	}

	ctx := r.Context()
	var enrolled int64
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		var academicYearID string
		if err := tx.QueryRowContext(ctx,
			`SELECT id::text FROM academic_years WHERE school_id = $1 AND year = $2`,
			identity.SchoolID, req.AcademicYear).Scan(&academicYearID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errAcademicYearMissing
			}
			return err
		}

		result, err := tx.ExecContext(ctx,
			`INSERT INTO enrolments (school_id, student_id, academic_year_id, grade_level_id, class_id, status)
			 SELECT s.school_id, s.id, $2, c.grade_level_id, s.current_class_id, 'active'
			   FROM students s
			   JOIN classes c ON c.id = s.current_class_id AND c.school_id = s.school_id
			  WHERE s.school_id = $1 AND s.status = 'active'
			 ON CONFLICT (school_id, student_id, academic_year_id) DO NOTHING`,
			identity.SchoolID, academicYearID)
		if err != nil {
			return err
		}
		enrolled, _ = result.RowsAffected()

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "students.re_enrolled",
				ResourceType: "academic_year",
				ResourceID:   academicYearID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"academic_year": req.AcademicYear, "enrolled": enrolled},
			})
		}
		return nil
	})
	if errors.Is(err, errAcademicYearMissing) {
		middleware.RespondError(w, http.StatusNotFound, "NOT_FOUND", "The target academic year does not exist.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to re-enrol students.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"enrolled": enrolled})
}

func validateAdmitRequest(req admitStudentRequest) error {
	switch {
	case req.AdmissionNumber == "":
		return validationErrorf("admission_number is required")
	case req.FullName == "":
		return validationErrorf("full_name is required")
	case req.Gender != "M" && req.Gender != "F":
		return validationErrorf("gender must be M or F")
	case !uuidPattern.MatchString(req.ClassID):
		return validationErrorf("class_id must be a UUID")
	case req.DormitoryID != "" && !uuidPattern.MatchString(req.DormitoryID):
		return validationErrorf("dormitory_id must be a UUID when provided")
	case req.GuardianName == "":
		return validationErrorf("guardian_name is required")
	case !auth.ValidPhoneNumber(req.GuardianPhone):
		return validationErrorf("guardian_phone must be a Kenyan number in +2547XXXXXXXX form")
	}
	if _, err := time.Parse(dateLayout, req.DOB); err != nil {
		return validationErrorf("dob must be YYYY-MM-DD")
	}
	return nil
}

func (s *Service) chargeStudent(ctx context.Context, schoolID, studentID, gradeLevelID, academicYearID string) bool {
	if s.charges == nil {
		return false
	}
	return s.charges.StudentChargeHook(ctx, schoolID, studentID, gradeLevelID, academicYearID) == nil
}
