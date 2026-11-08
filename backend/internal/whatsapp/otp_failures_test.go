package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type failedOTPStore struct {
	Store
	failAt string
}

var storeFailure = errors.New("synthetic storage failure")

func (s failedOTPStore) IsSenderBlocked(ctx context.Context, sender string) (bool, time.Time, error) {
	if s.failAt == "blocked" {
		return false, time.Time{}, storeFailure
	}
	return s.Store.IsSenderBlocked(ctx, sender)
}
func (s failedOTPStore) GetActiveSession(ctx context.Context, school, sender string) (*ChatSession, error) {
	if s.failAt == "session" {
		return nil, storeFailure
	}
	return s.Store.GetActiveSession(ctx, school, sender)
}
func (s failedOTPStore) SaveSession(ctx context.Context, session *ChatSession) error {
	if s.failAt == "save_session" {
		return storeFailure
	}
	return s.Store.SaveSession(ctx, session)
}
func (s failedOTPStore) HasRecentOTPForSender(ctx context.Context, sender string, w time.Duration) (bool, error) {
	if s.failAt == "sender_cooldown" {
		return false, storeFailure
	}
	return s.Store.HasRecentOTPForSender(ctx, sender, w)
}
func (s failedOTPStore) HasRecentOTPForStudent(ctx context.Context, school, admission string, w time.Duration) (bool, error) {
	if s.failAt == "student_cooldown" {
		return false, storeFailure
	}
	return s.Store.HasRecentOTPForStudent(ctx, school, admission, w)
}
func (s failedOTPStore) GetRecentOTPRequestCount(ctx context.Context, sender string, w time.Duration) (int, error) {
	if s.failAt == "count" {
		return 0, storeFailure
	}
	if s.failAt == "block_write" {
		return 5, nil
	}
	return s.Store.GetRecentOTPRequestCount(ctx, sender, w)
}
func (s failedOTPStore) BlockSender(ctx context.Context, school, sender, reason string, until time.Time) error {
	if s.failAt == "block_write" {
		return storeFailure
	}
	return s.Store.BlockSender(ctx, school, sender, reason, until)
}
func (s failedOTPStore) SaveOTP(ctx context.Context, id, hash string, expiry time.Time) error {
	if s.failAt == "otp_write" {
		return storeFailure
	}
	return s.Store.SaveOTP(ctx, id, hash, expiry)
}
func (s failedOTPStore) VerifyOTP(ctx context.Context, id, code string) (bool, error) {
	if s.failAt == "verify" {
		return false, storeFailure
	}
	return s.Store.VerifyOTP(ctx, id, code)
}

type failedSMS struct{}

func (failedSMS) SendTransactionalSMS(context.Context, string, string) error {
	return errors.New("provider-secret OTP 123456")
}

func TestOTPWorkflowPropagatesFailures(t *testing.T) {
	for _, phase := range []string{"blocked", "session", "save_session", "sender_cooldown", "student_cooldown", "count", "block_write", "otp_write", "sms", "missing_sms", "verify"} {
		t.Run(phase, func(t *testing.T) {
			service, store, sms, _ := setupTestWhatsApp()
			store.Bots["bot"] = &SchoolWhatsAppBot{SchoolID: "school", IsActive: true}
			store.Students["school:SYNTHETIC"] = struct{ ID, Name string }{"student", "Synthetic Student"}
			store.Guardians["school:student"] = []GuardianCandidate{{UserID: "guardian", PhoneNumber: "+254700000001", MaskedPhone: "07•• ••• 001"}}
			service.store = failedOTPStore{Store: store, failAt: phase}
			if phase == "sms" {
				service.sms = failedSMS{}
			}
			if phase == "missing_sms" {
				service.sms = nil
			}
			text := "SYNTHETIC"
			if phase == "verify" {
				if err := store.SaveSession(context.Background(), &ChatSession{SchoolID: "school", SenderWANumber: "sender", State: StateAwaitingOTP}); err != nil {
					t.Fatal(err)
				}
				text = "123456"
			}
			reply, err := service.ProcessIncomingMessage(context.Background(), "bot", "sender", text)
			if err == nil || reply != "" {
				t.Fatalf("reported success: %q %v", reply, err)
			}
			if strings.Contains(err.Error(), "provider-secret") || strings.Contains(err.Error(), "123456") {
				t.Fatal("sensitive error exposed")
			}
			if len(sms.sent) != 0 {
				t.Fatal("unexpected SMS dispatch")
			}
		})
	}
}

func TestOTPRejectsHashReplayExpiredAndConsumedCodes(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if err := store.SaveSession(ctx, &ChatSession{ID: "session", SchoolID: "school"}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		code     string
		expires  time.Time
		accepted bool
	}{
		{hashOTP("123456"), time.Now().Add(time.Minute), false},
		{"123456", time.Now().Add(-time.Second), false},
		{"123456", time.Now().Add(time.Minute), true},
	} {
		if err := store.SaveOTP(ctx, "session", hashOTP("123456"), tc.expires); err != nil {
			t.Fatal(err)
		}
		valid, _ := store.VerifyOTP(ctx, "session", tc.code)
		if valid != tc.accepted {
			t.Fatalf("OTP acceptance=%v", valid)
		}
		if valid {
			if replay, _ := store.VerifyOTP(ctx, "session", tc.code); replay {
				t.Fatal("consumed OTP replay accepted")
			}
		}
	}
}

func TestWebhookDoesNotAcknowledgeFailedWorkflow(t *testing.T) {
	service, store, _, _ := setupTestWhatsApp()
	store.Bots["bot"] = &SchoolWhatsAppBot{SchoolID: "school", IsActive: true}
	service.store = failedOTPStore{Store: store, failAt: "session"}
	service.SetWebhookAuth("synthetic-token", "synthetic-secret")
	body := `{"entry":[{"changes":[{"value":{"metadata":{"display_phone_number":"bot"},"messages":[{"from":"sender","text":{"body":"hello"}}]}}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/whatsapp/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signTestWebhook([]byte(body), "synthetic-secret"))
	response := httptest.NewRecorder()
	service.handleWebhook(response, req)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", response.Code)
	}
}

func signTestWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookRejectsMalformedAndUndeliverableReply(t *testing.T) {
	for _, body := range []string{"invalid", `{"entry":[{"changes":[{"value":{"metadata":{"display_phone_number":"bot"},"messages":[{"from":"sender","text":{"body":"unknown"}}]}}]}]}`} {
		service, store, _, _ := setupTestWhatsApp()
		store.Bots["bot"] = &SchoolWhatsAppBot{SchoolID: "school", IsActive: true}
		service.waClient = nil
		service.SetWebhookAuth("synthetic-token", "synthetic-secret")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/whatsapp/webhook", strings.NewReader(body))
		req.Header.Set("X-Hub-Signature-256", signTestWebhook([]byte(body), "synthetic-secret"))
		response := httptest.NewRecorder()
		service.handleWebhook(response, req)
		if response.Code != http.StatusBadRequest && response.Code != http.StatusServiceUnavailable {
			t.Fatalf("undeliverable webhook acknowledged: %d", response.Code)
		}
	}
}
