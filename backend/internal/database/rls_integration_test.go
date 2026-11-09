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
