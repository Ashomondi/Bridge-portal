package main

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
)

func mapEnv(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

func TestLoadWorkerConfigValidation(t *testing.T) {
	base := map[string]string{
		"DATABASE_URL":            "postgres://user:pass@localhost:5432/ikonex?sslmode=disable",
		"REDIS_URL":               "redis://localhost:6379/0",
		"AFRICASTALKING_USERNAME": "sandbox",
	}

	cfg, err := LoadWorkerConfig(mapEnv(base))
	if err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if cfg.Concurrency != 10 || cfg.ShutdownTimeout != 15*time.Second || cfg.PurgeRetentionDays != 7 || cfg.EnablePurgeScheduler {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	cases := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{
			name:    "missing database",
			mutate:  func(m map[string]string) { delete(m, "DATABASE_URL") },
			wantErr: "DATABASE_URL",
		},
		{
			name:    "missing redis",
			mutate:  func(m map[string]string) { delete(m, "REDIS_URL") },
			wantErr: "REDIS_URL",
		},
		{
			name:    "malformed redis url",
			mutate:  func(m map[string]string) { m["REDIS_URL"] = "://not-a-valid-uri" },
			wantErr: "invalid REDIS_URL",
		},
		{
			name:    "invalid concurrency",
			mutate:  func(m map[string]string) { m["WORKER_CONCURRENCY"] = "0" },
			wantErr: "WORKER_CONCURRENCY",
		},
		{
			name:    "invalid shutdown timeout",
			mutate:  func(m map[string]string) { m["WORKER_SHUTDOWN_TIMEOUT"] = "not-a-duration" },
			wantErr: "WORKER_SHUTDOWN_TIMEOUT",
		},
		{
			name:    "invalid retention days",
			mutate:  func(m map[string]string) { m["PURGE_RETENTION_DAYS"] = "-3" },
			wantErr: "PURGE_RETENTION_DAYS",
		},
		{
			name:    "invalid enable purge bool",
			mutate:  func(m map[string]string) { m["ENABLE_PURGE_SCHEDULER"] = "maybe" },
			wantErr: "ENABLE_PURGE_SCHEDULER",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := make(map[string]string, len(base))
			for k, v := range base {
				env[k] = v
			}
			tc.mutate(env)
			_, err := LoadWorkerConfig(mapEnv(env))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestRunWorkerFailsFastOnUnreachableRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := RunWorker(ctx, WorkerConfig{
		DatabaseURL: "postgres://user:pass@127.0.0.1:1/ikonex?sslmode=disable",
		RedisAddr:   "127.0.0.1:1",
		Concurrency: 2,
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "redis ping failed") {
		t.Fatalf("expected fast redis ping failure, got %v", err)
	}
}

func requireTestRedisAndDB(t *testing.T) (string, string) {
	t.Helper()
	redisAddr := strings.TrimSpace(os.Getenv("DEV05_TEST_REDIS_ADDR"))
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	conn, err := net.DialTimeout("tcp", redisAddr, 500*time.Millisecond)
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("Redis is required in CI at %s: %v", redisAddr, err)
		}
		t.Skipf("skipping live worker lifecycle test; Redis unavailable at %s", redisAddr)
	}
	_ = conn.Close()

	dbURL := strings.TrimSpace(os.Getenv("DEV05_TEST_DATABASE_URL"))
	if dbURL == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("DEV05_TEST_DATABASE_URL is required in CI")
		}
		t.Skip("skipping live worker lifecycle test; DEV05_TEST_DATABASE_URL not set")
	}
	return redisAddr, dbURL
}

func TestRunWorkerLifecycleAndValidation(t *testing.T) {
	redisAddr, dbURL := requireTestRedisAndDB(t)

	// 1. Reachable Redis + unreachable Postgres fails fast.
	ctxPG, cancelPG := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelPG()
	err := RunWorker(ctxPG, WorkerConfig{
		DatabaseURL: "postgres://user:pass@127.0.0.1:1/ikonex?sslmode=disable",
		RedisAddr:   redisAddr,
		Concurrency: 2,
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "postgres ping failed") {
		t.Fatalf("expected fast postgres ping failure, got %v", err)
	}

	// 2. ENABLE_PURGE_SCHEDULER=true with nil purger fails fast.
	ctxPurge, cancelPurge := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPurge()
	err = RunWorker(ctxPurge, WorkerConfig{
		DatabaseURL:          dbURL,
		RedisAddr:            redisAddr,
		SMSConfig:            notifications.SMSConfig{Username: "sandbox"},
		Concurrency:          2,
		ShutdownTimeout:      2 * time.Second,
		EnablePurgeScheduler: true,
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "ENABLE_PURGE_SCHEDULER is true") {
		t.Fatalf("expected purge scheduler guard error, got %v", err)
	}

	// 3. Valid live startup and clean graceful shutdown on context cancel.
	ctxRun, cancelRun := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunWorker(ctxRun, WorkerConfig{
			DatabaseURL:          dbURL,
			RedisAddr:            redisAddr,
			SMSConfig:            notifications.SMSConfig{Username: "sandbox"},
			Concurrency:          2,
			ShutdownTimeout:      3 * time.Second,
			EnablePurgeScheduler: false,
		}, nil, nil)
	}()

	time.Sleep(250 * time.Millisecond)
	cancelRun()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected clean shutdown without error, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for RunWorker graceful shutdown")
	}
}
