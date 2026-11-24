package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// WorkerConfig holds validated runtime settings for the background worker process.
type WorkerConfig struct {
	DatabaseURL          string
	RedisURL             string
	RedisAddr            string
	RedisPassword        string
	SMSConfig            notifications.SMSConfig
	DarajaBaseURL        string
	DarajaConsumerKey    string
	DarajaConsumerSecret string
	DarajaCallbackURL    string
	Concurrency          int
	ShutdownTimeout      time.Duration
	PurgeRetentionDays   int
	PurgeCron            string
	EnablePurgeScheduler bool
}

// LoadWorkerConfig reads and strictly validates worker configuration from env.
func LoadWorkerConfig(getenv func(string) string) (WorkerConfig, error) {
	if getenv == nil {
		getenv = os.Getenv
	}

	dbURL := strings.TrimSpace(getenv("DATABASE_URL"))
	if dbURL == "" {
		dbHost := strings.TrimSpace(getenv("DB_HOST"))
		dbName := strings.TrimSpace(getenv("DB_NAME"))
		dbUser := strings.TrimSpace(getenv("DB_USER"))
		dbPass := getenv("DB_PASSWORD")
		dbPort := strings.TrimSpace(getenv("DB_PORT"))
		if dbPort == "" {
			dbPort = "5432"
		}
		sslMode := strings.TrimSpace(getenv("DB_SSLMODE"))
		if sslMode == "" {
			sslMode = "disable"
		}
		if dbHost != "" && dbName != "" && dbUser != "" {
			u := url.URL{
				Scheme:   "postgres",
				User:     url.UserPassword(dbUser, dbPass),
				Host:     net.JoinHostPort(dbHost, dbPort),
				Path:     dbName,
				RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
			}
			dbURL = u.String()
		}
	}
	if dbURL == "" {
		return WorkerConfig{}, errors.New("worker config: DATABASE_URL (or DB_HOST/DB_USER/DB_NAME) is required")
	}

	redisURL := strings.TrimSpace(getenv("REDIS_URL"))
	redisAddr := strings.TrimSpace(getenv("REDIS_ADDR"))
	redisPassword := getenv("REDIS_PASSWORD")
	if redisURL == "" && redisAddr == "" {
		if host := strings.TrimSpace(getenv("REDIS_HOST")); host != "" {
			port := strings.TrimSpace(getenv("REDIS_PORT"))
			if port == "" {
				port = "6379"
			}
			redisAddr = net.JoinHostPort(host, port)
		}
	}
	if redisURL == "" && redisAddr == "" {
		return WorkerConfig{}, errors.New("worker config: REDIS_URL (or REDIS_ADDR / REDIS_HOST) is required")
	}
	if _, err := parseRedisConnOpt(redisURL, redisAddr, redisPassword); err != nil {
		return WorkerConfig{}, err
	}

	concurrency := 10
	if raw := strings.TrimSpace(getenv("WORKER_CONCURRENCY")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return WorkerConfig{}, fmt.Errorf("worker config: WORKER_CONCURRENCY must be a positive integer, got %q", raw)
		}
		concurrency = n
	}

	shutdownTimeout := 15 * time.Second
	if raw := strings.TrimSpace(getenv("WORKER_SHUTDOWN_TIMEOUT")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return WorkerConfig{}, fmt.Errorf("worker config: WORKER_SHUTDOWN_TIMEOUT must be a positive duration, got %q", raw)
		}
		shutdownTimeout = d
	}

	retentionDays := jobs.DefaultPurgeRetentionDays
	if raw := strings.TrimSpace(getenv("PURGE_RETENTION_DAYS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return WorkerConfig{}, fmt.Errorf("worker config: PURGE_RETENTION_DAYS must be a positive integer, got %q", raw)
		}
		retentionDays = n
	}

	purgeCron := strings.TrimSpace(getenv("PURGE_CRON"))
	if purgeCron == "" {
		purgeCron = jobs.DefaultPurgeCronExpression
	}

	enablePurge := false
	if raw := strings.TrimSpace(getenv("ENABLE_PURGE_SCHEDULER")); raw != "" {
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return WorkerConfig{}, fmt.Errorf("worker config: ENABLE_PURGE_SCHEDULER must be a boolean, got %q", raw)
		}
		enablePurge = b
	}

	return WorkerConfig{
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		SMSConfig: notifications.SMSConfig{
			Username: strings.TrimSpace(getenv("AFRICASTALKING_USERNAME")),
			APIKey:   strings.TrimSpace(getenv("AFRICASTALKING_API_KEY")),
			SenderID: strings.TrimSpace(getenv("AFRICASTALKING_SENDER_ID")),
		},
		DarajaBaseURL:        strings.TrimSpace(getenv("DARAJA_BASE_URL")),
		DarajaConsumerKey:    strings.TrimSpace(getenv("DARAJA_CONSUMER_KEY")),
		DarajaConsumerSecret: strings.TrimSpace(getenv("DARAJA_CONSUMER_SECRET")),
		DarajaCallbackURL:    strings.TrimSpace(getenv("DARAJA_CALLBACK_URL")),
		Concurrency:          concurrency,
		ShutdownTimeout:      shutdownTimeout,
		PurgeRetentionDays:   retentionDays,
		PurgeCron:            purgeCron,
		EnablePurgeScheduler: enablePurge,
	}, nil
}

func parseRedisConnOpt(redisURL, redisAddr, redisPassword string) (asynq.RedisConnOpt, error) {
	if redisURL != "" {
		opt, err := asynq.ParseRedisURI(redisURL)
		if err != nil {
			return nil, fmt.Errorf("worker config: invalid REDIS_URL: %w", err)
		}
		return opt, nil
	}
	if redisAddr == "" {
		return nil, errors.New("worker config: redis address is required")
	}
	return asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword}, nil
}

func pingRedis(ctx context.Context, opt asynq.RedisConnOpt) error {
	raw := opt.MakeRedisClient()
	client, ok := raw.(redis.UniversalClient)
	if !ok {
		return errors.New("worker: unexpected redis client type")
	}
	defer func() { _ = client.Close() }()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("worker: redis ping failed: %w", err)
	}
	return nil
}

// RunWorker initializes all dependencies, verifies PostgreSQL and Redis
// connectivity, runs the Asynq server (and scheduler when enabled), and shuts
// down cleanly on context cancellation. Any startup or runtime error is returned.
func RunWorker(ctx context.Context, cfg WorkerConfig, purger jobs.UnverifiedSubmissionPurger, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	redisOpt, err := parseRedisConnOpt(cfg.RedisURL, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		return err
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()

	if err := pingRedis(pingCtx, redisOpt); err != nil {
		return err
	}

	pool, err := pgxpool.New(pingCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("worker: create postgres pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("worker: postgres ping failed: %w", err)
	}

	tenantAdapter, err := jobs.NewTenantDBAdapter(pool)
	if err != nil {
		return err
	}
	tenantDB := tenantAdapter.AsContractsTenantDB()

	smsClient, err := notifications.NewSMSDispatcher(cfg.SMSConfig, &http.Client{Timeout: 10 * time.Second}, logger)
	if err != nil {
		return fmt.Errorf("worker: init sms dispatcher: %w", err)
	}

	if cfg.EnablePurgeScheduler && purger == nil {
		return errors.New("worker: ENABLE_PURGE_SCHEDULER is true but no onboarding purger is wired")
	}

	asynqClient := asynq.NewClient(redisOpt)
	defer func() { _ = asynqClient.Close() }()
	queueProducer := jobs.NewClient(asynqClient)

	darajaClient := &billing.HTTPDarajaClient{
		BaseURL:        cfg.DarajaBaseURL,
		ConsumerKey:    cfg.DarajaConsumerKey,
		ConsumerSecret: cfg.DarajaConsumerSecret,
		CallbackURL:    cfg.DarajaCallbackURL,
		HTTPClient:     &http.Client{Timeout: 15 * time.Second},
		Now:            time.Now,
	}

	billingSvc := &billing.Service{
		DB:        tenantDB,
		Audit:     billing.NewPostgresAuditLogger(),
		Notifier:  &notifications.Service{DB: tenantDB},
		Daraja:    darajaClient,
		PollQueue: queueProducer,
		Now:       time.Now,
	}

	workerDeps := jobs.WorkerDeps{
		SMSDispatcher:  smsClient,
		STKPoller:      billingSvc,
		STKRescheduler: queueProducer,
		TenantDB:       tenantAdapter,
		Purger:         purger,
		Logger:         logger,
	}

	mux := asynq.NewServeMux()
	jobs.RegisterHandlers(mux, workerDeps)

	serverCfg := jobs.ServerConfigWithOptions(logger, cfg.Concurrency, cfg.ShutdownTimeout)
	server := jobs.NewServer(redisOpt, serverCfg)

	var scheduler *asynq.Scheduler
	if cfg.EnablePurgeScheduler {
		scheduler = jobs.NewScheduler(redisOpt, time.UTC)
		entryID, err := jobs.RegisterSchedulerTasks(scheduler, cfg.PurgeRetentionDays, cfg.PurgeCron)
		if err != nil {
			return fmt.Errorf("worker: register purge scheduler: %w", err)
		}
		logger.Info("registered unverified submissions purge schedule",
			slog.String("cron", cfg.PurgeCron),
			slog.String("timezone", "UTC"),
			slog.Int("retention_days", cfg.PurgeRetentionDays),
			slog.String("entry_id", entryID))
	} else {
		logger.Warn("unverified submissions purge scheduler disabled (ENABLE_PURGE_SCHEDULER=false or onboarding purger not wired)")
	}

	logger.Info("starting asynq worker server", slog.Int("concurrency", cfg.Concurrency))
	if err := server.Start(mux); err != nil {
		return fmt.Errorf("worker: start asynq server: %w", err)
	}

	if scheduler != nil {
		logger.Info("starting asynq scheduler")
		if err := scheduler.Start(); err != nil {
			server.Shutdown()
			return fmt.Errorf("worker: start asynq scheduler: %w", err)
		}
	}

	<-ctx.Done()
	logger.Info("shutdown signal received; draining in-flight tasks")

	if scheduler != nil {
		scheduler.Shutdown()
	}
	server.Shutdown()
	logger.Info("worker stopped cleanly")
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := LoadWorkerConfig(os.Getenv)
	if err != nil {
		logger.Error("worker configuration error", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := RunWorker(ctx, cfg, nil, logger); err != nil {
		logger.Error("worker exited with error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
