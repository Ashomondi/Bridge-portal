package jobs

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
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
