package whatsapp

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"
)

func TestSQLOTPConsumptionLimitsAndFailure(t *testing.T) {
	url := os.Getenv("DEV5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("disposable PostgreSQL database and registered driver required")
	}
	app, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	admin, err := sql.Open("pgx", os.Getenv("DEV5_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	ctx := context.Background()
	school := "00000000-0000-4000-8000-000000000001"
	session := "00000000-0000-4000-8000-000000000701"
	if _, err := admin.Exec(`DELETE FROM chat_otp_requests; DELETE FROM chat_sessions;`); err != nil {
		t.Fatal(err)
	}
	store := NewSQLStore(app)
	with := func(fn func(Store) error) {
		t.Helper()
		if err := store.WithTx(ctx, school, fn); err != nil {
			t.Fatal(err)
		}
	}
	with(func(tx Store) error {
		return tx.SaveSession(ctx, &ChatSession{ID: session, SchoolID: school, Channel: "whatsapp", SenderWANumber: "synthetic-sender", StudentID: "00000000-0000-4000-8000-000000000100", GuardianUserID: "00000000-0000-4000-8000-000000000300", State: StateAwaitingOTP})
	})
	save := func(expiry time.Time) {
		with(func(tx Store) error { return tx.SaveOTP(ctx, session, hashOTP("123456"), expiry) })
	}
	verify := func(code string, want bool) {
		with(func(tx Store) error {
			valid, err := tx.VerifyOTP(ctx, session, code)
			if valid != want || (err != nil && !errors.Is(err, ErrInvalidOTP)) {
				t.Fatalf("verify: %v %v", valid, err)
			}
			return nil
		})
	}
	save(time.Now().Add(time.Minute))
	verify(hashOTP("123456"), false)
	verify("123456", true)
	verify("123456", false)
	with(func(tx Store) error {
		count, err := tx.GetRecentOTPRequestCount(ctx, "synthetic-sender", 30*time.Minute)
		if count != 1 {
			t.Fatalf("consumption removed rate history: %d", count)
		}
		return err
	})
	save(time.Now().Add(-time.Second))
	verify("123456", false)
	save(time.Now().Add(time.Minute))
	for i := 0; i < 4; i++ {
		verify("000000", false)
	}
	verify("123456", false)
	save(time.Now().Add(time.Minute))
	if _, err := admin.Exec(`CREATE OR REPLACE FUNCTION dev5_otp_reject() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic OTP storage failure'; END $$; CREATE TRIGGER dev5_otp_failure BEFORE UPDATE ON chat_otp_requests FOR EACH ROW EXECUTE FUNCTION dev5_otp_reject();`); err != nil {
		t.Fatal(err)
	}
	err = store.WithTx(ctx, school, func(tx Store) error {
		valid, err := tx.VerifyOTP(ctx, session, "123456")
		if valid {
			t.Fatal("OTP accepted despite consumption failure")
		}
		return err
	})
	if _, dropErr := admin.Exec(`DROP TRIGGER dev5_otp_failure ON chat_otp_requests`); dropErr != nil {
		t.Fatal(dropErr)
	}
	if err == nil {
		t.Fatal("ignored verification persistence failure")
	}
	verify("123456", true)
	with(func(tx Store) error {
		if err := tx.BlockSender(ctx, school, "synthetic-sender", "synthetic test", time.Now().Add(24*time.Hour)); err != nil {
			return err
		}
		blocked, _, err := tx.IsSenderBlocked(ctx, "synthetic-sender")
		if !blocked {
			t.Fatal("block was not persisted")
		}
		return err
	})
	if err := store.WithTx(ctx, "00000000-0000-4000-8000-000000000002", func(tx Store) error {
		if _, err := tx.GetSessionByID(ctx, session); !errors.Is(err, ErrNotFound) {
			t.Fatalf("cross-school session lookup: %v", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSQLFailedSMSRollsBackOTPAndSession(t *testing.T) {
	url := os.Getenv("DEV5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("disposable PostgreSQL database and registered driver required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	school := "00000000-0000-4000-8000-000000000001"
	sender := "synthetic-failed-delivery"
	store := NewSQLStore(db)
	if err := store.WithTx(ctx, school, func(tx Store) error {
		_, err := tx.(*SQLStore).tx.ExecContext(ctx, `DELETE FROM chat_otp_requests WHERE school_id=$1 AND admission_number=$2`, school, "SYNTHETIC-100")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	err = store.WithTx(ctx, school, func(tx Store) error {
		service := NewService(tx, failedSMS{}, nil, nil)
		reply, err := service.processSchoolMessage(ctx, school, sender, "SYNTHETIC-100")
		if err == nil || reply != "" {
			t.Fatalf("failed delivery reported success: %q %v", reply, err)
		}
		return err
	})
	if err == nil {
		t.Fatal("SMS failure was swallowed")
	}
	if err := store.WithTx(ctx, school, func(tx Store) error {
		if _, err := tx.GetActiveSession(ctx, school, sender); !errors.Is(err, ErrNotFound) {
			t.Fatalf("failed delivery session committed: %v", err)
		}
		count, err := tx.GetRecentOTPRequestCount(ctx, sender, 30*time.Minute)
		if count != 0 {
			t.Fatalf("undelivered OTP committed: %d", count)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}
