package whatsapp_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/Ashomondi/ikonex-academy/backend/internal/whatsapp"
)

type spySMSDispatcher struct {
	mu       sync.Mutex
	messages []notifications.SMSMessage
}

func (s *spySMSDispatcher) DispatchSMS(_ context.Context, msg notifications.SMSMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

func (s *spySMSDispatcher) lastOTP(t *testing.T) (string, string) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		t.Fatal("expected at least 1 SMS message")
	}
	last := s.messages[len(s.messages)-1]
	re := regexp.MustCompile(`\b(\d{6})\b`)
	match := re.FindStringSubmatch(last.Body)
	if len(match) != 2 {
		t.Fatalf("could not extract 6-digit OTP from SMS body %q", last.Body)
	}
	return last.To, match[1]
}

type spyAIForwarder struct {
	calls []whatsapp.AIParentQueryInput
}

func (a *spyAIForwarder) ForwardParentQuery(_ context.Context, in whatsapp.AIParentQueryInput) (string, error) {
	a.calls = append(a.calls, in)
	return "AI reply for " + in.Message, nil
}

func TestMaskPhone(t *testing.T) {
	masked, err := whatsapp.MaskPhone("+254712345678")
	if err != nil || masked != "07•• ••• 678" {
		t.Fatalf("got %q (%v), want 07•• ••• 678", masked, err)
	}
	if _, err := whatsapp.MaskPhone("invalid-phone-secret"); err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected safe error on invalid phone, got %v", err)
	}
}

func TestWhatsAppStateMachineTenantIsolationAndRateLimits(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	// Create sender_rate_blocks table (from migration 000002 contract) in disposable DB to verify DB persistence of 24h sender blocks
	if _, err := db.Admin.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sender_rate_blocks (
			sender_wa_number VARCHAR(30) PRIMARY KEY,
			blocked_until TIMESTAMPTZ NOT NULL,
			reason VARCHAR(100),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		t.Fatal(err)
	}
	db.Grant(t)

	// Set distinct admission numbers on School A (ADM-A-100) and School B (ADM-B-200)
	if _, err := db.Admin.Exec(ctx, `UPDATE students SET admission_number = 'ADM-A-100' WHERE school_id = $1 AND id = $2`, schoolA.ID, schoolA.StudentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `UPDATE students SET admission_number = 'ADM-B-200' WHERE school_id = $1 AND id = $2`, schoolB.ID, schoolB.StudentID); err != nil {
		t.Fatal(err)
	}

	// Add a second guardian to School A's student (+254722111482 and +254733222910)
	if _, err := db.Admin.Exec(ctx, `UPDATE users SET phone_number = '+254722111482' WHERE id = $1`, schoolA.UserID); err != nil {
		t.Fatal(err)
	}
	var secondGuardianID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO users(phone_number,full_name) VALUES('+254733222910','Second Guardian') RETURNING id`).Scan(&secondGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'parent','active')`, schoolA.ID, secondGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'mother')`, schoolA.ID, schoolA.StudentID, secondGuardianID); err != nil {
		t.Fatal(err)
	}

	// Register per-school WhatsApp bots
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO school_whatsapp_bots(school_id, phone_number_id, display_phone_number, is_active)
		VALUES ($1, 'pnid-school-a', '+254700111001', true),
		       ($2, 'pnid-school-b', '+254700222002', true)`,
		schoolA.ID, schoolB.ID,
	); err != nil {
		t.Fatal(err)
	}

	smsSpy := &spySMSDispatcher{}
	aiSpy := &spyAIForwarder{}
	currentNow := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	svc := &whatsapp.Service{
		DB:  db,
		SMS: smsSpy,
		AI:  aiSpy,
		ResolveBot: func(_ context.Context, phoneNumberID, displayPhone string) (string, error) {
			switch {
			case phoneNumberID == "pnid-school-a" || displayPhone == "+254700111001":
				return schoolA.ID, nil
			case phoneNumberID == "pnid-school-b" || displayPhone == "+254700222002":
				return schoolB.ID, nil
			default:
				return "", errors.New("unknown bot")
			}
		},
		Now: func() time.Time { return currentNow },
	}

	unregisteredSender := "+254799000111"

	// 1. Cross-school isolation (AC-12) and unknown admission number (C14):
	// School A's bot queried with School B's admission number returns generic unknown response and sends ZERO OTPs.
	res, err := svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, "ADM-B-200")
	if err != nil {
		t.Fatalf("unknown admission lookup failed: %v", err)
	}
	if res.State != whatsapp.StateAwaitingAdmissionNumber || res.Reply != whatsapp.UnknownAdmissionReply || len(smsSpy.messages) != 0 {
		t.Fatalf("expected generic unknown response and 0 SMS, got state=%s reply=%q sms=%d", res.State, res.Reply, len(smsSpy.messages))
	}

	// 2. Valid admission number in School A returns masked guardian numbers without exposing full numbers (AC-08, AC-09)
	res, err = svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, "ADM-A-100")
	if err != nil {
		t.Fatalf("valid admission lookup failed: %v", err)
	}
	if res.State != whatsapp.StateAwaitingGuardianSelection || len(res.GuardianCandidates) != 2 {
		t.Fatalf("expected AWAITING_GUARDIAN_SELECTION with 2 masked options, got %+v", res)
	}
	if strings.Contains(res.Reply, "+254722111482") || strings.Contains(res.Reply, "+254733222910") {
		t.Fatalf("full phone number leaked in reply: %s", res.Reply)
	}
	if !strings.Contains(res.Reply, "07•• ••• 482") || !strings.Contains(res.Reply, "07•• ••• 910") {
		t.Fatalf("expected masked guardian numbers in reply, got: %s", res.Reply)
	}

	// 3. Selecting guardian #2 sends OTP ONLY to that registered guardian phone, NEVER to unregisteredSender
	res, err = svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, "2")
	if err != nil {
		t.Fatalf("guardian selection failed: %v", err)
	}
	if res.State != whatsapp.StateAwaitingOTP {
		t.Fatalf("expected AWAITING_OTP, got %s", res.State)
	}
	sentTo, otpCode := smsSpy.lastOTP(t)
	if sentTo == unregisteredSender {
		t.Fatal("OTP was sent to requesting WhatsApp number instead of registered guardian number")
	}
	if sentTo != "+254722111482" && sentTo != "+254733222910" {
		t.Fatalf("unexpected OTP recipient %s", sentTo)
	}

	// 4. Cooldown enforcement (AC-10 + A3):
	// Requesting another OTP within 60 seconds from the same sender OR for the same (school_id, admission_number) from another sender returns 429 OTP_COOLDOWN_ACTIVE
	currentNow = currentNow.Add(30 * time.Second)
	otherSender := "+254799000222"
	if _, err := svc.HandleIncomingText(ctx, "", "pnid-school-a", "", otherSender, "ADM-A-100"); err != nil {
		t.Fatal(err)
	}
	_, err = svc.HandleIncomingText(ctx, "", "pnid-school-a", "", otherSender, "1")
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTooManyRequests || apiErr.Code != "OTP_COOLDOWN_ACTIVE" {
		t.Fatalf("expected 429 OTP_COOLDOWN_ACTIVE on per-admission cooldown within 60s, got %v", err)
	}

	// 5. Verify OTP establishes a 6-hour VERIFIED session and forwards queries to AI service (AC-11)
	res, err = svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, otpCode)
	if err != nil {
		t.Fatalf("OTP verification failed: %v", err)
	}
	if res.State != whatsapp.StateVerified || res.ExpiresAt == nil || !res.ExpiresAt.Equal(currentNow.Add(6*time.Hour)) {
		t.Fatalf("expected VERIFIED session expiring in 6h, got %+v", res)
	}

	// Query 1 minute before 6h expiry succeeds and forwards to AI
	currentNow = res.ExpiresAt.Add(-1 * time.Minute)
	queryRes, err := svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, "What is the fee balance?")
	if err != nil || queryRes.State != whatsapp.StateVerified || queryRes.Reply != "AI reply for What is the fee balance?" {
		t.Fatalf("expected AI reply before 6h expiry, got %+v (%v)", queryRes, err)
	}

	// Exact 6-hour expiry boundary (NOW() == expires_at) requires fresh OTP verification and does NOT call AI
	aiCallsBefore := len(aiSpy.calls)
	currentNow = *res.ExpiresAt
	expiredRes, err := svc.HandleIncomingText(ctx, "", "pnid-school-a", "", unregisteredSender, "What is the fee balance?")
	if err != nil {
		t.Fatalf("expired session call failed: %v", err)
	}
	if expiredRes.State != whatsapp.StateAwaitingAdmissionNumber || len(aiSpy.calls) != aiCallsBefore {
		t.Fatalf("expected expired session to reset to AWAITING_ADMISSION_NUMBER without calling AI, got state=%s aiCalls=%d", expiredRes.State, len(aiSpy.calls))
	}

	// 6. Rate Limit 2: 6th OTP request within 30 minutes blocks ONLY the requesting WhatsApp number for 24 hours,
	// leaving the guardian account and other WhatsApp numbers unblocked (AC-10, C20).
	abuserWA := "+254788999000"
	baseTime := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	choiceOne := 1
	for i := 1; i <= 5; i++ {
		currentNow = baseTime.Add(time.Duration(i*2) * time.Minute)
		r, err := svc.StartOrAdvanceSession(ctx, whatsapp.StartSessionInput{
			SchoolID:            schoolA.ID,
			SenderWANumber:      abuserWA,
			AdmissionNumber:     "ADM-A-100",
			GuardianChoiceIndex: &choiceOne,
		})
		if err != nil || r.State != whatsapp.StateAwaitingOTP {
			t.Fatalf("request %d within 30m should succeed, got %+v (%v)", i, r, err)
		}
		// Reset session back so next OTP request can be made after cooldown
		_, _ = db.Admin.Exec(ctx, `UPDATE chat_sessions SET state = 'AWAITING_ADMISSION_NUMBER' WHERE school_id = $1 AND sender_wa_number = $2`, schoolA.ID, abuserWA)
	}

	// 6th request within 30 minutes (at +12 minutes) triggers 24-hour sender block
	currentNow = baseTime.Add(12 * time.Minute)
	_, err = svc.StartOrAdvanceSession(ctx, whatsapp.StartSessionInput{
		SchoolID:            schoolA.ID,
		SenderWANumber:      abuserWA,
		AdmissionNumber:     "ADM-A-100",
		GuardianChoiceIndex: &choiceOne,
	})
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTooManyRequests || apiErr.Code != "WA_SENDER_BLOCKED_24H" {
		t.Fatalf("expected 429 WA_SENDER_BLOCKED_24H on 6th request in 30m, got %v", err)
	}

	// Verify guardian's user account is still active AND guardian can request OTP from their own WhatsApp number
	var guardianActive bool
	if err := db.Admin.QueryRow(ctx, `SELECT is_active FROM users WHERE id = $1`, schoolA.UserID).Scan(&guardianActive); err != nil || !guardianActive {
		t.Fatalf("guardian account must remain active when senderWA is blocked: active=%v err=%v", guardianActive, err)
	}
	currentNow = currentNow.Add(2 * time.Minute)
	legitRes, err := svc.StartOrAdvanceSession(ctx, whatsapp.StartSessionInput{
		SchoolID:            schoolA.ID,
		SenderWANumber:      "+254722111482",
		AdmissionNumber:     "ADM-A-100",
		GuardianChoiceIndex: &choiceOne,
	})
	if err != nil || legitRes.State != whatsapp.StateAwaitingOTP {
		t.Fatalf("expected guardian's own number to remain unblocked, got %+v (%v)", legitRes, err)
	}

	// Verify REST testing endpoints POST /api/v1/chat/sessions and POST /api/v1/chat/sessions/{id}/otp
	mux := http.NewServeMux()
	whatsapp.RegisterRoutes(mux, svc, nil)
	_, legitOTP := smsSpy.lastOTP(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/chat/sessions/"+legitRes.SessionID+"/otp?school_id="+schoolA.ID,
		bytes.NewBufferString(`{"otp":"`+legitOTP+`"}`)))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"state":"VERIFIED"`) {
		t.Fatalf("POST /api/v1/chat/sessions/{id}/otp failed: %d %s", rec.Code, rec.Body.String())
	}
}
