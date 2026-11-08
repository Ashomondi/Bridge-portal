package billing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func billingSQLTestDB(t *testing.T) (*sql.DB, *sql.DB) {
	t.Helper()
	url := os.Getenv("DEV5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("DEV5_TEST_DATABASE_URL and a registered PostgreSQL driver are required")
	}
	app, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := sql.Open("pgx", os.Getenv("DEV5_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); admin.Close() })
	if err := app.Ping(); err != nil {
		t.Fatal(err)
	}
	return app, admin
}

func TestSQLManualResolutionRollbackAndConcurrency(t *testing.T) {
	app, admin := billingSQLTestDB(t)
	ctx := context.Background()
	school := "00000000-0000-4000-8000-000000000001"
	student := "00000000-0000-4000-8000-000000000100"
	actor := "00000000-0000-4000-8000-000000000300"
	store := NewSQLStore(app)
	service := NewService(store, nil, nil, nil)
	for i, stage := range []string{"payment", "ledger", "entry", "queue", "success"} {
		t.Run(stage, func(t *testing.T) {
			paymentID := fmt.Sprintf("00000000-0000-4000-8000-%012d", 500+i)
			if _, err := admin.Exec(`DELETE FROM payments; DELETE FROM ledger_entries; DELETE FROM fee_ledgers; DELETE FROM fee_charges;`); err != nil {
				t.Fatal(err)
			}
			if _, err := admin.Exec(`INSERT INTO fee_ledgers(school_id,student_id,academic_year_id,term_id,billed_amount,current_balance) VALUES ($1,$2,'00000000-0000-4000-8000-000000000010','00000000-0000-4000-8000-000000000040',1000,1000)`, school, student); err != nil {
				t.Fatal(err)
			}
			if err := store.WithTx(ctx, school, func(tx Store) error {
				if err := tx.CreatePayment(ctx, &Payment{ID: paymentID, SchoolID: school, TransactionRef: "synthetic-receipt", Provider: "mpesa_c2b", Amount: 100.25, MatchedStatus: "unmatched"}); err != nil {
					return err
				}
				return tx.CreateUnmatchedPayment(ctx, &UnmatchedPayment{SchoolID: school, PaymentID: paymentID, TransactionRef: "synthetic-receipt", Amount: 100.25, Status: "unmatched"})
			}); err != nil {
				t.Fatal(err)
			}
			table := map[string]string{"payment": "payments", "ledger": "fee_ledgers", "entry": "ledger_entries", "queue": "unmatched_payments"}[stage]
			if table != "" {
				if _, err := admin.Exec(`CREATE OR REPLACE FUNCTION dev5_reject_write() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic write failure'; END $$;`); err != nil {
					t.Fatal(err)
				}
				if _, err := admin.Exec(`CREATE TRIGGER dev5_failure BEFORE INSERT OR UPDATE ON ` + table + ` FOR EACH ROW EXECUTE FUNCTION dev5_reject_write()`); err != nil {
					t.Fatal(err)
				}
				_, err := service.MatchUnmatchedPayment(ctx, school, paymentID, student, "synthetic resolution", actor)
				if _, dropErr := admin.Exec(`DROP TRIGGER dev5_failure ON ` + table); dropErr != nil {
					t.Fatal(dropErr)
				}
				if err == nil {
					t.Fatal("expected injected SQL failure")
				}
				var matched, status string
				var paid float64
				var entries int
				if err := admin.QueryRow(`SELECT matched_status FROM payments WHERE id=$1`, paymentID).Scan(&matched); err != nil {
					t.Fatal(err)
				}
				if err := admin.QueryRow(`SELECT status FROM unmatched_payments WHERE payment_id=$1`, paymentID).Scan(&status); err != nil {
					t.Fatal(err)
				}
				if err := admin.QueryRow(`SELECT paid_amount FROM fee_ledgers WHERE school_id=$1 AND student_id=$2`, school, student).Scan(&paid); err != nil {
					t.Fatal(err)
				}
				if err := admin.QueryRow(`SELECT COUNT(*) FROM ledger_entries`).Scan(&entries); err != nil {
					t.Fatal(err)
				}
				if matched != "unmatched" || status != "unmatched" || paid != 0 || entries != 0 {
					t.Fatalf("partial SQL resolution: %s %s %.2f %d", matched, status, paid, entries)
				}
			}
			var wg sync.WaitGroup
			results := make(chan error, 4)
			for j := 0; j < 4; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := service.MatchUnmatchedPayment(ctx, school, paymentID, student, "synthetic retry", actor)
					results <- err
				}()
			}
			wg.Wait()
			close(results)
			success := 0
			for err := range results {
				if err == nil {
					success++
				} else if !strings.Contains(err.Error(), "PAYMENT_ALREADY_RESOLVED") {
					t.Error(err)
				}
			}
			if err := store.WithTx(ctx, school, func(tx Store) error { return tx.UpdatePaymentStatus(ctx, school, paymentID, "failed") }); err != nil {
				t.Fatal(err)
			}
			var matched string
			if err := admin.QueryRow(`SELECT matched_status FROM payments WHERE id=$1`, paymentID).Scan(&matched); err != nil || matched != "matched" {
				t.Fatalf("late failure reopened resolution: %s %v", matched, err)
			}
			if success != 1 {
				t.Fatalf("successful duplicate requests=%d", success)
			}
			var paid float64
			var entries int
			if err := admin.QueryRow(`SELECT paid_amount FROM fee_ledgers WHERE school_id=$1 AND student_id=$2`, school, student).Scan(&paid); err != nil {
				t.Fatal(err)
			}
			if err := admin.QueryRow(`SELECT COUNT(*) FROM ledger_entries`).Scan(&entries); err != nil {
				t.Fatal(err)
			}
			if paid != 100.25 || entries != 1 {
				t.Fatalf("duplicate credit: %.2f %d", paid, entries)
			}
		})
	}
}

func TestSQLPrepaymentAndTenantIsolation(t *testing.T) {
	app, admin := billingSQLTestDB(t)
	ctx := context.Background()
	store := NewSQLStore(app)
	service := NewService(store, nil, nil, nil)
	school := "00000000-0000-4000-8000-000000000001"
	student := "00000000-0000-4000-8000-000000000100"
	if _, err := admin.Exec(`DELETE FROM payments; DELETE FROM ledger_entries; DELETE FROM fee_ledgers; DELETE FROM fee_charges;`); err != nil {
		t.Fatal(err)
	}
	if err := service.creditStudentPayment(ctx, school, student, "", "synthetic-prepaid", "mpesa_c2b", 1500.25, "", ""); err != nil {
		t.Fatal(err)
	}
	balance, err := service.GetStudentFeeBalance(ctx, school, student, "admin", RoleIkonexAdmin)
	if err != nil || balance.CreditBalance != 1500.25 {
		t.Fatalf("prepaid credit missing: %+v %v", balance, err)
	}
	_, err = service.LoadFeeStructure(ctx, LoadFeeStructureRequest{SchoolID: school, AcademicYearID: "00000000-0000-4000-8000-000000000010", GradeLevelID: "00000000-0000-4000-8000-000000000020", TermID: "00000000-0000-4000-8000-000000000040", TermStartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Items: []FeeStructureItemInput{{Title: "synthetic tuition", Amount: 1000}}})
	if err != nil {
		t.Fatal(err)
	}
	balance, err = service.GetStudentFeeBalance(ctx, school, student, "admin", RoleIkonexAdmin)
	if err != nil || balance.CreditBalance != 500.25 || balance.CurrentBalance != 0 || balance.PaidAmount != 1500.25 {
		t.Fatalf("first term lost credit: %+v %v", balance, err)
	}
	var count int
	if err := app.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unset RLS returned rows: %d %v", count, err)
	}
	if err := store.WithTx(ctx, school, func(tx Store) error {
		if err := tx.(*SQLStore).tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM students`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			t.Fatalf("RLS leaked other school: %d", count)
		}
		if err := tx.(*SQLStore).WithTx(ctx, "00000000-0000-4000-8000-000000000002", func(Store) error { t.Fatal("nested tenant switch executed"); return nil }); err == nil {
			t.Fatal("nested tenant switch accepted")
		}
		if _, _, _, err := tx.GetStudent(ctx, school, "00000000-0000-4000-8000-000000000200"); err == nil {
			t.Fatal("cross-school student read accepted")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("tenant context leaked across pooled connection: %d %v", count, err)
	}
}
