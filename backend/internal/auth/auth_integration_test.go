package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
)

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the auth integration tests")
	}
	db, err := database.Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

var testIPCounter atomic.Uint32

func uniquePhone() string {
	return fmt.Sprintf("+2547%08d", time.Now().UnixNano()%100000000)
}

func uniqueIP() string {
	return fmt.Sprintf("198.51.100.%d", testIPCounter.Add(1)%250+1)
}

func newTestService(db *database.DB) *Service {
	return NewService(db.DB, NewTokenManager(strings.Repeat("s", 32), time.Hour), nil, nil, false)
}

func cleanupPhone(t *testing.T, db *database.DB, phone string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM otp_requests WHERE phone_number = $1`, phone)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM sender_rate_blocks WHERE sender_identifier = $1`, phone)
	})
}

func cleanupIP(t *testing.T, db *database.DB, ip string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM otp_requests WHERE request_ip = $1`, ip)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM sender_rate_blocks WHERE sender_identifier = $1`, "ip:"+ip)
	})
}

func TestIntegrationOTPAttemptLimit(t *testing.T) {
	db := openTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phone := uniquePhone()
	cleanupPhone(t, db, phone)

	var requestID string
	err := db.QueryRowContext(ctx,
		`INSERT INTO otp_requests (phone_number, sender_identifier, otp_hash, purpose, expires_at)
		 VALUES ($1, $1, $2, 'app_login', NOW() + INTERVAL '5 minutes')
		 RETURNING id::text`,
		phone, HashOTP("123456")).Scan(&requestID)
	if err != nil {
		t.Fatalf("seed otp request: %v", err)
	}

	service := newTestService(db)
	for attempt := 1; attempt <= maxOTPAttempts; attempt++ {
		rec := httptest.NewRecorder()
		body := fmt.Sprintf(`{"phone_number":%q,"otp":"000000"}`, phone)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/verify", strings.NewReader(body))
		service.handleVerifyOTP(rec, req)

		want := http.StatusUnauthorized
		if attempt == maxOTPAttempts {
			want = http.StatusTooManyRequests
		}
		if rec.Code != want {
			t.Fatalf("attempt %d status = %d, want %d (body: %s)", attempt, rec.Code, want, rec.Body.String())
		}
	}

	var consumed bool
	if err := db.QueryRowContext(ctx,
		`SELECT consumed_at IS NOT NULL FROM otp_requests WHERE id = $1`, requestID).Scan(&consumed); err != nil {
		t.Fatalf("read consumed_at: %v", err)
	}
	if !consumed {
		t.Fatal("OTP was not invalidated after reaching the attempt limit")
	}
}

func TestIntegrationOTPRequestCooldownIsAtomic(t *testing.T) {
	db := openTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phone := uniquePhone()
	ip := uniqueIP()
	cleanupPhone(t, db, phone)
	cleanupIP(t, db, ip)

	service := newTestService(db)
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.reserveOTP(ctx, phone, ip, HashOTP("123456"))
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	var successes, cooldowns int
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		var limit *rateLimitError
		if errors.As(err, &limit) && limit.Code == "OTP_COOLDOWN" {
			cooldowns++
			continue
		}
		t.Fatalf("unexpected reservation error: %v", err)
	}
	if successes != 1 || cooldowns != 1 {
		t.Fatalf("successes = %d, cooldowns = %d; want 1 and 1", successes, cooldowns)
	}
}

func TestIntegrationOTPRequestBlocksAfterThreshold(t *testing.T) {
	db := openTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phone := uniquePhone()
	ip := uniqueIP()
	cleanupPhone(t, db, phone)
	cleanupIP(t, db, ip)

	for i := 0; i < blockThreshold; i++ {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO otp_requests (phone_number, sender_identifier, otp_hash, purpose, request_ip, expires_at, created_at)
			 VALUES ($1, $1, $2, 'app_login', $3, NOW() + INTERVAL '5 minutes', NOW() - INTERVAL '2 minutes')`,
			phone, HashOTP("123456"), ip); err != nil {
			t.Fatalf("seed otp request %d: %v", i, err)
		}
	}

	service := newTestService(db)
	_, err := service.reserveOTP(ctx, phone, ip, HashOTP("654321"))
	var limit *rateLimitError
	if !errors.As(err, &limit) || limit.Code != "SENDER_BLOCKED" {
		t.Fatalf("reserveOTP error = %v, want SENDER_BLOCKED", err)
	}

	var blocked bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM sender_rate_blocks WHERE sender_identifier = $1 AND blocked_until > NOW())`,
		phone).Scan(&blocked); err != nil {
		t.Fatalf("read sender block: %v", err)
	}
	if !blocked {
		t.Fatal("sender was not blocked after the threshold")
	}
}

func TestIntegrationOTPRequestBlocksIPAfterThreshold(t *testing.T) {
	db := openTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ip := uniqueIP()
	targetPhone := uniquePhone()
	cleanupIP(t, db, ip)
	cleanupPhone(t, db, targetPhone)

	// Five requests from five different phone numbers on the same IP.
	for i := 0; i < blockThreshold; i++ {
		rotatedPhone := fmt.Sprintf("+2547%08d", 10000000+i)
		if _, err := db.ExecContext(ctx,
			`INSERT INTO otp_requests (phone_number, sender_identifier, otp_hash, purpose, request_ip, expires_at, created_at)
			 VALUES ($1, $1, $2, 'app_login', $3, NOW() + INTERVAL '5 minutes', NOW() - INTERVAL '2 minutes')`,
			rotatedPhone, HashOTP("123456"), ip); err != nil {
			t.Fatalf("seed otp request %d: %v", i, err)
		}
	}

	service := newTestService(db)
	_, err := service.reserveOTP(ctx, targetPhone, ip, HashOTP("654321"))
	var limit *rateLimitError
	if !errors.As(err, &limit) || limit.Code != "IP_BLOCKED" {
		t.Fatalf("reserveOTP error = %v, want IP_BLOCKED", err)
	}

	var blocked bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM sender_rate_blocks WHERE sender_identifier = $1 AND blocked_until > NOW())`,
		"ip:"+ip).Scan(&blocked); err != nil {
		t.Fatalf("read IP block: %v", err)
	}
	if !blocked {
		t.Fatal("IP was not blocked after the threshold")
	}

	otherIP := uniqueIP()
	cleanupIP(t, db, otherIP)
	if _, err := service.reserveOTP(ctx, targetPhone, otherIP, HashOTP("654321")); err != nil {
		t.Fatalf("different IP was blocked: %v", err)
	}
}
