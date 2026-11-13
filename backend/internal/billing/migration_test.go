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

func TestMigrationPreservesLegacyCreditAcrossRoundTrip(t *testing.T) {
	db := testenv.New(t, false)
	ctx := context.Background()
	var schools []testenv.School
	balances := []string{"-123.45", "55.00", "0.00"}
	for i, code := range []string{"01", "02", "03"} {
		school := db.SeedSchool(t, code)
		schools = append(schools, school)
		if _, err := db.Admin.Exec(ctx, `INSERT INTO fee_ledgers
			(school_id,student_id,academic_year_id,term_id,current_balance)
			VALUES($1,$2,$3,$4,$5)`, school.ID, school.StudentID, school.YearID, school.TermID, balances[i]); err != nil {
			t.Fatal(err)
		}
	}
	checkCredit := func() {
		t.Helper()
		for i, school := range schools {
			var balance, credit, applied, transferred string
			err := db.Admin.QueryRow(ctx, `SELECT current_balance::text,credit_balance::text,
				credit_applied::text,credit_transferred::text FROM fee_ledgers
				WHERE school_id=$1 AND student_id=$2`, school.ID, school.StudentID).Scan(&balance, &credit, &applied, &transferred)
			wantBalance, wantCredit := balances[i], "0.00"
			if i == 0 {
				wantBalance, wantCredit = "0.00", "123.45"
			}
			if err != nil || balance != wantBalance || credit != wantCredit || applied != "0.00" || transferred != "0.00" {
				t.Fatalf("school %s: balance=%s credit=%s applied=%s transferred=%s: %v", school.ID, balance, credit, applied, transferred, err)
			}
		}
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	checkCredit()
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.down.sql")
	for i, school := range schools {
		var balance string
		err := db.Admin.QueryRow(ctx, "SELECT current_balance::text FROM fee_ledgers WHERE school_id=$1 AND student_id=$2", school.ID, school.StudentID).Scan(&balance)
		if err != nil || balance != balances[i] {
			t.Fatalf("legacy balance changed: %s, want %s: %v", balance, balances[i], err)
		}
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	checkCredit()
}

func TestMigrationRollbackCollapsesBoardingVariants(t *testing.T) {
	db := testenv.New(t, true)
	ctx := context.Background()
	type survivor struct {
		school              testenv.School
		id, variant, amount string
	}
	var survivors []survivor
	for i, variants := range [][]string{{"boarder", "day", "all"}, {"boarder", "day"}, {"boarder"}, {"day"}} {
		school := db.SeedSchool(t, []string{"01", "02", "03", "04"}[i])
		var keep survivor
		for j, variant := range variants {
			amount := []string{"300.00", "200.00", "100.00"}[j]
			var id string
			err := db.Admin.QueryRow(ctx, `INSERT INTO fee_structures
				(school_id,academic_year_id,grade_level_id,term_id,boarding_status,total_amount)
				VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, school.ID, school.YearID, school.GradeID, school.TermID, variant, amount).Scan(&id)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Admin.Exec(ctx, "INSERT INTO fee_structure_items(fee_structure_id,title,amount) VALUES($1,$2,$3)", id, variant, amount); err != nil {
				t.Fatal(err)
			}
			keep = survivor{school, id, variant, amount}
		}
		survivors = append(survivors, keep)
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.down.sql")
	for _, keep := range survivors {
		var count int
		if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM fee_structures WHERE school_id=$1", keep.school.ID).Scan(&count); err != nil || count != 1 {
			t.Fatalf("expected one structure per legacy key, got %d: %v", count, err)
		}
		var id, amount, title, itemAmount string
		err := db.Admin.QueryRow(ctx, `SELECT f.id,f.total_amount::text,i.title,i.amount::text
			FROM fee_structures f JOIN fee_structure_items i ON i.fee_structure_id=f.id
			WHERE f.school_id=$1`, keep.school.ID).Scan(&id, &amount, &title, &itemAmount)
		if err != nil || id != keep.id || amount != keep.amount || title != keep.variant || itemAmount != keep.amount {
			t.Fatalf("wrong survivor or changed items: %s %s %s %s: %v", id, amount, title, itemAmount, err)
		}
		_, err = db.Admin.Exec(ctx, `INSERT INTO fee_structures(school_id,academic_year_id,grade_level_id,term_id)
			VALUES($1,$2,$3,$4)`, keep.school.ID, keep.school.YearID, keep.school.GradeID, keep.school.TermID)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
			t.Fatalf("legacy uniqueness not restored: %v", err)
		}
	}
	var itemCount int
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM fee_structure_items").Scan(&itemCount); err != nil || itemCount != len(survivors) {
		t.Fatalf("discarded variant items remain: %d: %v", itemCount, err)
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	for _, keep := range survivors {
		var variant string
		if err := db.Admin.QueryRow(ctx, "SELECT boarding_status FROM fee_structures WHERE school_id=$1 AND id=$2", keep.school.ID, keep.id).Scan(&variant); err != nil || variant != "all" {
			t.Fatalf("surviving legacy structure did not upgrade: %s: %v", variant, err)
		}
	}
}

func TestMigrationVerifiedSessionRequiresExpiry(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()
	for _, tc := range []struct {
		name                        string
		guardian, verified, expires any
		valid                       bool
	}{
		{"missing expiry", school.UserID, "2026-01-01T10:00:00Z", nil, false},
		{"equal expiry", school.UserID, "2026-01-01T10:00:00Z", "2026-01-01T10:00:00Z", false},
		{"earlier expiry", school.UserID, "2026-01-01T10:00:00Z", "2026-01-01T09:00:00Z", false},
		{"missing guardian", nil, "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", false},
		{"missing verification", school.UserID, nil, "2026-01-01T11:00:00Z", false},
		{"valid", school.UserID, "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, update := range []bool{false, true} {
				err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
					var err error
					if update {
						var id string
						if err := tx.QueryRow(ctx, "INSERT INTO chat_sessions(school_id,channel) VALUES($1,'in_app') RETURNING id", school.ID).Scan(&id); err != nil {
							return err
						}
						_, err = tx.Exec(ctx, `UPDATE chat_sessions SET state='VERIFIED',guardian_user_id=$3,
							verified_at=$4::text::timestamptz,expires_at=$5::text::timestamptz WHERE school_id=$1 AND id=$2`, school.ID, id, tc.guardian, tc.verified, tc.expires)
					} else {
						_, err = tx.Exec(ctx, `INSERT INTO chat_sessions(school_id,channel,state,guardian_user_id,verified_at,expires_at)
							VALUES($1,'in_app','VERIFIED',$2,$3::text::timestamptz,$4::text::timestamptz)`, school.ID, tc.guardian, tc.verified, tc.expires)
					}
					return err
				})
				if tc.valid {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var pgErr *pgconn.PgError
					if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
						t.Fatalf("update=%t: expected CHECK violation, got %v", update, err)
					}
				}
			}
		})
	}
}
