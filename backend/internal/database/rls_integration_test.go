package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestRLSIsolation proves that FORCE ROW LEVEL SECURITY blocks cross-tenant
// reads and that an unset app.school_id yields zero rows (AC-27).
//
// It runs only when TEST_DATABASE_DSN points at a database with the migrations
// applied, so the default unit-test run does not require PostgreSQL.
func TestRLSIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the RLS integration test")
	}

	db, err := Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suffix := time.Now().UnixNano()
	schoolA := seedSchool(t, ctx, db, fmt.Sprintf("RLS-A-%d", suffix), "RLS School A")
	schoolB := seedSchool(t, ctx, db, fmt.Sprintf("RLS-B-%d", suffix), "RLS School B")

	studentA := seedStudent(t, ctx, db, schoolA, fmt.Sprintf("A-%d", suffix))
	studentB := seedStudent(t, ctx, db, schoolB, fmt.Sprintf("B-%d", suffix))

	if got := countStudent(t, ctx, db, schoolA, studentA); got != 1 {
		t.Fatalf("school A cannot read its own student: count = %d", got)
	}
	if got := countStudent(t, ctx, db, schoolA, studentB); got != 0 {
		t.Fatalf("school A read school B's student: count = %d", got)
	}
	if got := countStudent(t, ctx, db, schoolB, studentA); got != 0 {
		t.Fatalf("school B read school A's student: count = %d", got)
	}

	seedSchoolProfile(t, ctx, db, schoolA)
	seedSchoolProfile(t, ctx, db, schoolB)
	if got := countSchoolProfile(t, ctx, db, schoolA, schoolA); got != 1 {
		t.Fatalf("school A cannot read its own profile: count = %d", got)
	}
	if got := countSchoolProfile(t, ctx, db, schoolA, schoolB); got != 0 {
		t.Fatalf("school A read school B's profile: count = %d", got)
	}
	if got := countSchoolProfile(t, ctx, db, schoolB, schoolA); got != 0 {
		t.Fatalf("school B read school A's profile: count = %d", got)
	}

	var withoutTenant int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM students WHERE id = $1`, studentA).Scan(&withoutTenant); err != nil {
		t.Fatalf("query without tenant: %v", err)
	}
	if withoutTenant != 0 {
		t.Fatalf("query without app.school_id returned %d rows, want 0", withoutTenant)
	}
}

func seedSchool(t *testing.T, ctx context.Context, db *DB, code, name string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO schools (code, name) VALUES ($1, $2) RETURNING id::text`, code, name).Scan(&id); err != nil {
		t.Fatalf("seed school: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM schools WHERE id = $1`, id)
	})
	return id
}

func seedStudent(t *testing.T, ctx context.Context, db *DB, schoolID, admissionNumber string) string {
	t.Helper()
	var id string
	err := WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`INSERT INTO students (school_id, admission_number, full_name, dob, gender)
			 VALUES ($1, $2, 'RLS Test', '2011-01-01', 'F')
			 RETURNING id::text`, schoolID, admissionNumber).Scan(&id)
	})
	if err != nil {
		t.Fatalf("seed student: %v", err)
	}
	return id
}

func countStudent(t *testing.T, ctx context.Context, db *DB, schoolID, studentID string) int {
	t.Helper()
	var count int
	err := WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM students WHERE id = $1`, studentID).Scan(&count)
	})
	if err != nil {
		t.Fatalf("count student: %v", err)
	}
	return count
}

func seedSchoolProfile(t *testing.T, ctx context.Context, db *DB, schoolID string) {
	t.Helper()
	err := WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO school_profile (school_id, contact_phone, contact_email)
			 VALUES ($1, '+254700000000', 'test@ikonex.test')`, schoolID)
		return err
	})
	if err != nil {
		t.Fatalf("seed school profile: %v", err)
	}
}

func countSchoolProfile(t *testing.T, ctx context.Context, db *DB, tenantSchoolID, profileSchoolID string) int {
	t.Helper()
	var count int
	err := WithTenantTx(ctx, db.DB, tenantSchoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM school_profile WHERE school_id = $1`, profileSchoolID).Scan(&count)
	})
	if err != nil {
		t.Fatalf("count school profile: %v", err)
	}
	return count
}

func seedUser(t *testing.T, ctx context.Context, db *DB, phone, name string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO users (phone_number, full_name) VALUES ($1, $2) RETURNING id::text`, phone, name).Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func seedMembership(t *testing.T, ctx context.Context, db *DB, schoolID, userID, role string) string {
	t.Helper()
	var id string
	err := WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`INSERT INTO memberships (school_id, user_id, role) VALUES ($1, $2, $3) RETURNING id::text`,
			schoolID, userID, role).Scan(&id)
	})
	if err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	return id
}

func countMemberships(t *testing.T, ctx context.Context, db *DB, userID string) int {
	t.Helper()
	var count int
	err := WithUserTx(ctx, db.DB, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM memberships WHERE user_id = $1`, userID).Scan(&count)
	})
	if err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	return count
}

func TestMembershipsSelfRead(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the membership self-read test")
	}

	db, err := Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suffix := time.Now().UnixNano()
	schoolA := seedSchool(t, ctx, db, fmt.Sprintf("SELF-A-%d", suffix), "Self Read School A")
	schoolB := seedSchool(t, ctx, db, fmt.Sprintf("SELF-B-%d", suffix), "Self Read School B")
	userA := seedUser(t, ctx, db, fmt.Sprintf("+2547%08d", (suffix/1000)%100000000), "Self Read User A")
	userB := seedUser(t, ctx, db, fmt.Sprintf("+2547%08d", (suffix/1000+1)%100000000), "Self Read User B")

	seedMembership(t, ctx, db, schoolA, userA, "teacher")
	seedMembership(t, ctx, db, schoolB, userA, "parent")
	seedMembership(t, ctx, db, schoolA, userB, "teacher")

	if got := countMemberships(t, ctx, db, userA); got != 2 {
		t.Fatalf("user A memberships = %d, want 2", got)
	}
	if got := countMemberships(t, ctx, db, userB); got != 1 {
		t.Fatalf("user B memberships = %d, want 1", got)
	}

	var withoutUser int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM memberships WHERE user_id = $1`, userA).Scan(&withoutUser); err != nil {
		t.Fatalf("query without app.user_id: %v", err)
	}
	if withoutUser != 0 {
		t.Fatalf("query without app.user_id returned %d rows, want 0", withoutUser)
	}
}
