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
