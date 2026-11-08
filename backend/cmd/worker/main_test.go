package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/Ashomondi/ikonex-academy/backend/internal/onboarding"
	"github.com/hibiken/asynq"
)

type testPurger struct{ days int }

func (p *testPurger) PurgeExpiredUnverifiedSubmissions(_ context.Context, days int) (int, error) {
	p.days = days
	return 1, nil
}

func TestWorkerRejectsUnavailableDependencies(t *testing.T) {
	if _, err := workerMux(nil, &onboarding.Service{}); err == nil {
		t.Fatal("missing database accepted")
	}
	if _, err := workerMux(&sql.DB{}, &onboarding.Service{}); err == nil || !strings.Contains(err.Error(), "PurgeExpiredUnverifiedSubmissions") {
		t.Fatalf("missing cleanup implementation: %v", err)
	}
}

func TestWorkerRegistersSMSBillingAndPurge(t *testing.T) {
	t.Setenv("AFRICASTALKING_USERNAME", "sandbox")
	purger := &testPurger{}
	mux, err := workerMux(&sql.DB{}, purger)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(jobs.SMSTaskPayload{ToPhoneNumber: "+254700000001", Message: "Synthetic verification message"})
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeSMSTransactional, body)); err != nil {
		t.Fatal(err)
	}
	body, _ = json.Marshal(jobs.PurgeUnverifiedTaskPayload{RetentionDays: 7})
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeOnboardingPurgeUnverified, body)); err != nil || purger.days != 7 {
		t.Fatalf("purge task: %v %d", err, purger.days)
	}
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeMPesaSTKPoll, []byte("invalid"))); err == nil || !strings.Contains(err.Error(), "invalid stk poll payload") {
		t.Fatalf("billing handler not registered: %v", err)
	}
}
