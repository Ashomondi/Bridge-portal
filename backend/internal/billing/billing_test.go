package billing

import (
	"context"
	"errors"
	"net/http"
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

type mockDarajaClient struct {
	initiateFunc func(ctx context.Context, cfg *SchoolMpesaConfig, params STKInitiateParams) (*DarajaSTKResponse, error)
	queryFunc    func(ctx context.Context, cfg *SchoolMpesaConfig, checkoutRequestID string) (*DarajaQueryResponse, error)
}

func (m *mockDarajaClient) InitiateSTKPush(ctx context.Context, cfg *SchoolMpesaConfig, params STKInitiateParams) (*DarajaSTKResponse, error) {
	if m.initiateFunc != nil {
		return m.initiateFunc(ctx, cfg, params)
	}
	return &DarajaSTKResponse{
		MerchantRequestID:   "mock-merchant-req-1",
		CheckoutRequestID:   "ws_CO_MOCK_12345",
		ResponseCode:        "0",
		ResponseDescription: "Success. Request accepted for processing",
		CustomerMessage:     "Success",
	}, nil
}

func (m *mockDarajaClient) QuerySTKStatus(ctx context.Context, cfg *SchoolMpesaConfig, checkoutRequestID string) (*DarajaQueryResponse, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, cfg, checkoutRequestID)
	}
	return &DarajaQueryResponse{
		ResponseCode: "0",
		ResultCode:   "0",
		ResultDesc:   "The service request is processed successfully.",
	}, nil
}

type mockAsynqEnqueuer struct {
	polls []string
}

func (m *mockAsynqEnqueuer) EnqueueSTKPoll(ctx context.Context, schoolID, checkoutRequestID string, delay time.Duration) error {
	m.polls = append(m.polls, schoolID+":"+checkoutRequestID)
	return nil
}

func setupTestBillingService() (*Service, *MemoryStore, *mockNotifier, *mockAuditLogger) {
	store := NewMemoryStore()
	notifier := &mockNotifier{}
	auditLogger := &mockAuditLogger{}
	enqueuer := &mockAsynqEnqueuer{}
	svc := NewService(store, notifier, auditLogger, enqueuer)
	svc.SetDarajaClient(&mockDarajaClient{})
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
		ID:             "period-1",
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
	store.Students[schoolID+":student-1"] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: "student-1", Admission: "ADM001", Name: "Student One", GradeID: "g-7", YearID: "2026",
	}
	store.Guardians[schoolID+":student-1"] = []string{"guardian-user-1"}

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

	_ = store.SaveSchoolMpesaConfig(ctx, &SchoolMpesaConfig{
		SchoolID:  schoolID,
		Shortcode: "600100",
	})

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

	// Verify unmatched queue is now empty after resolution
	remainingUnmatched, err := svc.ListUnmatchedPayments(ctx, schoolID)
	if err != nil || len(remainingUnmatched) != 0 {
		t.Errorf("expected 0 unmatched payments after manual match, got %v", remainingUnmatched)
	}
}

func TestFeeChangeAndDueDateScopeIsolation(t *testing.T) {
	svc, store, notifier, _ := setupTestBillingService()
	ctx := context.Background()

	school1 := "school-1"
	school2 := "school-2"
	year := "2026"

	studentG7T1 := "student-s1-g7"
	studentG8T1 := "student-s1-g8"
	studentS2G7 := "student-s2-g7"

	guardianG7 := "guardian-s1-g7"
	guardianG8 := "guardian-s1-g8"
	guardianS2 := "guardian-s2-g7"

	// Seed enrollments and student profiles across grades and schools
	store.EnrolledStudents[school1+":grade-7:"+year] = []string{studentG7T1}
	store.EnrolledStudents[school1+":grade-8:"+year] = []string{studentG8T1}
	store.EnrolledStudents[school2+":grade-7:"+year] = []string{studentS2G7}

	store.Students[school1+":"+studentG7T1] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentG7T1, Admission: "S1/G7/001", Name: "Grade 7 Student", GradeID: "grade-7", YearID: year,
	}
	store.Students[school1+":"+studentG8T1] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentG8T1, Admission: "S1/G8/001", Name: "Grade 8 Student", GradeID: "grade-8", YearID: year,
	}
	store.Students[school2+":"+studentS2G7] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentS2G7, Admission: "S2/G7/001", Name: "School 2 Grade 7", GradeID: "grade-7", YearID: year,
	}

	store.Guardians[school1+":"+studentG7T1] = []string{guardianG7}
	store.Guardians[school1+":"+studentG8T1] = []string{guardianG8}
	store.Guardians[school2+":"+studentS2G7] = []string{guardianS2}

	t1Due := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	t2Due := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	// Load structures for School 1 Grade 7 (Term 1 & Term 2), School 1 Grade 8 (Term 1), and School 2 Grade 7 (Term 1)
	_, err := svc.LoadFeeStructure(ctx, LoadFeeStructureRequest{
		SchoolID: school1, AcademicYearID: year, GradeLevelID: "grade-7", TermID: "term-1", TermStartDate: t1Due,
		Items: []FeeStructureItemInput{{Title: "Tuition T1", Amount: 20000, IsCompulsory: true}},
	})
	if err != nil {
		t.Fatalf("failed loading S1 G7 T1: %v", err)
	}
	_, err = svc.LoadFeeStructure(ctx, LoadFeeStructureRequest{
		SchoolID: school1, AcademicYearID: year, GradeLevelID: "grade-7", TermID: "term-2", TermStartDate: t2Due,
		Items: []FeeStructureItemInput{{Title: "Tuition T2", Amount: 22000, IsCompulsory: true}},
	})
	if err != nil {
		t.Fatalf("failed loading S1 G7 T2: %v", err)
	}
	_, err = svc.LoadFeeStructure(ctx, LoadFeeStructureRequest{
		SchoolID: school1, AcademicYearID: year, GradeLevelID: "grade-8", TermID: "term-1", TermStartDate: t1Due,
		Items: []FeeStructureItemInput{{Title: "Tuition G8 T1", Amount: 25000, IsCompulsory: true}},
	})
	if err != nil {
		t.Fatalf("failed loading S1 G8 T1: %v", err)
	}
	_, err = svc.LoadFeeStructure(ctx, LoadFeeStructureRequest{
		SchoolID: school2, AcademicYearID: year, GradeLevelID: "grade-7", TermID: "term-1", TermStartDate: t1Due,
		Items: []FeeStructureItemInput{{Title: "Tuition S2 T1", Amount: 18000, IsCompulsory: true}},
	})
	if err != nil {
		t.Fatalf("failed loading S2 G7 T1: %v", err)
	}

	// 1. Apply mid-year fee change ONLY to School 1, Term 1, Grade 7 (+3000)
	note, err := svc.ApplyMidYearFeeChange(ctx, FeeChangeRequest{
		SchoolID:     school1,
		TermID:       "term-1",
		GradeLevelID: "grade-7",
		AmountDelta:  3000,
		Note:         "Grade 7 laboratory levy for Term 1",
	})
	if err != nil {
		t.Fatalf("ApplyMidYearFeeChange failed: %v", err)
	}
	if note.AffectedStudentsCount != 1 {
		t.Errorf("expected exactly 1 affected student, got %d", note.AffectedStudentsCount)
	}

	// Verify S1 G7 Term 1 charge & ledger increased by 3000 (20000 -> 23000)
	s1G7T1Ledger, err := store.GetFeeLedger(ctx, school1, studentG7T1, year, "term-1")
	if err != nil || s1G7T1Ledger.BilledAmount != 23000 {
		t.Errorf("expected S1 G7 Term 1 billed amount 23000, got %+v (err=%v)", s1G7T1Ledger, err)
	}

	// Verify S1 G7 Term 2 ledger & charge were NOT mutated (must remain 22000 billed)
	s1G7T2Ledger, err := store.GetFeeLedger(ctx, school1, studentG7T1, year, "term-2")
	if err != nil || s1G7T2Ledger.BilledAmount != 22000 {
		t.Errorf("term isolation failure: S1 G7 Term 2 billed amount mutated to %f (expected 22000)", s1G7T2Ledger.BilledAmount)
	}

	// Verify S1 G8 Term 1 ledger & charge were NOT mutated (must remain 25000)
	s1G8T1Ledger, err := store.GetFeeLedger(ctx, school1, studentG8T1, year, "term-1")
	if err != nil || s1G8T1Ledger.BilledAmount != 25000 {
		t.Errorf("grade isolation failure: S1 G8 Term 1 billed amount mutated to %f (expected 25000)", s1G8T1Ledger.BilledAmount)
	}

	// Verify School 2 G7 Term 1 ledger & charge were NOT mutated (must remain 18000)
	s2G7T1Ledger, err := store.GetFeeLedger(ctx, school2, studentS2G7, year, "term-1")
	if err != nil || s2G7T1Ledger.BilledAmount != 18000 {
		t.Errorf("tenant isolation failure: S2 G7 Term 1 billed amount mutated to %f (expected 18000)", s2G7T1Ledger.BilledAmount)
	}

	// Verify ONLY guardianG7 received notification
	if len(notifier.notifications) != 1 || notifier.notifications[0].RecipientUserID != guardianG7 {
		t.Errorf("expected only guardianG7 to be notified, got %+v", notifier.notifications)
	}

	// 2. Shift due date ONLY for School 1, Term 2
	shiftedT2Due := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	if err := svc.OnTermStartDateChanged(ctx, school1, "term-2", shiftedT2Due, "admin-1", "Term 2 opening shifted"); err != nil {
		t.Fatalf("OnTermStartDateChanged failed: %v", err)
	}

	// Verify S1 G7 Term 1 charge still has original t1Due, while Term 2 charge has shiftedT2Due
	s1G7Charges, _ := store.ListFeeChargesForStudent(ctx, school1, studentG7T1)
	t1Period, _ := store.GetFeePeriod(ctx, school1, year, "term-1")
	t2Period, _ := store.GetFeePeriod(ctx, school1, year, "term-2")
	for _, ch := range s1G7Charges {
		if ch.FeePeriodID == t1Period.ID && !ch.DueDate.Equal(t1Due) {
			t.Errorf("term-1 charge due date was wrongly shifted to %v (expected %v)", ch.DueDate, t1Due)
		}
		if ch.FeePeriodID == t2Period.ID && !ch.DueDate.Equal(shiftedT2Due) {
			t.Errorf("term-2 charge due date was not shifted to %v (got %v)", shiftedT2Due, ch.DueDate)
		}
	}

	// Verify S1 G8 Term 1 charge still has original t1Due
	s1G8Charges, _ := store.ListFeeChargesForStudent(ctx, school1, studentG8T1)
	if len(s1G8Charges) != 1 || !s1G8Charges[0].DueDate.Equal(t1Due) {
		t.Errorf("S1 G8 Term 1 charge due date wrongly mutated: %+v", s1G8Charges)
	}
}

func TestC2BFailsClosedOnUnknownOrCrossTenantShortcode(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()

	school1 := "school-1"
	school2 := "school-2"
	shortcode1 := "600111"
	shortcode2 := "600222"

	_ = store.SaveSchoolMpesaConfig(ctx, &SchoolMpesaConfig{SchoolID: school1, Shortcode: shortcode1})
	_ = store.SaveSchoolMpesaConfig(ctx, &SchoolMpesaConfig{SchoolID: school2, Shortcode: shortcode2})

	// Student ADM/SHARED/01 exists in School 1
	store.Students[school1+":st-1"] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: "st-1", Admission: "ADM/SHARED/01", Name: "School 1 Student", GradeID: "g-7", YearID: "2026",
	}
	store.StudentsByAdmin[school1+":ADM/SHARED/01"] = "st-1"

	// 1. Unknown shortcode MUST fail closed with validation error (never fallback to school-1)
	_, err := svc.HandleC2BConfirmation(ctx, C2BConfirmationPayload{
		TransactionType:   "Pay Bill",
		TransID:           "TX_UNKNOWN_SC",
		TransAmount:       "10000.00",
		BusinessShortCode: "999999",
		BillRefNumber:     "ADM/SHARED/01",
		MSISDN:            "254700000001",
	})
	if err == nil {
		t.Fatalf("expected error when BusinessShortCode is unknown, got nil")
	}

	// 2. Payment sent to School 2's shortcode with School 1's admission number must NOT match School 1's student
	p, err := svc.HandleC2BConfirmation(ctx, C2BConfirmationPayload{
		TransactionType:   "Pay Bill",
		TransID:           "TX_CROSS_TENANT",
		TransAmount:       "10000.00",
		BusinessShortCode: shortcode2,
		BillRefNumber:     "ADM/SHARED/01",
		MSISDN:            "254700000002",
	})
	if err != nil {
		t.Fatalf("unexpected error for valid School 2 shortcode: %v", err)
	}
	if p.SchoolID != school2 {
		t.Errorf("expected payment scoped to %s, got %s", school2, p.SchoolID)
	}
	if p.MatchedStatus != "unmatched" || p.StudentID != nil {
		t.Errorf("cross-tenant leak: School 2 shortcode matched School 1 student: %+v", p)
	}
}

func TestSTKCallbackIdempotencyAndFailureHandling(t *testing.T) {
	svc, store, notifier, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "student-stk-cb"
	guardianUserID := "guardian-stk-user"
	checkoutID := "ws_CO_TEST_IDEMPOTENCY_001"

	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: "IKX/CB/001", Name: "Callback Student", GradeID: "g-7", YearID: "2026",
	}
	store.Guardians[schoolID+":"+studentID] = []string{guardianUserID}

	// Create pending payment
	_ = store.CreatePayment(ctx, &Payment{
		SchoolID:       schoolID,
		StudentID:      &studentID,
		TransactionRef: checkoutID,
		Provider:       "mpesa_stk",
		Amount:         5000,
		MatchedStatus:  "pending",
	})

	// 1. Unknown CheckoutRequestID returns error
	unknownPayload := STKCallbackPayload{}
	unknownPayload.Body.StkCallback.CheckoutRequestID = "UNKNOWN_CHECKOUT_ID"
	unknownPayload.Body.StkCallback.ResultCode = 0
	err := svc.HandleSTKCallback(ctx, unknownPayload)
	if err == nil {
		t.Fatalf("expected error for unknown checkout request id, got nil")
	}

	// 2. Successful callback updates payment to matched, credits ledger, dispatches notification
	successPayload := STKCallbackPayload{}
	successPayload.Body.StkCallback.CheckoutRequestID = checkoutID
	successPayload.Body.StkCallback.ResultCode = 0
	successPayload.Body.StkCallback.ResultDesc = "The service request is processed successfully."
	successPayload.Body.StkCallback.CallbackMetadata = &struct {
		Item []struct {
			Name  string      `json:"Name"`
			Value interface{} `json:"Value"`
		} `json:"Item"`
	}{
		Item: []struct {
			Name  string      `json:"Name"`
			Value interface{} `json:"Value"`
		}{
			{Name: "Amount", Value: float64(5000)},
			{Name: "MpesaReceiptNumber", Value: "QEH1234567"},
			{Name: "PhoneNumber", Value: float64(254712345678)},
		},
	}

	err = svc.HandleSTKCallback(ctx, successPayload)
	if err != nil {
		t.Fatalf("HandleSTKCallback failed: %v", err)
	}
	p, _ := store.GetPaymentByCheckoutID(ctx, checkoutID)
	if p == nil || p.MatchedStatus != "matched" {
		t.Errorf("expected matched payment, got %v", p)
	}
	if len(notifier.notifications) != 1 {
		t.Errorf("expected 1 notification dispatched, got %d", len(notifier.notifications))
	}
	entries, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if len(entries) != 1 || entries[0].Amount != 5000 {
		t.Errorf("expected 1 ledger entry of 5000, got %v", entries)
	}

	// 3. Duplicate callback with the SAME checkout ID must be idempotent (no duplicate ledger entries, no duplicate notifications)
	err = svc.HandleSTKCallback(ctx, successPayload)
	if err != nil {
		t.Fatalf("duplicate HandleSTKCallback failed: %v", err)
	}
	pDup, _ := store.GetPaymentByCheckoutID(ctx, checkoutID)
	if pDup == nil || pDup.MatchedStatus != "matched" {
		t.Errorf("expected matched payment, got %v", pDup)
	}
	if len(notifier.notifications) != 1 {
		t.Errorf("duplicate callback created duplicate notification: %d", len(notifier.notifications))
	}
	entriesAfter, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if len(entriesAfter) != 1 {
		t.Errorf("duplicate callback created duplicate ledger entry: %d entries", len(entriesAfter))
	}

	// 4. Cancelled/Failed callback (ResultCode != 0) marks payment as failed without crediting ledger
	checkoutFailedID := "ws_CO_TEST_CANCELLED_002"
	_ = store.CreatePayment(ctx, &Payment{
		SchoolID:       schoolID,
		StudentID:      &studentID,
		TransactionRef: checkoutFailedID,
		Provider:       "mpesa_stk",
		Amount:         3000,
		MatchedStatus:  "pending",
	})

	failedPayload := STKCallbackPayload{}
	failedPayload.Body.StkCallback.CheckoutRequestID = checkoutFailedID
	failedPayload.Body.StkCallback.ResultCode = 1032
	failedPayload.Body.StkCallback.ResultDesc = "Request cancelled by user."

	err = svc.HandleSTKCallback(ctx, failedPayload)
	if err != nil {
		t.Fatalf("HandleSTKCallback failed: %v", err)
	}
	pFailed, _ := store.GetPaymentByCheckoutID(ctx, checkoutFailedID)
	if pFailed == nil || pFailed.MatchedStatus != "failed" {
		t.Errorf("expected failed status, got %v", pFailed)
	}
	entriesAfterFailed, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if len(entriesAfterFailed) != 1 {
		t.Errorf("failed callback incorrectly created ledger entry")
	}
}

func TestManualMatchDuplicateResolutionPrevented(t *testing.T) {
	svc, store, _, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	studentID := "student-resolve-1"
	adminUser := "bursar-1"
	paymentID := "pay-unmatched-1"

	store.Students[schoolID+":"+studentID] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: studentID, Admission: "IKX/RES/001", Name: "Resolve Student", GradeID: "g-7", YearID: "2026",
	}

	_ = store.CreatePayment(ctx, &Payment{
		ID:             paymentID,
		SchoolID:       schoolID,
		TransactionRef: "TX_MANUAL_MATCH_1",
		Provider:       "mpesa_c2b",
		Amount:         6000,
		MatchedStatus:  "unmatched",
	})
	_ = store.CreateUnmatchedPayment(ctx, &UnmatchedPayment{
		ID:             "unm-1",
		SchoolID:       schoolID,
		PaymentID:      paymentID,
		TransactionRef: "TX_MANUAL_MATCH_1",
		Amount:         6000,
		Status:         "unmatched",
	})

	// First manual resolution must succeed (C25, AC-25)
	p1, err := svc.MatchUnmatchedPayment(ctx, schoolID, "unm-1", studentID, "Parent provided bank slip", adminUser)
	if err != nil {
		t.Fatalf("MatchUnmatchedPayment failed: %v", err)
	}
	if p1.MatchedStatus != "matched" {
		t.Errorf("expected matched payment, got %s", p1.MatchedStatus)
	}
	entries, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if len(entries) != 1 || entries[0].Amount != 6000 {
		t.Errorf("expected 1 ledger entry credited, got %v", entries)
	}

	// Repeated manual match resolution MUST fail with 409 Conflict: PAYMENT_ALREADY_RESOLVED (AC-25)
	_, err = svc.MatchUnmatchedPayment(ctx, schoolID, "unm-1", studentID, "Duplicate resolution attempt", adminUser)
	if err == nil {
		t.Fatalf("expected 409 conflict on duplicate resolution, got nil")
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.Status != http.StatusConflict || apiErr.Code != "PAYMENT_ALREADY_RESOLVED" {
			t.Errorf("expected 409 Conflict PAYMENT_ALREADY_RESOLVED, got %d %s", apiErr.Status, apiErr.Code)
		}
	} else {
		t.Errorf("expected APIError, got %v", err)
	}

	// Verify no double ledger credit
	entries2, _ := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if len(entries2) != 1 {
		t.Errorf("SECURITY FAULT: duplicate resolution credited ledger a second time: %d entries", len(entries2))
	}
}

func TestTermStartDateShiftNotifiesOnlyGuardiansNotAdmin(t *testing.T) {
	svc, store, notifier, _ := setupTestBillingService()
	ctx := context.Background()
	schoolID := "school-1"
	termID := "term-shift-test"
	deanUserID := "dean-user-999"
	guardian1ID := "guardian-user-11"
	guardian2ID := "guardian-user-22"

	newDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	_ = store.CreateFeePeriod(ctx, &FeePeriod{
		ID:             "period-shift-1",
		SchoolID:       schoolID,
		AcademicYearID: "2026",
		TermID:         termID,
		DueDate:        time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	})
	_ = store.CreateFeeCharge(ctx, &FeeCharge{
		SchoolID:    schoolID,
		StudentID:   "student-shift-1",
		FeePeriodID: "period-shift-1",
		DueDate:     time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	})

	store.Students[schoolID+":student-shift-1"] = struct{ ID, Admission, Name, GradeID, YearID string }{
		ID: "student-shift-1", Admission: "IKX/SHIFT/001", Name: "Shift Student", GradeID: "g-7", YearID: "2026",
	}
	store.Guardians[schoolID+":student-shift-1"] = []string{guardian1ID, guardian2ID}

	err := svc.OnTermStartDateChanged(ctx, schoolID, termID, newDate, deanUserID, "Election adjustment")
	if err != nil {
		t.Fatalf("OnTermStartDateChanged failed: %v", err)
	}

	if len(notifier.notifications) != 2 {
		t.Fatalf("expected 2 notifications dispatched (to 2 guardians), got %d", len(notifier.notifications))
	}
	for _, n := range notifier.notifications {
		if n.RecipientUserID == deanUserID {
			t.Errorf("BUG: acting dean %s was sent a student term-date shift notification", deanUserID)
		}
		if n.RecipientUserID != guardian1ID && n.RecipientUserID != guardian2ID {
			t.Errorf("unexpected notification recipient: %s", n.RecipientUserID)
		}
	}
}
