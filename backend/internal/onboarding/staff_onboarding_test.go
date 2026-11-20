package onboarding

import (
	"testing"
	"time"
)

func TestDefaultStaffWindow(t *testing.T) {
	now := time.Date(2026, 11, 5, 13, 45, 0, 0, time.UTC)
	start, end := defaultStaffWindow(now)

	wantStart := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 11, 5, 18, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("window = %v..%v, want %v..%v", start, end, wantStart, wantEnd)
	}
}

func TestParseStaffWindowDefaults(t *testing.T) {
	now := time.Date(2026, 11, 5, 20, 0, 0, 0, time.UTC)
	start, end, err := parseStaffWindow("", "", now)
	if err != nil {
		t.Fatalf("parseStaffWindow: %v", err)
	}
	if start.Hour() != 8 || end.Hour() != 18 {
		t.Fatalf("default window = %v..%v", start, end)
	}
}

func TestParseStaffWindowExplicit(t *testing.T) {
	start, end, err := parseStaffWindow("2026-11-06T09:00:00Z", "2026-11-06T15:30:00Z", time.Now())
	if err != nil {
		t.Fatalf("parseStaffWindow: %v", err)
	}
	if start.Hour() != 9 || end.Hour() != 15 || end.Minute() != 30 {
		t.Fatalf("explicit window = %v..%v", start, end)
	}
}

func TestParseStaffWindowRejectsInvalid(t *testing.T) {
	if _, _, err := parseStaffWindow("tomorrow", "", time.Now()); err == nil {
		t.Fatalf("invalid valid_from accepted")
	}
	if _, _, err := parseStaffWindow("", "2999-01-01T09:00:00Z", time.Now()); err != nil {
		t.Fatalf("valid valid_to should be accepted with default from: %v", err)
	}
	// Explicit window where the end is before the start.
	if _, _, err := parseStaffWindow("2026-11-06T15:00:00Z", "2026-11-06T09:00:00Z", time.Now()); err == nil {
		t.Fatalf("inverted window accepted")
	}
}

func TestGenerateQRToken(t *testing.T) {
	const schoolID = "123e4567-e89b-12d3-a456-426614174000"
	first, err := GenerateQRToken(schoolID)
	if err != nil {
		t.Fatalf("GenerateQRToken: %v", err)
	}
	second, err := GenerateQRToken(schoolID)
	if err != nil {
		t.Fatalf("GenerateQRToken: %v", err)
	}
	if first == "" || first == second {
		t.Fatalf("tokens should be non-empty and unique")
	}
	if len(first) < 40 {
		t.Fatalf("token is unexpectedly short: %d", len(first))
	}
	if parsed, ok := ParseQRToken(first); !ok || parsed != schoolID {
		t.Fatalf("ParseQRToken(%q) = %q, %v", first, parsed, ok)
	}
}

func TestParseQRTokenRejectsMalformed(t *testing.T) {
	for _, token := range []string{"", "abc", "not-a-uuid.suffix", "123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174000."} {
		if _, ok := ParseQRToken(token); ok {
			t.Fatalf("ParseQRToken(%q) accepted a malformed token", token)
		}
	}
}
