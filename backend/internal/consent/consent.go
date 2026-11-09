package consent

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var (
	validPurposes = map[string]bool{
		"biometric_attendance": true,
		"photo":                true,
		"sms":                  true,
		"whatsapp":             true,
	}
	uuidPattern    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	errNotGuardian = errors.New("consent: caller is not a guardian of the student")
)

// Service manages parent consent records for Release 1 synthetic testing.
type Service struct {
	db    *sql.DB
	audit *database.AuditLogger
}

// NewService builds the consent service.
func NewService(db *sql.DB, audit *database.AuditLogger) *Service {
	return &Service{db: db, audit: audit}
}

// RegisterRoutes mounts the consent endpoints. wrap should enforce
// authentication, tenant scope and the consents.manage permission.
func (s *Service) RegisterRoutes(mux *http.ServeMux, wrap func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/consents", wrap(http.HandlerFunc(s.handleList)))
	mux.Handle("POST /api/v1/consents", wrap(http.HandlerFunc(s.handleUpsert)))
}

type consentBody struct {
	StudentID string `json:"student_id"`
	Purpose   string `json:"purpose"`
	Granted   bool   `json:"granted"`
}

type consentResponse struct {
	StudentID string `json:"student_id"`
	Purpose   string `json:"purpose"`
	Granted   bool   `json:"granted"`
	Timestamp string `json:"timestamp"`
	RevokedAt string `json:"revoked_at,omitempty"`
}

func (s *Service) handleUpsert(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var body consentBody
	if !middleware.DecodeJSON(w, r, &body) {
		return
	}
	body.StudentID = strings.TrimSpace(body.StudentID)
	body.Purpose = strings.TrimSpace(body.Purpose)
	if !uuidPattern.MatchString(body.StudentID) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_STUDENT", "student_id must be a UUID.")
		return
	}
	if !validPurposes[body.Purpose] {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PURPOSE", "purpose must be biometric_attendance, photo, sms or whatsapp.")
		return
	}

	ctx := r.Context()
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		var isGuardian bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (
				SELECT 1 FROM student_guardians
				 WHERE school_id = $1 AND student_id = $2 AND user_id = $3
			)`, identity.SchoolID, body.StudentID, identity.UserID).Scan(&isGuardian); err != nil {
			return err
		}
		if !isGuardian {
			return errNotGuardian
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO consents (school_id, student_id, guardian_user_id, purpose, granted, timestamp)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			identity.SchoolID, body.StudentID, identity.UserID, body.Purpose, body.Granted)
		return err
	})
	switch {
	case errors.Is(err, errNotGuardian):
		middleware.RespondError(w, http.StatusForbidden, "NOT_GUARDIAN", "The caller is not a guardian of this student.")
		return
	case err != nil:
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to record consent.")
		return
	}

	if s.audit != nil {
		_ = s.audit.Log(ctx, nil, database.AuditEvent{
			SchoolID:     identity.SchoolID,
			UserID:       identity.UserID,
			Action:       "consent.update",
			ResourceType: "student",
			ResourceID:   body.StudentID,
			IPAddress:    r.RemoteAddr,
			UserAgent:    r.UserAgent(),
			Details:      map[string]any{"purpose": body.Purpose, "granted": body.Granted},
		})
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"student_id": body.StudentID,
		"purpose":    body.Purpose,
		"granted":    body.Granted,
	})
}

func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}

	ctx := r.Context()
	consents := make([]consentResponse, 0)
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT student_id::text, purpose, granted, timestamp::text, COALESCE(revoked_at::text, '')
			   FROM consents
			  WHERE school_id = current_school_id() AND guardian_user_id = $1
			  ORDER BY timestamp DESC`, identity.UserID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var consent consentResponse
			if err := rows.Scan(&consent.StudentID, &consent.Purpose, &consent.Granted, &consent.Timestamp, &consent.RevokedAt); err != nil {
				return err
			}
			consents = append(consents, consent)
		}
		return rows.Err()
	})
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load consents.")
		return
	}
	middleware.RespondJSON(w, http.StatusOK, consents)
}
