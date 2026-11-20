package onboarding

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var (
	uuidPattern         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	errParentQRNotFound = errors.New("onboarding: parent QR not found")
	ErrParentQRExpired  = errors.New("onboarding: parent QR has expired")
	ErrParentQRRevoked  = errors.New("onboarding: parent QR has been revoked")
	ErrParentQRNotYet   = errors.New("onboarding: parent QR is not active yet")
)

type createParentQRRequest struct {
	Notes string `json:"notes"`
}

func (s *Service) handleCreateParentQR(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req createParentQRRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.Notes = strings.TrimSpace(req.Notes)

	token, err := GenerateQRToken(identity.SchoolID)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to generate the QR code.")
		return
	}

	ctx := r.Context()
	var qrID string
	var validUntil time.Time
	err = database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO parent_qr_codes (school_id, qr_token, created_by, notes, valid_from, valid_until)
			 VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, ''), NOW(), NOW() + INTERVAL '12 hours')
			 RETURNING id::text, valid_until`,
			identity.SchoolID, token, identity.UserID, req.Notes).Scan(&qrID, &validUntil); err != nil {
			return err
		}
		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "parent.qr.created",
				ResourceType: "parent_qr",
				ResourceID:   qrID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"valid_until": validUntil.UTC().Format(time.RFC3339)},
			})
		}
		return nil
	})
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create the parent QR code.")
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, map[string]any{
		"qr_id":       qrID,
		"qr_token":    token,
		"valid_until": validUntil.UTC().Format(time.RFC3339),
	})
}

func (s *Service) handleRevokeParentQR(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	qrID := strings.TrimSpace(r.PathValue("id"))
	if !uuidPattern.MatchString(qrID) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_REQUEST", "A valid parent QR id is required.")
		return
	}

	ctx := r.Context()
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		var revokedID string
		err := tx.QueryRowContext(ctx,
			`UPDATE parent_qr_codes
			    SET revoked_at = NOW(), revoked_by = NULLIF($3, '')::uuid
			  WHERE id = $1 AND school_id = $2 AND revoked_at IS NULL
			  RETURNING id::text`,
			qrID, identity.SchoolID, identity.UserID).Scan(&revokedID)
		if errors.Is(err, sql.ErrNoRows) {
			var exists bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS (SELECT 1 FROM parent_qr_codes WHERE id = $1 AND school_id = $2)`,
				qrID, identity.SchoolID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errParentQRNotFound
			}
		} else if err != nil {
			return err
		}

		// Invalidate any verifier badge granted against this QR (AC-03, AC-05).
		if _, err := tx.ExecContext(ctx,
			`UPDATE membership_badges mb
			    SET valid_until = NOW()
			   FROM badges b
			  WHERE mb.badge_id = b.id
			    AND b.school_id = $1
			    AND b.slug = 'delegated_verifier'
			    AND mb.scope_id = $2
			    AND (mb.valid_until IS NULL OR mb.valid_until > NOW())`,
			identity.SchoolID, qrID); err != nil {
			return err
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "parent.qr.revoked",
				ResourceType: "parent_qr",
				ResourceID:   qrID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
			})
		}
		return nil
	})
	if errors.Is(err, errParentQRNotFound) {
		middleware.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Parent QR code not found.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to revoke the parent QR code.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"revoked": true})
}
