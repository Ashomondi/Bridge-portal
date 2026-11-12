package billing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

func assertStudentReconciliation(t *testing.T, ctx context.Context, db *testenv.DB, schoolID, studentID string) (opening, billed, paid, current, credit billing.Money) {
	t.Helper()
	err := db.Admin.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(opening_balance), 0)::numeric(12,2),
			COALESCE(SUM(billed_amount), 0)::numeric(12,2),
			COALESCE(SUM(paid_amount), 0)::numeric(12,2),
			COALESCE(SUM(current_balance), 0)::numeric(12,2),
			COALESCE(SUM(credit_balance), 0)::numeric(12,2)
		FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2`,
		schoolID, studentID,
	).Scan(&opening, &billed, &paid, &current, &credit)
	if err != nil {
		t.Fatalf("query ledger summary: %v", err)
	}
	if current < 0 || credit < 0 {
		t.Fatalf("negative balance detected: current=%s credit=%s", current, credit)
	}

	var entriesBilled, entriesPaid, entriesCreditApplied billing.Money
	err = db.Admin.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN entry_type IN ('charge', 'adjustment') THEN amount ELSE 0 END), 0)::numeric(12,2),
			COALESCE(SUM(CASE WHEN entry_type = 'payment' THEN amount ELSE 0 END), 0)::numeric(12,2),
			COALESCE(SUM(CASE WHEN entry_type = 'credit_applied' THEN amount ELSE 0 END), 0)::numeric(12,2)
		FROM ledger_entries
		WHERE school_id = $1 AND student_id = $2`,
		schoolID, studentID,
	).Scan(&entriesBilled, &entriesPaid, &entriesCreditApplied)
	if err != nil {
		t.Fatalf("query ledger entries: %v", err)
	}

	if billed != entriesBilled {
		t.Fatalf("reconciliation mismatch: summary billed=%s vs entries(charge+adjustment)=%s", billed, entriesBilled)
	}
	if paid != entriesPaid {
		t.Fatalf("reconciliation mismatch: summary paid=%s vs entries(payment)=%s", paid, entriesPaid)
	}
	wantNet := opening + billed - paid
	if wantNet >= 0 {
		if current != wantNet || credit != 0 {
			t.Fatalf("reconciliation mismatch: want current=%s credit=0, got current=%s credit=%s", wantNet, current, credit)
		}
	} else {
		if current != 0 || credit != -wantNet {
			t.Fatalf("reconciliation mismatch: want current=0 credit=%s, got current=%s credit=%s", -wantNet, current, credit)
		}
	}
	return opening, billed, paid, current, credit
}

func TestWorkedExampleCreditCarryForwardAndReconciliation(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()

	var term2ID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current)
		VALUES($1,$2,2,'2026-05-04','2026-08-07',false)
		RETURNING id`,
		school.ID, school.YearID,
	).Scan(&term2ID); err != nil {
		t.Fatal(err)
	}

	svc := &billing.Service{DB: db}
	admin := contracts.Principal{
		UserID:       school.UserID,
		SchoolID:     school.ID,
		MembershipID: school.MembershipID,
		Role:         "ikonex_admin",
	}
	rawAmt := func(s string) json.RawMessage { return json.RawMessage(`"` + s + `"`) }

	// Step 1: Term 1 initial charge 10,000.00
	t1Struct, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
	})
	if err != nil {
		t.Fatalf("load term 1 structure: %v", err)
	}
	_, billed, paid, current, credit := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "10000.00" || paid.String() != "0.00" || current.String() != "10000.00" || credit.String() != "0.00" {
		t.Fatalf("unexpected after initial charge: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// Step 2: Payment 15,000.00 -> billed=10000.00, paid=15000.00, current=0.00, credit=5000.00
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentTx(ctx, tx, school.ID, school.StudentID, "RCPT-WORKED-1", "M-Pesa payment", 1500000)
	}); err != nil {
		t.Fatalf("record payment: %v", err)
	}
	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "10000.00" || paid.String() != "15000.00" || current.String() != "0.00" || credit.String() != "5000.00" {
		t.Fatalf("worked example after 15k payment: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// Step 3: Next term charge 8,000.00 -> billed=18000.00, paid=15000.00, current=3000.00, credit=0.00
	t2Struct, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         term2ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Term 2 Tuition", Amount: rawAmt("8000.00")}},
	})
	if err != nil {
		t.Fatalf("load term 2 structure: %v", err)
	}
	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "18000.00" || paid.String() != "15000.00" || current.String() != "3000.00" || credit.String() != "0.00" {
		t.Fatalf("worked example after 8k next term charge: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// Verify informational credit_applied entry = 5,000.00
	var creditEntriesCount int
	var creditEntryAmount billing.Money
	if err := db.Admin.QueryRow(ctx, `
		SELECT count(*), COALESCE(SUM(amount), 0)::numeric(12,2)
		FROM ledger_entries
		WHERE school_id = $1 AND student_id = $2 AND entry_type = 'credit_applied'`,
		school.ID, school.StudentID,
	).Scan(&creditEntriesCount, &creditEntryAmount); err != nil {
		t.Fatal(err)
	}
	if creditEntriesCount != 1 || creditEntryAmount.String() != "5000.00" {
		t.Fatalf("expected 1 credit_applied entry of 5000.00, got count=%d amount=%s", creditEntriesCount, creditEntryAmount)
	}

	// Step 4: Exact replay of ChargeGradeForTerm and ChargeStudent must not change anything
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		if err := svc.ChargeGradeForTerm(ctx, tx, t1Struct.ID); err != nil {
			return err
		}
		if err := svc.ChargeGradeForTerm(ctx, tx, t2Struct.ID); err != nil {
			return err
		}
		return svc.ChargeStudent(ctx, tx, school.StudentID, t2Struct.FeePeriodID)
	}); err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "18000.00" || paid.String() != "15000.00" || current.String() != "3000.00" || credit.String() != "0.00" {
		t.Fatalf("replay mutated balances: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}
}

func TestFullCreditAbsorptionAndEligibilityRules(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()

	// Create Grade 8 + student in Grade 8 + student without active enrolment
	var grade8ID, class8ID, studentGrade8ID, studentInactiveID, term2ID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO grade_levels(school_id,number,phase,name) VALUES($1,8,'junior','Grade 8') RETURNING id`, school.ID).Scan(&grade8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO classes(school_id,grade_level_id,stream_name) VALUES($1,$2,'East') RETURNING id`, school.ID, grade8ID).Scan(&class8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'G8-001','Grade 8 Student','2012-02-02','male',$2) RETURNING id`, school.ID, class8ID).Scan(&studentGrade8ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, studentGrade8ID, school.YearID, grade8ID, class8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'INACT-001','Inactive Student','2013-03-03','female',$2) RETURNING id`, school.ID, school.ClassID).Scan(&studentInactiveID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'withdrawn')`, school.ID, studentInactiveID, school.YearID, school.GradeID, school.ClassID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,2,'2026-05-04','2026-08-07',false) RETURNING id`, school.ID, school.YearID).Scan(&term2ID); err != nil {
		t.Fatal(err)
	}

	svc := &billing.Service{DB: db}
	admin := contracts.Principal{UserID: school.UserID, SchoolID: school.ID, MembershipID: school.MembershipID, Role: "ikonex_admin"}
	rawAmt := func(s string) json.RawMessage { return json.RawMessage(`"` + s + `"`) }

	// Load Grade 7 Term 1 (5,000.00) and Grade 8 Term 1 (7,500.00)
	if _, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition G7", Amount: rawAmt("5000.00")}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   grade8ID,
		Items:          []billing.FeeItemInput{{Title: "Tuition G8", Amount: rawAmt("7500.00")}},
	}); err != nil {
		t.Fatal(err)
	}

	// Verify inactive student was skipped (0 charges)
	var inactiveCharges int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_charges WHERE school_id = $1 AND student_id = $2`, school.ID, studentInactiveID).Scan(&inactiveCharges); err != nil || inactiveCharges != 0 {
		t.Fatalf("expected 0 charges for inactive student, got %d (%v)", inactiveCharges, err)
	}

	// Verify Grade 8 student charged 7,500.00 and Grade 7 student charged 5,000.00
	_, billedG8, _, currentG8, _ := assertStudentReconciliation(t, ctx, db, school.ID, studentGrade8ID)
	if billedG8.String() != "7500.00" || currentG8.String() != "7500.00" {
		t.Fatalf("grade 8 student expected 7500.00, got billed=%s current=%s", billedG8, currentG8)
	}

	// Full credit absorption on Grade 7 student: pay 20,000.00 (credit = 15,000.00), then charge Term 2 = 6,000.00 -> credit = 9,000.00, current = 0.00, credit_applied = 6,000.00
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentTx(ctx, tx, school.ID, school.StudentID, "RCPT-FULL-1", "Overpayment", 2000000)
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         term2ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition G7 T2", Amount: rawAmt("6000.00")}},
	}); err != nil {
		t.Fatal(err)
	}
	_, billedG7, paidG7, currentG7, creditG7 := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billedG7.String() != "11000.00" || paidG7.String() != "20000.00" || currentG7.String() != "0.00" || creditG7.String() != "9000.00" {
		t.Fatalf("full credit absorption mismatch: billed=%s paid=%s current=%s credit=%s", billedG7, paidG7, currentG7, creditG7)
	}
	var absorbed billing.Money
	if err := db.Admin.QueryRow(ctx, `SELECT amount FROM ledger_entries WHERE school_id=$1 AND student_id=$2 AND entry_type='credit_applied'`, school.ID, school.StudentID).Scan(&absorbed); err != nil || absorbed.String() != "6000.00" {
		t.Fatalf("expected credit_applied=6000.00, got %s (%v)", absorbed, err)
	}
}

func TestConcurrentTermCharging20Goroutines(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()

	svc := &billing.Service{DB: db}
	// Pre-fund the student with 3,000.00 credit before the 20 concurrent term charges hit
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentTx(ctx, tx, school.ID, school.StudentID, "RCPT-PREFUND", "Prefund", 300000)
	}); err != nil {
		t.Fatal(err)
	}

	// Create structure in an initially non-current year so we can trigger 20 concurrent ChargeGradeForTerm calls simultaneously
	var year2ID, term2ID, structureID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2027,false) RETURNING id`, school.ID).Scan(&year2ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,1,'2027-01-04','2027-04-09',false) RETURNING id`, school.ID, year2ID).Scan(&term2ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, school.StudentID, year2ID, school.GradeID, school.ClassID); err != nil {
		t.Fatal(err)
	}

	admin := contracts.Principal{UserID: school.UserID, SchoolID: school.ID, MembershipID: school.MembershipID, Role: "ikonex_admin"}
	loaded, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: year2ID,
		TermID:         term2ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: json.RawMessage(`"10000.00"`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	structureID = loaded.ID

	const workers = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
				if idx%2 == 0 {
					return svc.ChargeGradeForTerm(ctx, tx, structureID)
				}
				return svc.ChargeStudent(ctx, tx, school.StudentID, loaded.FeePeriodID)
			})
			if err != nil {
				errs <- fmt.Errorf("worker %d: %w", idx, err)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent charge error: %v", err)
	}

	var chargeCount, chargeEntryCount, creditEntryCount int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_charges WHERE school_id=$1 AND student_id=$2`, school.ID, school.StudentID).Scan(&chargeCount); err != nil || chargeCount != 1 {
		t.Fatalf("expected 1 fee_charges row, got %d (%v)", chargeCount, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM ledger_entries WHERE school_id=$1 AND student_id=$2 AND entry_type='charge'`, school.ID, school.StudentID).Scan(&chargeEntryCount); err != nil || chargeEntryCount != 1 {
		t.Fatalf("expected 1 charge ledger entry, got %d (%v)", chargeEntryCount, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM ledger_entries WHERE school_id=$1 AND student_id=$2 AND entry_type='credit_applied'`, school.ID, school.StudentID).Scan(&creditEntryCount); err != nil || creditEntryCount != 1 {
		t.Fatalf("expected 1 credit_applied ledger entry, got %d (%v)", creditEntryCount, err)
	}

	_, billed, paid, current, credit := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "10000.00" || paid.String() != "3000.00" || current.String() != "7000.00" || credit.String() != "0.00" {
		t.Fatalf("unexpected post-concurrency balances: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}
}
