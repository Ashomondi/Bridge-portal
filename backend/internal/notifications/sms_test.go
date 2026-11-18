package notifications

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
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

func TestSandboxSMSRedactsSensitiveData(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	}))
	dispatcher := NewSandboxSMSDispatcher(logger)
	message := SMSMessage{To: "0712345678", Body: "Your synthetic OTP is 654321", Purpose: WhatsAppSessionOTP}
	if err := dispatcher.DispatchSMS(context.Background(), message); err != nil {
		t.Fatal(err)
	}
	first := output.String()
	for _, secret := range []string{message.To, "+254712345678", message.Body, "654321"} {
		if strings.Contains(first, secret) {
			t.Fatal("sandbox exposed sensitive message data")
		}
	}
	if !strings.Contains(first, `"recipient":"***678"`) || !strings.Contains(first, `"purpose":"whatsapp_session"`) {
		t.Fatalf("safe dispatch metadata missing: %s", first)
	}
	output.Reset()
	if err := dispatcher.DispatchSMS(context.Background(), message); err != nil || output.String() != first {
		t.Fatalf("nondeterministic sandbox output: %v", err)
	}
	output.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := dispatcher.DispatchSMS(ctx, message); !errors.Is(err, context.Canceled) || output.Len() != 0 {
		t.Fatalf("canceled dispatch: %v, %s", err, output.String())
	}
	message.Purpose = "absence"
	if err := dispatcher.DispatchSMS(context.Background(), message); !errors.Is(err, ErrInvalidSMS) || output.Len() != 0 {
		t.Fatalf("invalid dispatch: %v, %s", err, output.String())
	}
}

func TestSandboxSMSConcurrentDispatch(t *testing.T) {
	var output bytes.Buffer
	dispatcher := NewSandboxSMSDispatcher(slog.New(slog.NewJSONHandler(&output, nil)))
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := dispatcher.DispatchSMS(context.Background(), SMSMessage{
				To: "0712345678", Body: "Synthetic test", Purpose: AppLoginOTP,
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if strings.Count(output.String(), "sandbox transactional SMS accepted") != 20 {
		t.Fatal("sandbox lost dispatch metadata")
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
