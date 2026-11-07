package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
)

func main() {
	log.Println("Starting Ikonex Academy Background Worker...")

	qm := jobs.NewQueueManager(500)

	// Register Task Handlers
	qm.RegisterHandler(jobs.TypeSMSTransactional, func(ctx context.Context, payload []byte) error {
		log.Printf("[Worker] Executing %s", jobs.TypeSMSTransactional)
		return nil
	})

	qm.RegisterHandler(jobs.TypeMPesaSTKPoll, func(ctx context.Context, payload []byte) error {
		log.Printf("[Worker] Executing %s", jobs.TypeMPesaSTKPoll)
		return nil
	})

	qm.RegisterHandler(jobs.TypeOnboardingPurgeUnverified, func(ctx context.Context, payload []byte) error {
		log.Printf("[Worker] Executing %s", jobs.TypeOnboardingPurgeUnverified)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	qm.StartWorker(ctx)
	log.Println("Worker ready. Processing jobs queue...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	qm.Stop()
	log.Println("Worker stopped gracefully.")
}
