package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

type recordingPollQueue struct {
	mu      sync.Mutex
	tasks   []billing.STKPollTask
	delays  []time.Duration
	failErr error
}

func (q *recordingPollQueue) EnqueueSTKPoll(_ context.Context, task billing.STKPollTask, delay time.Duration) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.failErr != nil {
		return q.failErr
	}
	q.tasks = append(q.tasks, task)
	q.delays = append(q.delays, delay)
	return nil
}

type mpesaFailingAuditLogger struct {
	err error
}

func (f mpesaFailingAuditLogger) Log(_ context.Context, _ pgx.Tx, _ contracts.AuditEvent) error {
	return f.err
}

func assertDarajaSuccessBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON body: %v (%s)", err, rec.Body.String())
	}
	if len(raw) != 2 {
		t.Fatalf("expected exactly 2 flat Daraja fields, got %v (%s)", raw, rec.Body.String())
	}
	if _, wrapped := raw["status"]; wrapped {
		t.Fatalf("response must not use application envelope: %s", rec.Body.String())
	}
	code, ok := raw["ResultCode"].(float64)
	if !ok || int(code) != 0 {
		t.Fatalf("expected ResultCode=0 (number), got %T (%v)", raw["ResultCode"], raw["ResultCode"])
	}
	desc, ok := raw["ResultDesc"].(string)
	if !ok || desc != "Accepted" {
		t.Fatalf("expected ResultDesc=\"Accepted\", got %T (%v)", raw["ResultDesc"], raw["ResultDesc"])
	}
}

func assertDarajaErrorBody(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("expected HTTP %d, got %d: %s", wantStatus, rec.Code, rec.Body.String())
	}
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON body: %v (%s)", err, rec.Body.String())
	}
	if _, wrapped := raw["status"]; wrapped {
		t.Fatalf("error response must not use application envelope: %s", rec.Body.String())
	}
	code, ok := raw["ResultCode"].(float64)
	if !ok || int(code) != 1 {
		t.Fatalf("expected ResultCode=1 (number), got %T (%v)", raw["ResultCode"], raw["ResultCode"])
	}
	if desc, ok := raw["ResultDesc"].(string); !ok || desc == "" {
		t.Fatalf("expected non-empty string ResultDesc, got %T (%v)", raw["ResultDesc"], raw["ResultDesc"])
	}
}

func TestMpesaDarajaSTKCallbackC2BAndUnmatchedResolution(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	// Configure per-school M-Pesa Paybill shortcodes (migration 000005 already supplies users.chat_payment_paused_until).
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO school_mpesa_configs(school_id, shortcode, environment, passkey_encrypted, test_payment_verified)
		VALUES ($1, '600001', 'sandbox', 'enc:passkey-school-a', true),
		       ($2, '600002', 'sandbox', 'enc:passkey-school-b', true)`,
		schoolA.ID, schoolB.ID,
	); err != nil {
		t.Fatal(err)
	}

	// Set up httptest Daraja sandbox server
	var mu sync.Mutex
	var lastSTKBody map[string]any
	var stkCalls int
	var nextQueryResp string
	darajaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/oauth/v1/generate"):
			user, pass, ok := r.BasicAuth()
			if !ok || user != "test-consumer-key" || pass != "test-consumer-secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"sandbox-oauth-token","expires_in":"3599"}`))
		case r.URL.Path == "/mpesa/stkpush/v1/processrequest":
			if r.Header.Get("Authorization") != "Bearer sandbox-oauth-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mu.Lock()
			lastSTKBody = body
			stkCalls++
			seq := stkCalls
			mu.Unlock()
			suffix := fmt.Sprintf("100%d", seq)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"MerchantRequestID":"MR-` + suffix + `","CheckoutRequestID":"ws_CO_` + suffix + `","ResponseCode":"0","ResponseDescription":"Success","CustomerMessage":"Success. Request accepted for processing"}`))
		case r.URL.Path == "/mpesa/stkpushquery/v1/query":
			mu.Lock()
			respBody := nextQueryResp
			mu.Unlock()
			if respBody == "" {
				respBody = `{"MerchantRequestID":"MR-1001","CheckoutRequestID":"ws_CO_1001","ResultCode":"0","ResultDesc":"The service request is processed successfully.","MpesaReceiptNumber":"QWE123STK"}`
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(respBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer darajaServer.Close()

	// 2026-02-10 22:00:00 UTC => 2026-02-11 01:00:00 EAT (Africa/Nairobi)
	fixedNow := time.Date(2026, 2, 10, 22, 0, 0, 0, time.UTC)

	darajaClient := &billing.HTTPDarajaClient{
		BaseURL:        darajaServer.URL,
		ConsumerKey:    "test-consumer-key",
		ConsumerSecret: "test-consumer-secret",
		CallbackURL:    "https://portal.ikonex.ac.ke/api/v1/billing/mpesa/callback",
		DecryptPasskey: func(ciphertext string) (string, error) {
			return strings.TrimPrefix(ciphertext, "enc:"), nil
		},
		HTTPClient: darajaServer.Client(),
		Now:        func() time.Time { return fixedNow },
	}

	pollQueue := &recordingPollQueue{}
	tenantDB := contracts.TenantTxFunc(db.WithTenantTx)
	auditLog := billing.NewPostgresAuditLogger()
	notifSvc := &notifications.Service{DB: tenantDB}

	svc := &billing.Service{
		DB:       tenantDB,
		SystemDB: db.WithTenantTx,
		Audit:    auditLog,
		Notifier: notifSvc,
		Daraja:   darajaClient,
		PollQueue: pollQueue,
		ResolveShortcode: func(ctx context.Context, shortcode string) (string, error) {
			var sid string
			err := db.Admin.QueryRow(ctx, `SELECT school_id FROM school_mpesa_configs WHERE shortcode = $1`, shortcode).Scan(&sid)
			return sid, err
		},
		ResolveCheckoutRequest: func(ctx context.Context, checkoutID string) (string, error) {
			var sid string
			err := db.Admin.QueryRow(ctx, `SELECT school_id FROM mpesa_stk_requests WHERE checkout_request_id = $1`, checkoutID).Scan(&sid)
			return sid, err
		},
		ResolveCallbackToken: func(ctx context.Context, tokenHash string) (string, string, error) {
			var sid, checkoutID string
			err := db.Admin.QueryRow(ctx, `SELECT school_id, COALESCE(checkout_request_id, '') FROM mpesa_stk_requests WHERE callback_token_hash = $1`, tokenHash).Scan(&sid, &checkoutID)
			return sid, checkoutID, err
		},
		CallbackAuthToken:      "trusted-gateway-secret",
		TrustedCallbackProxies: []string{"127.0.0.1/32", "10.0.0.0/8"},
		Now:                    func() time.Time { return fixedNow },
	}

	adminA := seedMemberWithPermission(t, db, schoolA.ID, "+254711000021", "Fee Admin A", "bursar", billing.PermFeeStructuresManage)
	bursarUnmatchedA := seedMemberWithPermission(t, db, schoolA.ID, "+254711000023", "Bursar Unmatched A", "bursar", billing.PermUnmatchedPaymentsManage)
	bursarReadOnlyA := seedMemberWithPermission(t, db, schoolA.ID, "+254711000024", "Bursar ReadOnly A", "bursar", billing.PermLedgersView)
	bursarB := seedMemberWithPermission(t, db, schoolB.ID, "+254711000025", "Bursar B", "bursar", billing.PermUnmatchedPaymentsManage)
	noPermStaff := seedMemberWithPermission(t, db, schoolA.ID, "+254711000022", "Teacher A", "teacher", "")
	guardianA := contracts.Principal{UserID: schoolA.UserID, SchoolID: schoolA.ID, MembershipID: schoolA.MembershipID, Role: "parent"}
	guardianB := contracts.Principal{UserID: schoolB.UserID, SchoolID: schoolB.ID, MembershipID: schoolB.MembershipID, Role: "parent"}

	// Seed a non-fee-paying guardian on schoolA.StudentID
	var nonFeePayerID, nonFeePayerMembID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO users(phone_number,full_name) VALUES('+254799554433','Non Fee Guardian') RETURNING id`).Scan(&nonFeePayerID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'parent','active') RETURNING id`, schoolA.ID, nonFeePayerID).Scan(&nonFeePayerMembID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO student_guardians(school_id,student_id,user_id,relationship,is_fee_payer) VALUES($1,$2,$3,'aunt',false)`, schoolA.ID, schoolA.StudentID, nonFeePayerID); err != nil {
		t.Fatal(err)
	}
	nonFeeGuardian := contracts.Principal{UserID: nonFeePayerID, SchoolID: schoolA.ID, MembershipID: nonFeePayerMembID, Role: "parent"}

	// Load Term 1 fee structure in School A = 10,000.00
	if _, err := svc.LoadFeeStructure(ctx, adminA, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: json.RawMessage(`"10000.00"`)}},
	}); err != nil {
		t.Fatal(err)
	}

	// Build HTTP mux with switchable identity for end-to-end HTTP handler verification.
	var activePrincipal contracts.Principal
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return activePrincipal, nil
	})

	// 1. Cross-school guardian, non-fee-payer guardian, zero/negative/fractional KES rejected
	var apiErr *contracts.APIError
	if _, err := svc.InitiateSTKPush(ctx, guardianB, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
	}, false); !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-school student STK push, got %v", err)
	}
	if _, err := svc.InitiateSTKPush(ctx, nonFeeGuardian, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
	}, false); !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 for non-fee-paying guardian STK push, got %v", err)
	}
	for _, badAmt := range []string{`"0.00"`, `"-500.00"`, `"0.50"`, `"500.50"`} {
		if _, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
			StudentID: schoolA.StudentID,
			Amount:    json.RawMessage(badAmt),
		}, false); !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for STK amount %s, got %v", badAmt, err)
		}
	}

	// 2. Chat payments via HTTP route ALWAYS require explicit parent_confirmed=true (even when omitted!)
	activePrincipal = guardianA
	for _, body := range []string{
		`{"student_id":"` + schoolA.StudentID + `","amount":"500.00","source":"chat"}`,
		`{"student_id":"` + schoolA.StudentID + `","amount":"500.00","source":"chat","parent_confirmed":false}`,
		`{"student_id":"` + schoolA.StudentID + `","amount":"500.00","source":"in_app"}`,
		`{"student_id":"` + schoolA.StudentID + `","amount":"500.00","source":"whatsapp","parent_confirmed":false}`,
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/payments/stk-push", bytes.NewBufferString(body)))
		if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "PARENT_CONFIRMATION_REQUIRED") {
			t.Fatalf("expected 422 PARENT_CONFIRMATION_REQUIRED for %s, got %d: %s", body, rec.Code, rec.Body.String())
		}
	}

	// 3. Phone-change freeze: blocks chat payment during 24h window, allows after expiry, and allows web payment
	pauseUntil := fixedNow.Add(12 * time.Hour)
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2`, pauseUntil, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	recFreeze := httptest.NewRecorder()
	mux.ServeHTTP(recFreeze, httptest.NewRequest(http.MethodPost, "/api/v1/payments/stk-push",
		bytes.NewBufferString(`{"student_id":"`+schoolA.StudentID+`","amount":"500.00","source":"chat","parent_confirmed":true}`)))
	if recFreeze.Code != http.StatusForbidden || !strings.Contains(recFreeze.Body.String(), "CHAT_PAYMENT_PAUSED_PHONE_CHANGE") {
		t.Fatalf("expected 403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE via HTTP route, got %d: %s", recFreeze.Code, recFreeze.Body.String())
	}

	// Expired freeze allows chat payment with explicit parent_confirmed=true and valid registered phone
	expiredPause := fixedNow.Add(-1 * time.Minute)
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2`, expiredPause, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	recChatOK := httptest.NewRecorder()
	chatReq := httptest.NewRequest(http.MethodPost, "/api/v1/payments/stk-push",
		bytes.NewBufferString(`{"student_id":"`+schoolA.StudentID+`","amount":"500.00","source":"chat","parent_confirmed":true}`))
	chatReq.Header.Set("Idempotency-Key", "chat-idem-001")
	mux.ServeHTTP(recChatOK, chatReq)
	if recChatOK.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for confirmed chat payment after freeze expired, got %d: %s", recChatOK.Code, recChatOK.Body.String())
	}

	// Re-apply freeze and verify ordinary web payment (whole KES, overpayment > balance) succeeds + Idempotency works!
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2`, pauseUntil, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	stkRes, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:      schoolA.StudentID,
		Amount:         json.RawMessage(`"12000.00"`),
		Source:         "web",
		IdempotencyKey: "web-idem-12000",
	}, false)
	if err != nil {
		t.Fatalf("expected web STK push to succeed during chat freeze, got %v", err)
	}
	if !stkRes.PollScheduled {
		t.Fatalf("expected PollScheduled=true")
	}
	mu.Lock()
	gotTimestamp := lastSTKBody["Timestamp"]
	gotAccRef := lastSTKBody["AccountReference"]
	gotCBURL, _ := lastSTKBody["CallBackURL"].(string)
	callsBeforeReplay := stkCalls
	mu.Unlock()
	if gotAccRef != "DEMO-001" {
		t.Fatalf("expected AccountReference=DEMO-001, got %v", gotAccRef)
	}
	// 2026-02-10 22:00:00 UTC => 2026-02-11 01:00:00 EAT
	if gotTimestamp != "20260211010000" {
		t.Fatalf("expected Africa/Nairobi Daraja Timestamp 20260211010000, got %v", gotTimestamp)
	}
	if !strings.HasPrefix(gotCBURL, "https://portal.ikonex.ac.ke/api/v1/billing/mpesa/callback/") {
		t.Fatalf("expected per-request callback token path in CallBackURL, got %q", gotCBURL)
	}
	perReqToken := strings.TrimPrefix(gotCBURL, "https://portal.ikonex.ac.ke/api/v1/billing/mpesa/callback/")
	if len(perReqToken) != 64 {
		t.Fatalf("expected 64-char hex callback token, got %q", perReqToken)
	}

	// Idempotency replay: identical request returns same CheckoutRequestID without calling Daraja again
	stkReplay, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:      schoolA.StudentID,
		Amount:         json.RawMessage(`"12000.00"`),
		Source:         "web",
		IdempotencyKey: "web-idem-12000",
	}, false)
	if err != nil || stkReplay.CheckoutRequestID != stkRes.CheckoutRequestID || stkReplay.STKRequestID != stkRes.STKRequestID {
		t.Fatalf("expected identical idempotent replay, got %+v (%v)", stkReplay, err)
	}
	mu.Lock()
	if stkCalls != callsBeforeReplay {
		t.Fatalf("idempotent replay must not call Daraja again: calls=%d want=%d", stkCalls, callsBeforeReplay)
	}
	mu.Unlock()

	// Idempotency conflict: same key with different amount rejected with 409 IDEMPOTENCY_KEY_CONFLICT
	if _, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:      schoolA.StudentID,
		Amount:         json.RawMessage(`"15000.00"`),
		Source:         "web",
		IdempotencyKey: "web-idem-12000",
	}, false); !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict || apiErr.Code != "IDEMPOTENCY_KEY_CONFLICT" {
		t.Fatalf("expected 409 IDEMPOTENCY_KEY_CONFLICT, got %v", err)
	}

	// 4. Genuine Safaricom STK callback fixture: NO BusinessShortCode, NO SchoolID, NO school_id query param!
	// Authenticated via per-request callback path token (/api/v1/billing/mpesa/callback/{token}).
	genuineCallbackJSON := fmt.Sprintf(`{
		"Body": {
			"stkCallback": {
				"MerchantRequestID": %q,
				"CheckoutRequestID": %q,
				"ResultCode": 0,
				"ResultDesc": "The service request is processed successfully.",
				"CallbackMetadata": {
					"Item": [
						{"Name": "Amount", "Value": 12000},
						{"Name": "MpesaReceiptNumber", "Value": "QWE123STK"},
						{"Name": "TransactionDate", "Value": 20260211010100},
						{"Name": "PhoneNumber", "Value": 254700000001}
					]
				}
			}
		}
	}`, stkRes.MerchantRequestID, stkRes.CheckoutRequestID)

	// Wrong callback token rejected with 401 flat Daraja JSON
	recWrongTok := httptest.NewRecorder()
	mux.ServeHTTP(recWrongTok, httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback/0000000000000000000000000000000000000000000000000000000000000000", bytes.NewBufferString(genuineCallbackJSON)))
	assertDarajaErrorBody(t, recWrongTok, http.StatusUnauthorized)

	// Valid per-request callback token succeeds and returns flat {"ResultCode":0,"ResultDesc":"Accepted"}
	recSTK := httptest.NewRecorder()
	mux.ServeHTTP(recSTK, httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback/"+perReqToken, bytes.NewBufferString(genuineCallbackJSON)))
	assertDarajaSuccessBody(t, recSTK)

	// Duplicate STK callback replay returns flat Daraja 200 and does NOT double-credit
	recSTKDup := httptest.NewRecorder()
	mux.ServeHTTP(recSTKDup, httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback/"+perReqToken, bytes.NewBufferString(genuineCallbackJSON)))
	assertDarajaSuccessBody(t, recSTKDup)

	_, billed, paid, current, credit := assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if billed.String() != "10000.00" || paid.String() != "12000.00" || current.String() != "0.00" || credit.String() != "2000.00" {
		t.Fatalf("unexpected balances after STK callback: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// Failed callback arriving AFTER successful settlement must preserve succeeded state!
	failedAfterSuccessJSON := fmt.Sprintf(`{
		"Body": {
			"stkCallback": {
				"MerchantRequestID": %q,
				"CheckoutRequestID": %q,
				"ResultCode": 1032,
				"ResultDesc": "Request cancelled by user"
			}
		}
	}`, stkRes.MerchantRequestID, stkRes.CheckoutRequestID)
	recFailAfterOK := httptest.NewRecorder()
	mux.ServeHTTP(recFailAfterOK, httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback/"+perReqToken, bytes.NewBufferString(failedAfterSuccessJSON)))
	assertDarajaSuccessBody(t, recFailAfterOK)
	var statusAfterFail string
	if err := db.Admin.QueryRow(ctx, `SELECT status FROM mpesa_stk_requests WHERE id = $1`, stkRes.STKRequestID).Scan(&statusAfterFail); err != nil || statusAfterFail != "succeeded" {
		t.Fatalf("expected status to remain succeeded after late failure callback, got %q (%v)", statusAfterFail, err)
	}

	// 5. Earlier failed attempt receiving a valid provider-confirmed success callback MUST settle!
	stkResRetry, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"1000.00"`),
		Source:    "web",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	// First send a failure callback via trusted proxy on static /api/v1/billing/mpesa/callback (NO shortcode, NO school_id!)
	failReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(fmt.Sprintf(`{
		"Body": {"stkCallback": {"MerchantRequestID": %q, "CheckoutRequestID": %q, "ResultCode": 1037, "ResultDesc": "DS timeout"}}
	}`, stkResRetry.MerchantRequestID, stkResRetry.CheckoutRequestID)))
	failReq.RemoteAddr = "10.0.0.5:44300"
	failReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recFail := httptest.NewRecorder()
	mux.ServeHTTP(recFail, failReq)
	assertDarajaSuccessBody(t, recFail)

	// Now send valid success callback for that same CheckoutRequestID -> supersedes timeout/failed and settles!
	okAfterFailReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(fmt.Sprintf(`{
		"Body": {
			"stkCallback": {
				"MerchantRequestID": %q,
				"CheckoutRequestID": %q,
				"ResultCode": 0,
				"ResultDesc": "Success",
				"CallbackMetadata": {
					"Item": [
						{"Name": "Amount", "Value": 1000},
						{"Name": "MpesaReceiptNumber", "Value": "QWE1000LATE"},
						{"Name": "PhoneNumber", "Value": 254700000001}
					]
				}
			}
		}
	}`, stkResRetry.MerchantRequestID, stkResRetry.CheckoutRequestID)))
	okAfterFailReq.RemoteAddr = "10.0.0.5:44301"
	okAfterFailReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recOKLate := httptest.NewRecorder()
	mux.ServeHTTP(recOKLate, okAfterFailReq)
	assertDarajaSuccessBody(t, recOKLate)

	_, _, paid, _, credit = assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if paid.String() != "13000.00" || credit.String() != "3000.00" {
		t.Fatalf("expected paid=13000.00 credit=3000.00 after late success settlement, got paid=%s credit=%s", paid, credit)
	}

	// 6. Unknown CheckoutRequestID & invalid shortcode association rejected
	unknownReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(`{
		"Body": {"stkCallback": {"MerchantRequestID": "MR-NONE", "CheckoutRequestID": "ws_CO_UNKNOWN", "ResultCode": 0}}
	}`))
	unknownReq.RemoteAddr = "10.0.0.5:44302"
	unknownReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recUnknown := httptest.NewRecorder()
	mux.ServeHTTP(recUnknown, unknownReq)
	assertDarajaErrorBody(t, recUnknown, http.StatusNotFound)

	// 7. STK Callback amount integrity: missing, invalid, negative, zero, mismatched amounts rejected
	stkResAmt, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
		Source:    "web",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	badCallbackPayloads := []string{
		// Mismatched amount
		fmt.Sprintf(`{"Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":99999},{"Name":"MpesaReceiptNumber","Value":"BADAMT1"}]}}}}`, stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID),
		// Missing amount
		fmt.Sprintf(`{"Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"MpesaReceiptNumber","Value":"BADAMT2"}]}}}}`, stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID),
		// Invalid non-numeric amount
		fmt.Sprintf(`{"Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":"not-a-number"},{"Name":"MpesaReceiptNumber","Value":"BADAMT3"}]}}}}`, stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID),
		// Zero amount
		fmt.Sprintf(`{"Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":0},{"Name":"MpesaReceiptNumber","Value":"BADAMT4"}]}}}}`, stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID),
		// Negative amount
		fmt.Sprintf(`{"Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":-500},{"Name":"MpesaReceiptNumber","Value":"BADAMT5"}]}}}}`, stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID),
	}
	for i, badBody := range badCallbackPayloads {
		badReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(badBody))
		badReq.RemoteAddr = "10.0.0.5:44310"
		badReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
		recBad := httptest.NewRecorder()
		mux.ServeHTTP(recBad, badReq)
		assertDarajaErrorBody(t, recBad, http.StatusUnprocessableEntity)
		var st string
		if err := db.Admin.QueryRow(ctx, `SELECT status FROM mpesa_stk_requests WHERE id = $1`, stkResAmt.STKRequestID).Scan(&st); err != nil || st != "pending" {
			t.Fatalf("case %d: invalid callback must not mutate status, got %q (%v)", i, st, err)
		}
	}
	// Wrong BusinessShortCode explicitly supplied in STK callback rejected
	wrongSCReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(fmt.Sprintf(
		`{"BusinessShortCode":"600002","Body":{"stkCallback":{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":500},{"Name":"MpesaReceiptNumber","Value":"WRONGSC1"}]}}}}`,
		stkResAmt.MerchantRequestID, stkResAmt.CheckoutRequestID,
	)))
	wrongSCReq.RemoteAddr = "10.0.0.5:44311"
	wrongSCReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recWrongSC := httptest.NewRecorder()
	mux.ServeHTTP(recWrongSC, wrongSCReq)
	assertDarajaErrorBody(t, recWrongSC, http.StatusBadRequest)

	// 8. Direct Paybill C2B auto-matching via HTTP endpoint returns flat Daraja JSON, strictly school-scoped
	c2bReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/c2b-confirmation", bytes.NewBufferString(`{
		"TransactionType": "Pay Bill",
		"TransID": "C2B999MATCH",
		"TransAmount": "1500.00",
		"BusinessShortCode": "600001",
		"BillRefNumber": "  demo-001 ",
		"MSISDN": "254700000001",
		"FirstName": "Jane",
		"LastName": "Guardian"
	}`))
	c2bReq.RemoteAddr = "127.0.0.1:50001"
	c2bReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recC2B := httptest.NewRecorder()
	mux.ServeHTTP(recC2B, c2bReq)
	assertDarajaSuccessBody(t, recC2B)

	// Duplicate C2B receipt returns flat 200 and does NOT double-credit
	c2bDupReq := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/c2b-confirmation", bytes.NewBufferString(`{
		"TransactionType": "Pay Bill",
		"TransID": "C2B999MATCH",
		"TransAmount": "1500.00",
		"BusinessShortCode": "600001",
		"BillRefNumber": "demo-001",
		"MSISDN": "254700000001"
	}`))
	c2bDupReq.RemoteAddr = "127.0.0.1:50002"
	c2bDupReq.Header.Set("X-Callback-Token", "trusted-gateway-secret")
	recC2BDup := httptest.NewRecorder()
	mux.ServeHTTP(recC2BDup, c2bDupReq)
	assertDarajaSuccessBody(t, recC2BDup)

	_, _, paid, _, credit = assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if paid.String() != "14500.00" || credit.String() != "4500.00" {
		t.Fatalf("expected paid=14500.00 credit=4500.00 after C2B match, got paid=%s credit=%s", paid, credit)
	}

	// Cross-school admission collision: School B student also has DEMO-001 and must NOT be touched
	_, _, paidB, _, _ := assertStudentReconciliation(t, ctx, db, schoolB.ID, schoolB.StudentID)
	if paidB.String() != "0.00" {
		t.Fatalf("cross-school C2B leak: School B student paid=%s", paidB)
	}

	// Unsupported Till (Buy Goods) C2B transaction type explicitly rejected
	if _, err := svc.ProcessC2BConfirmation(ctx, billing.DarajaC2BConfirmation{
		TransactionType:   "Buy Goods",
		TransID:           "TILL999",
		TransAmount:       json.RawMessage(`"500.00"`),
		BusinessShortCode: "600001",
	}); !errors.As(err, &apiErr) || apiErr.Code != "UNSUPPORTED_MPESA_SHORTCODE_TYPE" {
		t.Fatalf("expected UNSUPPORTED_MPESA_SHORTCODE_TYPE for Till C2B, got %v", err)
	}

	// 9. C2B unknown admission number -> unmatched queue, bursar permissions & cross-school isolation
	unmatchedRes, err := svc.ProcessC2BConfirmation(ctx, billing.DarajaC2BConfirmation{
		TransID:           "C2B888UNMATCHED",
		TransAmount:       json.RawMessage(`"4000.00"`),
		BusinessShortCode: "600001",
		BillRefNumber:     "UNKNOWN-ADM-99",
		MSISDN:            "254712345678",
		FirstName:         "John",
		LastName:          "Payer",
	})
	if err != nil || unmatchedRes.MatchedStatus != "unmatched" {
		t.Fatalf("expected C2B unmatched status, got %+v (%v)", unmatchedRes, err)
	}

	// Bursar with billing.unmatched_payments.manage in School A can list unmatched payments
	page := contracts.Page{Page: 1, PerPage: 25}
	unmatchedList, err := svc.ListUnmatchedPayments(ctx, bursarUnmatchedA, &page)
	if err != nil || len(unmatchedList) != 1 || unmatchedList[0].TransactionRef != "C2B888UNMATCHED" {
		t.Fatalf("expected bursar with billing.unmatched_payments.manage to list 1 unmatched payment, got %+v (%v)", unmatchedList, err)
	}
	// Bursar in School B sees 0 unmatched payments from School A (strict tenant isolation)
	pageB := contracts.Page{Page: 1, PerPage: 25}
	unmatchedB, err := svc.ListUnmatchedPayments(ctx, bursarB, &pageB)
	if err != nil || len(unmatchedB) != 0 {
		t.Fatalf("cross-school unmatched leak: School B bursar saw %+v (%v)", unmatchedB, err)
	}
	// School A bursar cannot match payment to School B student
	if _, err := svc.MatchUnmatchedPayment(ctx, bursarUnmatchedA, unmatchedList[0].PaymentID, billing.MatchPaymentInput{
		StudentID:      schoolB.StudentID,
		ResolutionNote: "Cross-school attempt",
	}); !errors.As(err, &apiErr) || apiErr.Code != "STUDENT_NOT_FOUND" {
		t.Fatalf("expected STUDENT_NOT_FOUND when matching to another school's student, got %v", err)
	}
	// Teacher without badge and read-only bursar (billing.ledgers.view only) cannot manually match
	for _, unauthorized := range []contracts.Principal{noPermStaff, bursarReadOnlyA} {
		if _, err := svc.MatchUnmatchedPayment(ctx, unauthorized, unmatchedList[0].PaymentID, billing.MatchPaymentInput{
			StudentID:      schoolA.StudentID,
			ResolutionNote: "Unauthorized attempt",
		}); !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
			t.Fatalf("expected 403 for unauthorized staff %s, got %v", unauthorized.UserID, err)
		}
	}

	// Concurrent manual match attempts by authorized bursar: exactly one succeeds, all others get 409!
	activePrincipal = bursarUnmatchedA
	var wg sync.WaitGroup
	var okCount, conflictCount int32
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/payments/"+unmatchedList[0].PaymentID+"/match",
				bytes.NewBufferString(`{"student_id":"`+schoolA.StudentID+`","resolution_note":"Confirmed via M-Pesa statement"}`)))
			switch rec.Code {
			case http.StatusOK:
				atomic.AddInt32(&okCount, 1)
			case http.StatusConflict:
				atomic.AddInt32(&conflictCount, 1)
			default:
				t.Errorf("unexpected status %d: %s", rec.Code, rec.Body.String())
			}
		}()
	}
	wg.Wait()
	if okCount != 1 || conflictCount != 4 {
		t.Fatalf("expected 1 OK and 4 Conflict in concurrent manual match, got ok=%d conflict=%d", okCount, conflictCount)
	}

	_, _, paid, _, credit = assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if paid.String() != "18500.00" || credit.String() != "8500.00" {
		t.Fatalf("expected paid=18500.00 credit=8500.00 after manual match, got paid=%s credit=%s", paid, credit)
	}

	// Verify non-fee-paying guardian received zero payment notifications
	var nonFeeNotices int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2`, schoolA.ID, nonFeePayerID).Scan(&nonFeeNotices); err != nil || nonFeeNotices != 0 {
		t.Fatalf("expected 0 payment notifications for non-fee-paying guardian, got %d (%v)", nonFeeNotices, err)
	}

	// 10. Chat payment phone restrictions: override & invalid/missing registered phone rejected
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = NULL WHERE id = $1`, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	tr := true
	if _, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:       schoolA.StudentID,
		Amount:          json.RawMessage(`"500.00"`),
		PhoneNumber:     "0799999999",
		Source:          "chat",
		ParentConfirmed: &tr,
	}, true); !errors.As(err, &apiErr) || apiErr.Code != "CHAT_PHONE_OVERRIDE_FORBIDDEN" {
		t.Fatalf("expected CHAT_PHONE_OVERRIDE_FORBIDDEN, got %v", err)
	}

	// Corrupt registered phone -> must reject with INVALID_REGISTERED_PHONE (never bypass!)
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET phone_number = 'not-a-phone' WHERE id = $1`, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:       schoolA.StudentID,
		Amount:          json.RawMessage(`"500.00"`),
		Source:          "chat",
		ParentConfirmed: &tr,
	}, true); !errors.As(err, &apiErr) || apiErr.Code != "INVALID_REGISTERED_PHONE" {
		t.Fatalf("expected INVALID_REGISTERED_PHONE for corrupt registered phone, got %v", err)
	}
	// Restore valid registered phone
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET phone_number = '+254700000001' WHERE id = $1`, schoolA.UserID); err != nil {
		t.Fatal(err)
	}

	// 11. Polling queue failure is explicitly handled & recoverable via RetryUnscheduledSTKPolls
	pollQueue.mu.Lock()
	pollQueue.failErr = errors.New("redis unavailable")
	pollQueue.mu.Unlock()
	stkStranded, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
		Source:    "web",
	}, false)
	if err != nil {
		t.Fatalf("STK push should still return attempt when Daraja accepted but poll enqueue failed: %v", err)
	}
	if stkStranded.PollScheduled {
		t.Fatalf("expected PollScheduled=false when queue failed")
	}
	pollQueue.mu.Lock()
	pollQueue.failErr = nil
	pollQueue.mu.Unlock()
	recovered, err := svc.RetryUnscheduledSTKPolls(ctx, schoolA.ID)
	if err != nil || recovered != 1 {
		t.Fatalf("expected RetryUnscheduledSTKPolls to recover 1 stranded request, got recovered=%d err=%v", recovered, err)
	}

	// Poll result cannot overwrite already-succeeded STK request
	mu.Lock()
	nextQueryResp = fmt.Sprintf(`{"MerchantRequestID":%q,"CheckoutRequestID":%q,"ResultCode":"1032","ResultDesc":"Cancelled"}`, stkRes.MerchantRequestID, stkRes.CheckoutRequestID)
	mu.Unlock()
	if err := svc.PollSTKRequest(ctx, schoolA.ID, stkRes.STKRequestID); err != nil {
		t.Fatalf("PollSTKRequest on succeeded row failed: %v", err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT status FROM mpesa_stk_requests WHERE id = $1`, stkRes.STKRequestID).Scan(&statusAfterFail); err != nil || statusAfterFail != "succeeded" {
		t.Fatalf("poll must not overwrite succeeded status, got %q (%v)", statusAfterFail, err)
	}

	// 12. Mandatory dependencies: nil PollQueue, nil Notifier, and failing AuditLogger fail closed!
	svcNoQueue := *svc
	svcNoQueue.PollQueue = nil
	if _, err := svcNoQueue.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
	}, false); !errors.As(err, &apiErr) || apiErr.Code != "MPESA_POLL_QUEUE_NOT_CONFIGURED" {
		t.Fatalf("expected MPESA_POLL_QUEUE_NOT_CONFIGURED when PollQueue is nil, got %v", err)
	}

	svcNoNotifier := *svc
	svcNoNotifier.Notifier = nil
	if _, err := svcNoNotifier.ProcessC2BConfirmation(ctx, billing.DarajaC2BConfirmation{
		TransID:           "C2B_NO_NOTIF",
		TransAmount:       json.RawMessage(`"500.00"`),
		BusinessShortCode: "600001",
		BillRefNumber:     "DEMO-001",
	}); !errors.As(err, &apiErr) || apiErr.Code != "NOTIFIER_REQUIRED" {
		t.Fatalf("expected NOTIFIER_REQUIRED when Notifier is nil, got %v", err)
	}

	svcBadAudit := *svc
	svcBadAudit.Audit = mpesaFailingAuditLogger{err: errors.New("audit storage down")}
	if _, err := svcBadAudit.ProcessC2BConfirmation(ctx, billing.DarajaC2BConfirmation{
		TransID:           "C2B_AUDIT_FAIL",
		TransAmount:       json.RawMessage(`"500.00"`),
		BusinessShortCode: "600001",
		BillRefNumber:     "DEMO-001",
	}); err == nil {
		t.Fatalf("expected C2B confirmation to fail when mandatory audit logging fails")
	}
	// Verify transaction rolled back: payment must NOT exist!
	var leaked int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM payments WHERE transaction_ref = 'C2B_AUDIT_FAIL'`).Scan(&leaked); err != nil || leaked != 0 {
		t.Fatalf("audit failure must roll back payment mutation, found %d rows (%v)", leaked, err)
	}
}

func TestChatPaymentFreezeFailsClosedWhenColumnMissing(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "09")
	ctx := context.Background()

	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO school_mpesa_configs(school_id, shortcode, environment, passkey_encrypted, test_payment_verified)
		VALUES ($1, '600009', 'sandbox', 'enc:passkey', true)`,
		school.ID,
	); err != nil {
		t.Fatal(err)
	}
	// Simulate an unmigrated database missing users.chat_payment_paused_until
	if _, err := db.Admin.Exec(ctx, `ALTER TABLE users DROP COLUMN chat_payment_paused_until`); err != nil {
		t.Fatal(err)
	}

	tenantDB := contracts.TenantTxFunc(db.WithTenantTx)
	svc := &billing.Service{
		DB:        tenantDB,
		SystemDB:  db.WithTenantTx,
		Audit:     billing.NewPostgresAuditLogger(),
		Notifier:  &notifications.Service{DB: tenantDB},
		Daraja:    &billing.HTTPDarajaClient{},
		PollQueue: &recordingPollQueue{},
	}
	guardian := contracts.Principal{UserID: school.UserID, SchoolID: school.ID, MembershipID: school.MembershipID, Role: "parent"}
	tr := true
	_, err := svc.InitiateSTKPush(ctx, guardian, billing.InitiateSTKPushInput{
		StudentID:       school.StudentID,
		Amount:          json.RawMessage(`"500.00"`),
		Source:          "chat",
		ParentConfirmed: &tr,
	}, true)
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusServiceUnavailable || apiErr.Code != "CHAT_PAYMENT_FREEZE_UNAVAILABLE" {
		t.Fatalf("expected 503 CHAT_PAYMENT_FREEZE_UNAVAILABLE when freeze column is missing, got %v", err)
	}
}
