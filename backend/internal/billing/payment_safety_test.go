package billing

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type billingTransport func(*http.Request) (*http.Response, error)

func (fn billingTransport) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func TestDarajaPreservesFractionalAmount(t *testing.T) {
	client := NewHTTPDarajaClient(&http.Client{Transport: billingTransport(func(r *http.Request) (*http.Response, error) {
		body := `{"access_token":"synthetic-token"}`
		if r.Method == http.MethodPost {
			var payload map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if string(payload["Amount"]) != "1250.25" {
				t.Fatalf("amount altered: %s", payload["Amount"])
			}
			if r.Header.Get("Authorization") != "Bearer synthetic-token" {
				t.Fatal("missing OAuth token")
			}
			body = `{"CheckoutRequestID":"synthetic-checkout","ResponseCode":"0"}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})})
	_, err := client.InitiateSTKPush(context.Background(), &SchoolMpesaConfig{ConsumerKeyEncrypted: "synthetic-key", ConsumerSecretEncrypted: "synthetic-secret"}, STKInitiateParams{Amount: 1250.25})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPaymentAmountValidation(t *testing.T) {
	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1), 0.001, 1e10} {
		if _, err := paymentAmountCents(amount); err == nil {
			t.Fatalf("accepted invalid amount %v", amount)
		}
	}
	for _, amount := range []float64{0.01, 0.1, 1250.25} {
		if _, err := paymentAmountCents(amount); err != nil {
			t.Fatal(err)
		}
	}
}

type pausedLookupStore struct {
	Store
	lookupError error
	paused      *time.Time
}

func (s pausedLookupStore) GetUserChatPaymentPausedUntil(context.Context, string) (*time.Time, error) {
	return s.paused, s.lookupError
}

func TestChatPaymentPauseFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name        string
		lookupError error
		locked      bool
		wantError   bool
	}{
		{"unlocked", nil, false, false}, {"locked", nil, true, true}, {"lookup error", errors.New("lookup unavailable"), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, store, _, _ := setupTestBillingService()
			store.Students["school:student"] = struct{ ID, Admission, Name, GradeID, YearID string }{ID: "student", Admission: "SYNTHETIC"}
			store.MpesaConfigs["school"] = &SchoolMpesaConfig{SchoolID: "school", Shortcode: "123456"}
			var pause *time.Time
			if tc.locked {
				future := time.Now().Add(time.Hour)
				pause = &future
			}
			service.store = pausedLookupStore{Store: store, lookupError: tc.lookupError, paused: pause}
			calls := 0
			service.SetDarajaClient(&mockDarajaClient{initiateFunc: func(context.Context, *SchoolMpesaConfig, STKInitiateParams) (*DarajaSTKResponse, error) {
				calls++
				return &DarajaSTKResponse{CheckoutRequestID: "synthetic-checkout", ResponseCode: "0"}, nil
			}})
			ctx := WithIdentity(context.Background(), Identity{UserID: "admin", SchoolID: "school", Role: RoleIkonexAdmin})
			_, err := service.InitiateSTKPush(ctx, STKPushRequest{SchoolID: "school", StudentID: "student", PhoneNumber: "+254700000001", Amount: 10.25, Source: "chat", ParentConfirmed: true})
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v", err)
			}
			if tc.wantError && (calls != 0 || len(store.Payments) != 0) {
				t.Fatal("payment initiated while pause could not be cleared")
			}
		})
	}
}

func TestPrepaymentIsVisibleAndAppliedOnce(t *testing.T) {
	for _, paymentAmount := range []float64{500.25, 1500.25} {
		service, store, _, _ := setupTestBillingService()
		ctx := context.Background()
		store.Students["school:student"] = struct{ ID, Admission, Name, GradeID, YearID string }{ID: "student", Admission: "SYNTHETIC"}
		store.EnrolledStudents["school:grade:year"] = []string{"student"}
		if err := service.creditStudentPayment(ctx, "school", "student", "", "synthetic-receipt", "mpesa_c2b", paymentAmount, "", ""); err != nil {
			t.Fatal(err)
		}
		balance, err := service.GetStudentFeeBalance(ctx, "school", "student", "admin", RoleIkonexAdmin)
		if err != nil || balance.CreditBalance != paymentAmount || balance.PaidAmount != paymentAmount {
			t.Fatalf("missing prepaid credit: %+v %v", balance, err)
		}
		if err := service.creditStudentPayment(ctx, "school", "student", "", "synthetic-receipt", "mpesa_c2b", paymentAmount, "", ""); err != nil {
			t.Fatal(err)
		}
		req := LoadFeeStructureRequest{SchoolID: "school", AcademicYearID: "year", GradeLevelID: "grade", TermID: "term", TermStartDate: time.Now(), Items: []FeeStructureItemInput{{Title: "synthetic tuition", Amount: 1000}}}
		if _, err := service.LoadFeeStructure(ctx, req); err != nil {
			t.Fatal(err)
		}
		if _, err := service.LoadFeeStructure(ctx, req); err != nil {
			t.Fatal(err)
		}
		ledger, err := store.GetLatestFeeLedger(ctx, "school", "student")
		if err != nil || ledger.PaidAmount != paymentAmount || ledger.CurrentBalance != math.Max(1000-paymentAmount, 0) || ledger.CreditBalance != math.Max(paymentAmount-1000, 0) {
			t.Fatalf("incorrect first ledger: %+v %v", ledger, err)
		}
		if err := service.creditStudentPayment(ctx, "school", "student", "", "synthetic-receipt-2", "mpesa_c2b", 100.25, "", ""); err != nil {
			t.Fatal(err)
		}
		ledger, _ = store.GetLatestFeeLedger(ctx, "school", "student")
		if ledger.PaidAmount != paymentAmount+100.25 || ledger.CreditBalance != math.Max(paymentAmount+100.25-1000, 0) {
			t.Fatalf("credit lost after later payment: %+v", ledger)
		}
	}
}

type resolutionStore struct {
	*MemoryStore
	transactionMu sync.Mutex
	failAt        string
}

func (s *resolutionStore) WithTx(ctx context.Context, school string, fn func(Store) error) error {
	s.transactionMu.Lock()
	defer s.transactionMu.Unlock()
	data, err := json.Marshal(s.MemoryStore)
	if err != nil {
		return err
	}
	backup := NewMemoryStore()
	if err := json.Unmarshal(data, backup); err != nil {
		return err
	}
	if err := fn(s); err != nil {
		s.Payments = backup.Payments
		s.UnmatchedPayments = backup.UnmatchedPayments
		s.FeeLedgers = backup.FeeLedgers
		s.LedgerEntries = backup.LedgerEntries
		return err
	}
	return nil
}
func (s *resolutionStore) CreateLedgerEntry(ctx context.Context, e *LedgerEntry) error {
	if s.failAt == "entry" {
		return errors.New("entry failed")
	}
	return s.MemoryStore.CreateLedgerEntry(ctx, e)
}
func (s *resolutionStore) SaveFeeLedger(ctx context.Context, l *FeeLedger) error {
	if s.failAt == "ledger" {
		return errors.New("ledger failed")
	}
	return s.MemoryStore.SaveFeeLedger(ctx, l)
}
func (s *resolutionStore) ResolveUnmatchedPayment(ctx context.Context, school, id, student, user, note string) error {
	if s.failAt == "queue" {
		return errors.New("queue failed")
	}
	return s.MemoryStore.ResolveUnmatchedPayment(ctx, school, id, student, user, note)
}
func (s *resolutionStore) UpdatePaymentMatched(ctx context.Context, school, id, student, user string) error {
	if s.failAt == "matched_by" {
		return errors.New("payment failed")
	}
	return s.MemoryStore.UpdatePaymentMatched(ctx, school, id, student, user)
}

func TestManualResolutionRollbackAndDuplicateRequests(t *testing.T) {
	for _, phase := range []string{"entry", "ledger", "queue", "matched_by", "success"} {
		t.Run(phase, func(t *testing.T) {
			store := &resolutionStore{MemoryStore: NewMemoryStore(), failAt: phase}
			ctx := context.Background()
			store.Students["school:student"] = struct{ ID, Admission, Name, GradeID, YearID string }{ID: "student"}
			if err := store.MemoryStore.SaveFeeLedger(ctx, &FeeLedger{SchoolID: "school", StudentID: "student", AcademicYearID: "year", TermID: "term", BilledAmount: 1000, CurrentBalance: 1000}); err != nil {
				t.Fatal(err)
			}
			if err := store.CreatePayment(ctx, &Payment{ID: "payment", SchoolID: "school", TransactionRef: "synthetic-receipt", Provider: "mpesa_c2b", Amount: 100.25, MatchedStatus: "unmatched"}); err != nil {
				t.Fatal(err)
			}
			if err := store.CreateUnmatchedPayment(ctx, &UnmatchedPayment{ID: "queue", SchoolID: "school", PaymentID: "payment", Status: "unmatched"}); err != nil {
				t.Fatal(err)
			}
			notifier := &mockNotifier{}
			service := NewService(store, notifier, nil, nil)
			payment, err := service.MatchUnmatchedPayment(ctx, "school", "payment", "student", "synthetic resolution", "admin")
			if phase != "success" {
				if err == nil {
					t.Fatal("expected failed resolution")
				}
				if store.Payments["payment"].MatchedStatus != "unmatched" || store.UnmatchedPayments["queue"].Status != "unmatched" || len(store.LedgerEntries["school:student"]) != 0 || len(notifier.notifications) != 0 {
					t.Fatal("partial resolution survived rollback")
				}
				ledger, _ := store.GetLatestFeeLedger(ctx, "school", "student")
				if ledger.PaidAmount != 0 {
					t.Fatal("ledger credit survived rollback")
				}
				store.failAt = "success"
				payment, err = service.MatchUnmatchedPayment(ctx, "school", "payment", "student", "retry", "admin")
			}
			if err != nil || payment.MatchedBy == nil || *payment.MatchedBy != "admin" {
				t.Fatalf("resolution failed: %+v %v", payment, err)
			}
			if _, err := service.MatchUnmatchedPayment(ctx, "school", "payment", "student", "duplicate", "admin"); err == nil {
				t.Fatal("duplicate resolution succeeded")
			}
			if len(store.LedgerEntries["school:student"]) != 1 {
				t.Fatal("duplicate credit")
			}
		})
	}
}

func TestSTKPollingAndFractionalReconciliation(t *testing.T) {
	for _, scenario := range []string{"poll success", "provider failure", "declined", "wrong checkout", "wrong merchant", "fractional callback", "truncated callback"} {
		t.Run(scenario, func(t *testing.T) {
			service, store, _, _ := setupTestBillingService()
			ctx := context.Background()
			store.Students["school:student"] = struct{ ID, Admission, Name, GradeID, YearID string }{ID: "student", Admission: "SYNTHETIC"}
			store.MpesaConfigs["school"] = &SchoolMpesaConfig{SchoolID: "school", Shortcode: "123456"}
			student := "student"
			payment := &Payment{ID: "payment", SchoolID: "school", StudentID: &student, TransactionRef: "checkout", Provider: "mpesa_stk", Amount: 100.25, PayerPhone: "254700000001", MatchedStatus: "pending", RawPayload: `{"checkout_request_id":"checkout","merchant_request_id":"merchant"}`}
			if err := store.CreatePayment(ctx, payment); err != nil {
				t.Fatal(err)
			}
			service.SetDarajaClient(&mockDarajaClient{queryFunc: func(context.Context, *SchoolMpesaConfig, string) (*DarajaQueryResponse, error) {
				result := &DarajaQueryResponse{CheckoutRequestID: "checkout", MerchantRequestID: "merchant", ResponseCode: "0", ResultCode: "0"}
				switch scenario {
				case "provider failure":
					return nil, errors.New("synthetic provider failure")
				case "declined":
					result.ResultCode = "1032"
				case "wrong checkout":
					result.CheckoutRequestID = "other"
				case "wrong merchant":
					result.MerchantRequestID = "other"
				}
				return result, nil
			}})
			var err error
			if strings.Contains(scenario, "callback") {
				amount := "100.25"
				if scenario == "truncated callback" {
					amount = "100"
				}
				var payload STKCallbackPayload
				body := `{"Body":{"stkCallback":{"MerchantRequestID":"merchant","CheckoutRequestID":"checkout","ResultCode":0,"CallbackMetadata":{"Item":[{"Name":"Amount","Value":` + amount + `},{"Name":"MpesaReceiptNumber","Value":"synthetic-receipt"},{"Name":"PhoneNumber","Value":"254700000001"}]}}}}`
				if err := json.Unmarshal([]byte(body), &payload); err != nil {
					t.Fatal(err)
				}
				err = service.HandleSTKCallback(ctx, payload)
			} else {
				err = service.QuerySTKPushStatus(ctx, "school", "checkout")
			}
			success := scenario == "poll success" || scenario == "fractional callback"
			if (err == nil) != (success || scenario == "declined") {
				t.Fatalf("unexpected result: %v", err)
			}
			if success {
				if err := store.UpdatePaymentStatus(ctx, "school", "payment", "failed"); err != nil {
					t.Fatal(err)
				}
				if payment.MatchedStatus != "matched" {
					t.Fatal("late failure reopened matched payment")
				}
				if err := service.QuerySTKPushStatus(ctx, "school", "checkout"); err != nil {
					t.Fatal(err)
				}
				entries, _ := store.ListLedgerEntriesForStudent(ctx, "school", "student")
				if len(entries) != 1 || entries[0].Amount != 100.25 {
					t.Fatalf("credit incorrect or repeated: %+v", entries)
				}
			} else if len(store.LedgerEntries) != 0 {
				t.Fatal("unconfirmed checkout credited")
			}
			if err := service.QuerySTKPushStatus(ctx, "other-school", "checkout"); err == nil {
				t.Fatal("cross-school polling accepted")
			}
		})
	}
}

type failedPollQueue struct{}

func (failedPollQueue) EnqueueSTKPoll(context.Context, string, string, time.Duration) error {
	return errors.New("synthetic Redis failure")
}
func TestSTKQueueFailureKeepsPendingPayment(t *testing.T) {
	service, store, _, _ := setupTestBillingService()
	service.enqueuer = failedPollQueue{}
	store.Students["school:student"] = struct{ ID, Admission, Name, GradeID, YearID string }{ID: "student", Admission: "SYNTHETIC"}
	store.MpesaConfigs["school"] = &SchoolMpesaConfig{SchoolID: "school", Shortcode: "123456"}
	ctx := WithIdentity(context.Background(), Identity{UserID: "admin", SchoolID: "school", Role: RoleIkonexAdmin})
	response, err := service.InitiateSTKPush(ctx, STKPushRequest{SchoolID: "school", StudentID: "student", Amount: 100.25, PhoneNumber: "+254700000001", Source: "portal"})
	if err == nil || response == nil {
		t.Fatalf("queue failure ignored: %+v %v", response, err)
	}
	pending, err := store.GetPaymentByCheckoutID(ctx, response.CheckoutRequestID)
	if err != nil || pending.Amount != 100.25 || pending.MatchedStatus != "pending" {
		t.Fatalf("pending provider request lost: %+v %v", pending, err)
	}
}
