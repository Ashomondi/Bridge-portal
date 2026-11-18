package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/academics"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func rawAmount(s string) json.RawMessage {
	return json.RawMessage(`"` + s + `"`)
}

func TestAuthorizationMatrixAndFeePayerVisibility(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()
	svc := newTestBillingService(db)

	feeManager := seedMemberWithPermission(t, db, schoolA.ID, "+254788000001", "Fee Manager", "bursar", billing.PermFeeStructuresManage)
	bursarReader := seedMemberWithPermission(t, db, schoolA.ID, "+254788000002", "Bursar Reader", "bursar", billing.PermLedgersView)
	bursarNoBadge := seedMemberWithPermission(t, db, schoolA.ID, "+254788000003", "Bursar No Badge", "bursar", "")
	// Even with an invented role string on Principal, missing badge permission must be denied.
	fakePlatformRole := bursarNoBadge
	fakePlatformRole.Role = "ikonex_admin"

	// Expired badge must be denied.
	expiredBadgeMember := seedMemberWithPermission(t, db, schoolA.ID, "+254788000004", "Expired Badge", "bursar", billing.PermLedgersView)
	if _, err := db.Admin.Exec(ctx, `UPDATE membership_badges SET valid_until = NOW() - interval '1 hour' WHERE membership_id = $1`, expiredBadgeMember.MembershipID); err != nil {
		t.Fatal(err)
	}

	// Inactive membership must be denied even with a valid badge.
	inactiveMember := seedMemberWithPermission(t, db, schoolA.ID, "+254788000005", "Suspended Staff", "bursar", billing.PermFeeStructuresManage)
	if _, err := db.Admin.Exec(ctx, `UPDATE memberships SET status = 'inactive' WHERE id = $1`, inactiveMember.MembershipID); err != nil {
		t.Fatal(err)
	}

	// Non-fee-paying guardian and fee-paying guardian for schoolA.StudentID.
	feePayerGuardian := contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "parent",
	}
	nonFeePayerGuardian := seedMemberWithPermission(t, db, schoolA.ID, "+254788000006", "Non Fee Payer Guardian", "parent", "")
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO student_guardians (school_id, student_id, user_id, relationship, is_primary, is_fee_payer)
		VALUES ($1, $2, $3, 'mother', false, false)`,
		schoolA.ID, schoolA.StudentID, nonFeePayerGuardian.UserID,
	); err != nil {
		t.Fatal(err)
	}
	unlinkedGuardian := seedMemberWithPermission(t, db, schoolA.ID, "+254788000007", "Unlinked Parent", "parent", "")
	studentRoleUser := seedMemberWithPermission(t, db, schoolA.ID, "+254788000008", "Other Role User", "teacher", "")

	loadInput := billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmount("10000.00")}},
	}

	// Denied callers for LoadFeeStructure:
	for _, tc := range []struct {
		name      string
		principal contracts.Principal
		wantCode  int
	}{
		{"missing principal", contracts.Principal{}, http.StatusUnauthorized},
		{"bursar with only ledgers.view", bursarReader, http.StatusForbidden},
		{"bursar without badge", bursarNoBadge, http.StatusForbidden},
		{"invented ikonex_admin role without badge", fakePlatformRole, http.StatusForbidden},
		{"inactive membership", inactiveMember, http.StatusForbidden},
		{"fee payer guardian", feePayerGuardian, http.StatusForbidden},
	} {
		t.Run("LoadFeeStructure/"+tc.name, func(t *testing.T) {
			_, err := svc.LoadFeeStructure(ctx, tc.principal, loadInput)
			var apiErr *contracts.APIError
			if !errors.As(err, &apiErr) || apiErr.Status != tc.wantCode {
				t.Fatalf("got %v, want HTTP %d", err, tc.wantCode)
			}
		})
	}

	// Authorized feeManager loads the structure.
	if _, err := svc.LoadFeeStructure(ctx, feeManager, loadInput); err != nil {
		t.Fatalf("authorized feeManager LoadFeeStructure failed: %v", err)
	}

	// Balance & ledger read matrix:
	for _, tc := range []struct {
		name      string
		principal contracts.Principal
		studentID string
		wantCode  int
	}{
		{"fee-paying guardian allowed", feePayerGuardian, schoolA.StudentID, http.StatusOK},
		{"bursar with ledgers.view allowed", bursarReader, schoolA.StudentID, http.StatusOK},
		{"non-fee-payer guardian forbidden", nonFeePayerGuardian, schoolA.StudentID, http.StatusForbidden},
		{"unlinked guardian forbidden", unlinkedGuardian, schoolA.StudentID, http.StatusForbidden},
		{"bursar without badge forbidden", bursarNoBadge, schoolA.StudentID, http.StatusForbidden},
		{"expired badge forbidden", expiredBadgeMember, schoolA.StudentID, http.StatusForbidden},
		{"feeManager without ledgers.view forbidden", feeManager, schoolA.StudentID, http.StatusForbidden},
		{"teacher without badge forbidden", studentRoleUser, schoolA.StudentID, http.StatusForbidden},
		{"cross-school access returns 404", bursarReader, schoolB.StudentID, http.StatusNotFound},
		{"missing principal returns 401", contracts.Principal{}, schoolA.StudentID, http.StatusUnauthorized},
	} {
		t.Run("BalanceAndLedger/"+tc.name, func(t *testing.T) {
			_, balErr := svc.GetStudentFeeBalance(ctx, tc.principal, tc.studentID)
			_, ledErr := svc.GetStudentLedger(ctx, tc.principal, tc.studentID)
			for label, err := range map[string]error{"balance": balErr, "ledger": ledErr} {
				if tc.wantCode == http.StatusOK {
					if err != nil {
						t.Fatalf("%s unexpected error: %v", label, err)
					}
					continue
				}
				var apiErr *contracts.APIError
				if !errors.As(err, &apiErr) || apiErr.Status != tc.wantCode {
					t.Fatalf("%s got %v, want HTTP %d", label, err, tc.wantCode)
				}
			}
		})
	}
}

func TestBoardingVariantsRejectedAndAllDeterministic(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()
	svc := newTestBillingService(db)
	feeManager := seedMemberWithPermission(t, db, school.ID, "+254788000011", "Fee Manager", "bursar", billing.PermFeeStructuresManage)

	// Give the school a dormitory and a second enrolled boarder student.
	var dormID, boarderID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO dormitories(school_id,name,gender,capacity) VALUES($1,'Kilimanjaro','female',60) RETURNING id`, school.ID).Scan(&dormID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id,current_dormitory_id)
		VALUES($1,'BRD-001','Boarder Student','2013-04-04','female',$2,$3)
		RETURNING id`,
		school.ID, school.ClassID, dormID,
	).Scan(&boarderID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status)
		VALUES($1,$2,$3,$4,$5,'active')`,
		school.ID, boarderID, school.YearID, school.GradeID, school.ClassID,
	); err != nil {
		t.Fatal(err)
	}

	// Release 1 rejects boarder and day variants deterministically.
	for _, variant := range []string{"boarder", "day", "BOARDER"} {
		_, err := svc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
			AcademicYearID: school.YearID,
			TermID:         school.TermID,
			GradeLevelID:   school.GradeID,
			BoardingStatus: variant,
			Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmount("15000.00")}},
		})
		var apiErr *contracts.APIError
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnprocessableEntity || apiErr.Code != "UNSUPPORTED_BOARDING_VARIANT" {
			t.Fatalf("variant %q: expected 422 UNSUPPORTED_BOARDING_VARIANT, got %v", variant, err)
		}
	}

	// Default / 'all' charges both day and dormitory-assigned students identically.
	loaded, err := svc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmount("11500.00")}},
	})
	if err != nil {
		t.Fatalf("LoadFeeStructure all failed: %v", err)
	}
	if loaded.BoardingStatus != "all" {
		t.Fatalf("expected boarding_status='all', got %q", loaded.BoardingStatus)
	}
	for _, sid := range []string{school.StudentID, boarderID} {
		_, billed, _, current, _ := assertStudentReconciliation(t, ctx, db, school.ID, sid)
		if billed.String() != "11500.00" || current.String() != "11500.00" {
			t.Fatalf("student %s expected 11500.00, got billed=%s current=%s", sid, billed, current)
		}
	}
}

func TestPreloadedStructuresReplayAndChargeCurrentTrigger(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()
	svc := newTestBillingService(db)
	feeManager := seedMemberWithPermission(t, db, school.ID, "+254788000021", "Fee Manager", "bursar", billing.PermFeeStructuresManage)

	var futureYearID, futureTerm1ID, futureTerm2ID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2027,false) RETURNING id`, school.ID).Scan(&futureYearID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,1,'2027-01-04','2027-04-02',false) RETURNING id`, school.ID, futureYearID).Scan(&futureTerm1ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,2,'2027-05-03','2027-08-06',false) RETURNING id`, school.ID, futureYearID).Scan(&futureTerm2ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, school.StudentID, futureYearID, school.GradeID, school.ClassID); err != nil {
		t.Fatal(err)
	}

	term1Input := billing.LoadFeeStructureInput{
		AcademicYearID: futureYearID,
		TermID:         futureTerm1ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition 2027 T1", Amount: rawAmount("12000.00")}},
	}
	term2Input := billing.LoadFeeStructureInput{
		AcademicYearID: futureYearID,
		TermID:         futureTerm2ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition 2027 T2", Amount: rawAmount("9000.00")}},
	}

	// Preload both structures while 2027 is not current -> 0 charges yet.
	if _, err := svc.LoadFeeStructure(ctx, feeManager, term1Input); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoadFeeStructure(ctx, feeManager, term2Input); err != nil {
		t.Fatal(err)
	}
	var chargeCount int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_charges WHERE school_id=$1 AND student_id=$2`, school.ID, school.StudentID).Scan(&chargeCount); err != nil || chargeCount != 0 {
		t.Fatalf("expected 0 charges for non-current preloaded structures, got %d (%v)", chargeCount, err)
	}

	// Now Term 1 of 2027 becomes current.
	if _, err := db.Admin.Exec(ctx, `UPDATE terms SET is_current = true WHERE id = $1`, futureTerm1ID); err != nil {
		t.Fatal(err)
	}

	// Replaying identical Term 1 structure MUST now trigger charging!
	if _, err := svc.LoadFeeStructure(ctx, feeManager, term1Input); err != nil {
		t.Fatalf("identical replay in now-current term failed: %v", err)
	}
	_, billed, _, current, _ := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "12000.00" || current.String() != "12000.00" {
		t.Fatalf("expected replay to charge 12000.00, got billed=%s current=%s", billed, current)
	}

	// Now Term 2 becomes current and production trigger ChargeCurrentTermStructures (HTTP route) runs.
	if _, err := db.Admin.Exec(ctx, `UPDATE terms SET is_current = true WHERE id = $1`, futureTerm2ID); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return feeManager, nil
	})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/billing/fee-structures/charge-current", bytes.NewBuffer(nil)))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST charge-current returned %d: %s", rec.Code, rec.Body.String())
	}
	_, billed, _, current, _ = assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "21000.00" || current.String() != "21000.00" {
		t.Fatalf("expected charge-current to add Term 2 9000.00, got billed=%s current=%s", billed, current)
	}

	// Repeat call to charge-current must be idempotent (0 additional charges).
	res, err := svc.ChargeCurrentTermStructures(ctx, feeManager)
	if err != nil || res.ChargesCreated != 0 {
		t.Fatalf("expected 0 new charges on repeat charge-current, got %+v (%v)", res, err)
	}
}

func TestPaymentAllocationAndPerTermBalancePreservation(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()
	svc := newTestBillingService(db)

	feeManager := seedMemberWithPermission(t, db, school.ID, "+254788000031", "Fee Manager", "bursar", billing.PermFeeStructuresManage)
	bursar := seedMemberWithPermission(t, db, school.ID, "+254788000032", "Bursar", "bursar", billing.PermLedgersView)

	var term2ID, term3ID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,2,'2026-05-04','2026-08-07',true) RETURNING id`, school.ID, school.YearID).Scan(&term2ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,3,'2026-09-07','2026-11-20',true) RETURNING id`, school.ID, school.YearID).Scan(&term3ID); err != nil {
		t.Fatal(err)
	}

	// Precharge all 3 terms: T1=10,000, T2=8,000, T3=6,000 (total 24,000).
	for _, spec := range []struct {
		termID string
		title  string
		amt    string
	}{
		{school.TermID, "T1 Tuition", "10000.00"},
		{term2ID, "T2 Tuition", "8000.00"},
		{term3ID, "T3 Tuition", "6000.00"},
	} {
		if _, err := svc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
			AcademicYearID: school.YearID,
			TermID:         spec.termID,
			GradeLevelID:   school.GradeID,
			Items:          []billing.FeeItemInput{{Title: spec.title, Amount: rawAmount(spec.amt)}},
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Verify all 3 historical per-term balances are preserved (not zeroed into last term!).
	bal, err := svc.GetStudentFeeBalance(ctx, bursar, school.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bal.Terms) != 3 ||
		bal.Terms[0].CurrentBalance.String() != "10000.00" ||
		bal.Terms[1].CurrentBalance.String() != "8000.00" ||
		bal.Terms[2].CurrentBalance.String() != "6000.00" {
		t.Fatalf("expected preserved per-term balances [10000, 8000, 6000], got %+v", bal.Terms)
	}

	// 1. Explicit partial payment of 4,000.00 intended for Term 1.
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentForTermTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, school.TermID, "PAY-T1-PART", "Partial Term 1", 400000)
	}); err != nil {
		t.Fatal(err)
	}
	// 2. Explicit payment of 8,000.00 intended for Term 2.
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentForTermTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, term2ID, "PAY-T2-FULL", "Full Term 2", 800000)
	}); err != nil {
		t.Fatal(err)
	}
	bal, err = svc.GetStudentFeeBalance(ctx, bursar, school.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if bal.Terms[0].CurrentBalance.String() != "6000.00" ||
		bal.Terms[1].CurrentBalance.String() != "0.00" ||
		bal.Terms[2].CurrentBalance.String() != "6000.00" ||
		bal.CurrentBalance.String() != "12000.00" {
		t.Fatalf("unexpected per-term balances after explicit T1+T2 payments: %+v", bal)
	}

	// 3. Unallocated payment of 9,000.00 -> oldest-due-first settles Term 1 (6,000) and applies 3,000 to Term 3 (3,000 left).
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, "PAY-UNALLOC-1", "Unallocated", 900000)
	}); err != nil {
		t.Fatal(err)
	}
	bal, err = svc.GetStudentFeeBalance(ctx, bursar, school.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if bal.Terms[0].CurrentBalance.String() != "0.00" ||
		bal.Terms[1].CurrentBalance.String() != "0.00" ||
		bal.Terms[2].CurrentBalance.String() != "3000.00" ||
		bal.CurrentBalance.String() != "3000.00" {
		t.Fatalf("expected oldest-due-first [0, 0, 3000], got %+v", bal.Terms)
	}

	// 4. Unallocated overpayment of 10,000.00 -> settles Term 3 (3,000) and keeps 7,000 credit.
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, "PAY-OVER-1", "Overpayment", 1000000)
	}); err != nil {
		t.Fatal(err)
	}
	_, billed, paid, current, credit := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "24000.00" || paid.String() != "31000.00" || current.String() != "0.00" || credit.String() != "7000.00" {
		t.Fatalf("unexpected totals after overpayment: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}

	// 5. Add Year 2 Term 1 (5,000.00), then replay PAY-UNALLOC-1 and PAY-OVER-1 -> replay must be idempotent,
	// and Year 2 Term 1 must have transferred 5,000.00 credit leaving 2,000.00 credit.
	var year2ID, y2Term1ID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2027,true) RETURNING id`, school.ID).Scan(&year2ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,1,'2027-01-05','2027-04-02',true) RETURNING id`, school.ID, year2ID).Scan(&y2Term1ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, school.StudentID, year2ID, school.GradeID, school.ClassID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
		AcademicYearID: year2ID,
		TermID:         y2Term1ID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Y2 T1 Tuition", Amount: rawAmount("5000.00")}},
	}); err != nil {
		t.Fatal(err)
	}

	// Replay payment references after new term charge: must do nothing!
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		if err := svc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, "PAY-UNALLOC-1", "Replay", 900000); err != nil {
			return err
		}
		return svc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, bursar.UserID, school.StudentID, "PAY-OVER-1", "Replay", 1000000)
	}); err != nil {
		t.Fatal(err)
	}
	_, billed, paid, current, credit = assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billed.String() != "29000.00" || paid.String() != "31000.00" || current.String() != "0.00" || credit.String() != "2000.00" {
		t.Fatalf("unexpected totals after Y2 charge + payment replay: billed=%s paid=%s current=%s credit=%s", billed, paid, current, credit)
	}
}

func TestAcademicsTermDateSyncAtomicityAndNotifications(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	// Seed a fee structure & charge in School A Term 1, plus an unaffected Term 2.
	billingSvc := newTestBillingService(db)
	notifSvc := &notifications.Service{DB: contracts.TenantTxFunc(db.WithTenantTx)}
	billingSvc.Notifier = notifSvc

	feeManagerA := seedMemberWithPermission(t, db, schoolA.ID, "+254788000041", "Fee Manager A", "bursar", billing.PermFeeStructuresManage)
	feeManagerB := seedMemberWithPermission(t, db, schoolB.ID, "+254788000042", "Fee Manager B", "bursar", billing.PermFeeStructuresManage)
	deanA := seedMemberWithPermission(t, db, schoolA.ID, "+254788000043", "Dean A", academics.RoleDean, academics.PermCalendarManage)

	// Add a non-fee-payer guardian for schoolA.StudentID (must NOT receive fee notice).
	nonFeePayer := seedMemberWithPermission(t, db, schoolA.ID, "+254788000044", "Non Fee Payer", "parent", "")
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO student_guardians(school_id,student_id,user_id,relationship,is_fee_payer)
		VALUES($1,$2,$3,'aunt',false)`,
		schoolA.ID, schoolA.StudentID, nonFeePayer.UserID,
	); err != nil {
		t.Fatal(err)
	}

	var term2AID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,2,'2026-05-04','2026-08-07',false) RETURNING id`, schoolA.ID, schoolA.YearID).Scan(&term2AID); err != nil {
		t.Fatal(err)
	}

	if _, err := billingSvc.LoadFeeStructure(ctx, feeManagerA, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition A", Amount: rawAmount("10000.00")}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := billingSvc.LoadFeeStructure(ctx, feeManagerA, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         term2AID,
		GradeLevelID:   schoolA.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition A T2", Amount: rawAmount("8000.00")}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := billingSvc.LoadFeeStructure(ctx, feeManagerB, billing.LoadFeeStructureInput{
		AcademicYearID: schoolB.YearID,
		TermID:         schoolB.TermID,
		GradeLevelID:   schoolB.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition B", Amount: rawAmount("9000.00")}},
	}); err != nil {
		t.Fatal(err)
	}

	sqlDB := stdlib.OpenDBFromPool(db.Admin)
	t.Cleanup(func() { _ = sqlDB.Close() })
	acadSvc := academics.NewService(academics.Deps{
		Store:     academics.NewSQLStore(sqlDB),
		Notifier:  notifSvc,
		FeeSyncer: billingSvc,
	})
	deanCtx := academics.WithIdentity(ctx, academics.Identity{
		UserID:       deanA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: deanA.MembershipID,
		Role:         academics.RoleDean,
	})

	// 1. Update Term 1 start date via academics.PutCalendar -> updates fee_periods, fee_charges, fee_ledgers,
	// creates fee_change_notes + fee_change_note_students, notifies only fee-paying guardian, and writes audit_log.
	cal, err := acadSvc.PutCalendar(deanCtx, academics.CalendarRequest{
		TermUpdates: []academics.TermUpdateInput{{
			TermID:        schoolA.TermID,
			StartDate:     "2026-01-19",
			EndDate:       "2026-04-10",
			FeeChangeNote: "Delayed opening by two weeks",
		}},
	})
	if err != nil {
		t.Fatalf("PutCalendar failed: %v", err)
	}
	if !cal.FeeDueDatesShifted {
		t.Fatalf("expected FeeDueDatesShifted=true")
	}

	var periodDue, chargeDue, ledgerDue string
	if err := db.Admin.QueryRow(ctx, `SELECT due_date::text FROM fee_periods WHERE school_id=$1 AND term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&periodDue); err != nil || periodDue != "2026-01-19" {
		t.Fatalf("fee_periods due_date=%s (%v)", periodDue, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT fc.due_date::text FROM fee_charges fc JOIN fee_periods fp ON fp.id=fc.fee_period_id WHERE fc.school_id=$1 AND fp.term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&chargeDue); err != nil || chargeDue != "2026-01-19" {
		t.Fatalf("fee_charges due_date=%s (%v)", chargeDue, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT due_date::text FROM fee_ledgers WHERE school_id=$1 AND term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&ledgerDue); err != nil || ledgerDue != "2026-01-19" {
		t.Fatalf("fee_ledgers due_date=%s (%v)", ledgerDue, err)
	}

	// Unaffected Term 2 and School B must retain original dates.
	var t2Due, bDue string
	if err := db.Admin.QueryRow(ctx, `SELECT due_date::text FROM fee_periods WHERE school_id=$1 AND term_id=$2`, schoolA.ID, term2AID).Scan(&t2Due); err != nil || t2Due != "2026-05-04" {
		t.Fatalf("unaffected term2 due_date=%s (%v)", t2Due, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT due_date::text FROM fee_periods WHERE school_id=$1 AND term_id=$2`, schoolB.ID, schoolB.TermID).Scan(&bDue); err != nil || bDue != "2026-01-05" {
		t.Fatalf("schoolB due_date mutated: %s (%v)", bDue, err)
	}

	// Only the fee-paying guardian received a notification (not nonFeePayer).
	var feePayerNotices, nonPayerNotices int
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2 AND category='fee_due_date_change'`, schoolA.ID, schoolA.UserID).Scan(&feePayerNotices)
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2 AND category='fee_due_date_change'`, schoolA.ID, nonFeePayer.UserID).Scan(&nonPayerNotices)
	if feePayerNotices != 1 || nonPayerNotices != 0 {
		t.Fatalf("expected 1 fee-payer notice and 0 non-fee-payer notices, got %d / %d", feePayerNotices, nonPayerNotices)
	}

	// 2. Rollback test: if a mandatory guardian notification fails during PutCalendar,
	// BOTH terms.start_date AND billing tables (fee_periods, fee_charges, fee_ledgers,
	// fee_change_notes, audit_log) must roll back atomically.
	failingNotifSvc := &billing.Service{
		DB:       contracts.TenantTxFunc(db.WithTenantTx),
		Audit:    billing.NewPostgresAuditLogger(),
		Notifier: failingNotifier{},
	}
	failingAcad := academics.NewService(academics.Deps{
		Store:     academics.NewSQLStore(sqlDB),
		Notifier:  notifSvc,
		FeeSyncer: failingNotifSvc,
	})
	_, err = failingAcad.PutCalendar(deanCtx, academics.CalendarRequest{
		TermUpdates: []academics.TermUpdateInput{{
			TermID:        schoolA.TermID,
			StartDate:     "2026-01-26",
			EndDate:       "2026-04-10",
			FeeChangeNote: "Should roll back on notification failure",
		}},
	})
	if err == nil {
		t.Fatal("expected PutCalendar to fail when mandatory guardian notification fails")
	}
	var rolledBackTermStart, rolledBackPeriodDue, rolledBackChargeDue string
	var noteCount int
	_ = db.Admin.QueryRow(ctx, `SELECT start_date::text FROM terms WHERE id=$1`, schoolA.TermID).Scan(&rolledBackTermStart)
	_ = db.Admin.QueryRow(ctx, `SELECT due_date::text FROM fee_periods WHERE school_id=$1 AND term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&rolledBackPeriodDue)
	_ = db.Admin.QueryRow(ctx, `SELECT fc.due_date::text FROM fee_charges fc JOIN fee_periods fp ON fp.id=fc.fee_period_id WHERE fc.school_id=$1 AND fp.term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&rolledBackChargeDue)
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_change_notes WHERE school_id=$1 AND term_id=$2`, schoolA.ID, schoolA.TermID).Scan(&noteCount)
	if rolledBackTermStart != "2026-01-19" || rolledBackPeriodDue != "2026-01-19" || rolledBackChargeDue != "2026-01-19" || noteCount != 1 {
		t.Fatalf("rollback failed: termStart=%s periodDue=%s chargeDue=%s noteCount=%d", rolledBackTermStart, rolledBackPeriodDue, rolledBackChargeDue, noteCount)
	}

	// 3. Academics service without FeeSyncer must fail instead of falsely returning FeeDueDatesShifted=true.
	unwiredAcad := academics.NewService(academics.Deps{Store: academics.NewSQLStore(sqlDB)})
	if _, err := unwiredAcad.PutCalendar(deanCtx, academics.CalendarRequest{
		TermUpdates: []academics.TermUpdateInput{{
			TermID:        schoolA.TermID,
			StartDate:     "2026-01-26",
			EndDate:       "2026-04-10",
			FeeChangeNote: "No syncer configured",
		}},
	}); !errors.Is(err, academics.ErrFeeSyncerUnavailable) {
		t.Fatalf("expected ErrFeeSyncerUnavailable, got %v", err)
	}
}

func TestMandatoryAuditLoggerAndRollback(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()

	feeManager := seedMemberWithPermission(t, db, school.ID, "+254788000051", "Fee Manager", "bursar", billing.PermFeeStructuresManage)

	// Missing AuditLogger must fail explicitly on financial mutations.
	noAuditSvc := &billing.Service{DB: contracts.TenantTxFunc(db.WithTenantTx)}
	_, err := noAuditSvc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmount("10000.00")}},
	})
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "AUDIT_LOGGER_REQUIRED" {
		t.Fatalf("expected AUDIT_LOGGER_REQUIRED on LoadFeeStructure, got %v", err)
	}
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return noAuditSvc.RecordStudentPaymentTx(ctx, tx, school.ID, school.StudentID, "NO-AUDIT-1", "Payment", 100000)
	}); !errors.As(err, &apiErr) || apiErr.Code != "AUDIT_LOGGER_REQUIRED" {
		t.Fatalf("expected AUDIT_LOGGER_REQUIRED on RecordStudentPaymentTx, got %v", err)
	}
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return noAuditSvc.OnTermStartDateChangedTx(ctx, tx, school.ID, school.TermID, time.Now(), feeManager.UserID, "Note")
	}); !errors.As(err, &apiErr) || apiErr.Code != "AUDIT_LOGGER_REQUIRED" {
		t.Fatalf("expected AUDIT_LOGGER_REQUIRED on OnTermStartDateChangedTx, got %v", err)
	}

	// Real PostgresAuditLogger writes durable audit_log rows with actor & tenant attribution.
	svc := newTestBillingService(db)
	if _, err := svc.LoadFeeStructure(ctx, feeManager, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmount("10000.00")}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, feeManager.UserID, school.StudentID, "AUD-PAY-1", "Desk payment", 500000)
	}); err != nil {
		t.Fatal(err)
	}

	var loadAudits, chargeAudits, paymentAudits int
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE school_id=$1 AND user_id=$2 AND action='billing.fee_structure.load'`, school.ID, feeManager.UserID).Scan(&loadAudits)
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE school_id=$1 AND user_id=$2 AND action='billing.student.charge'`, school.ID, feeManager.UserID).Scan(&chargeAudits)
	_ = db.Admin.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE school_id=$1 AND user_id=$2 AND action='billing.payment.record'`, school.ID, feeManager.UserID).Scan(&paymentAudits)
	if loadAudits != 1 || chargeAudits != 1 || paymentAudits != 1 {
		t.Fatalf("expected 1 audit row each, got load=%d charge=%d payment=%d", loadAudits, chargeAudits, paymentAudits)
	}

	// Ensure no phone numbers appear in audit_log details.
	var detailsText string
	if err := db.Admin.QueryRow(ctx, `SELECT details::text FROM audit_log WHERE school_id=$1 AND action='billing.payment.record'`, school.ID).Scan(&detailsText); err != nil || strings.Contains(detailsText, "+254") {
		t.Fatalf("unexpected audit details: %s (%v)", detailsText, err)
	}

	// Failing AuditLogger rolls back payment ledger changes atomically.
	failAuditSvc := &billing.Service{
		DB:    contracts.TenantTxFunc(db.WithTenantTx),
		Audit: failingAuditLogger{},
	}
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return failAuditSvc.RecordStudentPaymentWithActorTx(ctx, tx, school.ID, feeManager.UserID, school.StudentID, "FAIL-AUD-PAY", "Should roll back", 200000)
	}); err == nil {
		t.Fatal("expected failing audit logger to roll back payment")
	}
	var rolledBackPay int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM ledger_entries WHERE school_id=$1 AND reference='FAIL-AUD-PAY'`, school.ID).Scan(&rolledBackPay); err != nil || rolledBackPay != 0 {
		t.Fatalf("expected 0 ledger_entries for rolled back payment, got %d (%v)", rolledBackPay, err)
	}
}

type failingNotifier struct{}

func (failingNotifier) CreateInAppNotification(context.Context, contracts.SQLTx, contracts.InAppNotification) error {
	return errors.New("simulated notification failure")
}
