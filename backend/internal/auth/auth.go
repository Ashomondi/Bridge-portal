package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

const (
	otpTTL            = 5 * time.Minute
	requestCooldown   = time.Minute
	blockWindow       = 30 * time.Minute
	blockThreshold    = 5
	blockDuration     = 24 * time.Hour
	maxOTPAttempts    = 5
	otpCodeLength     = 6
	otpSMSMessage     = "Your Ikonex Academy login code is %s. It expires in 5 minutes."
	otpRequestPurpose = "app_login"
)

var (
	errInvalidToken = errors.New("auth: invalid token")
	phonePattern    = regexp.MustCompile(`^\+254[17][0-9]{8}$`)
	otpPattern      = regexp.MustCompile(`^[0-9]{6}$`)
)

// SMSDispatcher sends transactional SMS messages.
type SMSDispatcher interface {
	SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error
}

// TokenManager issues and verifies HS256 JWTs.
type TokenManager struct {
	secret []byte
	expiry time.Duration
}

// NewTokenManager builds a token manager.
func NewTokenManager(secret string, expiry time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), expiry: expiry}
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	Phone     string `json:"phone"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// Issue returns a signed JWT for the user and the time it expires.
func (m *TokenManager) Issue(userID, phoneNumber string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expiry)
	claims := jwtClaims{
		Subject:   userID,
		Phone:     phoneNumber,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: encode claims: %w", err)
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(signingInput)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), expiresAt, nil
}

// Verify parses and validates a signed JWT.
func (m *TokenManager) Verify(token string) (*middleware.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "HS256" {
		return nil, errInvalidToken
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errInvalidToken
	}
	if !hmac.Equal(providedSignature, m.sign(parts[0]+"."+parts[1])) {
		return nil, errInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errInvalidToken
	}
	if claims.Subject == "" || claims.ExpiresAt <= time.Now().Unix() {
		return nil, errInvalidToken
	}
	return &middleware.Claims{
		UserID:      claims.Subject,
		PhoneNumber: claims.Phone,
		ExpiresAt:   claims.ExpiresAt,
	}, nil
}

func (m *TokenManager) sign(input string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}

// Service implements phone + SMS OTP authentication.
type Service struct {
	db                *sql.DB
	tokens            *TokenManager
	sms               SMSDispatcher
	audit             *database.AuditLogger
	trustProxyHeaders bool
}

// NewService builds the auth service.
func NewService(db *sql.DB, tokens *TokenManager, sms SMSDispatcher, audit *database.AuditLogger, trustProxyHeaders bool) *Service {
	return &Service{db: db, tokens: tokens, sms: sms, audit: audit, trustProxyHeaders: trustProxyHeaders}
}

// rateLimitError describes a request that was refused by the OTP rate limiter.
type rateLimitError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *rateLimitError) Error() string {
	return "auth: " + e.Code
}

// sqlQuerier is satisfied by both *sql.DB and *sql.Tx.
type sqlQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// RegisterRoutes mounts the authentication endpoints.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/otp/request", s.handleRequestOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/verify", s.handleVerifyOTP)
}

type otpRequestBody struct {
	PhoneNumber string `json:"phone_number"`
}

func (s *Service) handleRequestOTP(w http.ResponseWriter, r *http.Request) {
	var body otpRequestBody
	if !middleware.DecodeJSON(w, r, &body) {
		return
	}
	phone := strings.TrimSpace(body.PhoneNumber)
	if !ValidPhoneNumber(phone) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PHONE_NUMBER", "Use a Kenyan phone number in the form +2547XXXXXXXX.")
		return
	}

	ctx := r.Context()
	if s.sms == nil {
		middleware.RespondError(w, http.StatusServiceUnavailable, "SMS_UNAVAILABLE", "SMS delivery is not configured.")
		return
	}
	otp, err := GenerateOTP()
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to generate an OTP.")
		return
	}
	requestID, err := s.reserveOTP(ctx, phone, HashOTP(otp))
	if err != nil {
		var limit *rateLimitError
		if errors.As(err, &limit) {
			middleware.RespondErrorDetails(w, http.StatusTooManyRequests, limit.Code, limit.Message, limit.Details)
			return
		}
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to process the OTP request.")
		return
	}

	if err := s.sms.SendTransactionalSMS(ctx, phone, fmt.Sprintf(otpSMSMessage, otp)); err != nil {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM otp_requests WHERE id = $1`, requestID)
		middleware.RespondError(w, http.StatusBadGateway, "SMS_DELIVERY_FAILED", "Unable to send the OTP SMS.")
		return
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"message":          "OTP sent",
		"cooldown_seconds": int(requestCooldown.Seconds()),
	})
}

type otpVerifyBody struct {
	PhoneNumber string `json:"phone_number"`
	OTP         string `json:"otp"`
}

type userResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
}

type membershipResponse struct {
	ID         string   `json:"id"`
	SchoolID   string   `json:"school_id"`
	SchoolName string   `json:"school_name"`
	Role       string   `json:"role"`
	Badges     []string `json:"badges"`
}

func (s *Service) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var body otpVerifyBody
	if !middleware.DecodeJSON(w, r, &body) {
		return
	}
	phone := strings.TrimSpace(body.PhoneNumber)
	otp := strings.TrimSpace(body.OTP)
	if !ValidPhoneNumber(phone) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_PHONE_NUMBER", "Use a Kenyan phone number in the form +2547XXXXXXXX.")
		return
	}
	if !otpPattern.MatchString(otp) {
		middleware.RespondError(w, http.StatusUnprocessableEntity, "INVALID_OTP", "The OTP must be 6 digits.")
		return
	}

	ctx := r.Context()
	requestID, storedHash, err := s.latestOTP(ctx, phone)
	if err != nil {
		middleware.RespondError(w, http.StatusUnauthorized, "INVALID_OTP", "The OTP is invalid or has expired.")
		return
	}
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(HashOTP(otp))) != 1 {
		attempts, err := s.recordFailedAttempt(ctx, requestID)
		if err != nil {
			middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify the OTP.")
			return
		}
		if attempts >= maxOTPAttempts {
			middleware.RespondErrorDetails(w, http.StatusTooManyRequests, "OTP_ATTEMPTS_EXCEEDED",
				"Too many incorrect attempts. Request a new code.",
				map[string]any{"max_attempts": maxOTPAttempts})
			return
		}
		middleware.RespondError(w, http.StatusUnauthorized, "INVALID_OTP", "The OTP is invalid or has expired.")
		return
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE otp_requests SET consumed_at = NOW() WHERE id = $1 AND consumed_at IS NULL`, requestID)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify the OTP.")
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		middleware.RespondError(w, http.StatusUnauthorized, "INVALID_OTP", "The OTP has already been used.")
		return
	}

	user, err := s.userByPhone(ctx, phone)
	if err != nil {
		middleware.RespondError(w, http.StatusUnauthorized, "ACCOUNT_NOT_FOUND", "No active account uses this phone number.")
		return
	}

	memberships, err := s.membershipsForUser(ctx, user.ID)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load school memberships.")
		return
	}

	token, expiresAt, err := s.tokens.Issue(user.ID, user.PhoneNumber)
	if err != nil {
		middleware.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to issue a session token.")
		return
	}

	if s.audit != nil {
		_ = s.audit.Log(ctx, nil, database.AuditEvent{
			UserID:       user.ID,
			Action:       "auth.login",
			ResourceType: "user",
			ResourceID:   user.ID,
			IPAddress:    s.clientIP(r),
			UserAgent:    r.UserAgent(),
		})
	}

	middleware.RespondJSON(w, http.StatusOK, map[string]any{
		"token":       token,
		"expires_at":  expiresAt.UTC().Format(time.RFC3339),
		"user":        user,
		"memberships": memberships,
	})
}

// ValidPhoneNumber reports whether phone is a Kenyan mobile number in E.164 form.
func ValidPhoneNumber(phone string) bool {
	return phonePattern.MatchString(phone)
}

// GenerateOTP returns a cryptographically random 6-digit code.
func GenerateOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", otpCodeLength, value.Int64()), nil
}

// HashOTP returns the SHA-256 hex digest stored in otp_requests.
func HashOTP(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}

// reserveOTP checks the sender's rate limits and stores a new OTP in one
// transaction. A per-phone advisory lock serialises concurrent requests so two
// callers cannot both pass the cooldown check.
func (s *Service) reserveOTP(ctx context.Context, phone, otpHash string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "otp:"+phone); err != nil {
		return "", err
	}

	if retryAfter, blocked, err := activeBlock(ctx, tx, phone); err != nil {
		return "", err
	} else if blocked {
		return "", &rateLimitError{
			Code:    "SENDER_BLOCKED",
			Message: "Too many requests. Try again later.",
			Details: map[string]any{"retry_after_seconds": int(retryAfter.Seconds())},
		}
	}

	recent, err := recentRequestCount(ctx, tx, phone)
	if err != nil {
		return "", err
	}
	if recent >= blockThreshold {
		if err := blockSender(ctx, tx, phone); err != nil {
			return "", err
		}
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return "", &rateLimitError{
			Code:    "SENDER_BLOCKED",
			Message: "Too many requests. Try again later.",
			Details: map[string]any{"retry_after_seconds": int(blockDuration.Seconds())},
		}
	}

	if cooldown, waiting, err := cooldownRemaining(ctx, tx, phone); err != nil {
		return "", err
	} else if waiting {
		return "", &rateLimitError{
			Code:    "OTP_COOLDOWN",
			Message: "Please wait before requesting another OTP.",
			Details: map[string]any{"cooldown_seconds": int(cooldown.Seconds())},
		}
	}

	var id string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO otp_requests (phone_number, sender_identifier, otp_hash, purpose, expires_at)
		 VALUES ($1, $1, $2, $3, NOW() + INTERVAL '5 minutes')
		 RETURNING id::text`,
		phone, otpHash, otpRequestPurpose).Scan(&id); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

// recordFailedAttempt increments the failed-attempt counter and invalidates the
// OTP once the maximum is reached. It returns the new attempt count.
func (s *Service) recordFailedAttempt(ctx context.Context, requestID string) (int, error) {
	var attempts int
	err := s.db.QueryRowContext(ctx,
		`UPDATE otp_requests
		    SET failed_attempts = failed_attempts + 1,
		        consumed_at = CASE WHEN failed_attempts + 1 >= $2 THEN NOW() ELSE consumed_at END
		  WHERE id = $1 AND consumed_at IS NULL
		  RETURNING failed_attempts`, requestID, maxOTPAttempts).Scan(&attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return attempts, err
}

func activeBlock(ctx context.Context, q sqlQuerier, sender string) (time.Duration, bool, error) {
	var blockedUntil time.Time
	err := q.QueryRowContext(ctx,
		`SELECT blocked_until FROM sender_rate_blocks WHERE sender_identifier = $1 AND blocked_until > NOW()`,
		sender).Scan(&blockedUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return time.Until(blockedUntil), true, nil
}

func recentRequestCount(ctx context.Context, q sqlQuerier, sender string) (int, error) {
	var count int
	err := q.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM otp_requests WHERE sender_identifier = $1 AND created_at >= NOW() - INTERVAL '30 minutes'`,
		sender).Scan(&count)
	return count, err
}

func blockSender(ctx context.Context, q sqlQuerier, sender string) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO sender_rate_blocks (sender_identifier, blocked_until, reason)
		 VALUES ($1, NOW() + INTERVAL '24 hours', 'otp_request_threshold')
		 ON CONFLICT (sender_identifier) DO UPDATE
		   SET blocked_until = EXCLUDED.blocked_until,
		       reason = EXCLUDED.reason,
		       created_at = NOW()`,
		sender)
	return err
}

func cooldownRemaining(ctx context.Context, q sqlQuerier, phone string) (time.Duration, bool, error) {
	var last time.Time
	err := q.QueryRowContext(ctx,
		`SELECT created_at FROM otp_requests
		  WHERE phone_number = $1 AND purpose = $2
		  ORDER BY created_at DESC LIMIT 1`,
		phone, otpRequestPurpose).Scan(&last)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	remaining := requestCooldown - time.Since(last)
	if remaining <= 0 {
		return 0, false, nil
	}
	return remaining, true, nil
}

func (s *Service) latestOTP(ctx context.Context, phone string) (string, string, error) {
	var id, hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id::text, otp_hash FROM otp_requests
		  WHERE phone_number = $1 AND purpose = $2
		    AND consumed_at IS NULL AND expires_at > NOW()
		  ORDER BY created_at DESC LIMIT 1`,
		phone, otpRequestPurpose).Scan(&id, &hash)
	return id, hash, err
}

func (s *Service) userByPhone(ctx context.Context, phone string) (userResponse, error) {
	var user userResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT id::text, phone_number, full_name FROM users
		  WHERE phone_number = $1 AND is_active = TRUE`, phone).
		Scan(&user.ID, &user.PhoneNumber, &user.FullName)
	return user, err
}

func (s *Service) membershipsForUser(ctx context.Context, userID string) ([]membershipResponse, error) {
	var memberships []membershipResponse
	err := database.WithUserTx(ctx, s.db, userID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT m.id::text, m.school_id::text, sc.name, m.role
			   FROM memberships m
			   JOIN schools sc ON sc.id = m.school_id
			  WHERE m.user_id = $1 AND m.status = 'active'
			  ORDER BY sc.name`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var membership membershipResponse
			if err := rows.Scan(&membership.ID, &membership.SchoolID, &membership.SchoolName, &membership.Role); err != nil {
				return err
			}
			memberships = append(memberships, membership)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	for i := range memberships {
		badges, err := s.badgesForMembership(ctx, memberships[i].SchoolID, memberships[i].ID)
		if err != nil {
			return nil, err
		}
		memberships[i].Badges = badges
	}
	return memberships, nil
}

func (s *Service) badgesForMembership(ctx context.Context, schoolID, membershipID string) ([]string, error) {
	var badges []string
	err := database.WithTenantTx(ctx, s.db, schoolID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT b.slug FROM membership_badges mb
			   JOIN badges b ON b.id = mb.badge_id AND b.school_id = $2
			  WHERE mb.membership_id = $1
			    AND (mb.valid_from IS NULL OR mb.valid_from <= NOW())
			    AND (mb.valid_until IS NULL OR mb.valid_until > NOW())
			  ORDER BY b.slug`, membershipID, schoolID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err != nil {
				return err
			}
			badges = append(badges, slug)
		}
		return rows.Err()
	})
	return badges, err
}

func (s *Service) clientIP(r *http.Request) string {
	if s.trustProxyHeaders {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			first, _, _ := strings.Cut(forwarded, ",")
			if ip := net.ParseIP(strings.TrimSpace(first)); ip != nil {
				return ip.String()
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

var _ middleware.TokenVerifier = (*TokenManager)(nil)
