package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tenantDBAdapter struct {
	pool *pgxpool.Pool
}

func (a *tenantDBAdapter) WithTenantTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
	if a == nil || a.pool == nil {
		return errors.New("db pool is nil")
	}
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func parseRedisConnOpt(redisURL, redisAddr string) asynq.RedisConnOpt {
	if redisURL != "" {
		if opt, err := asynq.ParseRedisURI(redisURL); err == nil {
			return opt
		}
	}
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	return asynq.RedisClientOpt{Addr: redisAddr}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting Ikonex Academy background worker")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	redisURL := os.Getenv("REDIS_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisOpt := parseRedisConnOpt(redisURL, redisAddr)

	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()
	queueProducer := jobs.NewClient(asynqClient)

	var tenantRunner contracts.TenantDB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			logger.Error("failed to connect to database", slog.String("err", err.Error()))
		} else {
			defer pool.Close()
			tenantRunner = &tenantDBAdapter{pool: pool}
			logger.Info("connected to PostgreSQL for tenant transactions")
		}
	}

	atUsername := os.Getenv("AFRICASTALKING_USERNAME")
	atAPIKey := os.Getenv("AFRICASTALKING_API_KEY")
	atSenderID := os.Getenv("AFRICASTALKING_SENDER_ID")
	smsClient, err := notifications.NewSMSDispatcher(notifications.SMSConfig{
		Username: atUsername,
		APIKey:   atAPIKey,
		SenderID: atSenderID,
	}, &http.Client{Timeout: 10 * time.Second}, logger)
	if err != nil {
		logger.Warn("could not initialize SMS dispatcher, using nil dispatcher", slog.String("err", err.Error()))
	}

	darajaClient := &billing.HTTPDarajaClient{
		BaseURL:        os.Getenv("DARAJA_BASE_URL"),
		ConsumerKey:    os.Getenv("DARAJA_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("DARAJA_CONSUMER_SECRET"),
		CallbackURL:    os.Getenv("DARAJA_CALLBACK_URL"),
		HTTPClient:     &http.Client{Timeout: 15 * time.Second},
		Now:            time.Now,
	}

	billingSvc := &billing.Service{
		DB:        tenantRunner,
		Daraja:    darajaClient,
		PollQueue: queueProducer,
		Now:       time.Now,
	}

	retentionDays := jobs.DefaultPurgeRetentionDays
	if rdStr := os.Getenv("PURGE_RETENTION_DAYS"); rdStr != "" {
		if parsed, err := strconv.Atoi(rdStr); err == nil && parsed > 0 {
			retentionDays = parsed
		}
	}

	purgeCron := os.Getenv("PURGE_CRON")
	if purgeCron == "" {
		purgeCron = jobs.DefaultPurgeCronExpression
	}

	workerDeps := jobs.WorkerDeps{
		SMSDispatcher: smsClient,
		STKPoller:     billingSvc,
		Logger:        logger,
	}

	mux := asynq.NewServeMux()
	jobs.RegisterHandlers(mux, workerDeps)

	serverCfg := jobs.DefaultServerConfig(logger)
	server := jobs.NewServer(redisOpt, serverCfg)

	scheduler := jobs.NewScheduler(redisOpt, time.UTC)
	if entryID, err := jobs.RegisterSchedulerTasks(scheduler, retentionDays, purgeCron); err != nil {
		logger.Error("failed to register scheduler tasks", slog.String("err", err.Error()))
	} else {
		logger.Info("registered unverified submissions purge schedule",
			slog.String("cron", purgeCron),
			slog.Int("retention_days", retentionDays),
			slog.String("entry_id", entryID))
	}

	errChan := make(chan error, 2)
	go func() {
		logger.Info("starting asynq worker server")
		if err := server.Run(mux); err != nil && !errors.Is(err, asynq.ErrServerClosed) {
			errChan <- err
		}
	}()

	go func() {
		logger.Info("starting asynq scheduler")
		if err := scheduler.Run(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received; terminating gracefully")
	case err := <-errChan:
		logger.Error("worker fatal error", slog.String("err", err.Error()))
	}

	scheduler.Shutdown()
	server.Shutdown()
	logger.Info("worker stopped cleanly")
}
