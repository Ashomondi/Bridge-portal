package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

// Task Types
const (
	TypeSMSTransactional          = "sms:transactional"
	TypeMPesaSTKPoll              = "mpesa:stk_poll"
	TypeOnboardingPurgeUnverified = "onboarding:purge_unverified"
)

// Task Payloads
type SMSTaskPayload struct {
	ToPhoneNumber string `json:"to_phone_number"`
	Message       string `json:"message"`
}

type STKPollTaskPayload struct {
	SchoolID          string `json:"school_id"`
	CheckoutRequestID string `json:"checkout_request_id"`
}

type PurgeUnverifiedTaskPayload struct {
	RetentionDays int `json:"retention_days"`
}

// Task represents an enqueued background job
type Task struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	ExecuteAt time.Time       `json:"execute_at"`
	Attempts  int             `json:"attempts"`
}

// SMSDispatcher interface for sending transactional SMS
type SMSDispatcher interface {
	SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error
}

// BillingQuerier interface for polling M-Pesa STK push status
type BillingQuerier interface {
	QuerySTKPushStatus(ctx context.Context, schoolID, checkoutRequestID string) error
}

// OnboardingPurger interface for purging unverified submissions
type OnboardingPurger interface {
	PurgeExpiredUnverifiedSubmissions(ctx context.Context, retentionDays int) (int, error)
}

// QueueManager handles enqueuing and processing background tasks (supporting both Asynq/Redis and in-memory execution)
type QueueManager struct {
	asynqClient *asynq.Client
	tasks       chan *Task
	handlers    map[string]func(ctx context.Context, payload []byte) error
	mu          sync.RWMutex
	stopSignal  chan struct{}
	stopped     bool
	running     bool
}

// NewQueueManager creates an in-memory QueueManager for testing or standalone execution
func NewQueueManager(bufferSize int) *QueueManager {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &QueueManager{
		tasks:      make(chan *Task, bufferSize),
		handlers:   make(map[string]func(ctx context.Context, payload []byte) error),
		stopSignal: make(chan struct{}),
	}
}

// NewAsynqQueueManager creates a QueueManager backed by an Asynq Redis client
func NewAsynqQueueManager(client *asynq.Client) *QueueManager {
	qm := NewQueueManager(100)
	qm.asynqClient = client
	return qm
}

// RegisterHandler registers a processor for a task type
func (qm *QueueManager) RegisterHandler(taskType string, handler func(ctx context.Context, payload []byte) error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.handlers[taskType] = handler
}

// Enqueue queues a task with execution delay (to Redis/Asynq if configured, otherwise in-memory)
func (qm *QueueManager) Enqueue(taskType string, payload any, delay time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if qm.asynqClient != nil {
		opts := []asynq.Option{
			asynq.MaxRetry(3),
		}
		if delay > 0 {
			opts = append(opts, asynq.ProcessIn(delay))
		}
		task := asynq.NewTask(taskType, data, opts...)
		_, err := qm.asynqClient.Enqueue(task)
		return err
	}

	task := &Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Type:      taskType,
		Payload:   data,
		ExecuteAt: time.Now().Add(delay),
	}
	select {
	case qm.tasks <- task:
		return nil
	default:
		return fmt.Errorf("task queue buffer full")
	}
}

// EnqueueSTKPoll fulfills billing.AsynqEnqueuer
func (qm *QueueManager) EnqueueSTKPoll(ctx context.Context, schoolID, checkoutRequestID string, delay time.Duration) error {
	return qm.Enqueue(TypeMPesaSTKPoll, STKPollTaskPayload{
		SchoolID:          schoolID,
		CheckoutRequestID: checkoutRequestID,
	}, delay)
}

// EnqueueSMS enqueues async transactional SMS
func (qm *QueueManager) EnqueueSMS(ctx context.Context, to, message string) error {
	return qm.Enqueue(TypeSMSTransactional, SMSTaskPayload{
		ToPhoneNumber: to,
		Message:       message,
	}, 0)
}

// EnqueuePurgeUnverified enqueues periodic purge of unverified onboarding submissions
func (qm *QueueManager) EnqueuePurgeUnverified(ctx context.Context, retentionDays int, delay time.Duration) error {
	return qm.Enqueue(TypeOnboardingPurgeUnverified, PurgeUnverifiedTaskPayload{
		RetentionDays: retentionDays,
	}, delay)
}

// StartWorker starts the in-memory background task processing loop
func (qm *QueueManager) StartWorker(ctx context.Context) {
	qm.mu.Lock()
	if qm.running {
		qm.mu.Unlock()
		return
	}
	if qm.stopped || qm.stopSignal == nil {
		qm.stopSignal = make(chan struct{})
		qm.stopped = false
	}
	stopCh := qm.stopSignal
	qm.running = true
	qm.mu.Unlock()

	go func(stop <-chan struct{}) {
		defer func() {
			qm.mu.Lock()
			if qm.stopSignal == stop {
				qm.running = false
			}
			qm.mu.Unlock()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case task := <-qm.tasks:
				if wait := time.Until(task.ExecuteAt); wait > 0 {
					select {
					case <-ctx.Done():
						return
					case <-stop:
						return
					case <-time.After(wait):
					}
				}
				qm.mu.RLock()
				handler, ok := qm.handlers[task.Type]
				qm.mu.RUnlock()

				if ok && handler != nil {
					if err := handler(ctx, task.Payload); err != nil {
						log.Printf("[Worker Error] Task %s failed: %v", task.Type, err)
					}
				}
			}
		}
	}(stopCh)
}

// Stop cleanly and idempotently stops the in-memory worker
func (qm *QueueManager) Stop() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	if !qm.stopped && qm.stopSignal != nil {
		close(qm.stopSignal)
		qm.stopped = true
		qm.running = false
	}
}

// NewAsynqServer creates an Asynq server configured with concurrency, queues, and retry policies
func NewAsynqServer(redisOpt asynq.RedisClientOpt, concurrency int) *asynq.Server {
	if concurrency <= 0 {
		concurrency = 10
	}
	return asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[Asynq Error] task %s failed: %v", task.Type(), err)
			}),
		},
	)
}

// RegisterAsynqHandlers configures an Asynq ServeMux with real task processors
func RegisterAsynqHandlers(mux *asynq.ServeMux, sms SMSDispatcher, bQuerier BillingQuerier, purger OnboardingPurger) error {
	if mux == nil || sms == nil || bQuerier == nil || purger == nil {
		return fmt.Errorf("Asynq handler mux, SMS dispatcher, STK status querier, and onboarding purger are required")
	}

	mux.HandleFunc(TypeSMSTransactional, func(ctx context.Context, t *asynq.Task) error {
		var p SMSTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid sms payload: %w", err)
		}
		return sms.SendTransactionalSMS(ctx, p.ToPhoneNumber, p.Message)
	})

	mux.HandleFunc(TypeMPesaSTKPoll, func(ctx context.Context, t *asynq.Task) error {
		var p STKPollTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid stk poll payload: %w", err)
		}
		return bQuerier.QuerySTKPushStatus(ctx, p.SchoolID, p.CheckoutRequestID)
	})

	mux.HandleFunc(TypeOnboardingPurgeUnverified, func(ctx context.Context, t *asynq.Task) error {
		var p PurgeUnverifiedTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid purge payload: %w", err)
		}
		_, err := purger.PurgeExpiredUnverifiedSubmissions(ctx, p.RetentionDays)
		return err
	})
	return nil
}
