package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
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

// QueueManager handles enqueuing and processing Asynq/Redis background tasks
type QueueManager struct {
	tasks      chan *Task
	handlers   map[string]func(ctx context.Context, payload []byte) error
	mu         sync.RWMutex
	stopSignal chan struct{}
	stopped    bool
	running    bool
}

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

// RegisterHandler registers a processor for a task type
func (qm *QueueManager) RegisterHandler(taskType string, handler func(ctx context.Context, payload []byte) error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.handlers[taskType] = handler
}

// Enqueue queues a task with execution delay
func (qm *QueueManager) Enqueue(taskType string, payload any, delay time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
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

// StartWorker starts the background task loop
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
				// If task is scheduled in future, wait until then or shutdown
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

// Stop cleanly and idempotently stops the worker
func (qm *QueueManager) Stop() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	if !qm.stopped && qm.stopSignal != nil {
		close(qm.stopSignal)
		qm.stopped = true
		qm.running = false
	}
}
