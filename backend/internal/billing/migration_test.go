package billing_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
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

	// Verify contracts.RequireIdentity + contracts.TenantTxFunc bind PostgreSQL RLS
	// solely to Principal.SchoolID, ignore request-supplied school IDs, reject
	// invalid principals, and preserve rollback/commit semantics.
	tenantDB := contracts.TenantTxFunc(db.WithTenantTx)
	var apiErr *contracts.APIError
	if err := tenantDB.WithTenantTx(ctx, contracts.Principal{}, func(pgx.Tx) error {
		t.Fatal("invalid principal should not open tenant transaction")
		return nil
	}); !errors.As(err, &apiErr) || apiErr.Code != "UNAUTHENTICATED" || apiErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 UNAUTHENTICATED for invalid principal, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/?school_id="+b.ID, strings.NewReader(`{"school_id":"`+b.ID+`"}`))
	req.Header.Set("X-School-ID", b.ID)
	req.Header.Set("X-User-ID", b.UserID)
	principal, err := contracts.RequireIdentity(req, func(*http.Request) (contracts.Principal, error) {
		return contracts.Principal{
			UserID:       a.UserID,
			SchoolID:     a.ID,
			MembershipID: a.MembershipID,
			Role:         "parent",
		}, nil
	})
	if err != nil {
		t.Fatalf("RequireIdentity failed: %v", err)
	}

	rollbackErr := errors.New("intentional rollback")
	if err := tenantDB.WithTenantTx(ctx, principal, func(tx pgx.Tx) error {
		var count int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications").Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			t.Fatalf("expected 1 isolated row for school a, got %d", count)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Rollback','Should roll back')", a.ID, a.UserID); err != nil {
			return err
		}
		return rollbackErr
	}); !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}

	if err := tenantDB.WithTenantTx(ctx, principal, func(tx pgx.Tx) error {
		var count int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications").Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			t.Fatalf("rolled-back write leaked; got %d rows, want 1", count)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Commit','Should commit')", a.ID, a.UserID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("commit tx failed: %v", err)
	}

	if err := tenantDB.WithTenantTx(ctx, principal, func(tx pgx.Tx) error {
		var count int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications").Scan(&count); err != nil {
			return err
		}
		if count != 2 {
			t.Fatalf("committed write missing; got %d rows, want 2", count)
		}
		return nil
	}); err != nil {
		t.Fatalf("post-commit verification failed: %v", err)
	}
}

func TestMigrationPreservesLegacyCreditAcrossRoundTrip(t *testing.T) {
	db := testenv.New(t, false)
	ctx := context.Background()
	type ledgerCase struct {
		schoolID, studentID, termID string
		legacyBalance               string
		wantBalance, wantCredit     string
	}
	var cases []ledgerCase

	// School 01 contains mixed records (negative, positive, and zero) across terms.
	school01 := db.SeedSchool(t, "01")
	var term2ID, term3ID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current)
		VALUES($1,$2,2,'2026-05-04','2026-08-07',false) RETURNING id`, school01.ID, school01.YearID).Scan(&term2ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current)
		VALUES($1,$2,3,'2026-09-07','2026-11-20',false) RETURNING id`, school01.ID, school01.YearID).Scan(&term3ID); err != nil {
		t.Fatal(err)
	}
	cases = append(cases,
		ledgerCase{school01.ID, school01.StudentID, school01.TermID, "-123.45", "0.00", "123.45"},
		ledgerCase{school01.ID, school01.StudentID, term2ID, "55.00", "55.00", "0.00"},
		ledgerCase{school01.ID, school01.StudentID, term3ID, "0.00", "0.00", "0.00"},
	)

	// Separate schools with positive and zero balances.
	school02 := db.SeedSchool(t, "02")
	cases = append(cases, ledgerCase{school02.ID, school02.StudentID, school02.TermID, "55.00", "55.00", "0.00"})
	school03 := db.SeedSchool(t, "03")
	cases = append(cases, ledgerCase{school03.ID, school03.StudentID, school03.TermID, "0.00", "0.00", "0.00"})

	for _, c := range cases {
		if _, err := db.Admin.Exec(ctx, `INSERT INTO fee_ledgers
			(school_id,student_id,academic_year_id,term_id,current_balance)
			SELECT $1,$2,academic_year_id,$3,$4 FROM terms WHERE school_id=$1 AND id=$3`,
			c.schoolID, c.studentID, c.termID, c.legacyBalance); err != nil {
			t.Fatal(err)
		}
	}
	checkCredit := func() {
		t.Helper()
		for _, c := range cases {
			var balance, credit, applied, transferred string
			err := db.Admin.QueryRow(ctx, `SELECT current_balance::text,credit_balance::text,
				credit_applied::text,credit_transferred::text FROM fee_ledgers
				WHERE school_id=$1 AND student_id=$2 AND term_id=$3`, c.schoolID, c.studentID, c.termID).Scan(&balance, &credit, &applied, &transferred)
			if err != nil || balance != c.wantBalance || credit != c.wantCredit || applied != "0.00" || transferred != "0.00" {
				t.Fatalf("school %s term %s: balance=%s credit=%s applied=%s transferred=%s: %v", c.schoolID, c.termID, balance, credit, applied, transferred, err)
			}
		}
	}
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	checkCredit()
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.down.sql")
	for _, c := range cases {
		var balance string
		err := db.Admin.QueryRow(ctx, "SELECT current_balance::text FROM fee_ledgers WHERE school_id=$1 AND student_id=$2 AND term_id=$3", c.schoolID, c.studentID, c.termID).Scan(&balance)
		if err != nil || balance != c.legacyBalance {
			t.Fatalf("legacy balance changed: %s, want %s: %v", balance, c.legacyBalance, err)
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
	precedence := func(v string) int {
		switch v {
		case "all":
			return 0
		case "day":
			return 1
		default:
			return 2
		}
	}
	variantAmounts := map[string]string{
		"all":     "100.00",
		"day":     "200.00",
		"boarder": "300.00",
	}
	combinations := [][]string{
		{"boarder", "day", "all"},
		{"day", "all"},
		{"boarder", "all"},
		{"boarder", "day"},
		{"day"},
		{"boarder"},
		{"all"},
	}
	var survivors []survivor
	for i, variants := range combinations {
		school := db.SeedSchool(t, []string{"01", "02", "03", "04", "05", "06", "07"}[i])
		var keep survivor
		bestRank := 99
		for _, variant := range variants {
			amount := variantAmounts[variant]
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
			if rank := precedence(variant); rank < bestRank {
				bestRank = rank
				keep = survivor{school, id, variant, amount}
			}
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
			t.Fatalf("wrong survivor or changed items: got (%s,%s,%s,%s) want (%s,%s,%s,%s): %v", id, amount, title, itemAmount, keep.id, keep.amount, keep.variant, keep.amount, err)
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

func TestMigrationUniqueSTKTransactionRef(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	ctx := context.Background()

	// School a can insert an STK request with a non-null transaction_ref.
	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mpesa_stk_requests
			(school_id,student_id,guardian_user_id,amount,phone_number,checkout_request_id,status,transaction_ref)
			VALUES($1,$2,$3,5000.00,'+254700000001','ws_CO_01','succeeded','QWE123RTY')`,
			a.ID, a.StudentID, a.UserID)
		return err
	}); err != nil {
		t.Fatalf("insert initial STK request: %v", err)
	}

	// Second request in the same school with the same transaction_ref must fail with 23505.
	err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mpesa_stk_requests
			(school_id,student_id,guardian_user_id,amount,phone_number,checkout_request_id,status,transaction_ref)
			VALUES($1,$2,$3,5000.00,'+254700000001','ws_CO_02','succeeded','QWE123RTY')`,
			a.ID, a.StudentID, a.UserID)
		return err
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("duplicate school transaction_ref: expected PostgreSQL 23505, got %v", err)
	}

	// Multiple NULL transaction_ref rows in the same school must succeed.
	var nullReqID string
	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO mpesa_stk_requests
			(school_id,student_id,guardian_user_id,amount,phone_number,checkout_request_id,status,transaction_ref)
			VALUES($1,$2,$3,1000.00,'+254700000001','ws_CO_03','pending',NULL) RETURNING id`,
			a.ID, a.StudentID, a.UserID).Scan(&nullReqID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO mpesa_stk_requests
			(school_id,student_id,guardian_user_id,amount,phone_number,checkout_request_id,status,transaction_ref)
			VALUES($1,$2,$3,2000.00,'+254700000001','ws_CO_04','timeout',NULL)`,
			a.ID, a.StudentID, a.UserID)
		return err
	}); err != nil {
		t.Fatalf("multiple NULL transaction_ref rows should succeed: %v", err)
	}

	// Updating a NULL transaction_ref row to an existing school ref must fail with 23505.
	err = db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE mpesa_stk_requests SET status='succeeded',transaction_ref='QWE123RTY' WHERE school_id=$1 AND id=$2`, a.ID, nullReqID)
		return err
	})
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("update to duplicate school transaction_ref: expected PostgreSQL 23505, got %v", err)
	}

	// Different school can use the same transaction_ref (school-scoped uniqueness).
	if err := db.WithTenantTx(ctx, b.ID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mpesa_stk_requests
			(school_id,student_id,guardian_user_id,amount,phone_number,checkout_request_id,status,transaction_ref)
			VALUES($1,$2,$3,5000.00,'+254700000002','ws_CO_B1','succeeded','QWE123RTY')`,
			b.ID, b.StudentID, b.UserID)
		return err
	}); err != nil {
		t.Fatalf("different school with same transaction_ref should succeed: %v", err)
	}
}

func TestMigrationVerifiedSessionRequiresExpiry(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()
	const validTokenHash = "sha256:synthetic-token-hash"
	for _, tc := range []struct {
		name                                   string
		guardian, tokenHash, verified, expires any
		valid                                  bool
	}{
		{"missing token hash", school.UserID, nil, "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", false},
		{"blank token hash", school.UserID, "   ", "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", false},
		{"missing expiry", school.UserID, validTokenHash, "2026-01-01T10:00:00Z", nil, false},
		{"equal expiry", school.UserID, validTokenHash, "2026-01-01T10:00:00Z", "2026-01-01T10:00:00Z", false},
		{"earlier expiry", school.UserID, validTokenHash, "2026-01-01T10:00:00Z", "2026-01-01T09:00:00Z", false},
		{"missing guardian", nil, validTokenHash, "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", false},
		{"missing verification", school.UserID, validTokenHash, nil, "2026-01-01T11:00:00Z", false},
		{"valid", school.UserID, validTokenHash, "2026-01-01T10:00:00Z", "2026-01-01T11:00:00Z", true},
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
						_, err = tx.Exec(ctx, `UPDATE chat_sessions SET state='VERIFIED',guardian_user_id=$3,access_token_hash=$4,
							verified_at=$5::text::timestamptz,expires_at=$6::text::timestamptz WHERE school_id=$1 AND id=$2`, school.ID, id, tc.guardian, tc.tokenHash, tc.verified, tc.expires)
					} else {
						_, err = tx.Exec(ctx, `INSERT INTO chat_sessions(school_id,channel,state,guardian_user_id,access_token_hash,verified_at,expires_at)
							VALUES($1,'in_app','VERIFIED',$2,$3,$4::text::timestamptz,$5::text::timestamptz)`, school.ID, tc.guardian, tc.tokenHash, tc.verified, tc.expires)
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

	t.Run("unverified session states allow null token hash", func(t *testing.T) {
		for _, state := range []string{"AWAITING_ADMISSION_NUMBER", "AWAITING_GUARDIAN_SELECTION", "AWAITING_OTP"} {
			if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO chat_sessions(school_id,channel,state,access_token_hash)
					VALUES($1,'in_app',$2,NULL)`, school.ID, state)
				return err
			}); err != nil {
				t.Fatalf("unverified state %s rejected null token hash: %v", state, err)
			}
		}
	})

	t.Run("verified session state transitions", func(t *testing.T) {
		var sessionID string
		if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `INSERT INTO chat_sessions
				(school_id,channel,state,guardian_user_id,access_token_hash,verified_at,expires_at)
				VALUES($1,'in_app','VERIFIED',$2,$3,'2026-01-01T10:00:00Z'::timestamptz,'2026-01-01T11:00:00Z'::timestamptz)
				RETURNING id`, school.ID, school.UserID, validTokenHash).Scan(&sessionID)
		}); err != nil {
			t.Fatalf("insert verified session: %v", err)
		}

		// Clearing access_token_hash while remaining VERIFIED must fail.
		err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "UPDATE chat_sessions SET access_token_hash=NULL WHERE school_id=$1 AND id=$2", school.ID, sessionID)
			return err
		})
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
			t.Fatalf("clearing access_token_hash while VERIFIED: expected 23514, got %v", err)
		}

		// Transitioning out of VERIFIED while clearing token/otp/timestamps must succeed.
		if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE chat_sessions
				SET state='AWAITING_ADMISSION_NUMBER',access_token_hash=NULL,otp_hash=NULL,verified_at=NULL,expires_at=NULL
				WHERE school_id=$1 AND id=$2`, school.ID, sessionID)
			return err
		}); err != nil {
			t.Fatalf("transition out of VERIFIED failed: %v", err)
		}
	})
}

func TestMigration000005UpDownUpPreservesSchema(t *testing.T) {
	db := testenv.New(t, false)
	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
	school := db.SeedSchool(t, "05")
	before05 := schemaSnapshot(t, db)

	db.Apply(t, "000005_r1_mpesa_hardening.up.sql")
	db.Apply(t, "000005_r1_mpesa_hardening.down.sql")
	if got := schemaSnapshot(t, db); !reflect.DeepEqual(got, before05) {
		t.Fatal("000005 rollback did not restore exact 000004 schema, constraints, or indexes")
	}

	db.Apply(t, "000005_r1_mpesa_hardening.up.sql")
	var exists bool
	if err := db.Admin.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'chat_payment_paused_until'
		)`).Scan(&exists); err != nil || !exists {
		t.Fatalf("expected users.chat_payment_paused_until after 000005 up, exists=%v err=%v", exists, err)
	}
	var userExists bool
	if err := db.Admin.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, school.UserID).Scan(&userExists); err != nil || !userExists {
		t.Fatalf("000005 migration lost seeded user data: %v", err)
	}
}
