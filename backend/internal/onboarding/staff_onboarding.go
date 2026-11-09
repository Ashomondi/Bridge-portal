package onboarding

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var (
	ErrQRNotFound   = errors.New("onboarding: staff QR not found")
	ErrQRNotStarted = errors.New("onboarding: staff QR is not active yet")
	ErrQRExpired    = errors.New("onboarding: staff QR has expired")
	ErrQRRevoked    = errors.New("onboarding: staff QR has been revoked")

	qrTokenPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\.[A-Za-z0-9_-]+$`)
)

// StaffInvite is a resolved staff QR invite.
type StaffInvite struct {
	ID          string
	SchoolID    string
	PhoneNumber string
	Role        string
	Status      string
	ValidFrom   time.Time
	ValidTo     time.Time
}

type createStaffInviteRequest struct {
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	ValidFrom   string `json:"valid_from"`
	ValidTo     string `json:"valid_to"`
}

func (s *Service) handleCreateStaffInvite(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req createStaffInviteRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.Role = strings.TrimSpace(req.Role)
	req.ValidFrom = strings.TrimSpace(req.ValidFrom)
	req.ValidTo = strings.TrimSpace(req.ValidTo)

	if !auth.ValidPhoneNumber(req.PhoneNumber) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PHONE_NUMBER", "phone_number must be a Kenyan number in +2547XXXXXXXX form.")
		return
	}
	if req.Role == "" || req.Role == "parent" {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_ROLE", "role must be a staff role (not parent).")
		return
	}

	validFrom, validTo, err := parseStaffWindow(req.ValidFrom, req.ValidTo, time.Now().UTC())
	if err != nil {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_WINDOW", err.Error())
		return
	}

	token, err := GenerateQRToken(identity.SchoolID)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to generate the QR code.")
		return
	}

	ctx := r.Context()
	var inviteID string
	err = database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO staff_invites
			    (school_id, phone_number, role, qr_token, valid_from, valid_to, status, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, 'pending', NULLIF($7, '')::uuid)
			 RETURNING id::text`,
			identity.SchoolID, req.PhoneNumber, req.Role, token, validFrom, validTo, identity.UserID).Scan(&inviteID); err != nil {
			return err
		}
		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "staff.invite.created",
				ResourceType: "staff_invite",
				ResourceID:   inviteID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"phone_number": req.PhoneNumber, "role": req.Role},
			})
		}
		return nil
	})
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create the staff invite.")
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"invite_id":  inviteID,
		"qr_token":   token,
		"valid_from": validFrom.UTC().Format(time.RFC3339),
		"valid_to":   validTo.UTC().Format(time.RFC3339),
	})
}

func (s *Service) handleRevokeStaffInvite(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	inviteID := strings.TrimSpace(r.PathValue("id"))
	if inviteID == "" {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", "A staff invite id is required.")
		return
	}

	ctx := r.Context()
	var revokedID string
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`UPDATE staff_invites
			    SET status = 'revoked'
			  WHERE id = $1 AND school_id = $2
			  RETURNING id::text`,
			inviteID, identity.SchoolID).Scan(&revokedID)
		if err != nil {
			return err
		}
		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "staff.invite.revoked",
				ResourceType: "staff_invite",
				ResourceID:   revokedID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
			})
		}
		return nil
	})
	if errors.Is(err, sql.ErrNoRows) {
		middleware.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Staff invite not found.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to revoke the staff invite.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"revoked": true})
}

// ValidateStaffInvite resolves a staff QR token and reports whether it may
// currently be used. The token carries its school id as a prefix so the tenant
// can be resolved before any RLS-protected row is read. BE-A3 maps the errors
// to 410 Gone.
func (s *Service) ValidateStaffInvite(ctx context.Context, qrToken string) (StaffInvite, error) {
	schoolID, ok := ParseQRToken(qrToken)
	if !ok {
		return StaffInvite{}, ErrQRNotFound
	}

	var invite StaffInvite
	err := database.WithTenantTx(ctx, s.db, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT id::text, school_id::text, phone_number, role, status, valid_from, valid_to
			   FROM staff_invites
			  WHERE school_id = $1 AND qr_token = $2`,
			schoolID, qrToken).
			Scan(&invite.ID, &invite.SchoolID, &invite.PhoneNumber, &invite.Role, &invite.Status, &invite.ValidFrom, &invite.ValidTo)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return StaffInvite{}, ErrQRNotFound
	}
	if err != nil {
		return StaffInvite{}, err
	}

	now := time.Now().UTC()
	switch {
	case invite.Status == "revoked":
		return invite, ErrQRRevoked
	case invite.Status != "pending":
		return invite, ErrQRExpired
	case now.Before(invite.ValidFrom):
		return invite, ErrQRNotStarted
	case !now.Before(invite.ValidTo):
		return invite, ErrQRExpired
	}
	return invite, nil
}

// GenerateQRToken returns a URL-safe QR token prefixed with the school id so a
// public scan can resolve the tenant before touching RLS-protected tables.
func GenerateQRToken(schoolID string) (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return schoolID + "." + base64.RawURLEncoding.EncodeToString(buffer), nil
}

// ParseQRToken splits a QR token into its school id prefix and random suffix.
func ParseQRToken(qrToken string) (string, bool) {
	if !qrTokenPattern.MatchString(qrToken) {
		return "", false
	}
	schoolID, _, ok := strings.Cut(qrToken, ".")
	if !ok {
		return "", false
	}
	return schoolID, true
}

func parseStaffWindow(rawFrom, rawTo string, now time.Time) (time.Time, time.Time, error) {
	defaultFrom, defaultTo := defaultStaffWindow(now)
	validFrom, validTo := defaultFrom, defaultTo

	if rawFrom != "" {
		parsed, err := time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, validationErrorf("valid_from must be an RFC3339 timestamp")
		}
		validFrom = parsed
	}
	if rawTo != "" {
		parsed, err := time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, validationErrorf("valid_to must be an RFC3339 timestamp")
		}
		validTo = parsed
	}
	if !validTo.After(validFrom) {
		return time.Time{}, time.Time{}, validationErrorf("valid_to must be after valid_from")
	}
	return validFrom, validTo, nil
}

func defaultStaffWindow(now time.Time) (time.Time, time.Time) {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return day.Add(8 * time.Hour), day.Add(18 * time.Hour)
}
