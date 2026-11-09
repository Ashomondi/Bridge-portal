package consent

import "testing"

func TestValidPurposes(t *testing.T) {
	for _, purpose := range []string{"biometric_attendance", "photo", "sms", "whatsapp"} {
		if !validPurposes[purpose] {
			t.Fatalf("purpose %q should be accepted", purpose)
		}
	}
	for _, purpose := range []string{"", "results", "attendance"} {
		if validPurposes[purpose] {
			t.Fatalf("purpose %q should be rejected", purpose)
		}
	}
}

func TestUUIDPattern(t *testing.T) {
	if !uuidPattern.MatchString("123e4567-e89b-12d3-a456-426614174000") {
		t.Fatalf("valid UUID was rejected")
	}
	if uuidPattern.MatchString("not-a-uuid") {
		t.Fatalf("invalid UUID was accepted")
	}
}
