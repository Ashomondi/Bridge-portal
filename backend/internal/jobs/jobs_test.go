package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/hibiken/asynq"
)

type mockSMSDispatcher struct {
	mu         sync.Mutex
	dispatched []notifications.SMSMessage
	err        error
}

func (m *mockSMSDispatcher) DispatchSMS(_ context.Context, msg notifications.SMSMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.dispatched = append(m.dispatched, msg)
	return nil
}

type mockSTKPoller struct {
	mu    sync.Mutex
	polls []struct{ schoolID, reqID string }
	err   error
}

func (m *mockSTKPoller) PollSTKRequest(_ context.Context, schoolID, stkRequestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.polls = append(m.polls, struct{ schoolID, reqID string }{schoolID, stkRequestID})
	return nil
}

type mockPurger struct {
	mu           sync.Mutex
	calledDays   []int
	purgedResult int64
	err          error
}

func (m *mockPurger) PurgeExpiredUnverifiedSubmissions(_ context.Context, retentionDays int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return 0, m.err
	}
	m.calledDays = append(m.calledDays, retentionDays)
	return m.purgedResult, nil
}

func TestInterfaceSatisfaction(t *testing.T) {
	var _ billing.STKPollQueue = (*jobs.Client)(nil)
	var _ jobs.UnverifiedSubmissionPurger = jobs.UnverifiedPurgerFunc(nil)
}

func TestNewTransactionalSMSTaskValidation(t *testing.T) {
	// Valid SMS
	task, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "0712345678",
		Body:    "Your OTP is 123456",
		Purpose: notifications.WhatsAppSessionOTP,
	})
	if err != nil {
		t.Fatalf("unexpected error creating sms task: %v", err)
	}
	if task.Type() != jobs.TypeTransactionalSMS {
		t.Fatalf("task type mismatch: got %s, want %s", task.Type(), jobs.TypeTransactionalSMS)
	}
	var payload jobs.SMSPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.To != "+254712345678" {
		t.Fatalf("expected normalized phone +254712345678, got %q", payload.To)
	}
	if payload.Purpose != notifications.WhatsAppSessionOTP {
		t.Fatalf("purpose mismatch: got %s", payload.Purpose)
	}

	// Invalid phone
	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "not-a-number",
		Body:    "OTP",
		Purpose: notifications.AppLoginOTP,
	}); err == nil {
		t.Fatal("expected error on invalid phone number")
	}

	// Empty body
	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "0712345678",
		Body:    "   ",
		Purpose: notifications.AppLoginOTP,
	}); err == nil {
		t.Fatal("expected error on empty body")
	}

	// Invalid purpose
	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "0712345678",
		Body:    "Marketing message",
		Purpose: "promotional_broadcast",
	}); err == nil {
		t.Fatal("expected error on non-transactional purpose")
	}
}

func TestNewSTKPollTaskValidation(t *testing.T) {
	schoolID := "11111111-1111-4111-8111-111111111111"
	reqID := "22222222-2222-4222-8222-222222222222"

	task, err := jobs.NewSTKPollTask(schoolID, reqID, "ws_CO_TEST_001")
	if err != nil {
		t.Fatalf("unexpected error creating stk poll task: %v", err)
	}
	if task.Type() != jobs.TypeSTKPoll {
		t.Fatalf("task type mismatch: got %s", task.Type())
	}

	var payload jobs.STKPollPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal stk payload: %v", err)
	}
	if payload.SchoolID != schoolID || payload.STKRequestID != reqID {
		t.Fatalf("payload mismatch: %+v", payload)
	}

	// Invalid IDs
	if _, err := jobs.NewSTKPollTask("bad-school-id", reqID, ""); err == nil {
		t.Fatal("expected error on invalid school_id")
	}
	if _, err := jobs.NewSTKPollTask(schoolID, "bad-req-id", ""); err == nil {
		t.Fatal("expected error on invalid stk_request_id")
	}
}

func TestNewPurgeUnverifiedTaskDefaults(t *testing.T) {
	taskDefault, err := jobs.NewPurgeUnverifiedTask(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var payloadDefault jobs.PurgeUnverifiedPayload
	if err := json.Unmarshal(taskDefault.Payload(), &payloadDefault); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payloadDefault.RetentionDays != jobs.DefaultPurgeRetentionDays {
		t.Fatalf("retention days default mismatch: got %d, want %d", payloadDefault.RetentionDays, jobs.DefaultPurgeRetentionDays)
	}

	taskCustom, err := jobs.NewPurgeUnverifiedTask(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var payloadCustom jobs.PurgeUnverifiedPayload
	if err := json.Unmarshal(taskCustom.Payload(), &payloadCustom); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payloadCustom.RetentionDays != 14 {
		t.Fatalf("custom retention days mismatch: got %d, want 14", payloadCustom.RetentionDays)
	}
}

func TestHandleTransactionalSMSRetryPolicy(t *testing.T) {
	ctx := context.Background()

	// 1. Success case
	dispatcher := &mockSMSDispatcher{}
	handler := jobs.HandleTransactionalSMS(jobs.WorkerDeps{SMSDispatcher: dispatcher})

	validTask, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "+254712345678",
		Body:    "Security alert",
		Purpose: notifications.PhoneChangeNotice,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := handler(ctx, validTask); err != nil {
		t.Fatalf("handler failed on valid task: %v", err)
	}
	if len(dispatcher.dispatched) != 1 || dispatcher.dispatched[0].To != "+254712345678" {
		t.Fatalf("unexpected dispatched sms: %+v", dispatcher.dispatched)
	}

	// 2. Malformed payload -> asynq.SkipRetry
	malformedTask := asynq.NewTask(jobs.TypeTransactionalSMS, []byte("not-json"))
	err = handler(ctx, malformedTask)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry on malformed payload, got %v", err)
	}

	// 3. Permanent error from dispatcher -> asynq.SkipRetry
	dispatcher.err = notifications.ErrInvalidPhone
	err = handler(ctx, validTask)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry on ErrInvalidPhone, got %v", err)
	}

	// 4. Temporary error -> retryable (not SkipRetry)
	dispatcher.err = errors.New("network timeout")
	err = handler(ctx, validTask)
	if err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable error on network timeout, got %v", err)
	}
}

func TestHandleSTKPollRetryPolicy(t *testing.T) {
	ctx := context.Background()
	poller := &mockSTKPoller{}
	handler := jobs.HandleSTKPoll(jobs.WorkerDeps{STKPoller: poller})

	schoolID := "11111111-1111-4111-8111-111111111111"
	reqID := "22222222-2222-4222-8222-222222222222"
	task, err := jobs.NewSTKPollTask(schoolID, reqID, "chk_123")
	if err != nil {
		t.Fatal(err)
	}

	// 1. Success case
	if err := handler(ctx, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(poller.polls) != 1 || poller.polls[0].schoolID != schoolID {
		t.Fatalf("unexpected polls recorded: %+v", poller.polls)
	}

	// 2. Permanent 4xx client error -> SkipRetry
	poller.err = contracts.Fail(http.StatusNotFound, "NOT_FOUND", "stk request not found")
	err = handler(ctx, task)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry on 404 client error, got %v", err)
	}

	// 3. Transient 5xx error -> retryable
	poller.err = contracts.Fail(http.StatusBadGateway, "UPSTREAM_DOWN", "Daraja upstream down")
	err = handler(ctx, task)
	if err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable error on 502, got %v", err)
	}
}

func TestHandlePurgeUnverified(t *testing.T) {
	ctx := context.Background()

	// 1. With injected purger
	purger := &mockPurger{purgedResult: 5}
	handler := jobs.HandlePurgeUnverified(jobs.WorkerDeps{Purger: purger})

	task, err := jobs.NewPurgeUnverifiedTask(7)
	if err != nil {
		t.Fatal(err)
	}
	if err := handler(ctx, task); err != nil {
		t.Fatalf("purge handler failed: %v", err)
	}
	if len(purger.calledDays) != 1 || purger.calledDays[0] != 7 {
		t.Fatalf("unexpected calledDays: %+v", purger.calledDays)
	}

	// 2. Nil purger skips cleanly without error
	nilPurgerHandler := jobs.HandlePurgeUnverified(jobs.WorkerDeps{Purger: nil})
	if err := nilPurgerHandler(ctx, task); err != nil {
		t.Fatalf("expected nil purger to skip safely, got %v", err)
	}
}

func TestAsynqRedisQueueAndWorkerEndToEnd(t *testing.T) {
	// Check if Redis is running on 127.0.0.1:6379
	conn, err := net.DialTimeout("tcp", "127.0.0.1:6379", 500*time.Millisecond)
	if err != nil {
		t.Skip("skipping Redis end-to-end integration test; Redis is not available on 127.0.0.1:6379")
	}
	_ = conn.Close()

	redisOpt := asynq.RedisClientOpt{Addr: "127.0.0.1:6379", DB: 15}
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	producer := jobs.NewClient(client)
	defer producer.Close()

	dispatcher := &mockSMSDispatcher{}
	deps := jobs.WorkerDeps{
		SMSDispatcher: dispatcher,
	}

	mux := asynq.NewServeMux()
	jobs.RegisterHandlers(mux, deps)

	server := jobs.NewServer(redisOpt, asynq.Config{
		Concurrency: 2,
		Queues:      jobs.DefaultQueuePriorities(),
	})

	serverStarted := make(chan struct{})
	go func() {
		close(serverStarted)
		_ = server.Run(mux)
	}()
	<-serverStarted
	defer server.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testMsg := notifications.SMSMessage{
		To:      "+254712345678",
		Body:    "Testing Redis background worker delivery",
		Purpose: notifications.AppLoginOTP,
	}

	info, err := producer.EnqueueTransactionalSMS(ctx, testMsg)
	if err != nil {
		t.Fatalf("failed to enqueue transactional sms: %v", err)
	}
	if info.Queue != jobs.QueueCritical {
		t.Fatalf("expected queue %s, got %s", jobs.QueueCritical, info.Queue)
	}

	// Wait for worker to process task
	deadline := time.Now().Add(5 * time.Second)
	var processed bool
	for time.Now().Before(deadline) {
		dispatcher.mu.Lock()
		count := len(dispatcher.dispatched)
		dispatcher.mu.Unlock()
		if count > 0 {
			processed = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !processed {
		t.Fatal("timed out waiting for Asynq worker to dispatch SMS task from Redis")
	}

	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	if dispatcher.dispatched[0].Body != testMsg.Body {
		t.Fatalf("dispatched body mismatch: got %q, want %q", dispatcher.dispatched[0].Body, testMsg.Body)
	}
}
