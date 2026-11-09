package notifications

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizePhone(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"0712345678", "+254712345678"},
		{"0112345678", "+254112345678"},
		{"254712345678", "+254712345678"},
		{" +254 (712) 345-678 ", "+254712345678"},
		{"+442079460000", "+442079460000"},
		{"", ""}, {"+", ""}, {"071234567", ""},
		{"+2547123456789", ""}, {"+254012345678", ""},
		{"+254212345678", ""}, {"+0712345678", ""},
		{"0712345678\napiKey=secret", ""}, {"+254７12345678", ""},
		{"+254+712345678", ""}, {"712345678", ""},
		{"+1234567890123456", ""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := NormalizePhone(tc.input)
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
			if err != nil && !errors.Is(err, ErrInvalidPhone) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestTransactionalSMSValidation(t *testing.T) {
	for _, purpose := range []SMSPurpose{AppLoginOTP, WhatsAppSessionOTP, PhoneChangeOTP, PhoneChangeNotice} {
		msg, err := (SMSMessage{To: "0712345678", Body: "Synthetic test message", Purpose: purpose}).normalized()
		if err != nil || msg.To != "+254712345678" || msg.Purpose != purpose {
			t.Fatalf("purpose %s: %+v %v", purpose, msg, err)
		}
	}
	for _, msg := range []SMSMessage{
		{To: "0712345678", Body: "Absence alert", Purpose: "absence"},
		{To: "0712345678", Body: "Sickbay alert", Purpose: "sickbay"},
		{To: "0712345678", Body: "hello"},
		{To: "0712345678", Body: " \n ", Purpose: AppLoginOTP},
		{To: "0712345678", Body: string([]byte{0xff}), Purpose: AppLoginOTP},
		{To: "private-invalid-number", Body: "654321", Purpose: AppLoginOTP},
	} {
		_, err := msg.normalized()
		if err == nil {
			t.Fatal("invalid message accepted")
		}
		if strings.Contains(err.Error(), msg.To) || strings.Contains(err.Error(), "654321") {
			t.Fatal("validation error exposed private message data")
		}
	}
}
