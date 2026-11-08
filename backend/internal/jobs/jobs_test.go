package jobs

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/hibiken/asynq"
)

func TestQueueManagerTaskProcessing(t *testing.T) {
	qm := NewQueueManager(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedSchoolID string
	qm.RegisterHandler(TypeMPesaSTKPoll, func(ctx context.Context, payload []byte) error {
		defer wg.Done()
		var p STKPollTaskPayload
		_ = json.Unmarshal(payload, &p)
		receivedSchoolID = p.SchoolID
		return nil
	})

	qm.StartWorker(ctx)

	err := qm.EnqueueSTKPoll(ctx, "school-poll-1", "ws_CO_TEST", 0)
	if err != nil {
		t.Fatalf("EnqueueSTKPoll failed: %v", err)
	}

	wg.Wait()

	if receivedSchoolID != "school-poll-1" {
		t.Errorf("expected schoolID school-poll-1, got %s", receivedSchoolID)
	}
	qm.Stop()
}

func TestQueueManagerIdempotentStopAndRestartCycles(t *testing.T) {
	qm := NewQueueManager(20)
	ctx := context.Background()

	// 1. Calling Stop before StartWorker must not panic
	qm.Stop()
	qm.Stop()

	// 2. Multiple StartWorker / Stop cycles must process tasks cleanly and never panic on repeated Stop calls
	for cycle := 1; cycle <= 3; cycle++ {
		var wg sync.WaitGroup
		wg.Add(1)

		var gotMsg string
		qm.RegisterHandler(TypeSMSTransactional, func(ctx context.Context, payload []byte) error {
			defer wg.Done()
			var p SMSTaskPayload
			_ = json.Unmarshal(payload, &p)
			gotMsg = p.Message
			return nil
		})

		qm.StartWorker(ctx)
		// Calling StartWorker again while already running should be a safe no-op
		qm.StartWorker(ctx)

		expectedMsg := "cycle-message"
		if err := qm.EnqueueSMS(ctx, "+254700000000", expectedMsg); err != nil {
			t.Fatalf("cycle %d: EnqueueSMS failed: %v", cycle, err)
		}

		wg.Wait()
		if gotMsg != expectedMsg {
			t.Errorf("cycle %d: expected message %q, got %q", cycle, expectedMsg, gotMsg)
		}

		// Concurrent repeated Stop calls must be idempotent and never panic
		var stopWG sync.WaitGroup
		for i := 0; i < 5; i++ {
			stopWG.Add(1)
			go func() {
				defer stopWG.Done()
				qm.Stop()
			}()
		}
		stopWG.Wait()
	}
}

type mockSMSDispatcher struct {
	sentTo  string
	sentMsg string
}

func (m *mockSMSDispatcher) SendTransactionalSMS(ctx context.Context, to, msg string) error {
	m.sentTo = to
	m.sentMsg = msg
	return nil
}

type mockBillingQuerier struct {
	queriedSchool   string
	queriedCheckout string
}

func (m *mockBillingQuerier) QuerySTKPushStatus(ctx context.Context, schoolID, checkoutRequestID string) error {
	m.queriedSchool = schoolID
	m.queriedCheckout = checkoutRequestID
	return nil
}

type mockOnboardingPurger struct {
	purgedRetention int
}

func (m *mockOnboardingPurger) PurgeExpiredUnverifiedSubmissions(ctx context.Context, retentionDays int) (int, error) {
	m.purgedRetention = retentionDays
	return 42, nil
}

func TestAsynqHandlerExecution(t *testing.T) {
	mux := NewAsynqServer(asynq.RedisClientOpt{}, 1)
	_ = mux // ensure NewAsynqServer compiles
	serveMux := asynq.NewServeMux()
	sms := &mockSMSDispatcher{}
	billing := &mockBillingQuerier{}
	purger := &mockOnboardingPurger{}

	if err := RegisterAsynqHandlers(serveMux, sms, billing, purger); err != nil {
		t.Fatalf("RegisterAsynqHandlers failed: %v", err)
	}

	ctx := context.Background()

	// 1. Process SMS task
	smsPayload, _ := json.Marshal(SMSTaskPayload{
		ToPhoneNumber: "+254712345678",
		Message:       "Your OTP is 123456",
	})
	smsTask := asynq.NewTask(TypeSMSTransactional, smsPayload)
	if err := serveMux.ProcessTask(ctx, smsTask); err != nil {
		t.Fatalf("ProcessTask for SMS failed: %v", err)
	}

	if sms.sentTo != "+254712345678" || sms.sentMsg != "Your OTP is 123456" {
		t.Errorf("unexpected SMS dispatch: %+v", sms)
	}

	// 2. Process STK Poll task
	stkPayload, _ := json.Marshal(STKPollTaskPayload{
		SchoolID:          "school-1",
		CheckoutRequestID: "ws_CO_TEST_POLL",
	})
	stkTask := asynq.NewTask(TypeMPesaSTKPoll, stkPayload)
	if err := serveMux.ProcessTask(ctx, stkTask); err != nil {
		t.Fatalf("ProcessTask for STK Poll failed: %v", err)
	}
	if billing.queriedSchool != "school-1" || billing.queriedCheckout != "ws_CO_TEST_POLL" {
		t.Errorf("unexpected STK poll: %+v", billing)
	}

	// 3. Process Purge Unverified task
	purgePayload, _ := json.Marshal(PurgeUnverifiedTaskPayload{
		RetentionDays: 7,
	})
	purgeTask := asynq.NewTask(TypeOnboardingPurgeUnverified, purgePayload)
	if err := serveMux.ProcessTask(ctx, purgeTask); err != nil {
		t.Fatalf("ProcessTask for Purge failed: %v", err)
	}
	if purger.purgedRetention != 7 {
		t.Errorf("unexpected retention days: %d", purger.purgedRetention)
	}
}

func TestAsynqHandlerRegistrationRequiresDependencies(t *testing.T) {
	if err := RegisterAsynqHandlers(asynq.NewServeMux(), nil, &mockBillingQuerier{}, &mockOnboardingPurger{}); err == nil {
		t.Fatal("expected missing SMS dispatcher to prevent handler registration")
	}
}
