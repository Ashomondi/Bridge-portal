package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/whatsapp"
)

func TestVerifiedWhatsAppChatRoutesThroughHTTPBridge(t *testing.T) {
	secret := "synthetic-service-secret"
	signer := MustNewContextSigner(secret)
	calls := 0
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/v1/ai/parent-query" || r.Header.Get(HeaderInternalServiceToken) != secret {
			t.Error("incorrect AI endpoint or authentication")
		}
		user, err := signer.Verify(r.Header.Get(HeaderUserContext))
		if err != nil || user.Channel != "whatsapp" || user.SessionID != "session" || user.SchoolID != "school" || user.UserID != "guardian" {
			t.Errorf("invalid signed context: %+v %v", user, err)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload["query"] != "synthetic fees question" || payload["channel"] != "whatsapp" || payload["student_id"] != "student" || payload["session_id"] != "session" {
			t.Errorf("incorrect AI payload: %+v", payload)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"reply":"Synthetic fee reply"}}`))
	}))
	defer ai.Close()
	facts := NewMemoryStore()
	facts.Guardians["school:student:guardian"] = true
	bridge := MustNewHandler(secret, signer, facts, nil, ai.URL)
	store := whatsapp.NewMemoryStore()
	store.Bots["bot"] = &whatsapp.SchoolWhatsAppBot{SchoolID: "school", IsActive: true}
	expires := time.Now().Add(time.Hour)
	if err := store.SaveSession(context.Background(), &whatsapp.ChatSession{ID: "session", SchoolID: "school", SenderWANumber: "sender", GuardianUserID: "guardian", StudentID: "student", State: whatsapp.StateVerified, ExpiresAt: &expires}); err != nil {
		t.Fatal(err)
	}
	service := whatsapp.NewService(store, nil, bridge, nil)
	reply, err := service.ProcessIncomingMessage(context.Background(), "bot", "sender", "synthetic fees question")
	if err != nil || reply != "Synthetic fee reply" || calls != 1 || len(facts.AuditLogs) != 1 {
		t.Fatalf("chat routing failed: %q %v calls=%d logs=%d", reply, err, calls, len(facts.AuditLogs))
	}
	if _, err := bridge.ForwardParentQuery(context.Background(), "other-school", "guardian", "student", "question", "session"); err == nil {
		t.Fatal("cross-school request accepted")
	}
	if calls != 1 {
		t.Fatal("forbidden request reached AI")
	}
}

func TestFrontendProxyPropagatesAIResponses(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body       string
		wantStatus int
	}{
		{"success", 200, `{"status":"success","data":{"reply":"synthetic reply"}}`, 200},
		{"unauthorized", 401, `{"error":"invalid context"}`, 401},
		{"rate limited", 429, `{"error":"rate limited"}`, 429},
		{"provider failed", 503, `{"error":"service unavailable"}`, 503},
		{"malformed success", 200, `broken`, 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer ai.Close()
			store := NewMemoryStore()
			store.Guardians["school:student:guardian"] = true
			handler := MustNewHandler("synthetic-service-secret", MustNewContextSigner("synthetic-service-secret"), store, nil, ai.URL)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/message", strings.NewReader(`{"student_id":"student","message":"question","session_id":"session"}`))
			req = req.WithContext(billing.WithIdentity(req.Context(), billing.Identity{UserID: "guardian", SchoolID: "school", Role: "parent", MembershipID: "membership"}))
			response := httptest.NewRecorder()
			handler.handleFrontendChatMessage(response, req)
			if response.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if tc.wantStatus != 502 && response.Body.String() != tc.body {
				t.Fatal("AI response altered")
			}
		})
	}
}

type failedBillingBridge struct {
	mockBillingBridge
	err error
}

func (b *failedBillingBridge) InitiateSTKPush(ctx context.Context, req billing.STKPushRequest) (*billing.STKPushResponse, error) {
	return nil, b.err
}
func TestInternalSTKPreservesPhonePauseStatus(t *testing.T) {
	handler, store, signer, _ := setupTestInternalAPI()
	store.Guardians["school:student:guardian"] = true
	handler.billing = &failedBillingBridge{err: &billing.APIError{Status: 403, Code: billing.CodeChatPaymentPaused, Message: "payment paused"}}
	signed, err := signer.Sign(UserContext{SchoolID: "school", UserID: "guardian", Role: "parent"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/payments/stk-push", bytes.NewBufferString(`{"student_id":"student","amount":10.25,"phone_number":"+254700000001","parent_confirmed":true}`))
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, signed)
	response := httptest.NewRecorder()
	handler.authenticateInternal(handler.handleSTKPush)(response, req)
	if response.Code != 403 || !strings.Contains(response.Body.String(), billing.CodeChatPaymentPaused) {
		t.Fatalf("pause error lost: %d %s", response.Code, response.Body.String())
	}
}

type auditFailureStore struct{ Store }

func (auditFailureStore) SaveAIAuditLog(context.Context, *AIAuditLogEntry) error {
	return errors.New("audit unavailable")
}
func TestAIProxyRejectsAuditFailureBeforeCallingAI(t *testing.T) {
	calls := 0
	ai := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer ai.Close()
	store := NewMemoryStore()
	store.Guardians["school:student:guardian"] = true
	handler := MustNewHandler("synthetic-service-secret", MustNewContextSigner("synthetic-service-secret"), auditFailureStore{store}, nil, ai.URL)
	if _, err := handler.ForwardParentQuery(context.Background(), "school", "guardian", "student", "question", "session"); err == nil || calls != 0 {
		t.Fatal("AI called without audit log")
	}
}
