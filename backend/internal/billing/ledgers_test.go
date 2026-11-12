package billing_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

func TestAdmissionHookAndLedgerAuthorization(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	svc := &billing.Service{DB: db}
	var chargeHook billing.StudentChargeHook = svc
	var billingHook billing.StudentBillingHook = svc

	adminA := contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "ikonex_admin",
	}

	// Load Term 1 fee structure in School A (12,000.00)
	if _, err := svc.LoadFeeStructure(ctx, adminA, billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items: []billing.FeeItemInput{
			{Title: "Tuition", Amount: json.RawMessage(`"10000.00"`)},
			{Title: "Laboratory", Amount: json.RawMessage(`"2000.00"`)},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Admit a second student in School A after structure was already loaded, and invoke StudentChargeHook / StudentBillingHook
	var newStudentID, guardianBUserID, guardianBMembershipID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id)
		VALUES($1,'DEMO-002','Newly Admitted Student','2013-06-15','male',$2)
		RETURNING id`,
		schoolA.ID, schoolA.ClassID,
	).Scan(&newStudentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status)
		VALUES($1,$2,$3,$4,$5,'active')`,
		schoolA.ID, newStudentID, schoolA.YearID, schoolA.GradeID, schoolA.ClassID,
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO users(phone_number,full_name) VALUES('+254711223344','Guardian B') RETURNING id`).Scan(&guardianBUserID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO memberships(school_id,user_id,role) VALUES($1,$2,'parent') RETURNING id`, schoolA.ID, guardianBUserID).Scan(&guardianBMembershipID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'father')`, schoolA.ID, newStudentID, guardianBUserID); err != nil {
		t.Fatal(err)
	}

	if err := db.WithTenantTx(ctx, schoolA.ID, func(tx pgx.Tx) error {
		if err := chargeHook.ChargeAdmittedStudent(ctx, tx, schoolA.ID, newStudentID, schoolA.YearID); err != nil {
			return err
		}
		// Replay via StudentBillingHook alias must be idempotent
		return billingHook.ChargeAdmittedStudent(ctx, tx, schoolA.ID, newStudentID, schoolA.YearID)
	}); err != nil {
		t.Fatalf("ChargeAdmittedStudent failed: %v", err)
	}

	guardianA := contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "parent",
	}
	guardianB := contracts.Principal{
		UserID:       guardianBUserID,
		SchoolID:     schoolA.ID,
		MembershipID: guardianBMembershipID,
		Role:         "parent",
	}
	guardianSchoolB := contracts.Principal{
		UserID:       schoolB.UserID,
		SchoolID:     schoolB.ID,
		MembershipID: schoolB.MembershipID,
		Role:         "parent",
	}

	// Linked Guardian A can view Student 1 balance and ledger
	bal1, err := svc.GetStudentFeeBalance(ctx, guardianA, schoolA.StudentID)
	if err != nil {
		t.Fatalf("Guardian A viewing own child balance: %v", err)
	}
	if bal1.BilledAmount.String() != "12000.00" || bal1.CurrentBalance.String() != "12000.00" || len(bal1.Items) != 2 {
		t.Fatalf("unexpected balance for student 1: %+v", bal1)
	}

	ledger1, err := svc.GetStudentLedger(ctx, guardianA, schoolA.StudentID)
	if err != nil {
		t.Fatalf("Guardian A viewing own child ledger: %v", err)
	}
	if len(ledger1.Entries) != 1 || ledger1.Entries[0].EntryType != "charge" || ledger1.Entries[0].Amount.String() != "12000.00" {
		t.Fatalf("unexpected ledger entries for student 1: %+v", ledger1.Entries)
	}

	// Linked Guardian B can view Newly Admitted Student
	bal2, err := svc.GetStudentFeeBalance(ctx, guardianB, newStudentID)
	if err != nil || bal2.CurrentBalance.String() != "12000.00" {
		t.Fatalf("Guardian B viewing own admitted child: %+v, %v", bal2, err)
	}

	// Unlinked Guardian A CANNOT view Guardian B's child (403 FORBIDDEN)
	_, err = svc.GetStudentFeeBalance(ctx, guardianA, newStudentID)
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 when Guardian A queries Guardian B's child, got %v", err)
	}
	_, err = svc.GetStudentLedger(ctx, guardianA, newStudentID)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 when Guardian A queries Guardian B's child ledger, got %v", err)
	}

	// Cross-school access must fail (404 STUDENT_NOT_FOUND due to tenant isolation)
	_, err = svc.GetStudentFeeBalance(ctx, guardianSchoolB, schoolA.StudentID)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 on cross-school student access, got %v", err)
	}

	// Ikonex team in School A can view any student in School A, but not School B's student
	if _, err := svc.GetStudentLedger(ctx, adminA, newStudentID); err != nil {
		t.Fatalf("expected adminA to view student in schoolA, got %v", err)
	}
	_, err = svc.GetStudentLedger(ctx, adminA, schoolB.StudentID)
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 when adminA queries schoolB student, got %v", err)
	}

	// Verify HTTP endpoints GET /api/v1/students/{id}/fee-balance and GET /api/v1/students/{id}/ledger
	var activePrincipal contracts.Principal
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return activePrincipal, nil
	})

	activePrincipal = guardianA
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/students/"+schoolA.StudentID+"/fee-balance", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET fee-balance returned %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/students/"+newStudentID+"/ledger", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GET unlinked student ledger expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}
