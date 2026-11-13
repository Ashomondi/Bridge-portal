package internalapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
)

const (
	HeaderInternalServiceToken = "X-Internal-Service-Token"
	HeaderUserContext          = "X-User-Context"
)

type UserContext struct {
	UserID       string `json:"user_id"`
	SchoolID     string `json:"school_id"`
	MembershipID string `json:"membership_id"`
	Role         string `json:"role"`
	Channel      string `json:"channel"`
	SessionID    string `json:"session_id"`
	Exp          int64  `json:"exp"`
}

func ValidateUserContext(ctx UserContext, now time.Time) error {
	if !contracts.ValidID(ctx.UserID) || !contracts.ValidID(ctx.SchoolID) || !contracts.ValidID(ctx.MembershipID) {
		return contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "User context is missing valid user, school, or membership identifiers.")
	}
	if strings.TrimSpace(ctx.Role) == "" {
		return contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "User context is missing a role.")
	}
	switch ctx.Channel {
	case "whatsapp", "in_app":
	default:
		return contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "User context channel must be 'whatsapp' or 'in_app'.")
	}
	if ctx.Channel == "whatsapp" && !contracts.ValidID(ctx.SessionID) {
		return contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "WhatsApp user context requires a valid session_id.")
	}
	if ctx.SessionID != "" && !contracts.ValidID(ctx.SessionID) {
		return contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "session_id must be a valid UUID.")
	}
	if ctx.Exp <= now.UTC().Unix() {
		return contracts.Fail(http.StatusUnauthorized, "EXPIRED_USER_CONTEXT", "User context has expired.")
	}
	return nil
}

// SignUserContext encodes and signs UserContext in the exact Master Integration Plan format:
// base64url_raw(json_payload) + "." + hex(hmac_sha256(base64url_payload, signingKey))
func SignUserContext(ctx UserContext, signingKey string) (string, error) {
	key := strings.TrimSpace(signingKey)
	if key == "" {
		return "", contracts.Fail(http.StatusInternalServerError, "MISSING_SIGNING_KEY", "Context signing key is not configured.")
	}
	if !contracts.ValidID(ctx.UserID) || !contracts.ValidID(ctx.SchoolID) || !contracts.ValidID(ctx.MembershipID) || strings.TrimSpace(ctx.Role) == "" || (ctx.Channel != "whatsapp" && ctx.Channel != "in_app") || ctx.Exp <= 0 {
		return "", contracts.Fail(http.StatusUnprocessableEntity, "INVALID_USER_CONTEXT", "Cannot sign an incomplete user context.")
	}
	payloadBytes, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(encodedPayload))
	signatureHex := hex.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signatureHex, nil
}

// VerifyUserContext verifies X-User-Context using constant-time HMAC-SHA256 comparison
// BEFORE trusting or decoding the JSON payload, and validates all required fields and expiration.
func VerifyUserContext(headerValue, signingKey string, now time.Time) (UserContext, error) {
	key := strings.TrimSpace(signingKey)
	if key == "" {
		return UserContext{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_INTERNAL_SERVICE", "Context signing key is not configured.")
	}
	trimmed := strings.TrimSpace(headerValue)
	if trimmed == "" {
		return UserContext{}, contracts.Fail(http.StatusUnauthorized, "MISSING_USER_CONTEXT", "X-User-Context header is required.")
	}
	parts := strings.Split(trimmed, ".")
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) != 64 {
		return UserContext{}, contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT_SIGNATURE", "Malformed X-User-Context header.")
	}

	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(parts[0]))
	expectedSigHex := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(strings.ToLower(parts[1])), []byte(expectedSigHex)) != 1 {
		return UserContext{}, contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT_SIGNATURE", "Invalid X-User-Context HMAC signature.")
	}

	rawJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		rawJSON, err = base64.URLEncoding.DecodeString(parts[0])
		if err != nil {
			return UserContext{}, contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "Invalid base64url payload in X-User-Context.")
		}
	}

	dec := json.NewDecoder(bytes.NewReader(rawJSON))
	dec.DisallowUnknownFields()
	var parsed UserContext
	if err := dec.Decode(&parsed); err != nil {
		return UserContext{}, contracts.Fail(http.StatusUnauthorized, "INVALID_USER_CONTEXT", "Malformed JSON payload in X-User-Context.")
	}
	if err := ValidateUserContext(parsed, now); err != nil {
		return UserContext{}, err
	}
	return parsed, nil
}

// VerifyServiceToken performs a constant-time comparison of X-Internal-Service-Token.
func VerifyServiceToken(providedToken, expectedToken string) error {
	expected := strings.TrimSpace(expectedToken)
	provided := strings.TrimSpace(providedToken)
	if expected == "" || provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_INTERNAL_SERVICE", "Valid X-Internal-Service-Token is required.")
	}
	return nil
}
