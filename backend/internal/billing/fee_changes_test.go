package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

type recordingAuditLogger struct {
	events []contracts.AuditEvent
}

func (r *recordingAuditLogger) Log(_ context.Context, _ pgx.Tx, ev contracts.AuditEvent) error {
	r.events = append(r.events, ev)
	return nil
}

func TestMidYearFeeChangesAndTermStartDateSync(t *testing.T) {
	db := testenv.New(t, true)
	school := db.SeedSchool(t, "01")
	ctx := context.Background()

	// Create a second student linked to the SAME guardian (to test notification deduplication)
	// and a third student in Grade 8 linked to an unaffected guardian (to test targeted notifications).
	var student2ID, grade8ID, class8ID, student3ID, unaffectedGuardianID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id)
		VALUES($1,'DEMO-SIBLING','Sibling Student','2014-04-04','female',$2)
		RETURNING id`,
		school.ID, school.ClassID,
	).Scan(&student2ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, student2ID, school.YearID, school.GradeID, school.ClassID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'guardian')`, school.ID, student2ID, school.UserID); err != nil {
		t.Fatal(err)
	}

	if err := db.Admin.QueryRow(ctx, `INSERT INTO grade_levels(school_id,number,phase,name) VALUES($1,8,'junior','Grade 8') RETURNING id`, school.ID).Scan(&grade8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO classes(school_id,grade_level_id,stream_name) VALUES($1,$2,'North') RETURNING id`, school.ID, grade8ID).Scan(&class8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'G8-UNAFFECTED','Grade 8 Student','2012-01-01','male',$2) RETURNING id`, school.ID, class8ID).Scan(&student3ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id,status) VALUES($1,$2,$3,$4,$5,'active')`, school.ID, student3ID, school.YearID, grade8ID, class8ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `INSERT INTO users(phone_number,full_name) VALUES('+254799887766','Unaffected Guardian') RETURNING id`).Scan(&unaffectedGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'parent','active')`, school.ID, unaffectedGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'guardian')`, school.ID, student3ID, unaffectedGuardianID); err != nil {
		t.Fatal(err)
	}

	auditLog := &recordingAuditLogger{}
	notifSvc := &notifications.Service{DB: db}
	svc := &billing.Service{
		DB:       db,
		Audit:    auditLog,
		Notifier: notifSvc,
	}

	admin := contracts.Principal{
		UserID:       school.UserID,
		SchoolID:     school.ID,
		MembershipID: school.MembershipID,
		Role:         "ikonex_admin",
	}

	// Load Grade 7 and Grade 8 structures for Term 1
	g7Struct, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   school.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: json.RawMessage(`"10000.00"`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.LoadFeeStructure(ctx, admin, billing.LoadFeeStructureInput{
		AcademicYearID: school.YearID,
		TermID:         school.TermID,
		GradeLevelID:   grade8ID,
		Items:          []billing.FeeItemInput{{Title: "Tuition G8", Amount: json.RawMessage(`"15000.00"`)}},
	}); err != nil {
		t.Fatal(err)
	}

	// 1. Missing or empty note returns 422 FEE_CHANGE_NOTE_REQUIRED
	_, err = svc.ApplyFeeChange(ctx, admin, billing.ApplyFeeChangeInput{
		FeePeriodID:  g7Struct.FeePeriodID,
		GradeLevelID: school.GradeID,
		AmountDelta:  json.RawMessage(`"1500.00"`),
		Note:         "   ",
	})
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusUnprocessableEntity || apiErr.Code != "FEE_CHANGE_NOTE_REQUIRED" {
		t.Fatalf("expected 422 FEE_CHANGE_NOTE_REQUIRED, got %v", err)
	}

	// 2. Non-admin role forbidden
	parentPrincipal := admin
	parentPrincipal.Role = "parent"
	_, err = svc.ApplyFeeChange(ctx, parentPrincipal, billing.ApplyFeeChangeInput{
		FeePeriodID: g7Struct.FeePeriodID,
		AmountDelta: json.RawMessage(`"1500.00"`),
		Note:        "Valid note",
	})
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 FORBIDDEN for parent role, got %v", err)
	}

	// 3. Valid Grade 7 fee increase (+1,500.00) updates both Grade 7 students, leaves Grade 8 untouched,
	// writes fee_change_notes + audit event, and sends exactly ONE notification to the shared Grade 7 guardian
	// and ZERO notifications to the Grade 8 guardian.
	noteView, err := svc.ApplyFeeChange(ctx, admin, billing.ApplyFeeChangeInput{
		FeePeriodID:  g7Struct.FeePeriodID,
		GradeLevelID: school.GradeID,
		AmountDelta:  json.RawMessage(`"1500.00"`),
		Note:         "Approved laboratory levy per BOM resolution",
	})
	if err != nil {
		t.Fatalf("ApplyFeeChange valid failed: %v", err)
	}
	if noteView.AffectedStudentsCount != 2 || noteView.AmountDelta.String() != "1500.00" {
		t.Fatalf("unexpected noteView: %+v", noteView)
	}

	_, billed1, _, current1, _ := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	_, billed2, _, current2, _ := assertStudentReconciliation(t, ctx, db, school.ID, student2ID)
	_, billed3, _, current3, _ := assertStudentReconciliation(t, ctx, db, school.ID, student3ID)
	if billed1.String() != "11500.00" || current1.String() != "11500.00" || billed2.String() != "11500.00" || current2.String() != "11500.00" {
		t.Fatalf("expected Grade 7 students billed=11500.00, got s1=%s s2=%s", billed1, billed2)
	}
	if billed3.String() != "15000.00" || current3.String() != "15000.00" {
		t.Fatalf("expected Grade 8 student untouched at 15000.00, got %s", billed3)
	}

	var affectedGuardianNotices, unaffectedGuardianNotices int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2`, school.ID, school.UserID).Scan(&affectedGuardianNotices); err != nil || affectedGuardianNotices != 1 {
		t.Fatalf("expected exactly 1 deduplicated notification to affected guardian, got %d (%v)", affectedGuardianNotices, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2`, school.ID, unaffectedGuardianID).Scan(&unaffectedGuardianNotices); err != nil || unaffectedGuardianNotices != 0 {
		t.Fatalf("expected 0 notifications to unaffected guardian, got %d (%v)", unaffectedGuardianNotices, err)
	}

	// Verify fee_change_notes appear on affected student's fee-balance, and NOT on unaffected student's fee-balance
	balAffected, err := svc.GetStudentFeeBalance(ctx, admin, school.StudentID)
	if err != nil || len(balAffected.FeeChangeNotes) != 1 || balAffected.FeeChangeNotes[0].Note != "Approved laboratory levy per BOM resolution" {
		t.Fatalf("expected fee change note on affected student, got %+v (%v)", balAffected.FeeChangeNotes, err)
	}
	balUnaffected, err := svc.GetStudentFeeBalance(ctx, admin, student3ID)
	if err != nil || len(balUnaffected.FeeChangeNotes) != 0 {
		t.Fatalf("expected 0 fee change notes on unaffected student, got %+v (%v)", balUnaffected.FeeChangeNotes, err)
	}

	// 4. OnTermStartDateChanged shifts due_date on fee_periods, fee_charges, and fee_ledgers,
	// records a due_date fee_change_note, and notifies affected guardians.
	newStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	if err := db.WithTenantTx(ctx, school.ID, func(tx pgx.Tx) error {
		return svc.OnTermStartDateChanged(ctx, tx, school.ID, school.TermID, newStart, school.UserID, "Term 1 opening pushed by 2 weeks")
	}); err != nil {
		t.Fatalf("OnTermStartDateChanged failed: %v", err)
	}

	balAfterShift, err := svc.GetStudentFeeBalance(ctx, admin, school.StudentID)
	if err != nil || balAfterShift.DueDate == nil || *balAfterShift.DueDate != "2026-01-19" {
		t.Fatalf("expected due_date=2026-01-19 after OnTermStartDateChanged, got %+v (%v)", balAfterShift.DueDate, err)
	}
	if len(balAfterShift.FeeChangeNotes) != 2 {
		t.Fatalf("expected 2 fee change notes after date shift, got %d", len(balAfterShift.FeeChangeNotes))
	}

	// Both guardians have students in Term 1, so unaffectedGuardianID (Grade 8 in Term 1) now has 1 due-date notice
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2`, school.ID, unaffectedGuardianID).Scan(&unaffectedGuardianNotices); err != nil || unaffectedGuardianNotices != 1 {
		t.Fatalf("expected 1 due-date notification for Grade 8 guardian, got %d (%v)", unaffectedGuardianNotices, err)
	}

	// 5. Verify HTTP endpoint POST /api/v1/fee-changes
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return admin, nil
	})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/fee-changes", bytes.NewBufferString(`{"fee_period_id":"`+g7Struct.FeePeriodID+`","amount_delta":"-500.00","note":""}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 on missing note via HTTP, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/fee-changes", bytes.NewBufferString(`{"fee_period_id":"`+g7Struct.FeePeriodID+`","student_ids":["`+school.StudentID+`"],"amount_delta":"-500.00","note":"Bursary adjustment"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on valid HTTP fee-change, got %d: %s", rec.Code, rec.Body.String())
	}
	_, billedAfterHTTP, _, currentAfterHTTP, _ := assertStudentReconciliation(t, ctx, db, school.ID, school.StudentID)
	if billedAfterHTTP.String() != "11000.00" || currentAfterHTTP.String() != "11000.00" {
		t.Fatalf("expected 11000.00 after -500.00 HTTP fee change, got billed=%s current=%s", billedAfterHTTP, currentAfterHTTP)
	}
}
