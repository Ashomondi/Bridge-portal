package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
)

const (
	TypeTransactionalSMS = "sms:transactional"
	TypeSTKPoll          = "mpesa:stk_poll"
	TypePurgeUnverified  = "onboarding:purge_unverified"

	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"

	DefaultSTKPollDelay         = 45 * time.Second
	DefaultSTKPollRetryInterval = 30 * time.Second
	DefaultSTKPollMaxAttempts   = 6
	DefaultOTPTTL               = 5 * time.Minute
	DefaultSMSUniqueTTL         = 10 * time.Minute
	DefaultPurgeRetentionDays   = 7
	DefaultPurgeCronExpression  = "0 2 * * *" // Daily at 02:00 UTC
)

var (
	// ErrMissingDependency indicates a misconfigured worker missing a required
	// handler dependency. It is intentionally NOT wrapped with asynq.SkipRetry so
	// configuration failures surface operationally rather than being silently dropped.
	ErrMissingDependency = errors.New("jobs: required worker dependency is not configured")
)

// DefaultQueuePriorities sets relative weights for worker queue processing.
func DefaultQueuePriorities() map[string]int {
	return map[string]int{
		QueueCritical: 6,
		QueueDefault:  3,
		QueueLow:      1,
	}
}

// SMSPayload carries transactional SMS details. Never log Body or full To.
type SMSPayload struct {
	To             string                   `json:"to"`
	Body           string                   `json:"body"`
	Purpose        notifications.SMSPurpose `json:"purpose"`
	IdempotencyKey string                   `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time                `json:"created_at,omitempty"`
	ExpiresAt      *time.Time               `json:"expires_at,omitempty"`
}

// SMSTaskOptions configures optional expiry and deduplication for SMS tasks.
type SMSTaskOptions struct {
	IdempotencyKey string
	ExpiresAt      *time.Time
	Now            func() time.Time
}

// STKPollPayload carries correlation and rescheduling state for Daraja STK polling.
type STKPollPayload struct {
	SchoolID          string    `json:"school_id"`
	STKRequestID      string    `json:"stk_request_id"`
	CheckoutRequestID string    `json:"checkout_request_id,omitempty"`
	Attempt           int       `json:"attempt,omitempty"`
	MaxAttempts       int       `json:"max_attempts,omitempty"`
	EnqueuedAt        time.Time `json:"enqueued_at"`
}

// PurgeUnverifiedPayload carries configuration for purging unverified onboarding submissions.
type PurgeUnverifiedPayload struct {
	RetentionDays int `json:"retention_days"`
}

func isOTPPurpose(purpose notifications.SMSPurpose) bool {
	switch purpose {
	case notifications.AppLoginOTP, notifications.WhatsAppSessionOTP, notifications.PhoneChangeOTP:
		return true
	default:
		return false
	}
}

// NewTransactionalSMSTask creates an Asynq task for delivering transactional SMS.
// OTP purposes automatically receive a 5-minute expiry deadline so stale login
// codes are never dispatched after retries.
func NewTransactionalSMSTask(msg notifications.SMSMessage, opts ...asynq.Option) (*asynq.Task, error) {
	return NewTransactionalSMSTaskWithOptions(msg, SMSTaskOptions{}, opts...)
}

// NewTransactionalSMSTaskWithOptions creates a transactional SMS task with
// explicit idempotency key and expiry controls.
func NewTransactionalSMSTaskWithOptions(msg notifications.SMSMessage, cfg SMSTaskOptions, opts ...asynq.Option) (*asynq.Task, error) {
	norm, err := notifications.NormalizePhone(msg.To)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", notifications.ErrInvalidPhone, err)
	}
	if strings.TrimSpace(msg.Body) == "" {
		return nil, fmt.Errorf("%w: sms body is required", notifications.ErrInvalidSMS)
	}
	switch msg.Purpose {
	case notifications.AppLoginOTP, notifications.WhatsAppSessionOTP, notifications.PhoneChangeOTP, notifications.PhoneChangeNotice:
	default:
		return nil, fmt.Errorf("%w: unrecognized purpose %q", notifications.ErrInvalidSMS, msg.Purpose)
	}

	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn().UTC()

	expiresAt := cfg.ExpiresAt
	if expiresAt == nil && isOTPPurpose(msg.Purpose) {
		exp := now.Add(DefaultOTPTTL)
		expiresAt = &exp
	}

	idempotencyKey := strings.TrimSpace(cfg.IdempotencyKey)
	payload, err := json.Marshal(SMSPayload{
		To:             norm,
		Body:           msg.Body,
		Purpose:        msg.Purpose,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal sms payload: %w", err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(3),
		asynq.Timeout(20 * time.Second),
		// Do not retain completed sensitive SMS payloads in Redis.
		asynq.Retention(0),
	}
	if expiresAt != nil {
		defaultOpts = append(defaultOpts, asynq.Deadline(*expiresAt))
	}
	if idempotencyKey != "" {
		defaultOpts = append(defaultOpts, asynq.TaskID("sms:"+idempotencyKey), asynq.Unique(DefaultSMSUniqueTTL))
	}
	allOpts := append(defaultOpts, opts...)
	return asynq.NewTask(TypeTransactionalSMS, payload, allOpts...), nil
}

// NewSTKPollTask creates an Asynq task to poll Daraja for a pending STK push request.
func NewSTKPollTask(schoolID, stkRequestID, checkoutRequestID string, opts ...asynq.Option) (*asynq.Task, error) {
	return newSTKPollTaskAttempt(schoolID, stkRequestID, checkoutRequestID, 1, DefaultSTKPollMaxAttempts, opts...)
}

func newSTKPollTaskAttempt(schoolID, stkRequestID, checkoutRequestID string, attempt, maxAttempts int, opts ...asynq.Option) (*asynq.Task, error) {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(stkRequestID) {
		return nil, contracts.Fail(http.StatusBadRequest, "INVALID_STK_TASK", "Valid school_id and stk_request_id are required.")
	}
	if attempt <= 0 {
		attempt = 1
	}
	if maxAttempts <= 0 {
		maxAttempts = DefaultSTKPollMaxAttempts
	}

	payload, err := json.Marshal(STKPollPayload{
		SchoolID:          schoolID,
		STKRequestID:      stkRequestID,
		CheckoutRequestID: checkoutRequestID,
		Attempt:           attempt,
		MaxAttempts:       maxAttempts,
		EnqueuedAt:        time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal stk poll payload: %w", err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
		asynq.TaskID(fmt.Sprintf("stk_poll:%s:%d", stkRequestID, attempt)),
	}
	allOpts := append(defaultOpts, opts...)
	return asynq.NewTask(TypeSTKPoll, payload, allOpts...), nil
}

// STKPollBackoff returns the delay before scheduling the next STK status poll.
func STKPollBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		return DefaultSTKPollRetryInterval
	}
	delay := DefaultSTKPollRetryInterval * time.Duration(attempt)
	if delay > 2*time.Minute {
		return 2 * time.Minute
	}
	return delay
}

// NewPurgeUnverifiedTask creates an Asynq task for unverified onboarding submission cleanup.
func NewPurgeUnverifiedTask(retentionDays int, opts ...asynq.Option) (*asynq.Task, error) {
	if retentionDays <= 0 {
		retentionDays = DefaultPurgeRetentionDays
	}
	payload, err := json.Marshal(PurgeUnverifiedPayload{
		RetentionDays: retentionDays,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal purge payload: %w", err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(QueueLow),
		asynq.MaxRetry(3),
		asynq.Timeout(2 * time.Minute),
	}
	allOpts := append(defaultOpts, opts...)
	return asynq.NewTask(TypePurgeUnverified, payload, allOpts...), nil
}

// STKPollRescheduler schedules follow-up STK polling attempts for non-terminal requests.
type STKPollRescheduler interface {
	EnqueueSTKPollAttempt(ctx context.Context, schoolID, stkRequestID, checkoutRequestID string, attempt, maxAttempts int, delay time.Duration) error
}

// Client wraps an *asynq.Client and implements billing.STKPollQueue and STKPollRescheduler.
type Client struct {
	client *asynq.Client
}

// NewClient constructs an Asynq queue producer client.
func NewClient(client *asynq.Client) *Client {
	return &Client{client: client}
}

// EnqueueSTKPoll satisfies billing.STKPollQueue. If delay <= 0, DefaultSTKPollDelay (45s) is used.
func (c *Client) EnqueueSTKPoll(ctx context.Context, task billing.STKPollTask, delay time.Duration) error {
	if delay <= 0 {
		delay = DefaultSTKPollDelay
	}
	return c.EnqueueSTKPollAttempt(ctx, task.SchoolID, task.STKRequestID, task.CheckoutRequestID, 1, DefaultSTKPollMaxAttempts, delay)
}

// EnqueueSTKPollAttempt enqueues a specific STK poll attempt, ignoring duplicate task IDs.
func (c *Client) EnqueueSTKPollAttempt(ctx context.Context, schoolID, stkRequestID, checkoutRequestID string, attempt, maxAttempts int, delay time.Duration) error {
	if c == nil || c.client == nil {
		return errors.New("asynq client is nil")
	}
	if delay <= 0 {
		delay = DefaultSTKPollDelay
	}

	asynqTask, err := newSTKPollTaskAttempt(schoolID, stkRequestID, checkoutRequestID, attempt, maxAttempts, asynq.ProcessIn(delay))
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, asynqTask)
	if errors.Is(err, asynq.ErrDuplicateTask) || errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}

// EnqueueTransactionalSMS enqueues a transactional SMS for background delivery.
func (c *Client) EnqueueTransactionalSMS(ctx context.Context, msg notifications.SMSMessage, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.EnqueueTransactionalSMSWithOptions(ctx, msg, SMSTaskOptions{}, opts...)
}

// EnqueueTransactionalSMSWithOptions enqueues a transactional SMS with explicit
// expiry and idempotency key controls.
func (c *Client) EnqueueTransactionalSMSWithOptions(ctx context.Context, msg notifications.SMSMessage, cfg SMSTaskOptions, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("asynq client is nil")
	}
	task, err := NewTransactionalSMSTaskWithOptions(msg, cfg, opts...)
	if err != nil {
		return nil, err
	}
	return c.client.EnqueueContext(ctx, task)
}

// EnqueuePurgeUnverified enqueues an unverified submission purge job.
func (c *Client) EnqueuePurgeUnverified(ctx context.Context, retentionDays int, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("asynq client is nil")
	}
	task, err := NewPurgeUnverifiedTask(retentionDays, opts...)
	if err != nil {
		return nil, err
	}
	return c.client.EnqueueContext(ctx, task)
}

// Close closes the underlying Asynq client.
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// STKPoller defines the billing service interface required by HandleSTKPoll.
type STKPoller interface {
	PollSTKRequest(ctx context.Context, schoolID, stkRequestID string) error
}

// STKStatusLookup inspects the current persisted status of an STK request so the
// worker can decide whether to schedule another poll.
type STKStatusLookup interface {
	LookupSTKStatus(ctx context.Context, schoolID, stkRequestID string) (string, error)
}

// STKStatusLookupFunc adapts a function to STKStatusLookup.
type STKStatusLookupFunc func(ctx context.Context, schoolID, stkRequestID string) (string, error)

func (f STKStatusLookupFunc) LookupSTKStatus(ctx context.Context, schoolID, stkRequestID string) (string, error) {
	return f(ctx, schoolID, stkRequestID)
}

// UnverifiedSubmissionPurger defines the onboarding cleanup interface provided by Evans (@eojuma).
type UnverifiedSubmissionPurger interface {
	PurgeExpiredUnverifiedSubmissions(ctx context.Context, retentionDays int) (int64, error)
}

// UnverifiedPurgerFunc adapts a function to UnverifiedSubmissionPurger.
type UnverifiedPurgerFunc func(ctx context.Context, retentionDays int) (int64, error)

func (f UnverifiedPurgerFunc) PurgeExpiredUnverifiedSubmissions(ctx context.Context, retentionDays int) (int64, error) {
	return f(ctx, retentionDays)
}

// WorkerDeps bundles dependencies for Asynq job execution.
type WorkerDeps struct {
	SMSDispatcher   notifications.SMSDispatcher
	STKPoller       STKPoller
	STKStatusLookup STKStatusLookup
	STKRescheduler  STKPollRescheduler
	TenantDB        *TenantDBAdapter
	Purger          UnverifiedSubmissionPurger
	Logger          *slog.Logger
	Now             func() time.Time
}

func (d WorkerDeps) logger() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}
	return slog.Default()
}

func (d WorkerDeps) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}

func (d WorkerDeps) lookupSTKStatus(ctx context.Context, schoolID, stkRequestID string) (string, error) {
	if d.STKStatusLookup != nil {
		return d.STKStatusLookup.LookupSTKStatus(ctx, schoolID, stkRequestID)
	}
	if d.TenantDB == nil {
		return "", nil
	}
	var status string
	err := d.TenantDB.WithSchoolTx(ctx, schoolID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT status
			FROM mpesa_stk_requests
			WHERE school_id = $1 AND id = $2`,
			schoolID, stkRequestID,
		).Scan(&status)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return status, err
}

// HandleTransactionalSMS unmarshals and dispatches an SMS, enforcing OTP expiry
// and classifying permanent vs retryable provider failures.
func HandleTransactionalSMS(deps WorkerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload SMSPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("%w: malformed sms payload: %v", asynq.SkipRetry, err)
		}
		if deps.SMSDispatcher == nil {
			return fmt.Errorf("%w: sms dispatcher is nil", ErrMissingDependency)
		}
		if payload.ExpiresAt != nil && deps.now().After(*payload.ExpiresAt) {
			deps.logger().WarnContext(ctx, "dropping expired transactional SMS task",
				slog.String("purpose", string(payload.Purpose)))
			return fmt.Errorf("%w: sms task expired before dispatch", asynq.SkipRetry)
		}

		msg := notifications.SMSMessage{
			To:      payload.To,
			Body:    payload.Body,
			Purpose: payload.Purpose,
		}

		if err := deps.SMSDispatcher.DispatchSMS(ctx, msg); err != nil {
			if errors.Is(err, notifications.ErrInvalidPhone) || errors.Is(err, notifications.ErrInvalidSMS) {
				return fmt.Errorf("%w: permanent sms validation error: %v", asynq.SkipRetry, err)
			}
			var deliveryErr *notifications.SMSDeliveryError
			if errors.As(err, &deliveryErr) {
				if !deliveryErr.Retryable {
					return fmt.Errorf("%w: non-retryable sms delivery error (%s)", asynq.SkipRetry, deliveryErr.Code)
				}
				return fmt.Errorf("retryable sms delivery error (%s): %w", deliveryErr.Code, err)
			}
			return fmt.Errorf("dispatch sms: %w", err)
		}
		return nil
	}
}

// HandleSTKPoll executes a Daraja STK query check and reschedules another poll
// when the transaction remains non-terminal within the configured attempt budget.
func HandleSTKPoll(deps WorkerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload STKPollPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("%w: malformed stk poll payload: %v", asynq.SkipRetry, err)
		}
		if !contracts.ValidID(payload.SchoolID) || !contracts.ValidID(payload.STKRequestID) {
			return fmt.Errorf("%w: invalid school_id or stk_request_id", asynq.SkipRetry)
		}
		if deps.STKPoller == nil {
			return fmt.Errorf("%w: stk poller dependency is nil", ErrMissingDependency)
		}

		if err := deps.STKPoller.PollSTKRequest(ctx, payload.SchoolID, payload.STKRequestID); err != nil {
			var apiErr *contracts.APIError
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
				return fmt.Errorf("%w: permanent client error on stk poll: %v", asynq.SkipRetry, err)
			}
			return fmt.Errorf("poll stk request: %w", err)
		}

		status, err := deps.lookupSTKStatus(ctx, payload.SchoolID, payload.STKRequestID)
		if err != nil {
			return fmt.Errorf("lookup stk status after poll: %w", err)
		}
		if status != "pending" && status != "uncertain" {
			return nil
		}

		attempt := payload.Attempt
		if attempt <= 0 {
			attempt = 1
		}
		maxAttempts := payload.MaxAttempts
		if maxAttempts <= 0 {
			maxAttempts = DefaultSTKPollMaxAttempts
		}
		if attempt >= maxAttempts {
			deps.logger().WarnContext(ctx, "stk poll reached maximum attempts while still non-terminal",
				slog.String("school_id", payload.SchoolID),
				slog.String("stk_request_id", payload.STKRequestID),
				slog.String("status", status),
				slog.Int("attempts", attempt))
			return nil
		}
		if deps.STKRescheduler == nil {
			return fmt.Errorf("%w: stk rescheduler is nil while request remains %s", ErrMissingDependency, status)
		}

		nextAttempt := attempt + 1
		delay := STKPollBackoff(attempt)
		if err := deps.STKRescheduler.EnqueueSTKPollAttempt(
			ctx,
			payload.SchoolID,
			payload.STKRequestID,
			payload.CheckoutRequestID,
			nextAttempt,
			maxAttempts,
			delay,
		); err != nil {
			return fmt.Errorf("reschedule stk poll attempt %d: %w", nextAttempt, err)
		}
		return nil
	}
}

// HandlePurgeUnverified invokes onboarding cleanup for expired unverified registration submissions.
// If no purger is wired, it fails explicitly rather than silently reporting success.
func HandlePurgeUnverified(deps WorkerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload PurgeUnverifiedPayload
		if len(t.Payload()) > 0 {
			if err := json.Unmarshal(t.Payload(), &payload); err != nil {
				return fmt.Errorf("%w: malformed purge payload: %v", asynq.SkipRetry, err)
			}
		}
		if payload.RetentionDays <= 0 {
			payload.RetentionDays = DefaultPurgeRetentionDays
		}

		if deps.Purger == nil {
			return fmt.Errorf("%w: onboarding purger dependency is not configured", ErrMissingDependency)
		}

		purgedCount, err := deps.Purger.PurgeExpiredUnverifiedSubmissions(ctx, payload.RetentionDays)
		if err != nil {
			return fmt.Errorf("purge unverified submissions: %w", err)
		}

		deps.logger().InfoContext(ctx, "purged expired unverified submissions",
			slog.Int("retention_days", payload.RetentionDays),
			slog.Int64("purged_count", purgedCount))
		return nil
	}
}

// RegisterHandlers registers task handlers for all three background job types.
func RegisterHandlers(mux *asynq.ServeMux, deps WorkerDeps) {
	mux.HandleFunc(TypeTransactionalSMS, HandleTransactionalSMS(deps))
	mux.HandleFunc(TypeSTKPoll, HandleSTKPoll(deps))
	mux.HandleFunc(TypePurgeUnverified, HandlePurgeUnverified(deps))
}

// DefaultServerConfig returns a configured asynq.Config with queue priorities,
// safe concurrency, and shutdown grace period.
func DefaultServerConfig(logger *slog.Logger) asynq.Config {
	return ServerConfigWithOptions(logger, 10, 15*time.Second)
}

// ServerConfigWithOptions builds an asynq.Config with explicit concurrency and shutdown timeout.
func ServerConfigWithOptions(logger *slog.Logger, concurrency int, shutdownTimeout time.Duration) asynq.Config {
	if concurrency <= 0 {
		concurrency = 10
	}
	if shutdownTimeout <= 0 {
		shutdownTimeout = 15 * time.Second
	}
	return asynq.Config{
		Concurrency:     concurrency,
		ShutdownTimeout: shutdownTimeout,
		Queues:          DefaultQueuePriorities(),
		StrictPriority:  false,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			if logger != nil {
				// Never log raw task payloads (SMS bodies/OTPs/phone numbers).
				logger.ErrorContext(ctx, "asynq task failed",
					slog.String("type", task.Type()),
					slog.String("err", err.Error()))
			}
		}),
	}
}

// NewServer creates a new background worker server.
func NewServer(opt asynq.RedisConnOpt, cfg asynq.Config) *asynq.Server {
	return asynq.NewServer(opt, cfg)
}

// NewScheduler creates an Asynq periodic task scheduler pinned to loc (default UTC).
func NewScheduler(opt asynq.RedisConnOpt, loc *time.Location) *asynq.Scheduler {
	if loc == nil {
		loc = time.UTC
	}
	return asynq.NewScheduler(opt, &asynq.SchedulerOpts{Location: loc})
}

// RegisterSchedulerTasks registers periodic cron jobs (daily purge of unverified submissions).
func RegisterSchedulerTasks(scheduler *asynq.Scheduler, retentionDays int, cronExpr string) (string, error) {
	if scheduler == nil {
		return "", errors.New("scheduler is nil")
	}
	if cronExpr == "" {
		cronExpr = DefaultPurgeCronExpression
	}
	task, err := NewPurgeUnverifiedTask(retentionDays)
	if err != nil {
		return "", err
	}
	return scheduler.Register(cronExpr, task)
}
