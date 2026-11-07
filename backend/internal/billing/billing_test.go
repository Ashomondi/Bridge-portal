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

func TestSTKPushValidationAnd24hChatLock(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "student-stk-1"
	guardianUserID := "guardian-user-1"
	guardianPhone := "+254712345678"

	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: "IKX/STK/001", Name: "Zawadi Akinyi", GradeID: "g-7", YearID: "2026",
	}
	store.Guardians[schoolID+":"+studentID] = []string{guardianUserID}

	guardianCtx := WithIdentity(ctx, Identity{
		UserID:   guardianUserID,
		SchoolID: schoolID,
		Role:     RoleParent,
	})

	// 1. Non-positive amount rejected
	_, err := svc.InitiateSTKPush(guardianCtx, STKPushRequest{
		SchoolID:    schoolID,
		StudentID:   studentID,
		Amount:      0,
		PhoneNumber: guardianPhone,
	})
	if err == nil {
		t.Fatalf("expected error for zero amount, got nil")
	}

	// 2. Chat source without explicit parent confirmation rejected (AC-23)
	_, err = svc.InitiateSTKPush(guardianCtx, STKPushRequest{
		SchoolID:        schoolID,
		StudentID:       studentID,
		Amount:          5000,
		PhoneNumber:     guardianPhone,
		Source:          "chat",
		ParentConfirmed: false,
	})
	if err == nil || !strings.Contains(err.Error(), CodeParentConfirmationReq) {
		t.Fatalf("expected error code %s, got %v", CodeParentConfirmationReq, err)
	}

	// 3. 24-Hour Phone Change Freeze (C27, AC-26)
	// Set chat payment paused until 12 hours from now
	pauseUntil := time.Now().Add(12 * time.Hour)
	_ = store.SetUserChatPaymentPausedUntil(ctx, guardianUserID, pauseUntil)

	// Chat payment must be blocked with 403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE
	_, err = svc.InitiateSTKPush(guardianCtx, STKPushRequest{
		SchoolID:        schoolID,
		StudentID:       studentID,
		Amount:          5000,
		PhoneNumber:     guardianPhone,
		Source:          "chat",
		ParentConfirmed: true,
	})
	if err == nil || !strings.Contains(err.Error(), CodeChatPaymentPaused) {
		t.Fatalf("expected 403 %s, got %v", CodeChatPaymentPaused, err)
	}

	// 4. Portal payment during chat lock is ALLOWED (portal not affected by chat freeze)
	resp, err := svc.InitiateSTKPush(guardianCtx, STKPushRequest{
		SchoolID:    schoolID,
		StudentID:   studentID,
		Amount:      5000,
		PhoneNumber: guardianPhone,
		Source:      "portal",
	})
	if err != nil || resp == nil {
		t.Fatalf("portal payment failed during chat lock: %v", err)
	}
	if resp.CheckoutRequestID == "" {
		t.Errorf("expected non-empty checkout request ID")
	}
}

func TestC2BAutoMatchingAndOverpaymentCredit(t *testing.T) {
	svc, store, notifier, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "student-c2b-1"
	guardianID := "guardian-c2b-user"
	shortcode := "600100"
	admissionNum := "IKX/2026/088"

	_ = store.SaveSchoolMpesaConfig(ctx, &SchoolMpesaConfig{
		SchoolID:  schoolID,
		Shortcode: shortcode,
	})

	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: admissionNum, Name: "Baraka Mwangi", GradeID: "g-7", YearID: "2026",
	}
	store.StudentsByAdmin[schoolID+":"+strings.ToUpper(admissionNum)] = studentID
	store.Guardians[schoolID+":"+studentID] = []string{guardianID}

	// Initial ledger: Billed 20,000, Paid 0, CurrentBalance = 20,000
	_ = store.SaveFeeLedger(ctx, &FeeLedger{
		SchoolID:       schoolID,
		StudentID:      studentID,
		AcademicYearID: "2026",
		TermID:         "term-1",
		BilledAmount:   20000,
		CurrentBalance: 20000,
		CreditBalance:  0,
	})

	// Direct Paybill C2B payment with OVERPAYMENT of 25,000 (AC-24, AC-25)
	payload := C2BConfirmationPayload{
		TransactionType:   "Pay Bill",
		TransID:           "QKH9876543",
		TransTime:         "20261107220000",
		TransAmount:       "25000.00",
		BusinessShortCode: shortcode,
		BillRefNumber:     "  ikx/2026/088  ", // Case-insensitive and trimmed match
		MSISDN:            "254722000111",
		FirstName:         "Mwangi",
	}

	payment, err := svc.HandleC2BConfirmation(ctx, payload)
	if err != nil {
		t.Fatalf("HandleC2BConfirmation failed: %v", err)
	}
	if payment.MatchedStatus != "matched" {
		t.Errorf("expected matched status, got %s", payment.MatchedStatus)
	}
	if payment.StudentID == nil || *payment.StudentID != studentID {
		t.Errorf("expected student ID %s, got %v", studentID, payment.StudentID)
	}

	// Verify FeeLedger overpayment: CurrentBalance = 0, CreditBalance = 5,000 (AC-24)
	ledger, _ := store.GetLatestFeeLedger(ctx, schoolID, studentID)
	if ledger.CurrentBalance != 0 {
		t.Errorf("expected current balance 0 after full payment, got %f", ledger.CurrentBalance)
	}
	if ledger.CreditBalance != 5000 {
		t.Errorf("expected credit balance 5000 from overpayment, got %f", ledger.CreditBalance)
	}
	if ledger.PaidAmount != 25000 {
		t.Errorf("expected paid amount 25000, got %f", ledger.PaidAmount)
	}

	// Verify payment receipt notification sent to guardian
	if len(notifier.notifications) == 0 || notifier.notifications[0].RecipientUserID != guardianID {
		t.Errorf("expected payment receipt notification sent to guardian %s", guardianID)
	}
}

func TestC2BUnmatchedQueueAndManualMatch(t *testing.T) {
	svc, store, _, auditLogger := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	shortcode := "600100"

	_ = store.SaveSchoolMpesaConfig(ctx, &SchoolMpesaConfig{
		SchoolID:  schoolID,
		Shortcode: shortcode,
	})

	// C2B payment with unknown admission number (C50, AC-25)
	payload := C2BConfirmationPayload{
		TransactionType:   "Pay Bill",
		TransID:           "XYZ1234567",
		TransTime:         "20261107223000",
		TransAmount:       "12000.00",
		BusinessShortCode: shortcode,
		BillRefNumber:     "UNKNOWN/NO/999",
		MSISDN:            "254711999888",
		FirstName:         "Unknown Payer",
	}

	payment, err := svc.HandleC2BConfirmation(ctx, payload)
	if err != nil {
		t.Fatalf("HandleC2BConfirmation failed: %v", err)
	}
	if payment.MatchedStatus != "unmatched" {
		t.Errorf("expected unmatched status, got %s", payment.MatchedStatus)
	}

	// Verify payment queued in unmatched_payments
	unmatchedList, err := svc.ListUnmatchedPayments(ctx, schoolID)
	if err != nil || len(unmatchedList) != 1 {
		t.Fatalf("expected 1 unmatched payment, got %v", unmatchedList)
	}
	if unmatchedList[0].RawAccountRef != "UNKNOWN/NO/999" {
		t.Errorf("expected raw account ref UNKNOWN/NO/999, got %s", unmatchedList[0].RawAccountRef)
	}

	// Now simulate Ikonex team / bursar manually resolving the payment (AC-25)
	targetStudentID := "student-resolved-1"
	store.Students[schoolID+":"+targetStudentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: targetStudentID, Admission: "IKX/RESOLVED/001", Name: "Resolved Student", GradeID: "g-7", YearID: "2026",
	}
	_ = store.SaveFeeLedger(ctx, &FeeLedger{
		SchoolID:       schoolID,
		StudentID:      targetStudentID,
		AcademicYearID: "2026",
		TermID:         "term-1",
		BilledAmount:   12000,
		CurrentBalance: 12000,
	})

	matchedPayment, err := svc.MatchUnmatchedPayment(ctx, schoolID, payment.ID, targetStudentID, "Parent confirmed typo in admission number via receipt", "bursar-user-1")
	if err != nil {
		t.Fatalf("MatchUnmatchedPayment failed: %v", err)
	}
	if matchedPayment.MatchedStatus != "matched" {
		t.Errorf("expected matched status, got %s", matchedPayment.MatchedStatus)
	}

	// Verify target student's ledger credited
	targetLedger, _ := store.GetLatestFeeLedger(ctx, schoolID, targetStudentID)
	if targetLedger.CurrentBalance != 0 || targetLedger.PaidAmount != 12000 {
		t.Errorf("expected balance 0, paid 12000; got bal=%f, paid=%f", targetLedger.CurrentBalance, targetLedger.PaidAmount)
	}

	// Verify audit log recorded
	foundAudit := false
	for _, a := range auditLogger.actions {
		if a == "payment.manual_matched" {
			foundAudit = true
			break
		}
	}
	if !foundAudit {
		t.Errorf("expected audit action payment.manual_matched")
	}
}
