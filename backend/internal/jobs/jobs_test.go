package jobs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
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

type mockSTKRescheduler struct {
	mu        sync.Mutex
	scheduled []struct {
		schoolID, reqID, checkoutID string
		attempt, maxAttempts        int
		delay                       time.Duration
	}
	err error
}

func (m *mockSTKRescheduler) EnqueueSTKPollAttempt(_ context.Context, schoolID, stkRequestID, checkoutRequestID string, attempt, maxAttempts int, delay time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.scheduled = append(m.scheduled, struct {
		schoolID, reqID, checkoutID string
		attempt, maxAttempts        int
		delay                       time.Duration
	}{schoolID, stkRequestID, checkoutRequestID, attempt, maxAttempts, delay})
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
	var _ jobs.STKPollRescheduler = (*jobs.Client)(nil)
	var _ jobs.UnverifiedSubmissionPurger = jobs.UnverifiedPurgerFunc(nil)
}

func TestNewTransactionalSMSTaskValidationAndOTPExpiry(t *testing.T) {
	fixedNow := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	task, err := jobs.NewTransactionalSMSTaskWithOptions(notifications.SMSMessage{
		To:      "0712345678",
		Body:    "Your OTP is 654321",
		Purpose: notifications.WhatsAppSessionOTP,
	}, jobs.SMSTaskOptions{
		IdempotencyKey: "otp-challenge-1",
		Now:            func() time.Time { return fixedNow },
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
	if payload.ExpiresAt == nil || !payload.ExpiresAt.Equal(fixedNow.Add(jobs.DefaultOTPTTL)) {
		t.Fatalf("expected OTP expiry %v, got %v", fixedNow.Add(jobs.DefaultOTPTTL), payload.ExpiresAt)
	}
	if payload.IdempotencyKey != "otp-challenge-1" {
		t.Fatalf("expected idempotency key otp-challenge-1, got %q", payload.IdempotencyKey)
	}

	// Non-OTP transactional notice has no forced OTP expiry unless explicitly set.
	noticeTask, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To:      "+254712345678",
		Body:    "Your phone number was changed.",
		Purpose: notifications.PhoneChangeNotice,
	})
	if err != nil {
		t.Fatalf("unexpected error creating notice task: %v", err)
	}
	var noticePayload jobs.SMSPayload
	if err := json.Unmarshal(noticeTask.Payload(), &noticePayload); err != nil {
		t.Fatal(err)
	}
	if noticePayload.ExpiresAt != nil {
		t.Fatalf("expected nil ExpiresAt for phone_change_notice, got %v", noticePayload.ExpiresAt)
	}

	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To: "not-a-number", Body: "OTP", Purpose: notifications.AppLoginOTP,
	}); err == nil {
		t.Fatal("expected error on invalid phone number")
	}
	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To: "0712345678", Body: "   ", Purpose: notifications.AppLoginOTP,
	}); err == nil {
		t.Fatal("expected error on empty body")
	}
	if _, err := jobs.NewTransactionalSMSTask(notifications.SMSMessage{
		To: "0712345678", Body: "Marketing", Purpose: "promotional_broadcast",
	}); err == nil {
		t.Fatal("expected error on non-transactional purpose")
	}
}

func TestHandleTransactionalSMSPolicyAndExpiry(t *testing.T) {
	ctx := context.Background()
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	fixedNow := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	validTask, err := jobs.NewTransactionalSMSTaskWithOptions(notifications.SMSMessage{
		To:      "+254712345678",
		Body:    "Your login code is 987654",
		Purpose: notifications.AppLoginOTP,
	}, jobs.SMSTaskOptions{Now: func() time.Time { return fixedNow }})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Nil dispatcher must fail operationally (NOT asynq.SkipRetry).
	nilHandler := jobs.HandleTransactionalSMS(jobs.WorkerDeps{SMSDispatcher: nil, Logger: logger})
	err = nilHandler(ctx, validTask)
	if !errors.Is(err, jobs.ErrMissingDependency) || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected non-SkipRetry ErrMissingDependency for nil dispatcher, got %v", err)
	}

	// 2. Success before expiry.
	dispatcher := &mockSMSDispatcher{}
	handler := jobs.HandleTransactionalSMS(jobs.WorkerDeps{
		SMSDispatcher: dispatcher,
		Logger:        logger,
		Now:           func() time.Time { return fixedNow.Add(2 * time.Minute) },
	})
	if err := handler(ctx, validTask); err != nil {
		t.Fatalf("handler failed on valid unexpired task: %v", err)
	}
	if len(dispatcher.dispatched) != 1 {
		t.Fatalf("expected 1 dispatched sms, got %d", len(dispatcher.dispatched))
	}

	// 3. Expired OTP task is dropped with SkipRetry without calling provider.
	dispatcher.dispatched = nil
	expiredHandler := jobs.HandleTransactionalSMS(jobs.WorkerDeps{
		SMSDispatcher: dispatcher,
		Logger:        logger,
		Now:           func() time.Time { return fixedNow.Add(6 * time.Minute) },
	})
	err = expiredHandler(ctx, validTask)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry for expired OTP, got %v", err)
	}
	if len(dispatcher.dispatched) != 0 {
		t.Fatalf("expected expired OTP not to call provider, got %d calls", len(dispatcher.dispatched))
	}
	if strings.Contains(logBuf.String(), "987654") || strings.Contains(logBuf.String(), "+254712345678") {
		t.Fatalf("OTP or full phone number leaked into logs: %s", logBuf.String())
	}

	// 4. Non-retryable SMSDeliveryError -> SkipRetry.
	dispatcher.err = &notifications.SMSDeliveryError{Code: "PROVIDER_REJECTED", Retryable: false}
	err = handler(ctx, validTask)
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry on non-retryable SMSDeliveryError, got %v", err)
	}

	// 5. Retryable SMSDeliveryError (5xx / 429 / transport) -> retryable.
	dispatcher.err = &notifications.SMSDeliveryError{Code: "HTTP_ERROR", Retryable: true}
	err = handler(ctx, validTask)
	if err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable error on retryable SMSDeliveryError, got %v", err)
	}

	// 6. Malformed payload -> SkipRetry.
	err = handler(ctx, asynq.NewTask(jobs.TypeTransactionalSMS, []byte("{bad")))
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry on malformed payload, got %v", err)
	}
}

func TestHandleSTKPollReschedulingAndErrors(t *testing.T) {
	ctx := context.Background()
	schoolID := "11111111-1111-4111-8111-111111111111"
	reqID := "22222222-2222-4222-8222-222222222222"

	task, err := jobs.NewSTKPollTask(schoolID, reqID, "chk_123")
	if err != nil {
		t.Fatal(err)
	}

	// 1. Nil poller returns ErrMissingDependency (not SkipRetry).
	nilHandler := jobs.HandleSTKPoll(jobs.WorkerDeps{STKPoller: nil})
	if err := nilHandler(ctx, task); !errors.Is(err, jobs.ErrMissingDependency) || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected non-SkipRetry ErrMissingDependency for nil poller, got %v", err)
	}

	// 2. Pending transaction after poll -> schedules attempt 2.
	poller := &mockSTKPoller{}
	rescheduler := &mockSTKRescheduler{}
	pendingHandler := jobs.HandleSTKPoll(jobs.WorkerDeps{
		STKPoller:       poller,
		STKStatusLookup: jobs.STKStatusLookupFunc(func(context.Context, string, string) (string, error) { return "pending", nil }),
		STKRescheduler:  rescheduler,
	})
	if err := pendingHandler(ctx, task); err != nil {
		t.Fatalf("unexpected error on pending poll: %v", err)
	}
	if len(rescheduler.scheduled) != 1 || rescheduler.scheduled[0].attempt != 2 {
		t.Fatalf("expected rescheduled attempt 2, got %+v", rescheduler.scheduled)
	}

	// 3. Terminal status (succeeded) -> does NOT reschedule.
	rescheduler.scheduled = nil
	succeededHandler := jobs.HandleSTKPoll(jobs.WorkerDeps{
		STKPoller:       poller,
		STKStatusLookup: jobs.STKStatusLookupFunc(func(context.Context, string, string) (string, error) { return "succeeded", nil }),
		STKRescheduler:  rescheduler,
	})
	if err := succeededHandler(ctx, task); err != nil {
		t.Fatalf("unexpected error on succeeded poll: %v", err)
	}
	if len(rescheduler.scheduled) != 0 {
		t.Fatalf("expected no reschedule for succeeded request, got %+v", rescheduler.scheduled)
	}

	// 4. Max attempts reached while still pending -> stops polling cleanly.
	maxPayload, _ := json.Marshal(jobs.STKPollPayload{
		SchoolID:          schoolID,
		STKRequestID:      reqID,
		CheckoutRequestID: "chk_123",
		Attempt:           jobs.DefaultSTKPollMaxAttempts,
		MaxAttempts:       jobs.DefaultSTKPollMaxAttempts,
	})
	maxTask := asynq.NewTask(jobs.TypeSTKPoll, maxPayload)
	if err := pendingHandler(ctx, maxTask); err != nil {
		t.Fatalf("unexpected error at max attempts: %v", err)
	}
	if len(rescheduler.scheduled) != 0 {
		t.Fatalf("expected polling to stop at max attempts, got %+v", rescheduler.scheduled)
	}

	// 5. 4xx error -> SkipRetry; 5xx error -> retryable.
	poller.err = contracts.Fail(http.StatusNotFound, "NOT_FOUND", "not found")
	if err := succeededHandler(ctx, task); !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected SkipRetry on 404, got %v", err)
	}
	poller.err = contracts.Fail(http.StatusBadGateway, "UPSTREAM_DOWN", "upstream down")
	if err := succeededHandler(ctx, task); err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected retryable error on 502, got %v", err)
	}
}

func TestHandlePurgeUnverifiedFailsOnNilPurger(t *testing.T) {
	ctx := context.Background()
	task, err := jobs.NewPurgeUnverifiedTask(7)
	if err != nil {
		t.Fatal(err)
	}

	// 1. With injected purger.
	purger := &mockPurger{purgedResult: 5}
	handler := jobs.HandlePurgeUnverified(jobs.WorkerDeps{Purger: purger})
	if err := handler(ctx, task); err != nil {
		t.Fatalf("purge handler failed: %v", err)
	}
	if len(purger.calledDays) != 1 || purger.calledDays[0] != 7 {
		t.Fatalf("unexpected calledDays: %+v", purger.calledDays)
	}

	// 2. Nil purger must fail explicitly (never silently return nil).
	nilHandler := jobs.HandlePurgeUnverified(jobs.WorkerDeps{Purger: nil})
	if err := nilHandler(ctx, task); !errors.Is(err, jobs.ErrMissingDependency) {
		t.Fatalf("expected ErrMissingDependency when purger is nil, got %v", err)
	}

	// 3. Purger error propagates.
	purger.err = errors.New("db unavailable")
	if err := handler(ctx, task); err == nil {
		t.Fatal("expected error when purger returns error")
	}
}

func TestAsynqRedisEndToEndAndDeduplication(t *testing.T) {
	addr := strings.TrimSpace(os.Getenv("DEV05_TEST_REDIS_ADDR"))
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("Redis is required in CI at %s: %v", addr, err)
		}
		t.Skipf("skipping Redis end-to-end integration test; Redis unavailable at %s", addr)
	}
	_ = conn.Close()

	redisOpt := asynq.RedisClientOpt{Addr: addr, DB: 15}
	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()
	for _, q := range []string{jobs.QueueCritical, jobs.QueueDefault, jobs.QueueLow} {
		_, _ = inspector.DeleteAllScheduledTasks(q)
		_, _ = inspector.DeleteAllPendingTasks(q)
		_, _ = inspector.DeleteAllRetryTasks(q)
		_, _ = inspector.DeleteAllArchivedTasks(q)
		_, _ = inspector.DeleteAllCompletedTasks(q)
	}

	client := asynq.NewClient(redisOpt)
	producer := jobs.NewClient(client)
	defer producer.Close()

	dispatcher := &mockSMSDispatcher{}
	deps := jobs.WorkerDeps{SMSDispatcher: dispatcher}
	mux := asynq.NewServeMux()
	jobs.RegisterHandlers(mux, deps)

	server := jobs.NewServer(redisOpt, jobs.ServerConfigWithOptions(nil, 2, 5*time.Second))
	if err := server.Start(mux); err != nil {
		t.Fatalf("failed to start asynq server: %v", err)
	}
	defer server.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idempKey := "e2e-sms-" + time.Now().Format("20060102150405.000000000")
	msg := notifications.SMSMessage{
		To:      "+254712345678",
		Body:    "Your login code is 112233",
		Purpose: notifications.AppLoginOTP,
	}

	info, err := producer.EnqueueTransactionalSMSWithOptions(ctx, msg, jobs.SMSTaskOptions{IdempotencyKey: idempKey})
	if err != nil {
		t.Fatalf("failed to enqueue transactional sms: %v", err)
	}
	if info.Queue != jobs.QueueCritical {
		t.Fatalf("expected queue %s, got %s", jobs.QueueCritical, info.Queue)
	}

	// Duplicate enqueue with same IdempotencyKey must be rejected by Asynq uniqueness/task ID.
	if _, err := producer.EnqueueTransactionalSMSWithOptions(ctx, msg, jobs.SMSTaskOptions{IdempotencyKey: idempKey}); !errors.Is(err, asynq.ErrDuplicateTask) && !errors.Is(err, asynq.ErrTaskIDConflict) {
		t.Fatalf("expected duplicate task rejection, got %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	dispatched := false
	for time.Now().Before(deadline) {
		dispatcher.mu.Lock()
		n := len(dispatcher.dispatched)
		dispatcher.mu.Unlock()
		if n > 0 {
			dispatched = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !dispatched {
		t.Fatal("timed out waiting for Asynq worker to dispatch SMS task from Redis")
	}

	// Verify asynq.Retention(0): completed SMS payload is not retained in Redis.
	completed, err := inspector.ListCompletedTasks(jobs.QueueCritical)
	if err != nil {
		t.Fatalf("list completed tasks: %v", err)
	}
	if len(completed) != 0 {
		t.Fatalf("expected 0 retained completed SMS tasks in Redis, got %d", len(completed))
	}
}
