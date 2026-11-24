package jobs_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/jobs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openAppPool(t *testing.T, db *testenv.DB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatalf("generate role suffix: %v", err)
	}
	roleName := "jobs_rls_" + hex.EncodeToString(random[:])
	rolePass := "jobs_rls_pass"
	ident := pgx.Identifier{roleName}.Sanitize()

	if _, err := db.Admin.Exec(ctx,
		"CREATE ROLE "+ident+" LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD '"+rolePass+"'",
	); err != nil {
		t.Fatalf("create app role: %v", err)
	}
	if _, err := db.Admin.Exec(ctx,
		"GRANT USAGE ON SCHEMA public TO "+ident+"; "+
			"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "+ident,
	); err != nil {
		t.Fatalf("grant app role: %v", err)
	}

	cfg := db.Admin.Config().Copy()
	cfg.ConnConfig.User = roleName
	cfg.ConnConfig.Password = rolePass
	cfg.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("open app pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = db.Admin.Exec(context.Background(),
			"REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM "+ident+"; "+
				"REVOKE USAGE ON SCHEMA public FROM "+ident+"; "+
				"DROP ROLE IF EXISTS "+ident)
	})
	return pool
}

func TestTenantDBAdapterIsolationAndRollback(t *testing.T) {
	db := testenv.New(t, true)
	ctx := context.Background()

	schoolA := db.SeedSchool(t, "11")
	schoolB := db.SeedSchool(t, "22")
	appPool := openAppPool(t, db)

	adapter, err := jobs.NewTenantDBAdapter(appPool)
	if err != nil {
		t.Fatalf("NewTenantDBAdapter: %v", err)
	}

	// 1. Missing and invalid school IDs fail closed.
	if err := adapter.WithSchoolTx(ctx, "", func(pgx.Tx) error { return nil }); err == nil {
		t.Fatal("expected error for empty school_id")
	}
	if err := adapter.WithSchoolTx(ctx, "not-a-uuid", func(pgx.Tx) error { return nil }); err == nil {
		t.Fatal("expected error for invalid school_id")
	}

	// 2. Verify School A sees its own membership and School B cannot see School A's membership under NOBYPASSRLS.
	var countA int
	if err := adapter.WithSchoolTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT COUNT(*) FROM memberships WHERE user_id = $1`, schoolA.UserID).Scan(&countA)
	}); err != nil {
		t.Fatalf("school A query: %v", err)
	}
	if countA != 1 {
		t.Fatalf("expected 1 membership visible in school A, got %d", countA)
	}

	var countB int
	if err := adapter.WithSchoolTx(ctx, schoolB.ID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT COUNT(*) FROM memberships WHERE user_id = $1`, schoolA.UserID).Scan(&countB)
	}); err != nil {
		t.Fatalf("school B query: %v", err)
	}
	if countB != 0 {
		t.Fatalf("cross-tenant leak: school B saw %d memberships from school A", countB)
	}

	// 3. Rollback discards writes on error.
	sentinel := errors.New("abort tx")
	err = adapter.WithSchoolTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO in_app_notifications (school_id, recipient_user_id, category, title, body, metadata)
			VALUES ($1, $2, 'system', 'Rollback test', 'Should roll back', '{}'::jsonb)`, schoolA.ID, schoolA.UserID); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	var notifCount int
	if err := adapter.WithSchoolTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT COUNT(*) FROM in_app_notifications WHERE title = 'Rollback test'`).Scan(&notifCount)
	}); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if notifCount != 0 {
		t.Fatalf("expected 0 notifications after rollback, got %d", notifCount)
	}

	// 4. Concurrent tenant jobs remain isolated across pooled connections.
	tenantDB := adapter.AsContractsTenantDB()
	var wg sync.WaitGroup
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		isA := i%2 == 0
		go func(useA bool) {
			defer wg.Done()
			sid := schoolB.ID
			want := 0
			if useA {
				sid = schoolA.ID
				want = 1
			}
			err := tenantDB.WithTenantTx(ctx, contracts.Principal{
				SchoolID: sid, UserID: schoolA.UserID, MembershipID: schoolA.MembershipID, Role: "parent",
			}, func(tx pgx.Tx) error {
				var got int
				if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM memberships WHERE user_id = $1`, schoolA.UserID).Scan(&got); err != nil {
					return err
				}
				if got != want {
					return errors.New("concurrent RLS isolation mismatch")
				}
				return nil
			})
			if err != nil {
				errCh <- err
			}
		}(isA)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent tenant tx error: %v", err)
	}
}
