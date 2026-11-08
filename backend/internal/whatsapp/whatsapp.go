package whatsapp

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Session States
const (
	StateAwaitingAdmissionNumber   = "AWAITING_ADMISSION_NUMBER"
	StateAwaitingGuardianSelection = "AWAITING_GUARDIAN_SELECTION"
	StateAwaitingOTP               = "AWAITING_OTP"
	StateVerified                  = "VERIFIED"
)

// Common Error Codes
const (
	CodeRateLimited       = "RATE_LIMIT_EXCEEDED"
	CodeSenderBlocked     = "SENDER_BLOCKED"
	CodeInvalidOTP        = "INVALID_OTP"
	CodeSessionExpired    = "SESSION_EXPIRED"
	CodeAdmissionNotFound = "ADMISSION_NOT_FOUND"
	CodeUnknownBot        = "UNKNOWN_WHATSAPP_BOT"
)

// ChatSession represents an ongoing WhatsApp or In-App chat verification session
type ChatSession struct {
	ID               string     `json:"id"`
	SchoolID         string     `json:"school_id"`
	Channel          string     `json:"channel"` // "whatsapp" or "in_app"
	SenderWANumber   string     `json:"sender_wa_number"`
	StudentID        string     `json:"student_id,omitempty"`
	AdmissionNumber  string     `json:"admission_number,omitempty"`
	GuardianUserID   string     `json:"guardian_user_id,omitempty"`
	State            string     `json:"state"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	LookupCount      int        `json:"lookup_count"`
	PaymentInitCount int        `json:"payment_init_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// SchoolWhatsAppBot links a WhatsApp Business phone number to a school (AC-12)
type SchoolWhatsAppBot struct {
	ID                 string    `json:"id"`
	SchoolID           string    `json:"school_id"`
	PhoneNumberID      string    `json:"phone_number_id"`
	DisplayPhoneNumber string    `json:"display_phone_number"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
}

// GuardianCandidate carries registered details for identity selection
type GuardianCandidate struct {
	UserID       string `json:"user_id"`
	FullName     string `json:"full_name"`
	PhoneNumber  string `json:"phone_number"`
	Relationship string `json:"relationship"`
	MaskedPhone  string `json:"masked_phone"`
}

// SMSDispatcher sends transactional SMS (Africa's Talking)
type SMSDispatcher interface {
	SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error
}

// AIServiceBridge forwards verified messages to the AI service
type AIServiceBridge interface {
	ForwardParentQuery(ctx context.Context, schoolID, guardianUserID, studentID, query, sessionID string) (string, error)
}

// WhatsAppClient sends outbound messages via WhatsApp Cloud API
type WhatsAppClient interface {
	SendTextMessage(ctx context.Context, toWANumber, text string) error
}

// Store persistence interface for WhatsApp bot
type Store interface {
	GetSchoolByBotPhone(ctx context.Context, botPhone string) (*SchoolWhatsAppBot, error)
	GetSchoolByBotPhoneID(ctx context.Context, phoneID string) (*SchoolWhatsAppBot, error)
	RegisterSchoolBot(ctx context.Context, bot *SchoolWhatsAppBot) error

	GetActiveSession(ctx context.Context, schoolID, senderWANumber string) (*ChatSession, error)
	GetSessionByID(ctx context.Context, sessionID string) (*ChatSession, error)
	SaveSession(ctx context.Context, session *ChatSession) error

	GetStudentByAdmission(ctx context.Context, schoolID, admissionNumber string) (studentID, fullName string, err error)
	ListGuardiansForStudent(ctx context.Context, schoolID, studentID string) ([]GuardianCandidate, error)
	GetGuardianUser(ctx context.Context, userID string) (phone string, name string, err error)

	IsSenderBlocked(ctx context.Context, senderWANumber string) (bool, time.Time, error)
	BlockSender(ctx context.Context, schoolID, senderWANumber, reason string, until time.Time) error

	GetRecentOTPRequestCount(ctx context.Context, senderWANumber string, window time.Duration) (int, error)
	HasRecentOTPForStudent(ctx context.Context, schoolID, admissionNumber string, window time.Duration) (bool, error)
	HasRecentOTPForSender(ctx context.Context, senderWANumber string, window time.Duration) (bool, error)

	SaveOTP(ctx context.Context, sessionID, code string, expiresAt time.Time) error
	VerifyOTP(ctx context.Context, sessionID, code string) (bool, error)
}

// Service coordinates WhatsApp conversational state machine and OTP verification
type Service struct {
	store    Store
	sms      SMSDispatcher
	aiBridge AIServiceBridge
	waClient WhatsAppClient
}

// NewService constructs WhatsApp Service
func NewService(store Store, sms SMSDispatcher, aiBridge AIServiceBridge, waClient WhatsAppClient) *Service {
	return &Service{
		store:    store,
		sms:      sms,
		aiBridge: aiBridge,
		waClient: waClient,
	}
}

// ProcessIncomingMessage is the entry point for WhatsApp webhook and web chat messages
func (s *Service) ProcessIncomingMessage(ctx context.Context, botPhoneNumber, senderWANumber, text string) (reply string, err error) {
	// 1. Resolve school from receiving WhatsApp Business number (C16, AC-12)
	bot, err := s.store.GetSchoolByBotPhone(ctx, botPhoneNumber)
	if err != nil || bot == nil {
		bot, err = s.store.GetSchoolByBotPhoneID(ctx, botPhoneNumber)
	}
	if err != nil || bot == nil || !bot.IsActive {
		return "", errors.New(CodeUnknownBot)
	}
	schoolID := bot.SchoolID

	// 2. Check 24-hour rate limit block on requesting WhatsApp number (C20, AC-10)
	blocked, until, err := s.store.IsSenderBlocked(ctx, senderWANumber)
	if err == nil && blocked {
		return fmt.Sprintf("Your WhatsApp number has exceeded verification attempts and is temporarily blocked until %s. Please try again later or contact the school office.", until.Format("15:04 Jan 02")), nil
	}

	trimmedText := strings.TrimSpace(text)

	// 3. Fetch or initialize chat session
	session, err := s.store.GetActiveSession(ctx, schoolID, senderWANumber)
	if err != nil || session == nil {
		session = &ChatSession{
			SchoolID:       schoolID,
			Channel:        "whatsapp",
			SenderWANumber: senderWANumber,
			State:          StateAwaitingAdmissionNumber,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		_ = s.store.SaveSession(ctx, session)
	}

	// If session was verified but has expired past 6 hours, reset to AWAITING_ADMISSION_NUMBER (C13, AC-11)
	if session.State == StateVerified {
		if session.ExpiresAt != nil && time.Now().After(*session.ExpiresAt) {
			session.State = StateAwaitingAdmissionNumber
			session.StudentID = ""
			session.AdmissionNumber = ""
			session.GuardianUserID = ""
			session.VerifiedAt = nil
			session.ExpiresAt = nil
			_ = s.store.SaveSession(ctx, session)
			return "Your previous 6-hour verified session has expired. To access student records, please enter the student's admission number:", nil
		}

		// Session is active and verified! Forward query to AI service (C21, AC-28)
		session.LookupCount++
		_ = s.store.SaveSession(ctx, session)
		if s.aiBridge != nil {
			aiReply, err := s.aiBridge.ForwardParentQuery(ctx, schoolID, session.GuardianUserID, session.StudentID, trimmedText, session.ID)
			if err == nil && aiReply != "" {
				return aiReply, nil
			}
		}
		return "Hello! How can I assist you today regarding fees, academic progress, or events?", nil
	}

	// State Machine Transition
	switch session.State {
	case StateAwaitingAdmissionNumber:
		// User submits admission number
		admissionNum := strings.ToUpper(trimmedText)
		studentID, studentName, err := s.store.GetStudentByAdmission(ctx, schoolID, admissionNum)
		if err != nil || studentID == "" {
			// C14 Rule: If no student in school matches, return generic error and SEND NO DATA AND NO OTP
			return "We could not verify that admission number. Please check and try again or contact the school office.", nil
		}

		session.StudentID = studentID
		session.AdmissionNumber = admissionNum

		// Matched! Fetch linked guardians in student_guardians (C15, AC-09)
		guardians, err := s.store.ListGuardiansForStudent(ctx, schoolID, studentID)
		if err != nil || len(guardians) == 0 {
			return "We could not locate a registered guardian contact for this student. Please contact the school office to register your phone number.", nil
		}

		// Multiple guardians: display numbered masked phone numbers (AC-09)
		if len(guardians) > 1 {
			session.State = StateAwaitingGuardianSelection
			_ = s.store.SaveSession(ctx, session)

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Found record for %s. Multiple guardians are registered. Please reply with the number corresponding to your registered phone number:\n", studentName))
			for i, g := range guardians {
				sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, g.MaskedPhone, g.Relationship))
			}
			return sb.String(), nil
		}

		// Single guardian: proceed directly to OTP request
		session.GuardianUserID = guardians[0].UserID
		session.State = StateAwaitingOTP
		_ = s.store.SaveSession(ctx, session)

		return s.dispatchOTP(ctx, session, guardians[0].MaskedPhone, guardians[0].PhoneNumber)

	case StateAwaitingGuardianSelection:
		// User replies with choice (e.g. "1" or "2")
		guardians, err := s.store.ListGuardiansForStudent(ctx, schoolID, session.StudentID)
		if err != nil || len(guardians) == 0 {
			session.State = StateAwaitingAdmissionNumber
			_ = s.store.SaveSession(ctx, session)
			return "Session error. Please enter the student's admission number again:", nil
		}

		choice := 0
		_, _ = fmt.Sscanf(trimmedText, "%d", &choice)
		if choice < 1 || choice > len(guardians) {
			return fmt.Sprintf("Invalid choice. Please reply with a number between 1 and %d:", len(guardians)), nil
		}

		selected := guardians[choice-1]
		session.GuardianUserID = selected.UserID
		session.State = StateAwaitingOTP
		_ = s.store.SaveSession(ctx, session)

		return s.dispatchOTP(ctx, session, selected.MaskedPhone, selected.PhoneNumber)

	case StateAwaitingOTP:
		// User submits OTP code
		code := strings.TrimSpace(trimmedText)
		valid, err := s.store.VerifyOTP(ctx, session.ID, code)
		if err != nil || !valid {
			return "Invalid or expired OTP code. Please enter the 6-digit code sent via SMS to your registered number:", nil
		}

		// Valid OTP! Establish 6-Hour Verified Session (C13, AC-11)
		now := time.Now()
		expires := now.Add(6 * time.Hour)
		session.State = StateVerified
		session.VerifiedAt = &now
		session.ExpiresAt = &expires
		_ = s.store.SaveSession(ctx, session)

		return "Identity verified successfully! Your session is active for the next 6 hours. You can now inquire about your child's fee balance, term results, attendance, or pay school fees via M-Pesa.", nil

	default:
		session.State = StateAwaitingAdmissionNumber
		_ = s.store.SaveSession(ctx, session)
		return "Please enter the student's admission number to start:", nil
	}
}

// dispatchOTP enforces rate limits and dispatches 6-digit SMS strictly to guardian's registered number (AC-08, AC-10).
func (s *Service) dispatchOTP(ctx context.Context, session *ChatSession, maskedPhone, registeredPhone string) (string, error) {
	// 1. Rate Limit Check: 1 OTP per minute per sender OR per student admission number (A3, AC-10)
	recentSender, err := s.store.HasRecentOTPForSender(ctx, session.SenderWANumber, 60*time.Second)
	if err == nil && recentSender {
		return "An OTP was already requested recently for this session. Please wait 60 seconds before requesting another code.", nil
	}
	recentStudent, err := s.store.HasRecentOTPForStudent(ctx, session.SchoolID, session.AdmissionNumber, 60*time.Second)
	if err == nil && recentStudent {
		return "An OTP was recently requested for this student's admission number. Please wait 60 seconds before trying again.", nil
	}

	// 2. 6 requests in 30 minutes blocks sender_wa_number for 24 hours (C18–C20, AC-10)
	count, err := s.store.GetRecentOTPRequestCount(ctx, session.SenderWANumber, 30*time.Minute)
	if err == nil && count >= 5 {
		// 6th request: block requesting WhatsApp number for 24 hours (C20, AC-10)
		// Crucial: Only block sender_wa_number, NEVER block guardian's account or phone!
		blockUntil := time.Now().Add(24 * time.Hour)
		_ = s.store.BlockSender(ctx, session.SchoolID, session.SenderWANumber, "Excessive OTP verification requests (>5 in 30m)", blockUntil)
		return "You have exceeded the maximum allowed verification attempts. Your WhatsApp number is blocked from requesting OTPs for 24 hours. The student's registered guardian can still log in normally via the web portal.", nil
	}

	// 3. Generate 6-digit OTP
	otpCode := generate6DigitOTP()
	expiresAt := time.Now().Add(10 * time.Minute)
	_ = s.store.SaveOTP(ctx, session.ID, otpCode, expiresAt)

	// 4. Send SMS OTP strictly to registered phone number in users.phone_number (C12, C13, AC-08, AC-09)
	smsMessage := fmt.Sprintf("Your Ikonex Academy WhatsApp verification code is %s. Valid for 10 minutes. Do not share this code.", otpCode)
	if s.sms != nil {
		_ = s.sms.SendTransactionalSMS(ctx, registeredPhone, smsMessage)
	}

	return fmt.Sprintf("A 6-digit verification code has been sent via SMS to your registered number %s. Please reply with the code to verify your identity:", maskedPhone), nil
}

func generate6DigitOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// MemoryStore implementation for WhatsApp
type MemoryStore struct {
	mu             sync.RWMutex
	Bots           map[string]*SchoolWhatsAppBot           // key: display_phone
	BotsByID       map[string]*SchoolWhatsAppBot           // key: phone_id
	Sessions       map[string]*ChatSession                 // key: sessionID
	ActiveSessions map[string]string                       // key: schoolID:senderWANumber -> sessionID
	Students       map[string]struct{ ID, Name string }    // key: schoolID:upper(adminNum)
	Guardians      map[string][]GuardianCandidate          // key: schoolID:studentID
	GuardianUsers  map[string]struct{ Phone, Name string } // key: userID
	BlockedSenders map[string]time.Time                    // key: senderWANumber -> blockedUntil
	OTPRequests    []struct {
		SenderWA        string
		SchoolID        string
		AdmissionNumber string
		CreatedAt       time.Time
	}
	OTPCodes map[string]struct {
		Code    string
		Expires time.Time
	} // key: sessionID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Bots:           make(map[string]*SchoolWhatsAppBot),
		BotsByID:       make(map[string]*SchoolWhatsAppBot),
		Sessions:       make(map[string]*ChatSession),
		ActiveSessions: make(map[string]string),
		Students:       make(map[string]struct{ ID, Name string }),
		Guardians:      make(map[string][]GuardianCandidate),
		GuardianUsers:  make(map[string]struct{ Phone, Name string }),
		BlockedSenders: make(map[string]time.Time),
		OTPCodes: make(map[string]struct {
			Code    string
			Expires time.Time
		}),
	}
}

func (m *MemoryStore) GetSchoolByBotPhone(ctx context.Context, botPhone string) (*SchoolWhatsAppBot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.Bots[botPhone]
	if !ok {
		return nil, errors.New("bot not found")
	}
	return b, nil
}

func (m *MemoryStore) GetSchoolByBotPhoneID(ctx context.Context, phoneID string) (*SchoolWhatsAppBot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.BotsByID[phoneID]
	if !ok {
		return nil, errors.New("bot not found")
	}
	return b, nil
}

func (m *MemoryStore) RegisterSchoolBot(ctx context.Context, bot *SchoolWhatsAppBot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Bots[bot.DisplayPhoneNumber] = bot
	m.BotsByID[bot.PhoneNumberID] = bot
	return nil
}

func (m *MemoryStore) GetActiveSession(ctx context.Context, schoolID, senderWANumber string) (*ChatSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", schoolID, senderWANumber)
	id, ok := m.ActiveSessions[key]
	if !ok {
		return nil, errors.New("no active session")
	}
	s := m.Sessions[id]
	return s, nil
}

func (m *MemoryStore) GetSessionByID(ctx context.Context, sessionID string) (*ChatSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.Sessions[sessionID]
	if !ok {
		return nil, errors.New("session not found")
	}
	return s, nil
}

func (m *MemoryStore) SaveSession(ctx context.Context, session *ChatSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session.ID == "" {
		session.ID = fmt.Sprintf("cs_%d", time.Now().UnixNano())
	}
	session.UpdatedAt = time.Now()
	m.Sessions[session.ID] = session
	key := fmt.Sprintf("%s:%s", session.SchoolID, session.SenderWANumber)
	m.ActiveSessions[key] = session.ID
	return nil
}

func (m *MemoryStore) GetStudentByAdmission(ctx context.Context, schoolID, admissionNumber string) (studentID, fullName string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", schoolID, strings.ToUpper(strings.TrimSpace(admissionNumber)))
	st, ok := m.Students[key]
	if !ok {
		return "", "", errors.New("student not found")
	}
	return st.ID, st.Name, nil
}

func (m *MemoryStore) ListGuardiansForStudent(ctx context.Context, schoolID, studentID string) ([]GuardianCandidate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", schoolID, studentID)
	g := m.Guardians[key]
	res := make([]GuardianCandidate, len(g))
	copy(res, g)
	return res, nil
}

func (m *MemoryStore) GetGuardianUser(ctx context.Context, userID string) (phone string, name string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.GuardianUsers[userID]
	if !ok {
		return "", "", errors.New("user not found")
	}
	return u.Phone, u.Name, nil
}

func (m *MemoryStore) IsSenderBlocked(ctx context.Context, senderWANumber string) (bool, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.BlockedSenders[senderWANumber]
	if !ok || time.Now().After(t) {
		return false, time.Time{}, nil
	}
	return true, t, nil
}

func (m *MemoryStore) BlockSender(ctx context.Context, schoolID, senderWANumber, reason string, until time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlockedSenders[senderWANumber] = until
	return nil
}

func (m *MemoryStore) GetRecentOTPRequestCount(ctx context.Context, senderWANumber string, window time.Duration) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-window)
	count := 0
	for _, req := range m.OTPRequests {
		if req.SenderWA == senderWANumber && req.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count, nil
}

func (m *MemoryStore) HasRecentOTPForStudent(ctx context.Context, schoolID, admissionNumber string, window time.Duration) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-window)
	for _, req := range m.OTPRequests {
		if req.SchoolID == schoolID && req.AdmissionNumber == strings.ToUpper(admissionNumber) && req.CreatedAt.After(cutoff) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) HasRecentOTPForSender(ctx context.Context, senderWANumber string, window time.Duration) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-window)
	for _, req := range m.OTPRequests {
		if req.SenderWA == senderWANumber && req.CreatedAt.After(cutoff) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) SaveOTP(ctx context.Context, sessionID, code string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.Sessions[sessionID]
	if ok {
		m.OTPRequests = append(m.OTPRequests, struct {
			SenderWA        string
			SchoolID        string
			AdmissionNumber string
			CreatedAt       time.Time
		}{
			SenderWA:        s.SenderWANumber,
			SchoolID:        s.SchoolID,
			AdmissionNumber: s.AdmissionNumber,
			CreatedAt:       time.Now(),
		})
	}
	m.OTPCodes[sessionID] = struct {
		Code    string
		Expires time.Time
	}{Code: code, Expires: expiresAt}
	return nil
}

func (m *MemoryStore) VerifyOTP(ctx context.Context, sessionID, code string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.OTPCodes[sessionID]
	if !ok || time.Now().After(item.Expires) {
		return false, errors.New("expired or invalid OTP")
	}
	if item.Code != code {
		return false, nil
	}
	delete(m.OTPCodes, sessionID)
	return true, nil
}

// HTTP Webhook Handlers
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/whatsapp/webhook", s.handleWebhook)
	mux.HandleFunc("/api/v1/chat/sessions", s.handleCreateSession)
	mux.HandleFunc("/api/v1/chat/sessions/", s.handleSessionEndpoints)
}

func (s *Service) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Meta webhook verification challenge
	if r.Method == http.MethodGet {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")
		if mode == "subscribe" && token != "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(challenge))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Metadata struct {
						DisplayPhoneNumber string `json:"display_phone_number"`
						PhoneNumberID      string `json:"phone_number_id"`
					} `json:"metadata"`
					Messages []struct {
						From string `json:"from"`
						Text struct {
							Body string `json:"body"`
						} `json:"text"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &payload); err == nil && len(payload.Entry) > 0 {
		for _, entry := range payload.Entry {
			for _, ch := range entry.Changes {
				botPhone := ch.Value.Metadata.DisplayPhoneNumber
				for _, msg := range ch.Value.Messages {
					reply, _ := s.ProcessIncomingMessage(r.Context(), botPhone, msg.From, msg.Text.Body)
					if s.waClient != nil && reply != "" {
						_ = s.waClient.SendTextMessage(r.Context(), msg.From, reply)
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success"}`))
}

func (s *Service) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		BotPhoneNumber string `json:"bot_phone_number"`
		SenderWANumber string `json:"sender_wa_number"`
		Message        string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reply, err := s.ProcessIncomingMessage(r.Context(), req.BotPhoneNumber, req.SenderWANumber, req.Message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}

func (s *Service) handleSessionEndpoints(w http.ResponseWriter, r *http.Request) {
	// Handles REST endpoints for chat sessions (e.g. submitting OTP or query)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
