package billing_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func schemaSnapshot(t *testing.T, db *testenv.DB) []string {
	t.Helper()
	rows, err := db.Admin.Query(context.Background(), `
		SELECT 'column:'||table_name||':'||column_name||':'||data_type||':'||is_nullable||':'||coalesce(column_default,'')
		FROM information_schema.columns WHERE table_schema='public'
		UNION ALL SELECT 'constraint:'||c.relname||':'||conname||':'||pg_get_constraintdef(k.oid)
		FROM pg_constraint k JOIN pg_class c ON c.oid=k.conrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public'
		UNION ALL SELECT 'index:'||indexname||':'||indexdef FROM pg_indexes WHERE schemaname='public'
		ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var snapshot []string
	for rows.Next() {
		var entry string
		if err := rows.Scan(&entry); err != nil {
			t.Fatal(err)
		}
		snapshot = append(snapshot, entry)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestMigrationUpDownUpPreservesBaseline(t *testing.T) {
	db := testenv.New(t, false)
	school := db.SeedSchool(t, "01")
	baseline := schemaSnapshot(t, db)
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.down.sql")
	if got := schemaSnapshot(t, db); !reflect.DeepEqual(got, baseline) {
		t.Fatal("owned rollback changed baseline schema, constraints, or indexes")
	}
	var name string
	if err := db.Admin.QueryRow(context.Background(), "SELECT full_name FROM students WHERE school_id=$1 AND id=$2", school.ID, school.StudentID).Scan(&name); err != nil || name != "Synthetic Student" {
		t.Fatalf("rollback changed baseline data: %q %v", name, err)
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	for _, table := range []string{"school_mpesa_configs", "fee_periods", "fee_charges", "fee_change_notes", "fee_change_note_students", "ledger_entries", "unmatched_payments", "mpesa_stk_requests", "school_whatsapp_bots", "chat_sessions", "in_app_notifications", "ai_action_logs"} {
		var enabled, forced, policy bool
		err := db.Admin.QueryRow(context.Background(), `SELECT c.relrowsecurity,c.relforcerowsecurity,
		EXISTS(SELECT 1 FROM pg_policy p WHERE p.polrelid=c.oid AND p.polcmd='*' AND p.polqual IS NOT NULL AND p.polwithcheck IS NOT NULL)
		FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname=$1`, table).Scan(&enabled, &forced, &policy)
		if err != nil || !enabled || !forced || !policy {
			t.Fatalf("RLS missing on %s: %v", table, err)
		}
	}
}

func TestMigrationTenantIsolationAndReferences(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	ctx := context.Background()
	if _, err := db.Admin.Exec(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Demo','Synthetic'),($3,$4,'test','Demo','Synthetic')", a.ID, a.UserID, b.ID, b.UserID); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []string{"", a.ID, b.ID} {
		err := db.WithTenantTx(ctx, tenant, func(tx pgx.Tx) error {
			var count int
			// Deliberately unscoped negative test: RLS itself must filter records.
			if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications").Scan(&count); err != nil {
				return err
			}
			want := 1
			if tenant == "" {
				want = 0
			}
			if count != want {
				t.Fatalf("RLS returned %d rows, want %d", count, want)
			}
			var superuser, bypass bool
			if err := tx.QueryRow(ctx, "SELECT rolsuper,rolbypassrls FROM pg_roles WHERE rolname=current_user").Scan(&superuser, &bypass); err != nil {
				return err
			}
			if superuser || bypass {
				t.Fatal("RLS tested with privileged role")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		sql, code string
		args      []any
	}{
		{"INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Demo','Synthetic')", "42501", []any{b.ID, b.UserID}},
		{"INSERT INTO fee_periods(school_id,academic_year_id,term_id,name,due_date) VALUES($1,$2,$3,'Term 1','2026-01-05')", "23503", []any{a.ID, b.YearID, b.TermID}},
		{"INSERT INTO fee_change_notes(school_id,change_type,note,created_by,affected_students_count) VALUES($1,'amount',' ',$2,0)", "23514", []any{a.ID, a.UserID}},
		{"INSERT INTO fee_ledgers(school_id,student_id,academic_year_id,term_id,current_balance) VALUES($1,$2,$3,$4,-1)", "23514", []any{a.ID, a.StudentID, a.YearID, a.TermID}},
	} {
		err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, tc.sql, tc.args...); return err })
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != tc.code {
			t.Fatalf("expected PostgreSQL %s, got %v", tc.code, err)
		}
	}
}
