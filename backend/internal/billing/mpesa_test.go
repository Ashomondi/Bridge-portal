package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
)

type recordingPollQueue struct {
	mu     sync.Mutex
	tasks  []billing.STKPollTask
	delays []time.Duration
}

func (q *recordingPollQueue) EnqueueSTKPoll(_ context.Context, task billing.STKPollTask, delay time.Duration) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, task)
	q.delays = append(q.delays, delay)
	return nil
}

func TestMpesaDarajaSTKCallbackC2BAndUnmatchedResolution(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	// Add optional chat_payment_paused_until column on users in the disposable test DB to verify AC-26 phone-change freeze
	if _, err := db.Admin.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_payment_paused_until TIMESTAMPTZ`); err != nil {
		t.Fatal(err)
	}

	// Configure per-school M-Pesa shortcodes
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO school_mpesa_configs(school_id, shortcode, environment, passkey_encrypted, test_payment_verified)
		VALUES ($1, '600001', 'sandbox', 'enc:passkey-school-a', true),
		       ($2, '600002', 'sandbox', 'enc:passkey-school-b', true)`,
		schoolA.ID, schoolB.ID,
	); err != nil {
		t.Fatal(err)
	}

	// Set up httptest Daraja sandbox server
	var lastSTKBody map[string]any
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
			lastSTKBody = body
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"MerchantRequestID":"MR-1001","CheckoutRequestID":"ws_CO_1001","ResponseCode":"0","ResponseDescription":"Success","CustomerMessage":"Success. Request accepted for processing"}`))
		case r.URL.Path == "/mpesa/stkpushquery/v1/query":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"MerchantRequestID":"MR-1001","CheckoutRequestID":"ws_CO_1001","ResultCode":"0","ResultDesc":"The service request is processed successfully.","MpesaReceiptNumber":"QWE123STK"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer darajaServer.Close()

	darajaClient := &billing.HTTPDarajaClient{
		BaseURL:        darajaServer.URL,
		ConsumerKey:    "test-consumer-key",
		ConsumerSecret: "test-consumer-secret",
		CallbackURL:    "https://portal.ikonex.ac.ke/api/v1/billing/mpesa/callback",
		DecryptPasskey: func(ciphertext string) (string, error) {
			return strings.TrimPrefix(ciphertext, "enc:"), nil
		},
		HTTPClient: darajaServer.Client(),
	}

	pollQueue := &recordingPollQueue{}
	auditLog := &recordingAuditLogger{}
	notifSvc := &notifications.Service{DB: db}
	fixedNow := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)

	svc := &billing.Service{
		DB:        db,
		Audit:     auditLog,
		Notifier:  notifSvc,
		Daraja:    darajaClient,
		PollQueue: pollQueue,
		ResolveShortcode: func(_ context.Context, shortcode string) (string, error) {
			switch shortcode {
			case "600001":
				return schoolA.ID, nil
			case "600002":
				return schoolB.ID, nil
			default:
				return "", errors.New("unknown shortcode")
			}
		},
		Now: func() time.Time { return fixedNow },
	}

	adminA := contracts.Principal{UserID: schoolA.UserID, SchoolID: schoolA.ID, MembershipID: schoolA.MembershipID, Role: "ikonex_admin"}
	guardianA := contracts.Principal{UserID: schoolA.UserID, SchoolID: schoolA.ID, MembershipID: schoolA.MembershipID, Role: "parent"}
	guardianB := contracts.Principal{UserID: schoolB.UserID, SchoolID: schoolB.ID, MembershipID: schoolB.MembershipID, Role: "parent"}

	// Load Term 1 fee structure in School A = 10,000.00
	if _, err := svc.LoadFeeStructure(ctx, adminA, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: json.RawMessage(`"10000.00"`)}},
	}); err != nil {
		t.Fatal(err)
	}

	// 1. Unlinked/Cross-school guardian rejected
	_, err := svc.InitiateSTKPush(ctx, guardianB, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"500.00"`),
	}, false)
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-school student STK push, got %v", err)
	}

	// 2. Internal AI payment requires parent_confirmed = true (422 if false or missing)
	f := false
	_, err = svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:       schoolA.StudentID,
		Amount:          json.RawMessage(`"500.00"`),
		Source:          "chat",
		ParentConfirmed: &f,
	}, true)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnprocessableEntity || apiErr.Code != "PARENT_CONFIRMATION_REQUIRED" {
		t.Fatalf("expected 422 PARENT_CONFIRMATION_REQUIRED, got %v", err)
	}

	// 3. Phone-change security freeze: blocks chat payment (403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE) but ALLOWS web payment
	pauseUntil := fixedNow.Add(12 * time.Hour)
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2`, pauseUntil, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	tr := true
	_, err = svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID:       schoolA.StudentID,
		Amount:          json.RawMessage(`"500.00"`),
		Source:          "chat",
		ParentConfirmed: &tr,
	}, true)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden || apiErr.Code != "CHAT_PAYMENT_PAUSED_PHONE_CHANGE" {
		t.Fatalf("expected 403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE on chat payment during freeze, got %v", err)
	}

	// Ordinary web payment succeeds even while chat_payment_paused_until > NOW(), and any positive amount (e.g. 0.50 or 12,000.00 > balance) is accepted
	stkRes, err := svc.InitiateSTKPush(ctx, guardianA, billing.InitiateSTKPushInput{
		StudentID: schoolA.StudentID,
		Amount:    json.RawMessage(`"12000.00"`),
		Source:    "web",
	}, false)
	if err != nil {
		t.Fatalf("expected web STK push to succeed during chat freeze, got %v", err)
	}
	if stkRes.AccountReference != "DEMO-001" || lastSTKBody["AccountReference"] != "DEMO-001" {
		t.Fatalf("expected AccountReference=DEMO-001, got %s (%v)", stkRes.AccountReference, lastSTKBody["AccountReference"])
	}
	if len(pollQueue.tasks) != 1 || pollQueue.delays[0] != 45*time.Second {
		t.Fatalf("expected 45s STK poll task enqueued, got %+v (%v)", pollQueue.tasks, pollQueue.delays)
	}

	// 4. Process STK Callback (12,000.00 against 10,000.00 balance -> current=0.00, credit=2,000.00) and verify duplicate callback idempotency
	zero := 0
	var cbPayload billing.DarajaSTKCallbackEnvelope
	cbPayload.BusinessShortCode = "600001"
	cbPayload.Body.StkCallback.MerchantRequestID = stkRes.MerchantRequestID
	cbPayload.Body.StkCallback.CheckoutRequestID = stkRes.CheckoutRequestID
	cbPayload.Body.StkCallback.ResultCode = &zero
	cbPayload.Body.StkCallback.CallbackMetadata = &struct {
		Item []struct {
			Name  string          `json:"Name"`
			Value json.RawMessage `json:"Value"`
		} `json:"Item"`
	}{
		Item: []struct {
			Name  string          `json:"Name"`
			Value json.RawMessage `json:"Value"`
		}{
			{Name: "Amount", Value: json.RawMessage(`12000.00`)},
			{Name: "MpesaReceiptNumber", Value: json.RawMessage(`"QWE123STK"`)},
			{Name: "PhoneNumber", Value: json.RawMessage(`254700000001`)},
		},
	}

	if err := svc.ProcessSTKCallback(ctx, "", cbPayload); err != nil {
		t.Fatalf("ProcessSTKCallback failed: %v", err)
	}
	// Duplicate callback replay must be a no-op
	if err := svc.ProcessSTKCallback(ctx, "", cbPayload); err != nil {
		t.Fatalf("ProcessSTKCallback replay failed: %v", err)
	}

	_, billed, paid, current, credit := assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if billed.String() != "10000.00" || paid.String() != "12000.00" || current.String() != "0.00" || credit.String() != "2000.00" {
		t.Fatalf("unexpected balances after STK callback: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// 5. Direct Paybill C2B auto-matching (school-scoped case/whitespace-insensitive admission number)
	c2bRes, err := svc.ProcessC2BConfirmation(ctx, "", billing.DarajaC2BConfirmation{
		TransID:           "C2B999MATCH",
		TransAmount:       json.RawMessage(`"1500.00"`),
		BusinessShortCode: "600001",
		BillRefNumber:     "  demo-001 ",
		MSISDN:            "254700000001",
		FirstName:         "Jane",
		LastName:          "Guardian",
	})
	if err != nil || c2bRes.MatchedStatus != "matched" {
		t.Fatalf("expected C2B auto-match, got %+v (%v)", c2bRes, err)
	}
	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if paid.String() != "13500.00" || credit.String() != "3500.00" {
		t.Fatalf("expected paid=13500.00 credit=3500.00 after C2B match, got paid=%s credit=%s", paid, credit)
	}

	// Verify School B's student was NOT touched even though School B also has admission number 'DEMO-001'
	_, _, paidB, _, _ := assertStudentReconciliation(t, ctx, db, schoolB.ID, schoolB.StudentID)
	if paidB.String() != "0.00" {
		t.Fatalf("cross-school C2B leak: School B student paid=%s", paidB)
	}

	// 6. C2B with unknown admission number creates unmatched_payments record and never discards money
	unmatchedRes, err := svc.ProcessC2BConfirmation(ctx, "", billing.DarajaC2BConfirmation{
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

	page := contracts.Page{Page: 1, PerPage: 25}
	unmatchedList, err := svc.ListUnmatchedPayments(ctx, adminA, &page)
	if err != nil || len(unmatchedList) != 1 || unmatchedList[0].TransactionRef != "C2B888UNMATCHED" {
		t.Fatalf("expected 1 unmatched payment, got %+v (%v)", unmatchedList, err)
	}

	// 7. Manual resolution via MatchUnmatchedPayment (and HTTP endpoints check)
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return adminA, nil
	})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/payments/"+unmatchedList[0].PaymentID+"/match",
		bytes.NewBufferString(`{"student_id":"`+schoolA.StudentID+`","resolution_note":"Confirmed via M-Pesa statement"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST match failed with %d: %s", rec.Code, rec.Body.String())
	}

	// Duplicate resolution rejected with 409 PAYMENT_ALREADY_RESOLVED
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/payments/"+unmatchedList[0].PaymentID+"/match",
		bytes.NewBufferString(`{"student_id":"`+schoolA.StudentID+`","resolution_note":"Duplicate"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate match resolution, got %d: %s", rec.Code, rec.Body.String())
	}

	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, schoolA.ID, schoolA.StudentID)
	if billed.String() != "10000.00" || paid.String() != "17500.00" || current.String() != "0.00" || credit.String() != "7500.00" {
		t.Fatalf("expected paid=17500.00 credit=7500.00 after manual match, got billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}
}
