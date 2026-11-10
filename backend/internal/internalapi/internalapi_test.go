package internalapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type stubAcademics struct {
	attendance internalapi.AttendanceSummary
	marks      internalapi.MarksSummary
}

func (s *stubAcademics) GetStudentAttendanceSummary(_ context.Context, _, studentID string) (internalapi.AttendanceSummary, error) {
	out := s.attendance
	out.StudentID = studentID
	return out, nil
}

func (s *stubAcademics) GetStudentMarksSummary(_ context.Context, _, studentID string) (internalapi.MarksSummary, error) {
	out := s.marks
	out.StudentID = studentID
	return out, nil
}

type stubDaraja struct{}

func (d *stubDaraja) InitiateSTKPush(_ context.Context, req billing.DarajaSTKPushRequest) (billing.DarajaSTKPushResponse, error) {
	return billing.DarajaSTKPushResponse{
		MerchantRequestID: "MR-INT-001",
		CheckoutRequestID: "ws_CO_INT_001",
		ResponseCode:      "0",
		ResponseDesc:      "Success. Request accepted for processing",
		CustomerMessage:   "Ref:" + req.AccountReference,
	}, nil
}

func (d *stubDaraja) QuerySTKStatus(_ context.Context, req billing.DarajaSTKQueryRequest) (billing.DarajaSTKQueryResponse, error) {
	return billing.DarajaSTKQueryResponse{
		CheckoutRequestID: req.CheckoutRequestID,
		ResultCode:        0,
		ResultDesc:        "Processed",
	}, nil
}

func TestUserContextSignerAndVerifier(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	token := "internal-secret-token-dev05"
	validCtx := internalapi.UserContext{
		UserID:       "11111111-1111-4111-8111-111111111111",
		SchoolID:     "22222222-2222-4222-8222-222222222222",
		MembershipID: "33333333-3333-4333-8333-333333333333",
		Role:         "parent",
		Channel:      "whatsapp",
		SessionID:    "44444444-4444-4444-8444-444444444444",
		Exp:          now.Add(5 * time.Minute).Unix(),
	}

	signed, err := internalapi.SignUserContext(validCtx, token)
	if err != nil {
		t.Fatalf("SignUserContext failed: %v", err)
	}
	parts := strings.Split(signed, ".")
	if len(parts) != 2 || len(parts[1]) != 64 {
		t.Fatalf("unexpected signed format: %q", signed)
	}

	parsed, err := internalapi.VerifyUserContext(signed, token, now)
	if err != nil {
		t.Fatalf("VerifyUserContext failed: %v", err)
	}
	if parsed != validCtx {
		t.Fatalf("parsed context mismatch: got %+v, want %+v", parsed, validCtx)
	}

	// Tampered signature must fail
	if _, err := internalapi.VerifyUserContext(parts[0]+".00"+parts[1][2:], token, now); err == nil {
		t.Fatal("expected tampered signature to fail verification")
	}

	// Tampered payload with original signature must fail before trusting JSON
	if _, err := internalapi.VerifyUserContext("e30."+parts[1], token, now); err == nil {
		t.Fatal("expected tampered payload to fail verification")
	}

	// Wrong service token must fail
	if _, err := internalapi.VerifyUserContext(signed, "wrong-secret", now); err == nil {
		t.Fatal("expected wrong service token to fail verification")
	}

	// Expired context must fail
	if _, err := internalapi.VerifyUserContext(signed, token, now.Add(6*time.Minute)); err == nil {
		t.Fatal("expected expired user context to fail verification")
	}

	// Missing session_id on whatsapp channel must fail
	invalidWA := validCtx
	invalidWA.SessionID = ""
	signedInvalidWA, err := internalapi.SignUserContext(invalidWA, token)
	if err != nil {
		t.Fatalf("SignUserContext failed: %v", err)
	}
	if _, err := internalapi.VerifyUserContext(signedInvalidWA, token, now); err == nil {
		t.Fatal("expected whatsapp context without session_id to fail verification")
	}
}

func TestInternalAPIEndpointsAuthOwnershipMarksGuardAndChatProxy(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	if _, err := db.Admin.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_payment_paused_until TIMESTAMPTZ`); err != nil {
		t.Fatalf("alter users: %v", err)
	}
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO school_mpesa_configs(school_id, shortcode, environment, passkey_encrypted, test_payment_verified)
		VALUES ($1, '600001', 'sandbox', 'enc:passkey-school-a', true)`,
		schoolA.ID,
	); err != nil {
		t.Fatalf("seed school_mpesa_configs: %v", err)
	}

	bursarUserID := "91010000-0000-4000-8000-000000000079"
	bursarMembID := "91010000-0000-4000-8000-000000000080"
	parentBUserID := "91010000-0000-4000-8000-000000000081"
	parentBMembID := "91010000-0000-4000-8000-000000000082"
	sessionAID := "91010000-0000-4000-8000-000000000083"
	sessionBID := "91010000-0000-4000-8000-000000000084"
	student2ID := "91010000-0000-4000-8000-000000000085"
	now := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)

	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO users (id, phone_number, full_name, is_active)
		VALUES ($1, '+254799000111', 'Bursar A', TRUE),
		       ($2, '+254799111222', 'Parent B', TRUE)`,
		bursarUserID, parentBUserID,
	); err != nil {
		t.Fatalf("insert users: %v", err)
	}

	err := db.WithTenantTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO memberships (id, school_id, user_id, role, status)
			VALUES ($1, $2, $3, 'bursar', 'active'),
			       ($4, $2, $5, 'parent', 'active')`,
			bursarMembID, schoolA.ID, bursarUserID,
			parentBMembID, parentBUserID,
		); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO students (id, school_id, admission_number, full_name, dob, gender, current_class_id)
			VALUES ($1, $2, 'DEMO-002', 'Second Student', '2013-02-02', 'male', $3)`,
			student2ID, schoolA.ID, schoolA.ClassID,
		); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO student_guardians (school_id, student_id, user_id, relationship)
			VALUES ($1, $2, $3, 'mother')`,
			schoolA.ID, student2ID, parentBUserID,
		); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO chat_sessions (
				id, school_id, channel, sender_wa_number, state, student_id, guardian_user_id, verified_at, expires_at
			) VALUES
				($1, $2, 'whatsapp', '+254711000111', 'VERIFIED', $3, $4, $5, $6),
				($7, $2, 'whatsapp', '+254799111222', 'VERIFIED', $8, $9, $5, $6)`,
			sessionAID, schoolA.ID, schoolA.StudentID, schoolA.UserID, now, now.Add(6*time.Hour),
			sessionBID, student2ID, parentBUserID,
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed internalapi fixtures: %v", err)
	}

	billingSvc := &billing.Service{
		DB:     db,
		Daraja: &stubDaraja{},
		Now:    func() time.Time { return now },
	}
	adminPrincipal := contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "ikonex_admin",
	}
	structure, err := billingSvc.LoadFeeStructure(ctx, adminPrincipal, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		GradeLevelID:   schoolA.GradeID,
		TermID:         schoolA.TermID,
		BoardingStatus: "all",
		Items: []billing.FeeItemInput{
			{Title: "Tuition", Amount: json.RawMessage(`"12000.00"`)},
		},
	})
	if err != nil {
		t.Fatalf("LoadFeeStructure: %v", err)
	}
	if err := db.WithTenantTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		return billingSvc.ChargeStudent(ctx, tx, schoolA.StudentID, structure.FeePeriodID)
	}); err != nil {
		t.Fatalf("ChargeStudent: %v", err)
	}

	academicsStub := &stubAcademics{
		attendance: internalapi.AttendanceSummary{
			TermID:          schoolA.TermID,
			PresentDays:     42,
			AbsentDays:      2,
			LateDays:        1,
			TotalRecorded:   45,
			LastRollCallDay: "2026-05-11",
		},
		marks: internalapi.MarksSummary{
			TermID:   schoolA.TermID,
			Released: true,
			Subjects: []internalapi.SubjectMarkSummary{
				{Subject: "Mathematics", PerformanceLevel: "EE", TeacherComments: "Excellent"},
			},
		},
	}

	var capturedAIToken, capturedAICtxHeader string
	var capturedAIBody map[string]any
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ai/parent-query" {
			http.NotFound(w, r)
			return
		}
		capturedAIToken = r.Header.Get(internalapi.HeaderInternalServiceToken)
		capturedAICtxHeader = r.Header.Get(internalapi.HeaderUserContext)
		_ = json.NewDecoder(r.Body).Decode(&capturedAIBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"reply":"Jambo! Your fee balance is KES 12,000.00."}`))
	}))
	defer aiServer.Close()

	const serviceToken = "dev05-internal-token-secret"
	svc := &internalapi.Service{
		DB:                   db,
		Billing:              billingSvc,
		Academics:            academicsStub,
		InternalServiceToken: serviceToken,
		AIServiceBaseURL:     aiServer.URL + "/api/v1/ai",
		PortalBaseURL:        "https://portal.ikonex.test",
		HTTPClient:           aiServer.Client(),
		Now:                  func() time.Time { return now },
	}

	var activePrincipal contracts.Principal
	mux := http.NewServeMux()
	internalapi.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		if activePrincipal.UserID == "" {
			return contracts.Principal{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED", "missing principal")
		}
		return activePrincipal, nil
	})

	makeSignedHeader := func(channel string) string {
		t.Helper()
		uctx := internalapi.UserContext{
			UserID:       schoolA.UserID,
			SchoolID:     schoolA.ID,
			MembershipID: schoolA.MembershipID,
			Role:         "parent",
			Channel:      channel,
			SessionID:    sessionAID,
			Exp:          now.Add(5 * time.Minute).Unix(),
		}
		signed, err := internalapi.SignUserContext(uctx, serviceToken)
		if err != nil {
			t.Fatalf("SignUserContext: %v", err)
		}
		return signed
	}

	doInternal := func(method, path, token, userCtx string, body []byte) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		if len(body) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set(internalapi.HeaderInternalServiceToken, token)
		}
		if userCtx != "" {
			req.Header.Set(internalapi.HeaderUserContext, userCtx)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	// 1. Missing or invalid X-Internal-Service-Token -> 401
	recNoToken := doInternal(http.MethodGet, "/api/v1/internal/school-profile", "", makeSignedHeader("whatsapp"), nil)
	if recNoToken.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want 401", recNoToken.Code)
	}
	recBadSig := doInternal(http.MethodGet, "/api/v1/internal/school-profile", serviceToken, makeSignedHeader("whatsapp")+"bad", nil)
	if recBadSig.Code != http.StatusUnauthorized {
		t.Fatalf("bad signature status = %d, want 401", recBadSig.Code)
	}

	// 2. GET /api/v1/internal/school-profile -> 200
	recProfile := doInternal(http.MethodGet, "/api/v1/internal/school-profile", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recProfile.Code != http.StatusOK || !strings.Contains(recProfile.Body.String(), schoolA.ID) {
		t.Fatalf("school-profile status = %d, body = %s", recProfile.Code, recProfile.Body.String())
	}

	// 3. GET /api/v1/internal/parents/{user_id}/children -> 200 for self, 403 for another parent
	recChildren := doInternal(http.MethodGet, "/api/v1/internal/parents/"+schoolA.UserID+"/children", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recChildren.Code != http.StatusOK || !strings.Contains(recChildren.Body.String(), schoolA.StudentID) || strings.Contains(recChildren.Body.String(), student2ID) {
		t.Fatalf("parent children status = %d, body = %s", recChildren.Code, recChildren.Body.String())
	}
	recOtherParentChildren := doInternal(http.MethodGet, "/api/v1/internal/parents/"+parentBUserID+"/children", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recOtherParentChildren.Code != http.StatusForbidden {
		t.Fatalf("other parent children status = %d, want 403", recOtherParentChildren.Code)
	}

	// 4. GET /api/v1/internal/students/{id}/fee-balance -> 200 for own child, 403 for Parent B's child, 404 for School B's child
	recFeeOwn := doInternal(http.MethodGet, "/api/v1/internal/students/"+schoolA.StudentID+"/fee-balance", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recFeeOwn.Code != http.StatusOK || !strings.Contains(recFeeOwn.Body.String(), `"current_balance":"12000.00"`) {
		t.Fatalf("own child fee-balance status = %d, body = %s", recFeeOwn.Code, recFeeOwn.Body.String())
	}
	recFeeOtherChild := doInternal(http.MethodGet, "/api/v1/internal/students/"+student2ID+"/fee-balance", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recFeeOtherChild.Code != http.StatusForbidden {
		t.Fatalf("other parent child fee-balance status = %d, want 403", recFeeOtherChild.Code)
	}
	recFeeCrossSchool := doInternal(http.MethodGet, "/api/v1/internal/students/"+schoolB.StudentID+"/fee-balance", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recFeeCrossSchool.Code != http.StatusNotFound {
		t.Fatalf("cross-school child fee-balance status = %d, want 404", recFeeCrossSchool.Code)
	}

	// 5. GET /api/v1/internal/students/{id}/attendance-summary -> 200 for own child, 403 for Parent B's child
	recAttOwn := doInternal(http.MethodGet, "/api/v1/internal/students/"+schoolA.StudentID+"/attendance-summary", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recAttOwn.Code != http.StatusOK || !strings.Contains(recAttOwn.Body.String(), `"present_days":42`) {
		t.Fatalf("attendance-summary status = %d, body = %s", recAttOwn.Code, recAttOwn.Body.String())
	}
	recAttOther := doInternal(http.MethodGet, "/api/v1/internal/students/"+student2ID+"/attendance-summary", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recAttOther.Code != http.StatusForbidden {
		t.Fatalf("other child attendance-summary status = %d, want 403", recAttOther.Code)
	}

	// 6. GET /api/v1/internal/students/{id}/marks-summary -> WhatsApp minimal-facts guard vs in_app full marks
	recMarksWA := doInternal(http.MethodGet, "/api/v1/internal/students/"+schoolA.StudentID+"/marks-summary", serviceToken, makeSignedHeader("whatsapp"), nil)
	if recMarksWA.Code != http.StatusOK {
		t.Fatalf("whatsapp marks-summary status = %d, body = %s", recMarksWA.Code, recMarksWA.Body.String())
	}
	if strings.Contains(recMarksWA.Body.String(), "Mathematics") || !strings.Contains(recMarksWA.Body.String(), "https://portal.ikonex.test/parent/results") {
		t.Fatalf("expected WhatsApp marks guard to suppress subjects and return portal URL, got %s", recMarksWA.Body.String())
	}

	recMarksInApp := doInternal(http.MethodGet, "/api/v1/internal/students/"+schoolA.StudentID+"/marks-summary", serviceToken, makeSignedHeader("in_app"), nil)
	if recMarksInApp.Code != http.StatusOK || !strings.Contains(recMarksInApp.Body.String(), "Mathematics") {
		t.Fatalf("in_app marks-summary status = %d, body = %s", recMarksInApp.Code, recMarksInApp.Body.String())
	}

	// 7. POST /api/v1/internal/payments/stk-push -> requires parent_confirmed=true & enforces 24h phone-change freeze
	unconfirmedBody := []byte(`{"student_id":"` + schoolA.StudentID + `","amount":500000,"phone_number":"+254712345678","parent_confirmed":false}`)
	recUnconfirmed := doInternal(http.MethodPost, "/api/v1/internal/payments/stk-push", serviceToken, makeSignedHeader("whatsapp"), unconfirmedBody)
	if recUnconfirmed.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unconfirmed internal stk-push status = %d, want 422", recUnconfirmed.Code)
	}

	confirmedBody := []byte(`{"student_id":"` + schoolA.StudentID + `","amount":500000,"phone_number":"+254712345678","parent_confirmed":true}`)
	recSTK := doInternal(http.MethodPost, "/api/v1/internal/payments/stk-push", serviceToken, makeSignedHeader("whatsapp"), confirmedBody)
	if recSTK.Code != http.StatusAccepted || !strings.Contains(recSTK.Body.String(), `"account_reference":"DEMO-001"`) {
		t.Fatalf("confirmed internal stk-push status = %d, body = %s", recSTK.Code, recSTK.Body.String())
	}

	// Set chat_payment_paused_until -> internal stk-push must return 403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2`, now.Add(23*time.Hour), schoolA.UserID); err != nil {
		t.Fatalf("update chat_payment_paused_until: %v", err)
	}
	recFrozenSTK := doInternal(http.MethodPost, "/api/v1/internal/payments/stk-push", serviceToken, makeSignedHeader("whatsapp"), confirmedBody)
	if recFrozenSTK.Code != http.StatusForbidden || !strings.Contains(recFrozenSTK.Body.String(), "CHAT_PAYMENT_PAUSED_PHONE_CHANGE") {
		t.Fatalf("expected 403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE, got %d (%s)", recFrozenSTK.Code, recFrozenSTK.Body.String())
	}

	// 8. POST /api/v1/internal/ai/audit-log -> persists and strips secrets/tokens/OTPs from metadata
	auditPayload := []byte(`{
		"session_id": "` + sessionAID + `",
		"feature": "parent_assistant",
		"intent": "check_fee_balance",
		"endpoints_read": ["/api/v1/internal/students/` + schoolA.StudentID + `/fee-balance"],
		"action_taken": "answered_balance",
		"metadata": {
			"latency_ms": 140,
			"internal_service_token": "should-be-stripped",
			"otp_code": "123456",
			"x-user-context": "should-also-be-stripped"
		}
	}`)
	recAudit := doInternal(http.MethodPost, "/api/v1/internal/ai/audit-log", serviceToken, makeSignedHeader("whatsapp"), auditPayload)
	if recAudit.Code != http.StatusCreated {
		t.Fatalf("ai audit-log status = %d, body = %s", recAudit.Code, recAudit.Body.String())
	}
	var storedMeta string
	err = db.WithTenantTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT metadata::text FROM ai_action_logs WHERE school_id = $1 LIMIT 1`, schoolA.ID).Scan(&storedMeta)
	})
	if err != nil {
		t.Fatalf("query ai_action_logs: %v", err)
	}
	if strings.Contains(storedMeta, "should-be-stripped") || strings.Contains(storedMeta, "123456") || !strings.Contains(storedMeta, "latency_ms") {
		t.Fatalf("expected sensitive keys stripped from ai_action_logs metadata, got %s", storedMeta)
	}

	// 9. POST /api/v1/chat/message -> authenticates in-app parent, signs X-User-Context, and forwards to AI service
	activePrincipal = contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "parent",
	}
	chatReq := httptest.NewRequest(http.MethodPost, "/api/v1/chat/message", bytes.NewReader([]byte(`{
		"student_id": "`+schoolA.StudentID+`",
		"message": "How much do I owe for Term 1?"
	}`)))
	chatReq.Header.Set("Content-Type", "application/json")
	chatRec := httptest.NewRecorder()
	mux.ServeHTTP(chatRec, chatReq)
	if chatRec.Code != http.StatusOK || !strings.Contains(chatRec.Body.String(), "KES 12,000.00") {
		t.Fatalf("chat proxy status = %d, body = %s", chatRec.Code, chatRec.Body.String())
	}
	if capturedAIToken != serviceToken {
		t.Fatalf("captured AI service token = %q, want %q", capturedAIToken, serviceToken)
	}
	verifiedForwardedCtx, err := internalapi.VerifyUserContext(capturedAICtxHeader, serviceToken, now)
	if err != nil {
		t.Fatalf("forwarded X-User-Context failed verification: %v", err)
	}
	if verifiedForwardedCtx.Channel != "in_app" || verifiedForwardedCtx.UserID != schoolA.UserID {
		t.Fatalf("unexpected forwarded UserContext: %+v", verifiedForwardedCtx)
	}
	if capturedAIBody["message"] != "How much do I owe for Term 1?" {
		t.Fatalf("unexpected forwarded AI body: %+v", capturedAIBody)
	}
}
