package onboarding

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var (
	errSubmissionNotFound    = errors.New("onboarding: registration submission not found")
	errSubmissionNotPending  = errors.New("onboarding: registration submission is not pending")
	errVerifierTargetParent  = errors.New("onboarding: verifier grants are for staff only")
	errVerifierBadgeMissing  = errors.New("onboarding: delegated_verifier badge is not configured")
	errDocumentTypeRequired  = errors.New("onboarding: a verified document type is required")
	errNoCurrentAcademicYear = errors.New("onboarding: no current academic year")
	errClassNotFound         = errors.New("onboarding: class not found for this school")
)

type registrationStudent struct {
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	DOB             string `json:"dob"`
	Gender          string `json:"gender"`
	ClassID         string `json:"class_id"`
	DormitoryID     string `json:"dormitory_id"`
}

type registrationRequest struct {
	QRToken        string              `json:"qr_token"`
	ApplicantName  string              `json:"applicant_name"`
	ApplicantPhone string              `json:"applicant_phone"`
	NationalID     string              `json:"national_id"`
	Relationship   string              `json:"relationship"`
	Student        registrationStudent `json:"student"`
	DocumentType   string              `json:"document_type"`
	DocumentBase64 string              `json:"document_base64"`
}

type registrationPayload struct {
	Student      registrationStudent `json:"student"`
	Relationship string              `json:"relationship"`
}

func (s *Service) handleRegistrationSubmission(w http.ResponseWriter, r *http.Request) {
	var req registrationRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.QRToken = strings.TrimSpace(req.QRToken)
	req.ApplicantName = strings.TrimSpace(req.ApplicantName)
	req.ApplicantPhone = strings.TrimSpace(req.ApplicantPhone)
	req.NationalID = strings.TrimSpace(req.NationalID)
	req.Relationship = strings.TrimSpace(req.Relationship)
	req.DocumentType = strings.TrimSpace(req.DocumentType)

	schoolID, ok := ParseQRToken(req.QRToken)
	if !ok {
		middleware.RespondError(w, http.StatusGone, "QR_INVALID", "The QR code is invalid or has expired.")
		return
	}
	if req.ApplicantName == "" || !auth.ValidPhoneNumber(req.ApplicantPhone) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_APPLICANT", "applicant_name is required and applicant_phone must be a Kenyan number.")
		return
	}

	var documentBlob []byte
	if req.DocumentBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.DocumentBase64)
		if err != nil {
			middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_DOCUMENT", "document_base64 is not valid base64.")
			return
		}
		documentBlob = decoded
	}

	payload, err := json.Marshal(registrationPayload{Student: req.Student, Relationship: req.Relationship})
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to encode the submission.")
		return
	}

	ctx := r.Context()
	var submissionID, matchedStudentID string
	var duplicate bool
	err = database.WithTenantTx(ctx, s.db, schoolID, func(tx *sql.Tx) error {
		submissionType, parentQRID, staffInviteID, err := resolveQRToken(ctx, tx, schoolID, req.QRToken)
		if err != nil {
			return err
		}

		if req.Student.AdmissionNumber != "" {
			err := tx.QueryRowContext(ctx,
				`SELECT id::text FROM students WHERE school_id = $1 AND admission_number = $2`,
				schoolID, req.Student.AdmissionNumber).Scan(&matchedStudentID)
			if err == nil {
				duplicate = true
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO registration_submissions
			    (school_id, submission_type, parent_qr_id, staff_invite_id, applicant_phone,
			     applicant_name, national_id, payload, document_type, document_blob,
			     matched_student_id, is_duplicate_flag, status, expires_for_purge_at)
			 VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6,
			         NULLIF($7, ''), $8::jsonb, NULLIF($9, ''), $10,
			         NULLIF($11, '')::uuid, $12, 'pending', NOW() + ($13::int * INTERVAL '1 day'))
			 RETURNING id::text`,
			schoolID, submissionType, parentQRID, staffInviteID, req.ApplicantPhone,
			req.ApplicantName, req.NationalID, string(payload), req.DocumentType, documentBlob,
			matchedStudentID, duplicate, s.retentionDays).Scan(&submissionID); err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     schoolID,
				Action:       "registration.submitted",
				ResourceType: "registration_submission",
				ResourceID:   submissionID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"submission_type": submissionType, "is_duplicate_flag": duplicate},
			})
		}
		return nil
	})
	if mapped, ok := mapQRStatusError(err); ok {
		middleware.RespondError(w, http.StatusGone, mapped.code, mapped.message)
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to record the registration.")
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"submission_id":     submissionID,
		"status":            "pending",
		"is_duplicate_flag": duplicate,
	})
}

type verifierGrantRequest struct {
	StaffMembershipID string `json:"staff_membership_id"`
	ParentQRID        string `json:"parent_qr_id"`
}

func (s *Service) handleVerifierGrant(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req verifierGrantRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.StaffMembershipID = strings.TrimSpace(req.StaffMembershipID)
	req.ParentQRID = strings.TrimSpace(req.ParentQRID)
	if !uuidPattern.MatchString(req.StaffMembershipID) || !uuidPattern.MatchString(req.ParentQRID) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", "staff_membership_id and parent_qr_id must be UUIDs.")
		return
	}

	ctx := r.Context()
	var grantID string
	var validUntil time.Time
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		var role string
		if err := tx.QueryRowContext(ctx,
			`SELECT role FROM memberships WHERE id = $1 AND school_id = $2 AND status = 'active'`,
			req.StaffMembershipID, identity.SchoolID).Scan(&role); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errVerifierTargetParent
			}
			return err
		}
		if role == "parent" {
			return errVerifierTargetParent
		}

		var revokedAt sql.NullTime
		if err := tx.QueryRowContext(ctx,
			`SELECT valid_until, revoked_at FROM parent_qr_codes WHERE id = $1 AND school_id = $2`,
			req.ParentQRID, identity.SchoolID).Scan(&validUntil, &revokedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errParentQRNotFound
			}
			return err
		}
		now := time.Now().UTC()
		switch {
		case revokedAt.Valid:
			return ErrParentQRRevoked
		case !now.Before(validUntil):
			return ErrParentQRExpired
		}

		var badgeID string
		if err := tx.QueryRowContext(ctx,
			`SELECT id::text FROM badges WHERE school_id = $1 AND slug = 'delegated_verifier'`,
			identity.SchoolID).Scan(&badgeID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errVerifierBadgeMissing
			}
			return err
		}

		err := tx.QueryRowContext(ctx,
			`UPDATE membership_badges
			    SET valid_from = NOW(), valid_until = $4
			  WHERE membership_id = $1 AND badge_id = $2 AND scope_id = $3
			  RETURNING id::text`,
			req.StaffMembershipID, badgeID, req.ParentQRID, validUntil).Scan(&grantID)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx,
				`INSERT INTO membership_badges (membership_id, badge_id, scope_id, valid_from, valid_until)
				 VALUES ($1, $2, $3, NOW(), $4)
				 RETURNING id::text`,
				req.StaffMembershipID, badgeID, req.ParentQRID, validUntil).Scan(&grantID)
		}
		if err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "verifier.badge.granted",
				ResourceType: "membership_badge",
				ResourceID:   grantID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"staff_membership_id": req.StaffMembershipID, "parent_qr_id": req.ParentQRID},
			})
		}
		return nil
	})
	switch {
	case errors.Is(err, errVerifierTargetParent):
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_TARGET", "Verifier grants are only allowed for staff memberships.")
		return
	case errors.Is(err, errVerifierBadgeMissing):
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "The delegated_verifier badge is not configured for this school.")
		return
	case errors.Is(err, errParentQRNotFound):
		middleware.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Parent QR code not found.")
		return
	}
	if mapped, ok := mapQRStatusError(err); ok {
		middleware.RespondError(w, http.StatusGone, mapped.code, mapped.message)
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to grant the verifier badge.")
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"badge_grant_id": grantID,
		"valid_until":    validUntil.UTC().Format(time.RFC3339),
	})
}

type verifyRegistrationRequest struct {
	Action               string `json:"action"`
	DocumentTypeVerified string `json:"document_type_verified"`
	Notes                string `json:"notes"`
}

func (s *Service) handleVerifyRegistration(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	submissionID := strings.TrimSpace(r.PathValue("id"))
	if !uuidPattern.MatchString(submissionID) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", "A valid submission id is required.")
		return
	}
	var req verifyRegistrationRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	req.DocumentTypeVerified = strings.TrimSpace(req.DocumentTypeVerified)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.Action != "approve" && req.Action != "reject" {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_ACTION", "action must be approve or reject.")
		return
	}

	ctx := r.Context()
	var (
		result               approvedResult
		verificationRecordID string
	)
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		submission, err := loadSubmission(ctx, tx, identity.SchoolID, submissionID)
		if err != nil {
			return err
		}
		if submission.Status != "pending" {
			return errSubmissionNotPending
		}

		if req.Action == "reject" {
			if _, err := tx.ExecContext(ctx,
				`UPDATE registration_submissions
				    SET status = 'rejected', document_blob = NULL, document_deleted_at = NOW(),
				        expires_for_purge_at = NOW() + ($3::int * INTERVAL '1 day')
				  WHERE id = $1 AND school_id = $2`,
				submissionID, identity.SchoolID, s.retentionDays); err != nil {
				return err
			}
			result = approvedResult{Status: "rejected", DocumentDeleted: true}
			if s.audit != nil {
				return s.audit.Log(ctx, tx, database.AuditEvent{
					SchoolID:     identity.SchoolID,
					UserID:       identity.UserID,
					Action:       "registration.rejected",
					ResourceType: "registration_submission",
					ResourceID:   submissionID,
					IPAddress:    r.RemoteAddr,
					UserAgent:    r.UserAgent(),
				})
			}
			return nil
		}

		documentType := req.DocumentTypeVerified
		if documentType == "" {
			documentType = submission.DocumentType
		}
		if documentType == "" {
			return errDocumentTypeRequired
		}

		verificationRecordID, err = approveSubmission(ctx, tx, s, identity, submission, documentType, req.Notes, &result)
		return err
	})

	switch {
	case errors.Is(err, errSubmissionNotFound):
		middleware.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Registration submission not found.")
		return
	case errors.Is(err, errSubmissionNotPending):
		middleware.RespondError(w, http.StatusConflict, "ALREADY_REVIEWED", "This registration has already been reviewed.")
		return
	case errors.Is(err, errDocumentTypeRequired):
		middleware.RespondError(w, http.StatusUnprocessableEntity, "DOCUMENT_TYPE_REQUIRED", "A verified document type is required.")
		return
	case errors.Is(err, errNoCurrentAcademicYear):
		middleware.RespondError(w, http.StatusConflict, "SETUP_REQUIRED", "Complete the school setup wizard before admitting students.")
		return
	case errors.Is(err, errClassNotFound):
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_CLASS", "The student's class does not belong to this school.")
		return
	}
	var validationErr *validationError
	if errors.As(err, &validationErr) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", validationErr.Error())
		return
	}
	if isUniqueViolation(err) {
		middleware.RespondError(w, http.StatusConflict, "DUPLICATE_STUDENT", "A student with this admission number already exists.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify the registration.")
		return
	}

	if req.Action == "approve" && s.charges != nil {
		if chargeErr := s.charges.StudentChargeHook(ctx, identity.SchoolID, result.StudentID, result.GradeLevelID, result.AcademicYearID); chargeErr != nil {
			result.BillingChargeCreated = false
		} else {
			result.BillingChargeCreated = true
		}
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"status":                 result.Status,
		"document_deleted":       result.DocumentDeleted,
		"verification_record_id": verificationRecordID,
		"student_id":             result.StudentID,
		"enrolment_id":           result.EnrolmentID,
		"billing_charge_created": result.BillingChargeCreated,
	})
}

type submissionRecord struct {
	ID               string
	SubmissionType   string
	ApplicantPhone   string
	ApplicantName    string
	NationalID       string
	Payload          []byte
	DocumentType     string
	MatchedStudentID string
	IsDuplicate      bool
	Status           string
}

type approvedResult struct {
	Status               string
	DocumentDeleted      bool
	StudentID            string
	EnrolmentID          string
	GradeLevelID         string
	AcademicYearID       string
	BillingChargeCreated bool
}

func loadSubmission(ctx context.Context, tx *sql.Tx, schoolID, submissionID string) (submissionRecord, error) {
	var (
		record         submissionRecord
		matchedStudent sql.NullString
	)
	err := tx.QueryRowContext(ctx,
		`SELECT id::text, submission_type, applicant_phone, applicant_name,
		        COALESCE(national_id, ''), payload, COALESCE(document_type, ''),
		        matched_student_id::text, is_duplicate_flag, status
		   FROM registration_submissions
		  WHERE id = $1 AND school_id = $2
		  FOR UPDATE`,
		submissionID, schoolID).
		Scan(&record.ID, &record.SubmissionType, &record.ApplicantPhone, &record.ApplicantName,
			&record.NationalID, &record.Payload, &record.DocumentType,
			&matchedStudent, &record.IsDuplicate, &record.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return submissionRecord{}, errSubmissionNotFound
	}
	if err != nil {
		return submissionRecord{}, err
	}
	if matchedStudent.Valid {
		record.MatchedStudentID = matchedStudent.String
	}
	return record, nil
}

func approveSubmission(ctx context.Context, tx *sql.Tx, service *Service, identity middleware.Identity, submission submissionRecord, documentType, notes string, result *approvedResult) (string, error) {
	var payload registrationPayload
	if len(submission.Payload) > 0 {
		if err := json.Unmarshal(submission.Payload, &payload); err != nil {
			return "", err
		}
	}

	var guardianUserID string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO users (phone_number, full_name, is_active)
		 VALUES ($1, $2, TRUE)
		 ON CONFLICT (phone_number) DO UPDATE
		   SET full_name = EXCLUDED.full_name, is_active = TRUE, updated_at = NOW()
		 RETURNING id::text`,
		submission.ApplicantPhone, submission.ApplicantName).Scan(&guardianUserID); err != nil {
		return "", err
	}

	classID := strings.TrimSpace(payload.Student.ClassID)
	studentID := submission.MatchedStudentID
	if studentID == "" {
		if err := validateNewStudent(payload.Student); err != nil {
			return "", err
		}
		dob, _ := time.Parse(dateLayout, payload.Student.DOB)
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO students
			    (school_id, admission_number, full_name, dob, gender, current_class_id, current_dormitory_id)
			 VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid, NULLIF($7, '')::uuid)
			 RETURNING id::text`,
			identity.SchoolID, strings.TrimSpace(payload.Student.AdmissionNumber),
			strings.TrimSpace(payload.Student.FullName), dob, payload.Student.Gender,
			classID, strings.TrimSpace(payload.Student.DormitoryID)).Scan(&studentID); err != nil {
			return "", err
		}
	} else if classID == "" {
		var currentClass sql.NullString
		if err := tx.QueryRowContext(ctx,
			`SELECT current_class_id::text FROM students WHERE id = $1 AND school_id = $2`,
			studentID, identity.SchoolID).Scan(&currentClass); err != nil {
			return "", err
		}
		if currentClass.Valid {
			classID = currentClass.String
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO memberships (school_id, user_id, role, status)
		 VALUES ($1, $2, 'parent', 'active')
		 ON CONFLICT (school_id, user_id, role) DO NOTHING`,
		identity.SchoolID, guardianUserID); err != nil {
		return "", err
	}

	relationship := strings.TrimSpace(payload.Relationship)
	if relationship == "" {
		relationship = "guardian"
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO student_guardians (school_id, student_id, user_id, relationship, is_primary)
		 VALUES ($1, $2, $3, $4, TRUE)
		 ON CONFLICT (school_id, student_id, user_id) DO NOTHING`,
		identity.SchoolID, studentID, guardianUserID, relationship); err != nil {
		return "", err
	}

	var gradeLevelID string
	if err := tx.QueryRowContext(ctx,
		`SELECT grade_level_id::text FROM classes WHERE id = $1 AND school_id = $2`,
		classID, identity.SchoolID).Scan(&gradeLevelID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errClassNotFound
		}
		return "", err
	}

	var academicYearID string
	if err := tx.QueryRowContext(ctx,
		`SELECT id::text FROM academic_years
		  WHERE school_id = $1 AND is_current = TRUE
		  ORDER BY year DESC LIMIT 1`,
		identity.SchoolID).Scan(&academicYearID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errNoCurrentAcademicYear
		}
		return "", err
	}

	var enrolmentID string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO enrolments (school_id, student_id, academic_year_id, grade_level_id, class_id, status)
		 VALUES ($1, $2, $3, $4, $5, 'active')
		 ON CONFLICT (school_id, student_id, academic_year_id) DO UPDATE
		   SET grade_level_id = EXCLUDED.grade_level_id, class_id = EXCLUDED.class_id, status = 'active'
		 RETURNING id::text`,
		identity.SchoolID, studentID, academicYearID, gradeLevelID, classID).Scan(&enrolmentID); err != nil {
		return "", err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE registration_submissions
		    SET status = 'approved', document_blob = NULL, document_deleted_at = NOW(),
		        matched_student_id = $3, expires_for_purge_at = NULL
		  WHERE id = $1 AND school_id = $2`,
		submission.ID, identity.SchoolID, studentID); err != nil {
		return "", err
	}

	var verificationRecordID string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO verification_records
		    (school_id, submission_id, verified_user_id, student_id, verifier_user_id, document_type, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
		 RETURNING id::text`,
		identity.SchoolID, submission.ID, guardianUserID, studentID, identity.UserID, documentType, notes).Scan(&verificationRecordID); err != nil {
		return "", err
	}

	*result = approvedResult{
		Status:          "approved",
		DocumentDeleted: true,
		StudentID:       studentID,
		EnrolmentID:     enrolmentID,
		GradeLevelID:    gradeLevelID,
		AcademicYearID:  academicYearID,
	}

	if service.audit != nil {
		if err := service.audit.Log(ctx, tx, database.AuditEvent{
			SchoolID:     identity.SchoolID,
			UserID:       identity.UserID,
			Action:       "registration.verified",
			ResourceType: "registration_submission",
			ResourceID:   submission.ID,
			Details:      map[string]any{"student_id": studentID, "document_deleted": true},
		}); err != nil {
			return "", err
		}
	}
	return verificationRecordID, nil
}

func validateNewStudent(student registrationStudent) error {
	switch {
	case strings.TrimSpace(student.AdmissionNumber) == "":
		return validationErrorf("student.admission_number is required")
	case strings.TrimSpace(student.FullName) == "":
		return validationErrorf("student.full_name is required")
	case student.Gender != "M" && student.Gender != "F":
		return validationErrorf("student.gender must be M or F")
	}
	if _, err := time.Parse(dateLayout, student.DOB); err != nil {
		return validationErrorf("student.dob must be YYYY-MM-DD")
	}
	if strings.TrimSpace(student.ClassID) == "" {
		return validationErrorf("student.class_id is required")
	}
	return nil
}

// resolveQRToken decides whether a scanned token belongs to a parent QR or a
// staff invite, validating its time window and revocation state.
func resolveQRToken(ctx context.Context, tx *sql.Tx, schoolID, qrToken string) (string, string, string, error) {
	var (
		parentID   string
		validFrom  time.Time
		validUntil time.Time
		revokedAt  sql.NullTime
	)
	err := tx.QueryRowContext(ctx,
		`SELECT id::text, valid_from, valid_until, revoked_at
		   FROM parent_qr_codes WHERE school_id = $1 AND qr_token = $2`,
		schoolID, qrToken).Scan(&parentID, &validFrom, &validUntil, &revokedAt)
	if err == nil {
		now := time.Now().UTC()
		switch {
		case revokedAt.Valid:
			return "", "", "", ErrParentQRRevoked
		case now.Before(validFrom):
			return "", "", "", ErrParentQRNotYet
		case !now.Before(validUntil):
			return "", "", "", ErrParentQRExpired
		}
		return "parent", parentID, "", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", "", err
	}

	var (
		inviteID     string
		inviteFrom   time.Time
		inviteTo     time.Time
		inviteStatus string
	)
	err = tx.QueryRowContext(ctx,
		`SELECT id::text, valid_from, valid_to, status
		   FROM staff_invites WHERE school_id = $1 AND qr_token = $2`,
		schoolID, qrToken).Scan(&inviteID, &inviteFrom, &inviteTo, &inviteStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", ErrQRNotFound
	}
	if err != nil {
		return "", "", "", err
	}

	now := time.Now().UTC()
	switch {
	case inviteStatus == "revoked":
		return "", "", "", ErrQRRevoked
	case inviteStatus != "pending":
		return "", "", "", ErrQRExpired
	case now.Before(inviteFrom):
		return "", "", "", ErrQRNotStarted
	case !now.Before(inviteTo):
		return "", "", "", ErrQRExpired
	}
	return "staff", "", inviteID, nil
}

type qrStatusError struct {
	code    string
	message string
}

func mapQRStatusError(err error) (qrStatusError, bool) {
	switch {
	case errors.Is(err, ErrQRNotFound):
		return qrStatusError{"QR_INVALID", "The QR code is invalid or has expired."}, true
	case errors.Is(err, ErrQRRevoked), errors.Is(err, ErrParentQRRevoked):
		return qrStatusError{"QR_REVOKED", "This QR code has been revoked."}, true
	case errors.Is(err, ErrQRExpired), errors.Is(err, ErrParentQRExpired):
		return qrStatusError{"QR_EXPIRED", "This QR code has expired."}, true
	case errors.Is(err, ErrQRNotStarted), errors.Is(err, ErrParentQRNotYet):
		return qrStatusError{"QR_NOT_ACTIVE", "This QR code is not active yet."}, true
	}
	return qrStatusError{}, false
}
