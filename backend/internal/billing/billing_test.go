package billing

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Mock InAppNotifier
type mockNotifier struct {
	notifications []struct {
		SchoolID        string
		RecipientUserID string
		Category        string
		Title           string
		Body            string
	}
}

func (m *mockNotifier) CreateInAppNotification(ctx context.Context, schoolID, recipientUserID, category, title, body string, metadata map[string]interface{}) error {
	m.notifications = append(m.notifications, struct {
		SchoolID        string
		RecipientUserID string
		Category        string
		Title           string
		Body            string
	}{schoolID, recipientUserID, category, title, body})
	return nil
}

// Mock AuditLogger
type mockAuditLogger struct {
	actions []string
}

func (m *mockAuditLogger) LogAction(ctx context.Context, schoolID, userID, action, resourceType, resourceID string, details map[string]interface{}) error {
	m.actions = append(m.actions, action)
	return nil
}

func setupTestBillingService() (*Service, *MemoryStore, *mockNotifier, *mockAuditLogger) {
	store := NewMemoryStore()
	notifier := &mockNotifier{}
	auditLogger := &mockAuditLogger{}
	svc := NewService(store, notifier, auditLogger, nil)
	return svc, store, notifier, auditLogger
}

func TestAutomatedTermChargingAndCreditRollover(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	academicYearID := "2026"
	gradeID := "grade-7"
	termID := "term-1"
	studentID := "student-101"

	// Seed enrolled student and student info
	store.EnrolledStudents[schoolID+":"+gradeID+":"+academicYearID] = []string{studentID}
	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: "IKX/2026/001", Name: "Amani Otieno", GradeID: gradeID, YearID: academicYearID,
	}

	termStartDate := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	// 1. Load Fee Structure for Term 1 (Amount: 35,000)
	req := LoadFeeStructureRequest{
		SchoolID:       schoolID,
		AcademicYearID: academicYearID,
		GradeLevelID:   gradeID,
		TermID:         termID,
		TermStartDate:  termStartDate,
		Items: []FeeStructureItemInput{
			{Title: "Tuition", Amount: 25000, IsCompulsory: true},
			{Title: "Activity", Amount: 10000, IsCompulsory: true},
		},
	}

	res, err := svc.LoadFeeStructure(ctx, req)
	if err != nil {
		t.Fatalf("LoadFeeStructure failed: %v", err)
	}
	if res.TotalAmount != 35000 {
		t.Fatalf("expected total amount 35000, got %f", res.TotalAmount)
	}

	// Verify fee charges created with due_date = termStartDate (C61, AC-20)
	charges, err := store.ListFeeChargesForStudent(ctx, schoolID, studentID)
	if err != nil || len(charges) != 1 {
		t.Fatalf("expected 1 charge, got %v", charges)
	}
	if charges[0].Amount != 35000 {
		t.Errorf("expected charge amount 35000, got %f", charges[0].Amount)
	}
	if !charges[0].DueDate.Equal(termStartDate) {
		t.Errorf("expected due date %v, got %v", termStartDate, charges[0].DueDate)
	}

	// Verify ledger created: current_balance = 35000, credit_balance = 0
	ledger, err := store.GetLatestFeeLedger(ctx, schoolID, studentID)
	if err != nil {
		t.Fatalf("GetLatestFeeLedger failed: %v", err)
	}
	if ledger.CurrentBalance != 35000 || ledger.CreditBalance != 0 {
		t.Fatalf("expected current_balance 35000, credit 0; got bal=%f, cred=%f", ledger.CurrentBalance, ledger.CreditBalance)
	}

	// 2. Simulate prior overpayment resulting in credit_balance = 15000
	// e.g., Parent paid 50,000 in Term 1 -> CurrentBalance = 0, CreditBalance = 15,000
	ledger.PaidAmount = 50000
	ledger.CurrentBalance = 0
	ledger.CreditBalance = 15000
	_ = store.SaveFeeLedger(ctx, ledger)

	// 3. Now load Fee Structure for Term 2 (Amount: 30,000) (A22, AC-20)
	term2StartDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	term2Req := LoadFeeStructureRequest{
		SchoolID:       schoolID,
		AcademicYearID: academicYearID,
		GradeLevelID:   gradeID,
		TermID:         "term-2",
		TermStartDate:  term2StartDate,
		Items: []FeeStructureItemInput{
			{Title: "Tuition", Amount: 20000, IsCompulsory: true},
			{Title: "Boarding", Amount: 10000, IsCompulsory: true},
		},
	}

	_, err = svc.LoadFeeStructure(ctx, term2Req)
	if err != nil {
		t.Fatalf("LoadFeeStructure for Term 2 failed: %v", err)
	}

	// In Term 2, the 15,000 credit should be automatically applied!
	// New Billed: 30,000. Applied Credit: 15,000. New Current Balance: 15,000. New Credit: 0.
	term2Ledger, err := store.GetFeeLedger(ctx, schoolID, studentID, academicYearID, "term-2")
	if err != nil {
		t.Fatalf("GetFeeLedger for Term 2 failed: %v", err)
	}
	if term2Ledger.PaidAmount != 15000 {
		t.Errorf("expected Term 2 paid amount from credit = 15000, got %f", term2Ledger.PaidAmount)
	}
	if term2Ledger.CurrentBalance != 15000 {
		t.Errorf("expected Term 2 current balance = 15000, got %f", term2Ledger.CurrentBalance)
	}
	if term2Ledger.CreditBalance != 0 {
		t.Errorf("expected Term 2 remaining credit = 0, got %f", term2Ledger.CreditBalance)
	}

	// Verify ledger entries include 'credit_applied'
	entries, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	foundCreditApplied := false
	for _, e := range entries {
		if e.EntryType == "credit_applied" && e.Amount == 15000 {
			foundCreditApplied = true
			break
		}
	}
	if !foundCreditApplied {
		t.Errorf("expected credit_applied entry of 15000 in ledger")
	}
}

func TestMidYearFeeChangeValidationAndTargetedNotification(t *testing.T) {
	svc, store, notifier, auditLogger := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	termID := "term-1"
	gradeID := "grade-7"

	studentA := "student-A"
	studentB := "student-B"
	guardianA := "user-guardian-A"
	guardianB := "user-guardian-B"

	store.Students[schoolID+":"+studentA] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentA, Admission: "ADM/A", Name: "Student A", GradeID: gradeID, YearID: "2026",
	}
	store.Students[schoolID+":"+studentB] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentB, Admission: "ADM/B", Name: "Student B", GradeID: gradeID, YearID: "2026",
	}
	store.Guardians[schoolID+":"+studentA] = []string{guardianA}
	store.Guardians[schoolID+":"+studentB] = []string{guardianB}

	// Seed charge for student A only (so only student A is in this grade)
	_ = store.CreateFeeCharge(ctx, &FeeCharge{
		SchoolID:    schoolID,
		StudentID:   studentA,
		FeePeriodID: "period-1",
		Amount:      30000,
	})
	_ = store.SaveFeeLedger(ctx, &FeeLedger{
		SchoolID:       schoolID,
		StudentID:      studentA,
		AcademicYearID: "2026",
		TermID:         termID,
		BilledAmount:   30000,
		CurrentBalance: 30000,
	})

	// 1. Mandatory Note Rule: Empty note must be rejected with 422 FEE_CHANGE_NOTE_REQUIRED (AC-21)
	reqNoNote := FeeChangeRequest{
		SchoolID:     schoolID,
		TermID:       termID,
		GradeLevelID: gradeID,
		AmountDelta:  2500,
		Note:         "   ", // empty
	}
	_, err := svc.ApplyMidYearFeeChange(ctx, reqNoNote)
	if err == nil {
		t.Fatalf("expected error for empty note, got nil")
	}
	if !strings.Contains(err.Error(), CodeFeeChangeNoteRequired) {
		t.Fatalf("expected error code %s, got %v", CodeFeeChangeNoteRequired, err)
	}

	// 2. Valid fee adjustment with note
	reqValid := FeeChangeRequest{
		SchoolID:     schoolID,
		TermID:       termID,
		GradeLevelID: gradeID,
		AmountDelta:  2500,
		Note:         "County council water tariff increase",
	}
	noteRec, err := svc.ApplyMidYearFeeChange(ctx, reqValid)
	if err != nil {
		t.Fatalf("ApplyMidYearFeeChange failed: %v", err)
	}
	if noteRec.AmountDelta != 2500 {
		t.Errorf("expected amount delta 2500, got %f", noteRec.AmountDelta)
	}

	// Verify affected student's ledger updated
	ledgerA, _ := store.GetLatestFeeLedger(ctx, schoolID, studentA)
	if ledgerA.BilledAmount != 32500 || ledgerA.CurrentBalance != 32500 {
		t.Errorf("expected updated balance 32500, got %f", ledgerA.CurrentBalance)
	}

	// Verify notification sent ONLY to guardian of affected student A (AC-21)
	if len(notifier.notifications) != 1 {
		t.Fatalf("expected exactly 1 notification, got %d", len(notifier.notifications))
	}
	if notifier.notifications[0].RecipientUserID != guardianA {
		t.Errorf("expected notification recipient to be guardian A, got %s", notifier.notifications[0].RecipientUserID)
	}

	// Verify audit log
	if len(auditLogger.actions) == 0 || auditLogger.actions[0] != "billing.fee_change.applied" {
		t.Errorf("expected audit action billing.fee_change.applied, got %v", auditLogger.actions)
	}
}

func TestTermStartDateShiftHook(t *testing.T) {
	svc, store, notifier, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	termID := "term-1"
	oldDate := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	newDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_ = store.CreateFeePeriod(ctx, &FeePeriod{
		SchoolID:       schoolID,
		AcademicYearID: "2026",
		TermID:         termID,
		DueDate:        oldDate,
	})
	_ = store.CreateFeeCharge(ctx, &FeeCharge{
		SchoolID:    schoolID,
		StudentID:   "student-1",
		FeePeriodID: "period-1",
		DueDate:     oldDate,
	})

	err := svc.OnTermStartDateChanged(ctx, schoolID, termID, newDate, "dean-user-1", "Ministry election delay")
	if err != nil {
		t.Fatalf("OnTermStartDateChanged failed: %v", err)
	}

	// Verify period due date updated
	period, _ := store.GetFeePeriod(ctx, schoolID, "2026", termID)
	if !period.DueDate.Equal(newDate) {
		t.Errorf("expected period due date %v, got %v", newDate, period.DueDate)
	}

	// Verify charges due date updated
	charges, _ := store.ListFeeChargesForStudent(ctx, schoolID, "student-1")
	if len(charges) > 0 && !charges[0].DueDate.Equal(newDate) {
		t.Errorf("expected charge due date %v, got %v", newDate, charges[0].DueDate)
	}

	// Verify fee change note recorded
	notes, _ := store.ListFeeChangeNotes(ctx, schoolID, termID)
	if len(notes) == 0 || notes[0].ChangeType != "term_date_shift" {
		t.Errorf("expected term_date_shift note, got %v", notes)
	}

	// Verify notification dispatched
	if len(notifier.notifications) == 0 {
		t.Errorf("expected notification dispatched on term start date shift")
	}
}

func TestStudentFeeBalanceAndLedgerPermissions(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "student-101"
	guardianID := "user-guardian-1"
	otherParentID := "user-other-parent"

	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: "IKX/101", Name: "Kipchoge Keino", GradeID: "g-7", YearID: "2026",
	}
	store.Guardians[schoolID+":"+studentID] = []string{guardianID}

	// 1. Linked guardian can access fee balance (C10)
	bal, err := svc.GetStudentFeeBalance(ctx, schoolID, studentID, guardianID, RoleParent)
	if err != nil {
		t.Fatalf("linked guardian failed to get fee balance: %v", err)
	}
	if bal.FullName != "Kipchoge Keino" {
		t.Errorf("expected student name Kipchoge Keino, got %s", bal.FullName)
	}

	// 2. Unlinked parent receives 403 Forbidden
	_, err = svc.GetStudentFeeBalance(ctx, schoolID, studentID, otherParentID, RoleParent)
	if err == nil {
		t.Fatalf("expected 403 Forbidden for unlinked parent, got nil")
	}

	// 3. Ikonex admin or Bursar can access
	adminBal, err := svc.GetStudentFeeBalance(ctx, schoolID, studentID, "admin-user", RoleIkonexAdmin)
	if err != nil || adminBal == nil {
		t.Fatalf("admin failed to get fee balance: %v", err)
	}
}

func TestStudentChargeHookOnAdmit(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "new-student-200"
	gradeID := "grade-8"
	academicYearID := "2026"

	// Seed fee structure for Term 1
	_ = store.CreateFeePeriod(ctx, &FeePeriod{
		SchoolID:       schoolID,
		AcademicYearID: academicYearID,
		TermID:         "term-1",
		DueDate:        time.Now(),
	})
	_ = store.CreateFeeStructure(ctx, &FeeStructure{
		SchoolID:       schoolID,
		AcademicYearID: academicYearID,
		GradeLevelID:   gradeID,
		TermID:         "term-1",
		TotalAmount:    42000,
	}, []FeeStructureItem{
		{Title: "Tuition", Amount: 42000, IsCompulsory: true},
	})

	err := svc.StudentChargeHook(ctx, schoolID, studentID, gradeID, academicYearID)
	if err != nil {
		t.Fatalf("StudentChargeHook failed: %v", err)
	}

	// Verify student was charged
	charges, err := store.ListFeeChargesForStudent(ctx, schoolID, studentID)
	if err != nil || len(charges) != 1 {
		t.Fatalf("expected 1 charge for newly admitted student, got %v", charges)
	}
	if charges[0].Amount != 42000 {
		t.Errorf("expected charge 42000, got %f", charges[0].Amount)
	}
}
