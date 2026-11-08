package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Ikonex Academy Background Worker...")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	}

	server := jobs.NewAsynqServer(redisOpt, 10)
	mux := asynq.NewServeMux()

	// Register Asynq task handlers
	jobs.RegisterAsynqHandlers(mux, nil, nil, nil)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down worker gracefully...")
		server.Shutdown()
	}()

	log.Printf("Worker starting on Redis %s. Processing queues (critical, default, low)...", redisAddr)
	if err := server.Run(mux); err != nil {
		log.Printf("Asynq server exited or could not connect: %v. Running in-memory queue manager fallback...", err)
		qm := jobs.NewQueueManager(500)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		qm.StartWorker(ctx)
		<-sigChan
		qm.Stop()
	}
	log.Println("Worker stopped.")
}
