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
)

const (
	TypeTransactionalSMS = "sms:transactional"
	TypeSTKPoll          = "mpesa:stk_poll"
	TypePurgeUnverified  = "onboarding:purge_unverified"

	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"

	DefaultSTKPollDelay        = 45 * time.Second
	DefaultPurgeRetentionDays  = 7
	DefaultPurgeCronExpression = "0 2 * * *" // Daily at 02:00
)

// DefaultQueuePriorities sets relative weights for worker queue processing.
func DefaultQueuePriorities() map[string]int {
	return map[string]int{
		QueueCritical: 6,
		QueueDefault:  3,
		QueueLow:      1,
	}
}

// SMSPayload carries transactional SMS details. Never log body or full recipient phone.
type SMSPayload struct {
	To      string                   `json:"to"`
	Body    string                   `json:"body"`
	Purpose notifications.SMSPurpose `json:"purpose"`
}

// STKPollPayload carries correlation details for polling delayed Daraja STK callbacks.
type STKPollPayload struct {
	SchoolID          string    `json:"school_id"`
	STKRequestID      string    `json:"stk_request_id"`
	CheckoutRequestID string    `json:"checkout_request_id,omitempty"`
	EnqueuedAt        time.Time `json:"enqueued_at"`
}

// PurgeUnverifiedPayload carries configuration for purging unverified onboarding submissions.
type PurgeUnverifiedPayload struct {
	RetentionDays int `json:"retention_days"`
}

// NewTransactionalSMSTask creates an Asynq task for delivering transactional SMS.
func NewTransactionalSMSTask(msg notifications.SMSMessage, opts ...asynq.Option) (*asynq.Task, error) {
	norm, err := notifications.NormalizePhone(msg.To)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", notifications.ErrInvalidPhone, err)
	}
	if strings.TrimSpace(msg.Body) == "" {
		return nil, errors.New("sms body is required")
	}
	switch msg.Purpose {
	case notifications.AppLoginOTP, notifications.WhatsAppSessionOTP, notifications.PhoneChangeOTP, notifications.PhoneChangeNotice:
	default:
		return nil, fmt.Errorf("%w: unrecognized purpose %q", notifications.ErrInvalidSMS, msg.Purpose)
	}

	payload, err := json.Marshal(SMSPayload{
		To:      norm,
		Body:    msg.Body,
		Purpose: msg.Purpose,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal sms payload: %w", err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(3),
		asynq.Timeout(20 * time.Second),
	}
	allOpts := append(defaultOpts, opts...)
	return asynq.NewTask(TypeTransactionalSMS, payload, allOpts...), nil
}

// NewSTKPollTask creates an Asynq task to poll Daraja for a pending STK push request.
func NewSTKPollTask(schoolID, stkRequestID, checkoutRequestID string, opts ...asynq.Option) (*asynq.Task, error) {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(stkRequestID) {
		return nil, contracts.Fail(http.StatusBadRequest, "INVALID_STK_TASK", "Valid school_id and stk_request_id are required.")
	}

	payload, err := json.Marshal(STKPollPayload{
		SchoolID:          schoolID,
		STKRequestID:      stkRequestID,
		CheckoutRequestID: checkoutRequestID,
		EnqueuedAt:        time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal stk poll payload: %w", err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(5),
		asynq.Timeout(30 * time.Second),
	}
	allOpts := append(defaultOpts, opts...)
	return asynq.NewTask(TypeSTKPoll, payload, allOpts...), nil
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

// Client wraps an *asynq.Client and implements billing.STKPollQueue.
type Client struct {
	client *asynq.Client
}

// NewClient constructs an Asynq queue producer client.
func NewClient(client *asynq.Client) *Client {
	return &Client{client: client}
}

// EnqueueSTKPoll satisfies billing.STKPollQueue. If delay <= 0, DefaultSTKPollDelay (45s) is used.
func (c *Client) EnqueueSTKPoll(ctx context.Context, task billing.STKPollTask, delay time.Duration) error {
	if c == nil || c.client == nil {
		return errors.New("asynq client is nil")
	}
	if delay <= 0 {
		delay = DefaultSTKPollDelay
	}

	asynqTask, err := NewSTKPollTask(task.SchoolID, task.STKRequestID, task.CheckoutRequestID, asynq.ProcessIn(delay))
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, asynqTask)
	return err
}

// EnqueueTransactionalSMS enqueues a transactional SMS for background delivery.
func (c *Client) EnqueueTransactionalSMS(ctx context.Context, msg notifications.SMSMessage, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("asynq client is nil")
	}
	task, err := NewTransactionalSMSTask(msg, opts...)
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
	SMSDispatcher notifications.SMSDispatcher
	STKPoller     STKPoller
	Purger        UnverifiedSubmissionPurger
	Logger        *slog.Logger
}

func (d WorkerDeps) logger() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}
	return slog.Default()
}

// HandleTransactionalSMS unmarshals and dispatches an SMS, marking permanent failures with asynq.SkipRetry.
func HandleTransactionalSMS(deps WorkerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload SMSPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("%w: malformed sms payload: %v", asynq.SkipRetry, err)
		}
		if deps.SMSDispatcher == nil {
			return fmt.Errorf("%w: sms dispatcher is nil", asynq.SkipRetry)
		}

		msg := notifications.SMSMessage{
			To:      payload.To,
			Body:    payload.Body,
			Purpose: payload.Purpose,
		}

		if err := deps.SMSDispatcher.DispatchSMS(ctx, msg); err != nil {
			if errors.Is(err, notifications.ErrInvalidPhone) || errors.Is(err, notifications.ErrInvalidSMS) {
				return fmt.Errorf("%w: permanent sms dispatch error: %v", asynq.SkipRetry, err)
			}
			return fmt.Errorf("dispatch sms: %w", err)
		}
		return nil
	}
}

// HandleSTKPoll unmarshals and executes a Daraja STK query check against pending STK requests.
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
			return fmt.Errorf("%w: stk poller dependency is nil", asynq.SkipRetry)
		}

		if err := deps.STKPoller.PollSTKRequest(ctx, payload.SchoolID, payload.STKRequestID); err != nil {
			var apiErr *contracts.APIError
			if errors.As(err, &apiErr) && apiErr.Status >= 400 && apiErr.Status < 500 {
				return fmt.Errorf("%w: permanent client error on stk poll: %v", asynq.SkipRetry, err)
			}
			return fmt.Errorf("poll stk request: %w", err)
		}
		return nil
	}
}

// HandlePurgeUnverified invokes onboarding cleanup for expired unverified registration submissions.
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
			deps.logger().Warn("onboarding purger dependency not provided; skipping purge task",
				slog.Int("retention_days", payload.RetentionDays))
			return nil
		}

		purgedCount, err := deps.Purger.PurgeExpiredUnverifiedSubmissions(ctx, payload.RetentionDays)
		if err != nil {
			return fmt.Errorf("purge unverified submissions: %w", err)
		}

		deps.logger().Info("purged expired unverified submissions",
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

// DefaultServerConfig returns a configured asynq.Config with queue priorities and safe concurrency.
func DefaultServerConfig(logger *slog.Logger) asynq.Config {
	return asynq.Config{
		Concurrency:    10,
		Queues:         DefaultQueuePriorities(),
		StrictPriority: false,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			if logger != nil {
				logger.Error("asynq task failed",
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

// NewScheduler creates an Asynq periodic task scheduler.
func NewScheduler(opt asynq.RedisConnOpt, loc *time.Location) *asynq.Scheduler {
	if loc == nil {
		loc = time.UTC
	}
	return asynq.NewScheduler(opt, &asynq.SchedulerOpts{Location: loc})
}

// RegisterSchedulerTasks registers periodic cron jobs (e.g. daily purge of unverified submissions).
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
