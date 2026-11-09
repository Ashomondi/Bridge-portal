package onboarding

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

var otpPattern = regexp.MustCompile(`^[0-9]{6}$`)

const (
	phoneChangePurpose = "phone_change"
	phoneChangeMessage = "Your Ikonex Academy login code is %s. It expires in 5 minutes."
	phoneChangedAlert  = "Your Ikonex Academy registered phone number was changed. If this was not you, contact your school immediately."
)

var (
	errOTPInvalid = errors.New("onboarding: OTP is invalid or expired")
	errPhoneTaken = errors.New("onboarding: phone number already belongs to another account")
	errOTPReused  = errors.New("onboarding: OTP has already been used")
)

type phoneChangeRequestOTPRequest struct {
	NewPhoneNumber string `json:"new_phone_number"`
}

func (s *Service) handlePhoneChangeRequestOTP(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req phoneChangeRequestOTPRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.NewPhoneNumber = strings.TrimSpace(req.NewPhoneNumber)
	if !auth.ValidPhoneNumber(req.NewPhoneNumber) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PHONE_NUMBER", "new_phone_number must be a Kenyan number in +2547XXXXXXXX form.")
		return
	}
	if req.NewPhoneNumber == identity.PhoneNumber {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "SAME_PHONE_NUMBER", "The new phone number must differ from the current one.")
		return
	}
	if s.sms == nil {
		middleware.RespondError(w, http.StatusServiceUnavailable, "SMS_UNAVAILABLE", "SMS delivery is not configured.")
		return
	}

	ctx := r.Context()
	taken, err := s.phoneTakenByOther(ctx, req.NewPhoneNumber, identity.UserID)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to check the phone number.")
		return
	}
	if taken {
		middleware.RespondError(w, http.StatusConflict, "PHONE_IN_USE", "That phone number already belongs to another account.")
		return
	}

	waiting, err := s.phoneChangeCooldown(ctx, req.NewPhoneNumber)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to check rate limits.")
		return
	}
	if waiting {
		middleware.RespondError(w, http.StatusTooManyRequests, "OTP_COOLDOWN", "Please wait before requesting another OTP.")
		return
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to generate an OTP.")
		return
	}
	var requestID string
	if err := s.db.QueryRowContext(ctx,
		`INSERT INTO otp_requests (phone_number, sender_identifier, otp_hash, purpose, expires_at)
		 VALUES ($1, $1, $2, $3, NOW() + INTERVAL '5 minutes')
		 RETURNING id::text`,
		req.NewPhoneNumber, auth.HashOTP(otp), phoneChangePurpose).Scan(&requestID); err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to store the OTP.")
		return
	}
	if err := s.sms.SendTransactionalSMS(ctx, req.NewPhoneNumber, fmt.Sprintf(phoneChangeMessage, otp)); err != nil {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM otp_requests WHERE id = $1`, requestID)
		middleware.RespondError(w, http.StatusBadGateway, "SMS_DELIVERY_FAILED", "Unable to send the OTP SMS.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{"otp_sent_to": req.NewPhoneNumber})
}

type phoneChangeConfirmRequest struct {
	NewPhoneNumber string `json:"new_phone_number"`
	OTP            string `json:"otp"`
}

func (s *Service) handlePhoneChangeConfirm(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	var req phoneChangeConfirmRequest
	if !middleware.DecodeJSON(w, r, &req) {
		return
	}
	req.NewPhoneNumber = strings.TrimSpace(req.NewPhoneNumber)
	req.OTP = strings.TrimSpace(req.OTP)
	if !auth.ValidPhoneNumber(req.NewPhoneNumber) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PHONE_NUMBER", "new_phone_number must be a Kenyan number in +2547XXXXXXXX form.")
		return
	}
	if !otpPattern.MatchString(req.OTP) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_OTP", "The OTP must be 6 digits.")
		return
	}

	ctx := r.Context()
	var (
		oldPhone    string
		pausedUntil time.Time
	)
	err := database.WithTenantTx(ctx, s.db, identity.SchoolID, func(tx *sql.Tx) error {
		requestID, storedHash, err := latestPhoneChangeOTP(ctx, tx, req.NewPhoneNumber)
		if err != nil {
			return errOTPInvalid
		}
		if subtle.ConstantTimeCompare([]byte(storedHash), []byte(auth.HashOTP(req.OTP))) != 1 {
			return errOTPInvalid
		}
		result, err := tx.ExecContext(ctx,
			`UPDATE otp_requests SET consumed_at = NOW() WHERE id = $1 AND consumed_at IS NULL`, requestID)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return errOTPReused
		}

		var taken bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM users WHERE phone_number = $1 AND id <> $2)`,
			req.NewPhoneNumber, identity.UserID).Scan(&taken); err != nil {
			return err
		}
		if taken {
			return errPhoneTaken
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT phone_number FROM users WHERE id = $1 AND is_active = TRUE`,
			identity.UserID).Scan(&oldPhone); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`UPDATE users
			    SET phone_number = $1, phone_changed_at = NOW(),
			        chat_payment_paused_until = NOW() + INTERVAL '24 hours', updated_at = NOW()
			  WHERE id = $2 AND is_active = TRUE
			  RETURNING chat_payment_paused_until`,
			req.NewPhoneNumber, identity.UserID).Scan(&pausedUntil); err != nil {
			return err
		}

		// Invalidate WhatsApp chat sessions when that schema is present.
		var hasChatSessions bool
		if err := tx.QueryRowContext(ctx, `SELECT to_regclass('public.chat_sessions') IS NOT NULL`).Scan(&hasChatSessions); err != nil {
			return err
		}
		if hasChatSessions {
			if _, err := tx.ExecContext(ctx,
				`UPDATE chat_sessions SET expires_at = NOW()
				  WHERE guardian_user_id = $1 AND school_id = $2`,
				identity.UserID, identity.SchoolID); err != nil {
				return err
			}
		}

		if s.audit != nil {
			return s.audit.Log(ctx, tx, database.AuditEvent{
				SchoolID:     identity.SchoolID,
				UserID:       identity.UserID,
				Action:       "guardian.phone_changed",
				ResourceType: "user",
				ResourceID:   identity.UserID,
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Details:      map[string]any{"payment_paused_until": pausedUntil.UTC().Format(time.RFC3339)},
			})
		}
		return nil
	})

	switch {
	case errors.Is(err, errOTPInvalid), errors.Is(err, errOTPReused):
		middleware.RespondError(w, http.StatusUnauthorized, "INVALID_OTP", "The OTP is invalid or has expired.")
		return
	case errors.Is(err, errPhoneTaken):
		middleware.RespondError(w, http.StatusConflict, "PHONE_IN_USE", "That phone number already belongs to another account.")
		return
	}
	if isUniqueViolation(err) {
		middleware.RespondError(w, http.StatusConflict, "PHONE_IN_USE", "That phone number already belongs to another account.")
		return
	}
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to change the phone number.")
		return
	}

	oldNumberNotified := false
	if s.sms != nil && oldPhone != "" {
		oldNumberNotified = s.sms.SendTransactionalSMS(ctx, oldPhone, phoneChangedAlert) == nil
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"phone_updated":             true,
		"old_number_notified":       oldNumberNotified,
		"chat_payment_paused_until": pausedUntil.UTC().Format(time.RFC3339),
	})
}

func latestPhoneChangeOTP(ctx context.Context, tx *sql.Tx, phone string) (string, string, error) {
	var id, hash string
	err := tx.QueryRowContext(ctx,
		`SELECT id::text, otp_hash FROM otp_requests
		  WHERE phone_number = $1 AND purpose = $2
		    AND consumed_at IS NULL AND expires_at > NOW()
		  ORDER BY created_at DESC LIMIT 1`,
		phone, phoneChangePurpose).Scan(&id, &hash)
	return id, hash, err
}

func (s *Service) phoneTakenByOther(ctx context.Context, phone, userID string) (bool, error) {
	var taken bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE phone_number = $1 AND id <> $2)`,
		phone, userID).Scan(&taken)
	return taken, err
}

func (s *Service) phoneChangeCooldown(ctx context.Context, phone string) (bool, error) {
	var last time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT created_at FROM otp_requests
		  WHERE phone_number = $1 AND purpose = $2
		  ORDER BY created_at DESC LIMIT 1`,
		phone, phoneChangePurpose).Scan(&last)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return time.Since(last) < time.Minute, nil
}
