package onboarding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgconn"
)

const dateLayout = "2006-01-02"

var errSchoolProfileMissing = errors.New("onboarding: school profile is not provisioned")

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

func validationErrorf(format string, args ...any) error {
	return &validationError{msg: fmt.Sprintf(format, args...)}
}

type permissionSeed struct {
	code     string
	name     string
	category string
}

type badgeSeed struct {
	slug        string
	name        string
	scopeType   string
	description string
	permissions []permissionSeed
}

// defaultBadges are seeded for every new school so the badge engine has real
// permission bundles from day one.
var defaultBadges = []badgeSeed{
	{
		slug:        "class_teacher",
		name:        "Class Teacher",
		scopeType:   "class",
		description: "Scoped to one class: roster, attendance, report comments and results.",
		permissions: []permissionSeed{
			{"class.students.view", "View class roster", "academics"},
			{"students.admit", "Admit students", "onboarding"},
			{"results.release", "Release class results", "academics"},
			{"report_comments.edit", "Edit report comments", "academics"},
			{"attendance.record", "Record daily attendance", "academics"},
		},
	},
	{
		slug:        "delegated_verifier",
		name:        "Delegated Verifier",
		scopeType:   "none",
		description: "Time-bound delegation to verify parent registrations in person.",
		permissions: []permissionSeed{
			{"registrations.verify", "Verify registrations", "onboarding"},
		},
	},
}

type createSchoolRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	County          string `json:"county"`
	PrincipalName   string `json:"principal_name"`
	PrincipalPhone  string `json:"principal_phone"`
	ContactPhone    string `json:"contact_phone"`
	ContactEmail    string `json:"contact_email"`
	OfficeHours     string `json:"office_hours"`
	PhysicalAddress string `json:"physical_address"`
}

func (s *Service) handleCreateSchool(w http.ResponseWriter, r *http.Request) {
	var req createSchoolRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req = normaliseCreateSchool(req)
	if err := validateCreateSchool(req); err != nil {
		middleware.RespondErrorDetails(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	ctx := r.Context()
	var schoolID string
	if err := s.db.QueryRowContext(ctx, `SELECT gen_random_uuid()::text`).Scan(&schoolID); err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to provision the school.")
		return
	}

	var principalUserID string
	err := database.WithTenantTx(ctx, s.db, schoolID, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schools (id, code, name, county)
			 VALUES ($1, $2, $3, NULLIF($4, ''))`,
			schoolID, req.Code, req.Name, req.County); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO users (phone_number, full_name, is_active)
			 VALUES ($1, $2, TRUE)
			 ON CONFLICT (phone_number) DO UPDATE
			   SET full_name = EXCLUDED.full_name, is_active = TRUE, updated_at = NOW()
			 RETURNING id::text`,
			req.PrincipalPhone, req.PrincipalName).Scan(&principalUserID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO school_profile (school_id, contact_phone, contact_email, office_hours, physical_address)
			 VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))`,
			schoolID, req.ContactPhone, req.ContactEmail, req.OfficeHours, req.PhysicalAddress); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO memberships (school_id, user_id, role, status)
			 VALUES ($1, $2, 'principal', 'active')
			 ON CONFLICT (school_id, user_id, role) DO NOTHING`,
			schoolID, principalUserID); err != nil {
			return err
		}

		if err := seedDefaultBadges(ctx, tx, schoolID); err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     schoolID,
				Action:       "school.created",
				ResourceType: "school",
				ResourceID:   schoolID,
				Details:      map[string]any{"code": req.Code, "principal_user_id": principalUserID},
			})
		}
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			middleware.RespondError(w, http.StatusConflict, "SCHOOL_EXISTS", "A school with this code already exists.")
			return
		}
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to provision the school.")
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"school_id":         schoolID,
		"principal_user_id": principalUserID,
	})
}

func normaliseCreateSchool(req createSchoolRequest) createSchoolRequest {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.County = strings.TrimSpace(req.County)
	req.PrincipalName = strings.TrimSpace(req.PrincipalName)
	req.PrincipalPhone = strings.TrimSpace(req.PrincipalPhone)
	req.ContactPhone = strings.TrimSpace(req.ContactPhone)
	req.ContactEmail = strings.TrimSpace(req.ContactEmail)
	req.OfficeHours = strings.TrimSpace(req.OfficeHours)
	req.PhysicalAddress = strings.TrimSpace(req.PhysicalAddress)
	return req
}

func validateCreateSchool(req createSchoolRequest) error {
	switch {
	case req.Code == "" || len(req.Code) > 50:
		return validationErrorf("code is required and must be at most 50 characters")
	case req.Name == "" || len(req.Name) > 255:
		return validationErrorf("name is required and must be at most 255 characters")
	case req.PrincipalName == "":
		return validationErrorf("principal_name is required")
	case !auth.ValidPhoneNumber(req.PrincipalPhone):
		return validationErrorf("principal_phone must be a Kenyan number in +2547XXXXXXXX form")
	case !auth.ValidPhoneNumber(req.ContactPhone):
		return validationErrorf("contact_phone must be a Kenyan number in +2547XXXXXXXX form")
	case !validEmail(req.ContactEmail):
		return validationErrorf("contact_email must be a valid email address")
	}
	return nil
}

func seedDefaultBadges(ctx context.Context, tx *sql.Tx, schoolID string) error {
	for _, seed := range defaultBadges {
		var badgeID string
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO badges (school_id, name, slug, scope_type, description)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (school_id, slug) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id::text`,
			schoolID, seed.name, seed.slug, seed.scopeType, seed.description).Scan(&badgeID); err != nil {
			return err
		}
		for _, permission := range seed.permissions {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO permissions (code, name, category)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (code) DO NOTHING`,
				permission.code, permission.name, permission.category); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO badge_permissions (badge_id, permission_id)
				 SELECT $1, id FROM permissions WHERE code = $2
				 ON CONFLICT DO NOTHING`,
				badgeID, permission.code); err != nil {
				return err
			}
		}
	}
	return nil
}

type contactDetails struct {
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	OfficeHours     string `json:"office_hours"`
	PhysicalAddress string `json:"physical_address"`
}

type termRequest struct {
	TermNumber int    `json:"term_number"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type classRequest struct {
	GradeLevel int    `json:"grade_level"`
	StreamName string `json:"stream_name"`
}

type dormitoryRequest struct {
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Capacity int    `json:"capacity"`
}

type gradingScaleRequest struct {
	EEMin float64 `json:"ee_min"`
	MEMin float64 `json:"me_min"`
	AEMin float64 `json:"ae_min"`
	BEMin float64 `json:"be_min"`
}

type setupWizardRequest struct {
	Contact      contactDetails      `json:"contact"`
	AcademicYear int                 `json:"academic_year"`
	Terms        []termRequest       `json:"terms"`
	GradeLevels  []int               `json:"grade_levels"`
	Classes      []classRequest      `json:"classes"`
	Dormitories  []dormitoryRequest  `json:"dormitories"`
	GradingScale gradingScaleRequest `json:"grading_scale"`
}

type parsedTerm struct {
	number int
	start  time.Time
	end    time.Time
}

type parsedSetup struct {
	academicYear int
	terms        []parsedTerm
	grades       []int
	classes      []classRequest
	dormitories  []dormitoryRequest
	gradingScale gradingScaleRequest
}

func (s *Service) handleSetupWizard(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req setupWizardRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.Contact.Phone = strings.TrimSpace(req.Contact.Phone)
	req.Contact.Email = strings.TrimSpace(req.Contact.Email)
	req.Contact.OfficeHours = strings.TrimSpace(req.Contact.OfficeHours)
	req.Contact.PhysicalAddress = strings.TrimSpace(req.Contact.PhysicalAddress)

	parsed, err := validateSetupWizard(req)
	if err != nil {
		middleware.RespondErrorDetails(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	ctx := r.Context()
	err = database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`UPDATE school_profile
			    SET contact_phone = $2, contact_email = $3,
			        office_hours = NULLIF($4, ''), physical_address = NULLIF($5, ''),
			        wizard_completed = TRUE, updated_at = NOW()
			  WHERE school_id = $1`,
			identity.SchoolID, req.Contact.Phone, req.Contact.Email, req.Contact.OfficeHours, req.Contact.PhysicalAddress)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return errSchoolProfileMissing
		}

		var academicYearID string
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO academic_years (school_id, year, is_current)
			 VALUES ($1, $2, TRUE)
			 ON CONFLICT (school_id, year) DO UPDATE SET is_current = TRUE
			 RETURNING id::text`,
			identity.SchoolID, parsed.academicYear).Scan(&academicYearID); err != nil {
			return err
		}

		for _, term := range parsed.terms {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO terms (school_id, academic_year_id, term_number, start_date, end_date, is_current)
				 VALUES ($1, $2, $3, $4, $5, $6)
				 ON CONFLICT (school_id, academic_year_id, term_number) DO UPDATE
				   SET start_date = EXCLUDED.start_date, end_date = EXCLUDED.end_date, is_current = EXCLUDED.is_current`,
				identity.SchoolID, academicYearID, term.number, term.start, term.end, term.number == 1); err != nil {
				return err
			}
		}

		gradeIDs := make(map[int]string, len(parsed.grades))
		for _, grade := range parsed.grades {
			phase := "junior"
			if grade >= 10 {
				phase = "senior"
			}
			var gradeID string
			if err := tx.QueryRowContext(ctx,
				`INSERT INTO grade_levels (school_id, number, phase, name)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (school_id, number) DO UPDATE SET phase = EXCLUDED.phase, name = EXCLUDED.name
				 RETURNING id::text`,
				identity.SchoolID, grade, phase, fmt.Sprintf("Grade %d", grade)).Scan(&gradeID); err != nil {
				return err
			}
			gradeIDs[grade] = gradeID
		}

		for _, class := range parsed.classes {
			gradeID, ok := gradeIDs[class.GradeLevel]
			if !ok {
				return validationErrorf("class grade level %d is not in grade_levels", class.GradeLevel)
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO classes (school_id, grade_level_id, stream_name)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (school_id, grade_level_id, stream_name) DO NOTHING`,
				identity.SchoolID, gradeID, class.StreamName); err != nil {
				return err
			}
		}

		for _, dormitory := range parsed.dormitories {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO dormitories (school_id, name, gender, capacity)
				 VALUES ($1, $2, $3, $4)`,
				identity.SchoolID, dormitory.Name, dormitory.Gender, dormitory.Capacity); err != nil {
				return err
			}
		}

		if err := seedGradingScale(ctx, tx, identity.SchoolID, identity.UserID, parsed.gradingScale); err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "school.setup_wizard.completed",
				ResourceType: "school",
				ResourceID:   identity.SchoolID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"academic_year": parsed.academicYear},
			})
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errSchoolProfileMissing) {
			middleware.RespondError(w, http.StatusConflict, "SCHOOL_NOT_PROVISIONED", "The school profile has not been provisioned.")
			return
		}
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to complete school setup.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"wizard_completed": true})
}

func validateSetupWizard(req setupWizardRequest) (parsedSetup, error) {
	if !auth.ValidPhoneNumber(req.Contact.Phone) {
		return parsedSetup{}, validationErrorf("contact.phone must be a Kenyan number in +2547XXXXXXXX form")
	}
	if !validEmail(req.Contact.Email) {
		return parsedSetup{}, validationErrorf("contact.email must be a valid email address")
	}
	if req.AcademicYear < 2000 || req.AcademicYear > 2100 {
		return parsedSetup{}, validationErrorf("academic_year must be between 2000 and 2100")
	}
	if len(req.Terms) != 3 {
		return parsedSetup{}, validationErrorf("exactly three terms are required")
	}

	seenTerms := make(map[int]bool, 3)
	terms := make([]parsedTerm, 0, 3)
	for _, term := range req.Terms {
		if term.TermNumber < 1 || term.TermNumber > 3 || seenTerms[term.TermNumber] {
			return parsedSetup{}, validationErrorf("term_number values must be 1, 2 and 3")
		}
		start, err := time.Parse(dateLayout, term.StartDate)
		if err != nil {
			return parsedSetup{}, validationErrorf("term %d start_date must be YYYY-MM-DD", term.TermNumber)
		}
		end, err := time.Parse(dateLayout, term.EndDate)
		if err != nil {
			return parsedSetup{}, validationErrorf("term %d end_date must be YYYY-MM-DD", term.TermNumber)
		}
		if !end.After(start) {
			return parsedSetup{}, validationErrorf("term %d end_date must be after start_date", term.TermNumber)
		}
		seenTerms[term.TermNumber] = true
		terms = append(terms, parsedTerm{number: term.TermNumber, start: start, end: end})
	}

	if len(req.GradeLevels) == 0 {
		return parsedSetup{}, validationErrorf("at least one grade level is required")
	}
	gradeSet := make(map[int]bool, len(req.GradeLevels))
	for _, grade := range req.GradeLevels {
		if grade < 7 || grade > 12 {
			return parsedSetup{}, validationErrorf("grade_levels must be between 7 and 12")
		}
		if gradeSet[grade] {
			return parsedSetup{}, validationErrorf("grade level %d is duplicated", grade)
		}
		gradeSet[grade] = true
	}

	for _, class := range req.Classes {
		if strings.TrimSpace(class.StreamName) == "" {
			return parsedSetup{}, validationErrorf("every class needs a stream_name")
		}
		if !gradeSet[class.GradeLevel] {
			return parsedSetup{}, validationErrorf("class grade level %d is not in grade_levels", class.GradeLevel)
		}
	}

	for _, dormitory := range req.Dormitories {
		if strings.TrimSpace(dormitory.Name) == "" {
			return parsedSetup{}, validationErrorf("every dormitory needs a name")
		}
		if dormitory.Gender != "male" && dormitory.Gender != "female" {
			return parsedSetup{}, validationErrorf("dormitory gender must be male or female")
		}
		if dormitory.Capacity < 0 {
			return parsedSetup{}, validationErrorf("dormitory capacity cannot be negative")
		}
	}

	scale := req.GradingScale
	if scale.EEMin > 100 || scale.EEMin < 0 || scale.BEMin < 0 {
		return parsedSetup{}, validationErrorf("grading scale thresholds must be between 0 and 100")
	}
	if !(scale.EEMin > scale.MEMin && scale.MEMin > scale.AEMin && scale.AEMin > scale.BEMin) {
		return parsedSetup{}, validationErrorf("grading scale must satisfy ee_min > me_min > ae_min > be_min")
	}

	return parsedSetup{
		academicYear: req.AcademicYear,
		terms:        terms,
		grades:       req.GradeLevels,
		classes:      req.Classes,
		dormitories:  req.Dormitories,
		gradingScale: scale,
	}, nil
}

func seedGradingScale(ctx context.Context, tx *sql.Tx, schoolID, userID string, scale gradingScaleRequest) error {
	var tableExists bool
	if err := tx.QueryRowContext(ctx, `SELECT to_regclass('public.grading_scale_versions') IS NOT NULL`).Scan(&tableExists); err != nil {
		return err
	}
	if !tableExists {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO grading_scale_versions
		    (school_id, version_number, ee_min, me_min, ae_min, be_min, is_active, created_by)
		 VALUES ($1, 1, $2, $3, $4, $5, TRUE, NULLIF($6, '')::uuid)
		 ON CONFLICT (school_id, version_number) DO NOTHING`,
		schoolID, scale.EEMin, scale.MEMin, scale.AEMin, scale.BEMin, userID)
	return err
}

func validEmail(value string) bool {
	if value == "" {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
