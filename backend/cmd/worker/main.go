package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/Ashomondi/ikonex-academy/backend/internal/onboarding"
	"github.com/hibiken/asynq"
)

func main() {
	log.Println("Starting Ikonex Academy Background Worker...")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		host, port := os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "6379"
		}
		redisAddr = net.JoinHostPort(host, port)
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	}

	server := jobs.NewAsynqServer(redisOpt, 10)
	db, err := openWorkerDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mux, err := workerMux(db, &onboarding.Service{})
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down worker gracefully...")
		server.Shutdown()
	}()

	log.Printf("Worker starting on Redis %s. Processing queues (critical, default, low)...", redisAddr)
	if err := server.Run(mux); err != nil {
		log.Fatalf("Asynq worker failed: %v", err)
	}
	log.Println("Worker stopped.")
}

func openWorkerDatabase() (*sql.DB, error) {
	driver := ""
	for _, registered := range sql.Drivers() {
		if registered == "pgx" || registered == "postgres" {
			driver = registered
			break
		}
	}
	if driver == "" {
		return nil, fmt.Errorf("worker: PostgreSQL driver is not registered; database dependency wiring is required")
	}
	host, port := os.Getenv("DB_HOST"), os.Getenv("DB_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	dsn := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port), Path: os.Getenv("DB_NAME"), User: url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))}
	dsn.RawQuery = url.Values{"sslmode": {os.Getenv("DB_SSLMODE")}}.Encode()
	db, err := sql.Open(driver, dsn.String())
	if err != nil {
		return nil, fmt.Errorf("worker: unable to open PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("worker: PostgreSQL is unavailable")
	}
	return db, nil
}

func workerMux(db *sql.DB, onboardingService any) (*asynq.ServeMux, error) {
	if db == nil {
		return nil, fmt.Errorf("worker: PostgreSQL connection is required")
	}
	purger, ok := onboardingService.(jobs.OnboardingPurger)
	if !ok || purger == nil {
		return nil, fmt.Errorf("worker: onboarding.Service must implement PurgeExpiredUnverifiedSubmissions before startup")
	}
	sms := notifications.NewAfricasTalkingSMSClient(os.Getenv("AFRICASTALKING_USERNAME"), os.Getenv("AFRICASTALKING_API_KEY"), os.Getenv("AFRICASTALKING_SENDER_ID"))
	notifier := notifications.NewService(notifications.NewSQLStore(db), sms)
	billingService := billing.NewService(billing.NewSQLStore(db), notifier, nil, nil)
	mux := asynq.NewServeMux()
	if err := jobs.RegisterAsynqHandlers(mux, sms, billingService, purger); err != nil {
		return nil, err
	}
	return mux, nil
}
