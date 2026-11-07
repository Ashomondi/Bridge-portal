package whatsapp

import (
	"context"
	"strings"
	"testing"
	"time"
)

type mockSMS struct {
	sent []struct {
		To      string
		Message string
	}
}

func (m *mockSMS) SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error {
	m.sent = append(m.sent, struct {
		To      string
		Message string
	}{toPhoneNumber, message})
	return nil
}

type mockAIBridge struct {
	lastQuery string
}

func (m *mockAIBridge) ForwardParentQuery(ctx context.Context, schoolID, guardianUserID, studentID, query, sessionID string) (string, error) {
	m.lastQuery = query
	return "Mock AI reply for: " + query, nil
}

func setupTestWhatsApp() (*Service, *MemoryStore, *mockSMS, *mockAIBridge) {
	store := NewMemoryStore()
	sms := &mockSMS{}
	ai := &mockAIBridge{}
	svc := NewService(store, sms, ai, nil)
	return svc, store, sms, ai
}

func TestMaskPhoneNumber(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"+254712345678", "07•• ••• 678"},
		{"0712345482", "07•• ••• 482"},
		{"254799000123", "07•• ••• 123"},
		{"+254112345910", "01•• ••• 910"},
	}

	for _, c := range cases {
		masked := MaskPhoneNumber(c.input)
		if masked != c.expected {
			t.Errorf("MaskPhoneNumber(%q) = %q; expected %q", c.input, masked, c.expected)
		}
	}
}

func TestUnknownAdmissionNumberSendsZeroDataAndZeroOTP(t *testing.T) {
	svc, store, sms, _ := setupTestWhatsApp()
	ctx := context.Background()
	botPhone := "+254700000001"
	senderWA := "+254788111222"

	// Register School A bot
	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           "school-A",
		DisplayPhoneNumber: botPhone,
		IsActive:           true,
	})

	// 1. Initial greeting
	_, _ = svc.ProcessIncomingMessage(ctx, botPhone, senderWA, "Hello")

	// 2. Submit non-existent admission number (C14)
	reply, err := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, "INVALID/ADM/999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must return generic message and send zero OTPs
	if !strings.Contains(reply, "could not verify that admission number") {
		t.Errorf("expected generic verification failure message, got: %s", reply)
	}
	if len(sms.sent) != 0 {
		t.Errorf("expected zero SMS OTPs sent for invalid admission, got %d", len(sms.sent))
	}
}

func TestMultipleGuardiansMaskedSelectionAndOTP(t *testing.T) {
	svc, store, sms, _ := setupTestWhatsApp()
	ctx := context.Background()
	botPhone := "+254700000001"
	senderWA := "+254799888777"
	schoolID := "school-A"
	studentID := "student-001"
	admissionNum := "IKX/2026/001"

	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           schoolID,
		DisplayPhoneNumber: botPhone,
		IsActive:           true,
	})

	store.Students[schoolID+":"+admissionNum] = struct{ ID, Name string }{
		ID: studentID, Name: "Zawadi Akinyi",
	}

	// Two registered guardians (C15, AC-09)
	guardian1 := GuardianCandidate{
		UserID:       "user-mother",
		FullName:     "Grace Akinyi",
		PhoneNumber:  "+254712345482",
		Relationship: "Mother",
		MaskedPhone:  MaskPhoneNumber("+254712345482"), // "07•• ••• 482"
	}
	guardian2 := GuardianCandidate{
		UserID:       "user-father",
		FullName:     "John Otieno",
		PhoneNumber:  "+254722999910",
		Relationship: "Father",
		MaskedPhone:  MaskPhoneNumber("+254722999910"), // "07•• ••• 910"
	}
	store.Guardians[schoolID+":"+studentID] = []GuardianCandidate{guardian1, guardian2}

	// 1. Send admission number -> bot must reply with masked phone numbers (AC-09)
	reply, err := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, admissionNum)
	if err != nil {
		t.Fatalf("ProcessIncomingMessage failed: %v", err)
	}
	if !strings.Contains(reply, "07•• ••• 482") || !strings.Contains(reply, "07•• ••• 910") {
		t.Fatalf("expected reply to contain masked phone numbers, got: %s", reply)
	}

	// 2. Sender chooses guardian 1 ("1")
	replyChoice, err := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, "1")
	if err != nil {
		t.Fatalf("choice failed: %v", err)
	}
	if !strings.Contains(replyChoice, "07•• ••• 482") {
		t.Errorf("expected OTP dispatch message mentioning 07•• ••• 482, got: %s", replyChoice)
	}

	// 3. Verify OTP was sent STRICTLY to guardian 1's registered phone, NOT to senderWA (AC-08, AC-09)
	if len(sms.sent) != 1 {
		t.Fatalf("expected 1 SMS dispatched, got %d", len(sms.sent))
	}
	if sms.sent[0].To != "+254712345482" {
		t.Errorf("expected SMS to be dispatched strictly to registered number +254712345482, got %s", sms.sent[0].To)
	}
	if sms.sent[0].To == senderWA {
		t.Errorf("SECURITY FAULT: OTP must NEVER be sent to the unregistered sender WhatsApp number")
	}

	// 4. Extract OTP from SMS text and verify (AC-11)
	words := strings.Fields(sms.sent[0].Message)
	var otpCode string
	for _, w := range words {
		cleaned := strings.Trim(w, ".,!?:; ")
		if len(cleaned) == 6 && strings.Trim(cleaned, "0123456789") == "" {
			otpCode = cleaned
			break
		}
	}
	if otpCode == "" {
		t.Fatalf("could not extract 6-digit OTP from message: %s", sms.sent[0].Message)
	}

	replyVerified, err := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, otpCode)
	if err != nil {
		t.Fatalf("ProcessIncomingMessage with OTP failed: %v", err)
	}
	if !strings.Contains(replyVerified, "Identity verified successfully") {
		t.Errorf("expected verification success message, got: %s", replyVerified)
	}
}

func TestRateLimiterCooldownAnd24hSenderBlock(t *testing.T) {
	svc, store, _, _ := setupTestWhatsApp()
	ctx := context.Background()
	botPhone := "+254700000001"
	senderWA := "+254733444555"
	schoolID := "school-A"
	studentID := "student-rate-1"
	admissionNum := "IKX/RATE/001"

	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           schoolID,
		DisplayPhoneNumber: botPhone,
		IsActive:           true,
	})
	store.Students[schoolID+":"+admissionNum] = struct{ ID, Name string }{ID: studentID, Name: "Rate Student"}
	store.Guardians[schoolID+":"+studentID] = []GuardianCandidate{
		{UserID: "user-g", FullName: "Guardian G", PhoneNumber: "+254711222333", MaskedPhone: "07•• ••• 333"},
	}

	// Request 1: successful OTP dispatch
	reply1, _ := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, admissionNum)
	if !strings.Contains(reply1, "07•• ••• 333") {
		t.Fatalf("expected request 1 to dispatch OTP, got: %s", reply1)
	}

	// Request 2 immediately (within 60s cooldown): must be refused (AC-10)
	session, _ := store.GetActiveSession(ctx, schoolID, senderWA)
	session.State = StateAwaitingAdmissionNumber
	_ = store.SaveSession(ctx, session)

	reply2, _ := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, admissionNum)
	if !strings.Contains(reply2, "wait 60 seconds") {
		t.Errorf("expected 60-second cooldown message on 2nd request, got: %s", reply2)
	}

	// Simulate 5 requests within the last 30 minutes
	store.OTPRequests = nil
	now := time.Now()
	for i := 0; i < 5; i++ {
		store.OTPRequests = append(store.OTPRequests, struct {
			SenderWA        string
			SchoolID        string
			AdmissionNumber string
			CreatedAt       time.Time
		}{
			SenderWA:        senderWA,
			SchoolID:        schoolID,
			AdmissionNumber: admissionNum,
			CreatedAt:       now.Add(-time.Duration(i*3+2) * time.Minute), // spread across last 15 min
		})
	}

	// 6th Request: must block requesting WhatsApp number for 24 hours (AC-10)
	session.State = StateAwaitingAdmissionNumber
	_ = store.SaveSession(ctx, session)

	reply6, _ := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, admissionNum)
	if !strings.Contains(reply6, "blocked from requesting OTPs for 24 hours") {
		t.Errorf("expected 24h block message on 6th request, got: %s", reply6)
	}

	// Verify sender is in sender_rate_blocks
	blocked, until, _ := store.IsSenderBlocked(ctx, senderWA)
	if !blocked || until.Before(time.Now().Add(23*time.Hour)) {
		t.Errorf("expected sender to be blocked for 24h")
	}

	// Crucial rule (C20, AC-10): The guardian's user account itself must NOT be blocked!
	// Guardian phone +254711222333 should not be blocked in the system.
	guardianBlocked, _, _ := store.IsSenderBlocked(ctx, "+254711222333")
	if guardianBlocked {
		t.Errorf("SECURITY FAULT: guardian's actual phone number was blocked; only senderWA should be blocked")
	}
}

func TestSessionExpirationAfter6Hours(t *testing.T) {
	svc, store, _, aiBridge := setupTestWhatsApp()
	ctx := context.Background()
	botPhone := "+254700000001"
	senderWA := "+254755666777"
	schoolID := "school-A"

	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           schoolID,
		DisplayPhoneNumber: botPhone,
		IsActive:           true,
	})

	// Create verified session that expired 5 minutes ago (C13, AC-11)
	pastExpires := time.Now().Add(-5 * time.Minute)
	session := &ChatSession{
		SchoolID:       schoolID,
		Channel:        "whatsapp",
		SenderWANumber: senderWA,
		StudentID:      "student-exp-1",
		GuardianUserID: "user-guardian-exp",
		State:          StateVerified,
		ExpiresAt:      &pastExpires,
	}
	_ = store.SaveSession(ctx, session)

	// Send message to expired session
	reply, err := svc.ProcessIncomingMessage(ctx, botPhone, senderWA, "What is my fee balance?")
	if err != nil {
		t.Fatalf("ProcessIncomingMessage failed: %v", err)
	}

	// Must notify expiration and prompt for admission number again
	if !strings.Contains(reply, "6-hour verified session has expired") {
		t.Errorf("expected expiration prompt, got: %s", reply)
	}
	if aiBridge.lastQuery != "" {
		t.Errorf("expired session must not forward queries to AI service")
	}
}

func TestCrossSchoolBotIsolation(t *testing.T) {
	svc, store, _, _ := setupTestWhatsApp()
	ctx := context.Background()
	botSchoolA := "+254700000001"
	botSchoolB := "+254700000002"
	senderWA := "+254711222333"

	// Register bot for School A and School B (AC-12)
	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           "school-A",
		DisplayPhoneNumber: botSchoolA,
		IsActive:           true,
	})
	_ = store.RegisterSchoolBot(ctx, &SchoolWhatsAppBot{
		SchoolID:           "school-B",
		DisplayPhoneNumber: botSchoolB,
		IsActive:           true,
	})

	// Student ADM/100 belongs STRICTLY to School B
	store.Students["school-B:ADM/100"] = struct{ ID, Name string }{
		ID: "student-B-100", Name: "School B Student",
	}

	// Attempt to lookup ADM/100 via School A's WhatsApp bot
	reply, _ := svc.ProcessIncomingMessage(ctx, botSchoolA, senderWA, "ADM/100")

	// School A bot MUST NOT resolve student from School B (AC-12)
	if !strings.Contains(reply, "could not verify that admission number") {
		t.Errorf("cross-school leak: School A resolved School B student: %s", reply)
	}

	// Looking up via School B's bot DOES match
	store.Guardians["school-B:student-B-100"] = []GuardianCandidate{
		{UserID: "u-b", FullName: "Parent B", PhoneNumber: "+254700111222", MaskedPhone: "07•• ••• 222"},
	}
	replyB, _ := svc.ProcessIncomingMessage(ctx, botSchoolB, senderWA, "ADM/100")
	if strings.Contains(replyB, "could not verify") {
		t.Errorf("School B bot should have found its own student, got: %s", replyB)
	}
}
