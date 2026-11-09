package auth

import (
	"strings"
	"testing"
	"time"
)

func TestValidPhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"safaricom", "+254712345678", true},
		{"airtel", "+254100000000", true},
		{"missing plus", "254712345678", false},
		{"local format", "0712345678", false},
		{"too short", "+25471234567", false},
		{"letters", "+2547abcdefgh", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidPhoneNumber(tc.phone); got != tc.want {
				t.Fatalf("ValidPhoneNumber(%q) = %v, want %v", tc.phone, got, tc.want)
			}
		})
	}
}

func TestGenerateOTPFormat(t *testing.T) {
	code, err := GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}
	if len(code) != otpCodeLength {
		t.Fatalf("len(OTP) = %d, want %d", len(code), otpCodeLength)
	}
	if !otpPattern.MatchString(code) {
		t.Fatalf("OTP %q is not 6 digits", code)
	}
}

func TestHashOTPStableAndSized(t *testing.T) {
	first := HashOTP("123456")
	second := HashOTP("123456")
	if first != second {
		t.Fatalf("HashOTP is not deterministic")
	}
	if len(first) != 64 {
		t.Fatalf("len(HashOTP) = %d, want 64", len(first))
	}
	if HashOTP("654321") == first {
		t.Fatalf("different OTPs must not share a hash")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-secret-at-least-32-bytes-long", time.Hour)
	token, expiresAt, err := manager.Issue("user-1", "+254712345678")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiry %v is not in the future", expiresAt)
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("token is not a three-part JWT: %q", token)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != "user-1" || claims.PhoneNumber != "+254712345678" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestTokenRejectsTamperingAndExpiry(t *testing.T) {
	manager := NewTokenManager("test-secret-at-least-32-bytes-long", time.Hour)
	token, _, err := manager.Issue("user-1", "+254712345678")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if _, err := manager.Verify(token + "x"); err == nil {
		t.Fatalf("tampered token was accepted")
	}
	if _, err := manager.Verify("not-a-token"); err == nil {
		t.Fatalf("malformed token was accepted")
	}

	expired := NewTokenManager("test-secret-at-least-32-bytes-long", -time.Hour)
	expiredToken, _, err := expired.Issue("user-1", "+254712345678")
	if err != nil {
		t.Fatalf("Issue (expired): %v", err)
	}
	if _, err := manager.Verify(expiredToken); err == nil {
		t.Fatalf("expired token was accepted")
	}
}
