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
	signer := NewContextSigner("test-secret-key-12345")
	store := NewMemoryStore()
	billingBridge := &mockBillingBridge{}
	handler := NewHandler("secret-service-token", signer, store, billingBridge, "http://localhost:8001")
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

	if len(store.AuditLogs) != 1 {
		t.Fatalf("expected 1 audit log saved, got %d", len(store.AuditLogs))
	}
	saved := store.AuditLogs[0]
	if saved.Feature != "parent_assistant" || saved.Channel != "whatsapp" || saved.SessionID != "session-abc-123" {
		t.Errorf("unexpected audit log values: %+v", saved)
	}
}
