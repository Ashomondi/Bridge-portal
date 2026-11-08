package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
)

type mockBillingBridge struct {
	lastSTKReq billing.STKPushRequest
}

func (m *mockBillingBridge) GetStudentFeeBalance(ctx context.Context, schoolID, studentID, callerUserID, callerRole string) (*billing.StudentBalance, error) {
	return &billing.StudentBalance{
		StudentID:      studentID,
		CurrentBalance: 15000,
	}, nil
}

func (m *mockBillingBridge) InitiateSTKPush(ctx context.Context, req billing.STKPushRequest) (*billing.STKPushResponse, error) {
	m.lastSTKReq = req
	return &billing.STKPushResponse{
		CheckoutRequestID: "ws_CO_INTERNAL_12345",
		ResponseCode:      "0",
		CustomerMessage:   "Payment initiated",
	}, nil
}

func setupTestInternalAPI() (*Handler, *MemoryStore, *ContextSigner, *mockBillingBridge) {
	signer := MustNewContextSigner("test-secret-key-12345")
	store := NewMemoryStore()
	billingBridge := &mockBillingBridge{}
	handler := MustNewHandler("secret-service-token", signer, store, billingBridge, "http://localhost:8001")
	return handler, store, signer, billingBridge
}

func TestInternalAPIServiceTokenAndContextAuth(t *testing.T) {
	handler, _, signer, _ := setupTestInternalAPI()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// 1. Missing X-Internal-Service-Token returns 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/school-profile", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing service token, got %d", rr.Code)
	}

	// 2. Invalid X-Internal-Service-Token returns 401
	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/school-profile", nil)
	req.Header.Set(HeaderInternalServiceToken, "wrong-token")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong service token, got %d", rr.Code)
	}

	// 3. Valid service token but missing X-User-Context returns 401
	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/school-profile", nil)
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing user context, got %d", rr.Code)
	}

	// 4. Tampered X-User-Context returns 401
	validCtx := UserContext{
		UserID:   "user-1",
		SchoolID: "school-1",
		Role:     "parent",
		Channel:  "whatsapp",
		Exp:      time.Now().Add(10 * time.Minute).Unix(),
	}
	signed, _ := signer.Sign(validCtx, 10*time.Minute)
	tampered := signed + "tampered"

	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/school-profile", nil)
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, tampered)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for tampered user context, got %d", rr.Code)
	}

	// 5. Expired X-User-Context returns 401
	expiredCtx := UserContext{
		UserID:   "user-1",
		SchoolID: "school-1",
		Role:     "parent",
		Exp:      time.Now().Add(-5 * time.Minute).Unix(),
	}
	expiredSigned, _ := signer.Sign(expiredCtx, -5*time.Minute)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/internal/school-profile", nil)
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, expiredSigned)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired user context, got %d", rr.Code)
	}
}

func TestGuardianChildOwnershipEnforcement(t *testing.T) {
	handler, store, signer, _ := setupTestInternalAPI()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	schoolID := "school-1"
	parentA := "parent-A-user"
	studentA := "student-A-child"
	studentB := "student-B-child" // child of another parent!

	// Register Parent A as guardian ONLY of Student A
	store.Guardians[schoolID+":"+studentA+":"+parentA] = true

	uCtx := UserContext{
		UserID:   parentA,
		SchoolID: schoolID,
		Role:     "parent",
		Channel:  "whatsapp",
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
	}
	signedCtx, _ := signer.Sign(uCtx, 15*time.Minute)

	// 1. Parent A queries their OWN child (student A) -> ALLOWED (200 OK)
	reqAllowed := httptest.NewRequest(http.MethodGet, "/api/v1/internal/students/"+studentA+"/fee-balance", nil)
	reqAllowed.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	reqAllowed.Header.Set(HeaderUserContext, signedCtx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, reqAllowed)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK for own child lookup, got %d", rr.Code)
	}

	// 2. Parent A queries Student B (not their child) -> FORBIDDEN (403)
	reqForbidden := httptest.NewRequest(http.MethodGet, "/api/v1/internal/students/"+studentB+"/fee-balance", nil)
	reqForbidden.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	reqForbidden.Header.Set(HeaderUserContext, signedCtx)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, reqForbidden)
	if rr.Code != http.StatusForbidden {
		t.Errorf("SECURITY FAULT: expected 403 Forbidden for querying another guardian's child, got %d", rr.Code)
	}
}

func TestInternalSTKPushAndConfirmationCheck(t *testing.T) {
	handler, store, signer, billingBridge := setupTestInternalAPI()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	schoolID := "school-1"
	parentID := "parent-user-1"
	studentID := "student-101"

	store.Guardians[schoolID+":"+studentID+":"+parentID] = true

	uCtx := UserContext{
		UserID:   parentID,
		SchoolID: schoolID,
		Role:     "parent",
		Channel:  "whatsapp",
		Exp:      time.Now().Add(10 * time.Minute).Unix(),
	}
	signedCtx, _ := signer.Sign(uCtx, 10*time.Minute)

	// 1. STK push without explicit parent confirmation: rejected with 422 (AC-23)
	bodyNoConfirm, _ := json.Marshal(map[string]interface{}{
		"student_id":       studentID,
		"amount":           8000,
		"phone_number":     "254712345678",
		"parent_confirmed": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/payments/stk-push", bytes.NewReader(bodyNoConfirm))
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, signedCtx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for unconfirmed payment, got %d", rr.Code)
	}

	// 2. STK push with explicit parent confirmation: processed (200 OK) (AC-23)
	bodyConfirmed, _ := json.Marshal(map[string]interface{}{
		"student_id":       studentID,
		"amount":           8000,
		"phone_number":     "254712345678",
		"parent_confirmed": true,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/internal/payments/stk-push", bytes.NewReader(bodyConfirmed))
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, signedCtx)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK for confirmed payment, got %d: %s", rr.Code, rr.Body.String())
	}
	if billingBridge.lastSTKReq.Amount != 8000 || !billingBridge.lastSTKReq.ParentConfirmed {
		t.Errorf("billing bridge did not receive expected STK request")
	}
}

func TestAIAuditActionLogging(t *testing.T) {
	handler, store, signer, _ := setupTestInternalAPI()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	uCtx := UserContext{
		UserID:    "parent-1",
		SchoolID:  "school-1",
		Role:      "parent",
		Channel:   "whatsapp",
		SessionID: "session-abc-123",
		Exp:       time.Now().Add(10 * time.Minute).Unix(),
	}
	signedCtx, _ := signer.Sign(uCtx, 10*time.Minute)

	body, _ := json.Marshal(AIAuditLogEntry{
		Feature:       "parent_assistant",
		Intent:        "fee_inquiry",
		EndpointsRead: []string{"/api/v1/internal/students/101/fee-balance"},
		ActionTaken:   "answered_fee_balance",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/ai/audit-log", bytes.NewReader(body))
	req.Header.Set(HeaderInternalServiceToken, "secret-service-token")
	req.Header.Set(HeaderUserContext, signedCtx)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for audit log, got %d", rr.Code)
	}

	if len(store.AuditLogs) != 2 {
		t.Fatalf("expected request and action audit logs, got %d", len(store.AuditLogs))
	}
	foundAction := false
	for _, saved := range store.AuditLogs {
		if saved.Feature == "parent_assistant" && saved.Channel == "whatsapp" && saved.SessionID == "session-abc-123" {
			foundAction = true
		}
	}
	if !foundAction {
		t.Errorf("expected explicit AI action log, got %+v", store.AuditLogs)
	}
}

func TestMemoryStoreDoesNotInventAttendanceOrMarks(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.GetAttendanceSummary(ctx, "school-1", "student-1", "term-1"); err == nil {
		t.Fatal("expected missing attendance facts to return an error")
	}
	if _, err := store.GetMarksSummary(ctx, "school-1", "student-1", "term-1"); err == nil {
		t.Fatal("expected missing marks facts to return an error")
	}
}

func TestSecretValidationNoFallback(t *testing.T) {
	// 1. NewContextSigner fails closed on empty or short secret (must be >= 16 bytes)
	_, err := NewContextSigner("")
	if err == nil {
		t.Errorf("expected error for empty signer secret, got nil")
	}
	_, err = NewContextSigner("short-key")
	if err == nil {
		t.Errorf("expected error for short signer secret (<16 bytes), got nil")
	}

	signer, err := NewContextSigner("valid-secret-key-16-bytes!")
	if err != nil {
		t.Fatalf("unexpected error for valid signer key: %v", err)
	}

	store := NewMemoryStore()
	billingBridge := &mockBillingBridge{}

	// 2. NewHandler fails closed on empty or short service token (must be >= 16 bytes)
	_, err = NewHandler("", signer, store, billingBridge, "http://localhost:8001")
	if err == nil {
		t.Errorf("expected error for empty service token, got nil")
	}
	_, err = NewHandler("short-token", signer, store, billingBridge, "http://localhost:8001")
	if err == nil {
		t.Errorf("expected error for short service token (<16 bytes), got nil")
	}
}

func TestChatMessageProxyAuthenticationAndForwarding(t *testing.T) {
	// Setup mock AI server
	var receivedUserContext string
	var receivedBody map[string]interface{}
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserContext = r.Header.Get("X-User-Context")
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"reply": "Hello from mock AI assistant!"}`))
	}))
	defer mockAIServer.Close()

	signer := MustNewContextSigner("test-secret-key-12345")
	store := NewMemoryStore()
	billingBridge := &mockBillingBridge{}
	handler := MustNewHandler("secret-service-token-12345", signer, store, billingBridge, mockAIServer.URL)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	schoolID := "school-1"
	parentID := "parent-proxy-1"
	childID := "child-proxy-101"
	otherChildID := "other-child-999"

	store.Guardians[schoolID+":"+childID+":"+parentID] = true

	// 1. Unauthenticated request (no user in context) -> 401 Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodPost, "/api/v1/chat/message", bytes.NewReader([]byte(`{"child_id": "child-101", "message": "Hi"}`)))
	rrUnauth := httptest.NewRecorder()
	mux.ServeHTTP(rrUnauth, reqUnauth)
	if rrUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for unauthenticated chat, got %d", rrUnauth.Code)
	}

	// 2. Non-guardian trying to query child -> 403 Forbidden
	reqForbidden := httptest.NewRequest(http.MethodPost, "/api/v1/chat/message", bytes.NewReader([]byte(`{"child_id": "`+otherChildID+`", "message": "Hi"}`)))
	reqForbidden = reqForbidden.WithContext(billing.WithIdentity(context.Background(), billing.Identity{
		UserID:   parentID,
		SchoolID: schoolID,
		Role:     "parent",
	}))
	rrForbidden := httptest.NewRecorder()
	mux.ServeHTTP(rrForbidden, reqForbidden)
	if rrForbidden.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for querying non-guardian child, got %d", rrForbidden.Code)
	}

	// 3. Valid parent and linked child -> proxies to AI service with signed X-User-Context
	chatPayload, _ := json.Marshal(map[string]string{
		"child_id": childID,
		"message":  "What is the fee balance?",
	})
	reqValid := httptest.NewRequest(http.MethodPost, "/api/v1/chat/message", bytes.NewReader(chatPayload))
	reqValid = reqValid.WithContext(billing.WithIdentity(context.Background(), billing.Identity{
		UserID:   parentID,
		SchoolID: schoolID,
		Role:     "parent",
	}))
	rrValid := httptest.NewRecorder()
	mux.ServeHTTP(rrValid, reqValid)

	if rrValid.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid chat proxy, got %d: %s", rrValid.Code, rrValid.Body.String())
	}
	if receivedUserContext == "" {
		t.Fatalf("expected X-User-Context header received at AI service")
	}
	// Verify signature of the forwarded X-User-Context
	verifiedCtx, err := signer.Verify(receivedUserContext)
	if err != nil {
		t.Fatalf("AI service received invalid signature on X-User-Context: %v", err)
	}
	if verifiedCtx.UserID != parentID || verifiedCtx.SchoolID != schoolID {
		t.Errorf("mismatch in signed user context: %+v", verifiedCtx)
	}
	if receivedBody["query"] != "What is the fee balance?" {
		t.Errorf("expected message forwarded to AI service as query, got %v", receivedBody)
	}
}
