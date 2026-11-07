package internalapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// UserContext carries the authenticated user's authorization and channel context
type UserContext struct {
	UserID    string `json:"user_id"`
	SchoolID  string `json:"school_id"`
	Role      string `json:"role"`
	Channel   string `json:"channel"` // "whatsapp" or "in_app"
	SessionID string `json:"session_id,omitempty"`
	Exp       int64  `json:"exp"` // Unix timestamp
}

// ContextSigner signs and verifies HMAC-SHA256 signatures for X-User-Context
type ContextSigner struct {
	SecretKey []byte
}

func NewContextSigner(secretKey string) *ContextSigner {
	if secretKey == "" {
		secretKey = "default-ikonex-internal-hmac-secret-change-in-prod"
	}
	return &ContextSigner{
		SecretKey: []byte(secretKey),
	}
}

// Sign serializes and generates an HMAC signature: base64(payload).signatureHex
func (cs *ContextSigner) Sign(u UserContext, ttl time.Duration) (string, error) {
	if u.Exp == 0 {
		if ttl <= 0 {
			ttl = 15 * time.Minute // Short-lived user context
		}
		u.Exp = time.Now().Add(ttl).Unix()
	}

	rawJSON, err := json.Marshal(u)
	if err != nil {
		return "", err
	}

	b64Payload := base64.RawURLEncoding.EncodeToString(rawJSON)

	mac := hmac.New(sha256.New, cs.SecretKey)
	mac.Write([]byte(b64Payload))
	sigHex := hex.EncodeToString(mac.Sum(nil))

	return b64Payload + "." + sigHex, nil
}

// Verify decodes and verifies signature and expiration of X-User-Context
func (cs *ContextSigner) Verify(signedContext string) (*UserContext, error) {
	parts := strings.Split(signedContext, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid signed context format")
	}

	b64Payload := parts[0]
	sigHex := parts[1]

	// Verify HMAC signature
	mac := hmac.New(sha256.New, cs.SecretKey)
	mac.Write([]byte(b64Payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigHex), []byte(expectedSig)) {
		return nil, errors.New("invalid context HMAC signature")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		return nil, errors.New("invalid base64 payload")
	}

	var u UserContext
	if err := json.Unmarshal(payloadBytes, &u); err != nil {
		return nil, errors.New("invalid context payload JSON")
	}

	// Verify expiration
	if time.Now().Unix() > u.Exp {
		return nil, errors.New("context has expired")
	}

	return &u, nil
}
