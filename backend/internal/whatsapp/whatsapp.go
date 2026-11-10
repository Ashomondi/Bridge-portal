package whatsapp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

const (
	StateAwaitingAdmissionNumber   = "AWAITING_ADMISSION_NUMBER"
	StateAwaitingGuardianSelection = "AWAITING_GUARDIAN_SELECTION"
	StateAwaitingOTP               = "AWAITING_OTP"
	StateVerified                  = "VERIFIED"

	UnknownAdmissionReply = "We could not verify that admission number. Please check and try again or contact the school office."
)

type BotResolver func(ctx context.Context, phoneNumberID, displayPhoneNumber string) (string, error)

type AIParentQueryInput struct {
	UserID       string    `json:"user_id"`
	SchoolID     string    `json:"school_id"`
	MembershipID string    `json:"membership_id"`
	Role         string    `json:"role"`
	Channel      string    `json:"channel"`
	SessionID    string    `json:"session_id"`
	StudentID    string    `json:"student_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	Message      string    `json:"message"`
}

type AIForwarder interface {
	ForwardParentQuery(ctx context.Context, input AIParentQueryInput) (string, error)
}

type OutboundSender interface {
	SendTextMessage(ctx context.Context, schoolID, phoneNumberID, toWhatsAppNumber, body string) error
}

type GuardianOption struct {
	Index       int    `json:"index"`
	UserID      string `json:"-"`
	PhoneNumber string `json:"-"`
	MaskedPhone string `json:"masked_phone"`
}

type SessionResponse struct {
	SessionID          string           `json:"session_id"`
	SchoolID           string           `json:"school_id"`
	State              string           `json:"state"`
	StudentID          *string          `json:"student_id,omitempty"`
	GuardianCandidates []GuardianOption `json:"guardian_candidates,omitempty"`
	Reply              string           `json:"reply"`
	VerifiedAt         *time.Time       `json:"verified_at,omitempty"`
	ExpiresAt          *time.Time       `json:"expires_at,omitempty"`
}

type StartSessionInput struct {
	SchoolID            string `json:"school_id"`
	PhoneNumberID       string `json:"phone_number_id"`
	DisplayPhoneNumber  string `json:"display_phone_number"`
	SenderWANumber      string `json:"sender_wa_number"`
	AdmissionNumber     string `json:"admission_number"`
	GuardianChoiceIndex *int   `json:"guardian_choice_index,omitempty"`
	Message             string `json:"message,omitempty"`
}

type SessionOTPInput struct {
	SenderWANumber      string `json:"sender_wa_number"`
	GuardianChoiceIndex *int   `json:"guardian_choice_index,omitempty"`
	OTP                 string `json:"otp,omitempty"`
}

type rateLimiterState struct {
	mu                  sync.Mutex
	senderRequests      map[string][]time.Time
	senderBlockedUntil  map[string]time.Time
	admissionLastSentAt map[string]time.Time
}

func newRateLimiterState() *rateLimiterState {
	return &rateLimiterState{
		senderRequests:      make(map[string][]time.Time),
		senderBlockedUntil:  make(map[string]time.Time),
		admissionLastSentAt: make(map[string]time.Time),
	}
}

type Service struct {
	DB                  contracts.TenantDB
	SMS                 notifications.SMSDispatcher
	ResolveBot          BotResolver
	AI                  AIForwarder
	Outbound            OutboundSender
	WebhookVerifyToken  string
	WebhookAppSecret    string
	OTPCooldown         time.Duration
	OTPWindow           time.Duration
	MaxOTPPerWindow     int
	SenderBlockDuration time.Duration
	SessionTTL          time.Duration
	OTPTTL              time.Duration
	Now                 func() time.Time

	once    sync.Once
	limiter *rateLimiterState
}

func (s *Service) initLimiter() *rateLimiterState {
	s.once.Do(func() {
		if s.limiter == nil {
			s.limiter = newRateLimiterState()
		}
	})
	return s.limiter
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) otpCooldown() time.Duration {
	if s != nil && s.OTPCooldown > 0 {
		return s.OTPCooldown
	}
	return 60 * time.Second
}

func (s *Service) otpWindow() time.Duration {
	if s != nil && s.OTPWindow > 0 {
		return s.OTPWindow
	}
	return 30 * time.Minute
}

func (s *Service) maxOTPPerWindow() int {
	if s != nil && s.MaxOTPPerWindow > 0 {
		return s.MaxOTPPerWindow
	}
	return 5
}

func (s *Service) senderBlockDuration() time.Duration {
	if s != nil && s.SenderBlockDuration > 0 {
		return s.SenderBlockDuration
	}
	return 24 * time.Hour
}

func (s *Service) sessionTTL() time.Duration {
	if s != nil && s.SessionTTL > 0 {
		return s.SessionTTL
	}
	return 6 * time.Hour
}

func (s *Service) otpTTL() time.Duration {
	if s != nil && s.OTPTTL > 0 {
		return s.OTPTTL
	}
	return 10 * time.Minute
}

func (s *Service) resolveSchoolID(ctx context.Context, explicitSchoolID, phoneNumberID, displayPhone string) (string, error) {
	if contracts.ValidID(strings.TrimSpace(explicitSchoolID)) {
		return strings.TrimSpace(explicitSchoolID), nil
	}
	if s != nil && s.ResolveBot != nil {
		schoolID, err := s.ResolveBot(ctx, strings.TrimSpace(phoneNumberID), strings.TrimSpace(displayPhone))
		if err == nil && contracts.ValidID(schoolID) {
			return schoolID, nil
		}
	}
	return "", contracts.Fail(http.StatusBadRequest, "UNKNOWN_WHATSAPP_BOT", "Could not resolve school for the WhatsApp Business number.")
}

func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func generateSixDigitOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HandleIncomingText drives the conversational WhatsApp state machine for both
// the webhook and the REST testing endpoints.
func (s *Service) HandleIncomingText(ctx context.Context, schoolID, phoneNumberID, displayPhone, rawSenderWA, text string) (SessionResponse, error) {
	resolvedSchoolID, err := s.resolveSchoolID(ctx, schoolID, phoneNumberID, displayPhone)
	if err != nil {
		return SessionResponse{}, err
	}
	senderWA, err := notifications.NormalizePhone(rawSenderWA)
	if err != nil {
		return SessionResponse{}, contracts.Fail(http.StatusBadRequest, "INVALID_SENDER_PHONE", "A valid sender WhatsApp number is required.")
	}
	trimmedText := strings.TrimSpace(text)

	if s == nil || s.DB == nil {
		return SessionResponse{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}

	var resp SessionResponse
	var pendingSMS *notifications.SMSMessage
	var pendingAI *AIParentQueryInput

	err = s.DB.WithTenantTx(ctx, resolvedSchoolID, func(tx pgx.Tx) error {
		if phoneNumberID != "" || displayPhone != "" {
			var botActive bool
			if err := tx.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM school_whatsapp_bots
					WHERE school_id = $1 AND is_active
					  AND ($2 = '' OR phone_number_id = $2)
					  AND ($3 = '' OR display_phone_number = $3)
				)`,
				resolvedSchoolID, strings.TrimSpace(phoneNumberID), strings.TrimSpace(displayPhone),
			).Scan(&botActive); err != nil {
				return err
			}
			if !botActive {
				return contracts.Fail(http.StatusBadRequest, "UNKNOWN_WHATSAPP_BOT", "WhatsApp bot is not active for this school.")
			}
		}

		session, createdNew, err := s.getOrCreateSessionTx(ctx, tx, resolvedSchoolID, senderWA)
		if err != nil {
			return err
		}

		now := s.now()
		// Check 6-hour session expiry boundary: NOW() >= expires_at requires fresh OTP verification.
		if session.State == StateVerified && (session.ExpiresAt == nil || !now.Before(*session.ExpiresAt)) {
			if err := s.resetSessionToAwaitingAdmissionTx(ctx, tx, resolvedSchoolID, session.ID); err != nil {
				return err
			}
			session.State = StateAwaitingAdmissionNumber
			session.StudentID = nil
			session.GuardianUserID = nil
			session.VerifiedAt = nil
			session.ExpiresAt = nil
			resp = SessionResponse{
				SessionID: session.ID,
				SchoolID:  resolvedSchoolID,
				State:     StateAwaitingAdmissionNumber,
				Reply:     "Your 6-hour verification session has expired. Please enter your child's admission number to verify again.",
			}
			return nil
		}

		switch session.State {
		case StateAwaitingAdmissionNumber:
			if createdNew && (trimmedText == "" || isGreetingInput(trimmedText)) {
				resp = SessionResponse{
					SessionID: session.ID,
					SchoolID:  resolvedSchoolID,
					State:     StateAwaitingAdmissionNumber,
					Reply:     "Welcome to the school WhatsApp portal. Please reply with your child's admission number.",
				}
				return nil
			}
			if trimmedText == "" {
				resp = SessionResponse{
					SessionID: session.ID,
					SchoolID:  resolvedSchoolID,
					State:     StateAwaitingAdmissionNumber,
					Reply:     "Please reply with your child's admission number.",
				}
				return nil
			}
			return s.transitionAdmissionLookupTx(ctx, tx, resolvedSchoolID, &session, trimmedText, &resp)

		case StateAwaitingGuardianSelection:
			if session.StudentID == nil {
				return s.transitionAdmissionLookupTx(ctx, tx, resolvedSchoolID, &session, trimmedText, &resp)
			}
			return s.transitionGuardianSelectionToOTPTx(ctx, tx, resolvedSchoolID, senderWA, &session, trimmedText, &resp, &pendingSMS)

		case StateAwaitingOTP:
			return s.transitionVerifyOTPTx(ctx, tx, resolvedSchoolID, &session, trimmedText, &resp)

		case StateVerified:
			var membershipID, role string
			if err := tx.QueryRow(ctx, `
				SELECT id, role
				FROM memberships
				WHERE school_id = $1 AND user_id = $2 AND status = 'active'
				ORDER BY CASE WHEN role = 'parent' THEN 0 ELSE 1 END
				LIMIT 1`,
				resolvedSchoolID, *session.GuardianUserID,
			).Scan(&membershipID, &role); err != nil {
				return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Active school membership not found for guardian.")
			}
			pendingAI = &AIParentQueryInput{
				UserID:       *session.GuardianUserID,
				SchoolID:     resolvedSchoolID,
				MembershipID: membershipID,
				Role:         role,
				Channel:      "whatsapp",
				SessionID:    session.ID,
				StudentID:    *session.StudentID,
				ExpiresAt:    *session.ExpiresAt,
				Message:      trimmedText,
			}
			resp = SessionResponse{
				SessionID:  session.ID,
				SchoolID:   resolvedSchoolID,
				State:      StateVerified,
				StudentID:  session.StudentID,
				VerifiedAt: session.VerifiedAt,
				ExpiresAt:  session.ExpiresAt,
			}
			return nil
		default:
			return contracts.Fail(http.StatusInternalServerError, "INVALID_SESSION_STATE", "Unrecognized chat session state.")
		}
	})
	if err != nil {
		return SessionResponse{}, err
	}

	if pendingSMS != nil && s.SMS != nil {
		if err := s.SMS.DispatchSMS(ctx, *pendingSMS); err != nil {
			return SessionResponse{}, err
		}
	}

	if pendingAI != nil {
		if s.AI != nil {
			aiReply, err := s.AI.ForwardParentQuery(ctx, *pendingAI)
			if err != nil {
				return SessionResponse{}, err
			}
			resp.Reply = aiReply
		} else {
			resp.Reply = "Your session is verified."
		}
	}

	if s.Outbound != nil && strings.TrimSpace(phoneNumberID) != "" && strings.TrimSpace(resp.Reply) != "" {
		_ = s.Outbound.SendTextMessage(ctx, resolvedSchoolID, phoneNumberID, senderWA, resp.Reply)
	}

	return resp, nil
}

func isGreetingInput(text string) bool {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "hi", "hello", "hey", "start", "menu", "help", "jambo", "habari":
		return true
	default:
		return false
	}
}

type chatSessionRow struct {
	ID             string
	SchoolID       string
	SenderWANumber string
	StudentID      *string
	GuardianUserID *string
	State          string
	OTPHash        *string
	OTPExpiresAt   *time.Time
	OTPAttempts    int
	VerifiedAt     *time.Time
	ExpiresAt      *time.Time
}

func (s *Service) getOrCreateSessionTx(ctx context.Context, tx pgx.Tx, schoolID, senderWA string) (chatSessionRow, bool, error) {
	var row chatSessionRow
	err := tx.QueryRow(ctx, `
		SELECT id, school_id, sender_wa_number, student_id, guardian_user_id, state,
		       otp_hash, otp_expires_at, otp_attempts, verified_at, expires_at
		FROM chat_sessions
		WHERE school_id = $1 AND channel = 'whatsapp' AND sender_wa_number = $2
		FOR UPDATE`,
		schoolID, senderWA,
	).Scan(
		&row.ID, &row.SchoolID, &row.SenderWANumber, &row.StudentID, &row.GuardianUserID,
		&row.State, &row.OTPHash, &row.OTPExpiresAt, &row.OTPAttempts, &row.VerifiedAt, &row.ExpiresAt,
	)
	if err == nil {
		return row, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return chatSessionRow{}, false, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO chat_sessions (school_id, channel, sender_wa_number, state)
		VALUES ($1, 'whatsapp', $2, $3)
		RETURNING id, school_id, sender_wa_number, student_id, guardian_user_id, state,
		          otp_hash, otp_expires_at, otp_attempts, verified_at, expires_at`,
		schoolID, senderWA, StateAwaitingAdmissionNumber,
	).Scan(
		&row.ID, &row.SchoolID, &row.SenderWANumber, &row.StudentID, &row.GuardianUserID,
		&row.State, &row.OTPHash, &row.OTPExpiresAt, &row.OTPAttempts, &row.VerifiedAt, &row.ExpiresAt,
	)
	return row, true, err
}

func (s *Service) resetSessionToAwaitingAdmissionTx(ctx context.Context, tx pgx.Tx, schoolID, sessionID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE chat_sessions
		SET state = $1,
		    student_id = NULL,
		    guardian_user_id = NULL,
		    otp_hash = NULL,
		    otp_expires_at = NULL,
		    otp_attempts = 0,
		    verified_at = NULL,
		    expires_at = NULL,
		    updated_at = NOW()
		WHERE school_id = $2 AND id = $3`,
		StateAwaitingAdmissionNumber, schoolID, sessionID,
	)
	return err
}

func (s *Service) transitionAdmissionLookupTx(ctx context.Context, tx pgx.Tx, schoolID string, session *chatSessionRow, admissionInput string, resp *SessionResponse) error {
	var studentID string
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM students
		WHERE school_id = $1
		  AND status = 'active'
		  AND UPPER(TRIM(admission_number)) = UPPER(TRIM($2))`,
		schoolID, admissionInput,
	).Scan(&studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		*resp = SessionResponse{
			SessionID: session.ID,
			SchoolID:  schoolID,
			State:     StateAwaitingAdmissionNumber,
			Reply:     UnknownAdmissionReply,
		}
		return nil
	}
	if err != nil {
		return err
	}

	options, err := loadMaskedGuardiansForStudentTx(ctx, tx, schoolID, studentID)
	if err != nil {
		return err
	}
	if len(options) == 0 {
		*resp = SessionResponse{
			SessionID: session.ID,
			SchoolID:  schoolID,
			State:     StateAwaitingAdmissionNumber,
			Reply:     UnknownAdmissionReply,
		}
		return nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE chat_sessions
		SET state = $1,
		    student_id = $2,
		    guardian_user_id = NULL,
		    otp_hash = NULL,
		    otp_expires_at = NULL,
		    otp_attempts = 0,
		    updated_at = NOW()
		WHERE school_id = $3 AND id = $4`,
		StateAwaitingGuardianSelection, studentID, schoolID, session.ID,
	); err != nil {
		return err
	}

	var replyBuilder strings.Builder
	if len(options) == 1 {
		fmt.Fprintf(&replyBuilder, "Confirm guardian phone number to receive SMS OTP:\n1. %s\nReply 1 to send OTP.", options[0].MaskedPhone)
	} else {
		replyBuilder.WriteString("Select your registered phone number to receive an SMS OTP:\n")
		for i, opt := range options {
			if i > 0 {
				replyBuilder.WriteString("\n")
			}
			fmt.Fprintf(&replyBuilder, "%d. %s", opt.Index, opt.MaskedPhone)
		}
	}

	*resp = SessionResponse{
		SessionID:          session.ID,
		SchoolID:           schoolID,
		State:              StateAwaitingGuardianSelection,
		StudentID:          &studentID,
		GuardianCandidates: options,
		Reply:              replyBuilder.String(),
	}
	return nil
}

func loadMaskedGuardiansForStudentTx(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]GuardianOption, error) {
	rows, err := tx.Query(ctx, `
		SELECT u.id, u.phone_number
		FROM student_guardians sg
		JOIN users u ON u.id = sg.user_id
		WHERE sg.school_id = $1 AND sg.student_id = $2 AND u.is_active
		ORDER BY sg.is_primary DESC, sg.created_at ASC, u.id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type rawGuardian struct {
		userID string
		phone  string
	}
	var rawList []rawGuardian
	for rows.Next() {
		var rg rawGuardian
		if err := rows.Scan(&rg.userID, &rg.phone); err != nil {
			return nil, err
		}
		rawList = append(rawList, rg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(rawList, func(i, j int) bool {
		return rawList[i].userID < rawList[j].userID
	})

	options := make([]GuardianOption, 0, len(rawList))
	for idx, rg := range rawList {
		norm, err := notifications.NormalizePhone(rg.phone)
		if err != nil {
			continue
		}
		masked, err := MaskPhone(norm)
		if err != nil {
			continue
		}
		options = append(options, GuardianOption{
			Index:       idx + 1,
			UserID:      rg.userID,
			PhoneNumber: norm,
			MaskedPhone: masked,
		})
	}
	return options, nil
}

func (s *Service) transitionGuardianSelectionToOTPTx(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, senderWA string,
	session *chatSessionRow,
	selectionInput string,
	resp *SessionResponse,
	pendingSMS **notifications.SMSMessage,
) error {
	options, err := loadMaskedGuardiansForStudentTx(ctx, tx, schoolID, *session.StudentID)
	if err != nil {
		return err
	}
	if len(options) == 0 {
		return contracts.Fail(http.StatusUnprocessableEntity, "NO_LINKED_GUARDIANS", "No active guardians are linked to this student.")
	}

	choiceIdx := 0
	trimmed := strings.TrimSpace(selectionInput)
	if len(options) == 1 && (trimmed == "1" || strings.EqualFold(trimmed, "yes") || strings.EqualFold(trimmed, "confirm")) {
		choiceIdx = 1
	} else if parsed, err := strconv.Atoi(trimmed); err == nil {
		choiceIdx = parsed
	}
	if choiceIdx < 1 || choiceIdx > len(options) {
		return contracts.Fail(http.StatusUnprocessableEntity, "INVALID_GUARDIAN_SELECTION", "Please select a valid numbered guardian option.")
	}
	chosen := options[choiceIdx-1]

	var admissionNumber string
	if err := tx.QueryRow(ctx, `SELECT admission_number FROM students WHERE school_id = $1 AND id = $2`, schoolID, *session.StudentID).Scan(&admissionNumber); err != nil {
		return err
	}

	if err := s.enforceOTPRateLimitsTx(ctx, tx, schoolID, senderWA, admissionNumber); err != nil {
		return err
	}

	otpCode, err := generateSixDigitOTP()
	if err != nil {
		return err
	}
	now := s.now()
	otpExpires := now.Add(s.otpTTL())
	otpHash := hashSecret(otpCode)

	if _, err := tx.Exec(ctx, `
		UPDATE chat_sessions
		SET state = $1,
		    guardian_user_id = $2,
		    otp_hash = $3,
		    otp_expires_at = $4,
		    otp_attempts = 0,
		    updated_at = NOW()
		WHERE school_id = $5 AND id = $6`,
		StateAwaitingOTP, chosen.UserID, otpHash, otpExpires, schoolID, session.ID,
	); err != nil {
		return err
	}

	// Send OTP strictly to chosen.PhoneNumber (the registered guardian number in users.phone_number),
	// NEVER to senderWA.
	*pendingSMS = &notifications.SMSMessage{
		To:      chosen.PhoneNumber,
		Purpose: notifications.WhatsAppSessionOTP,
		Body:    fmt.Sprintf("Your Ikonex Academy WhatsApp verification code is %s. Valid for 10 minutes.", otpCode),
	}

	*resp = SessionResponse{
		SessionID: session.ID,
		SchoolID:  schoolID,
		State:     StateAwaitingOTP,
		StudentID: session.StudentID,
		Reply:     fmt.Sprintf("An SMS verification code has been sent to %s. Please reply with the 6-digit code.", chosen.MaskedPhone),
	}
	return nil
}

func (s *Service) enforceOTPRateLimitsTx(ctx context.Context, tx pgx.Tx, schoolID, senderWA, admissionNumber string) error {
	now := s.now()
	lim := s.initLimiter()
	lim.mu.Lock()
	defer lim.mu.Unlock()

	// Check DB sender_rate_blocks if migration 000002 table exists
	hasBlocksTable, err := tableExistsTx(ctx, tx, "sender_rate_blocks")
	if err != nil {
		return err
	}
	if hasBlocksTable {
		var dbBlocked bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM sender_rate_blocks
				WHERE sender_wa_number = $1 AND blocked_until > $2
			)`,
			senderWA, now,
		).Scan(&dbBlocked); err != nil {
			return err
		}
		if dbBlocked {
			return contracts.Fail(http.StatusTooManyRequests, "WA_SENDER_BLOCKED_24H", "This WhatsApp number is temporarily blocked for 24 hours due to excessive OTP requests.")
		}
	}

	if until, blocked := lim.senderBlockedUntil[senderWA]; blocked && now.Before(until) {
		return contracts.Fail(http.StatusTooManyRequests, "WA_SENDER_BLOCKED_24H", "This WhatsApp number is temporarily blocked for 24 hours due to excessive OTP requests.")
	}

	// Prune requests outside the 30-minute window
	windowCutoff := now.Add(-s.otpWindow())
	history := lim.senderRequests[senderWA]
	validHistory := history[:0]
	for _, ts := range history {
		if !ts.Before(windowCutoff) {
			validHistory = append(validHistory, ts)
		}
	}
	lim.senderRequests[senderWA] = validHistory

	// Limit 2: 6th request within 30 minutes (> 5 in 30m) triggers a 24-hour block on senderWA ONLY.
	// Note: if the 6th request occurs within 30 minutes, check whether 5 requests already occurred in the window.
	if len(validHistory) >= s.maxOTPPerWindow() {
		blockedUntil := now.Add(s.senderBlockDuration())
		lim.senderBlockedUntil[senderWA] = blockedUntil
		if hasBlocksTable {
			_, _ = tx.Exec(ctx, `
				INSERT INTO sender_rate_blocks (sender_wa_number, blocked_until, reason)
				VALUES ($1, $2, 'otp_rate_limit')
				ON CONFLICT (sender_wa_number)
				DO UPDATE SET blocked_until = EXCLUDED.blocked_until`,
				senderWA, blockedUntil,
			)
		}
		return contracts.Fail(http.StatusTooManyRequests, "WA_SENDER_BLOCKED_24H", "Too many OTP requests in 30 minutes. This WhatsApp number is blocked for 24 hours.")
	}

	// Limit 1a: 60-second cooldown per requesting WhatsApp number
	if len(validHistory) > 0 {
		lastSenderReq := validHistory[len(validHistory)-1]
		if now.Sub(lastSenderReq) < s.otpCooldown() {
			return contracts.Fail(http.StatusTooManyRequests, "OTP_COOLDOWN_ACTIVE", "Please wait at least 60 seconds before requesting another OTP.")
		}
	}

	// Limit 1b: 60-second cooldown per (school_id, admission_number)
	admKey := schoolID + ":" + strings.ToUpper(strings.TrimSpace(admissionNumber))
	if lastAdmReq, ok := lim.admissionLastSentAt[admKey]; ok && now.Sub(lastAdmReq) < s.otpCooldown() {
		return contracts.Fail(http.StatusTooManyRequests, "OTP_COOLDOWN_ACTIVE", "An OTP was recently requested for this admission number. Please wait 60 seconds.")
	}

	lim.senderRequests[senderWA] = append(validHistory, now)
	lim.admissionLastSentAt[admKey] = now
	return nil
}

func tableExistsTx(ctx context.Context, tx pgx.Tx, tableName string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`,
		tableName,
	).Scan(&exists)
	return exists, err
}

func (s *Service) transitionVerifyOTPTx(ctx context.Context, tx pgx.Tx, schoolID string, session *chatSessionRow, codeInput string, resp *SessionResponse) error {
	now := s.now()
	code := strings.TrimSpace(codeInput)
	if len(code) != 6 || session.OTPHash == nil || session.OTPExpiresAt == nil || !now.Before(*session.OTPExpiresAt) {
		return contracts.Fail(http.StatusUnauthorized, "INVALID_OTP", "The verification code is invalid or has expired.")
	}
	candidateHash := hashSecret(code)
	if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(*session.OTPHash)) != 1 {
		newAttempts := session.OTPAttempts + 1
		_, _ = tx.Exec(ctx, `
			UPDATE chat_sessions
			SET otp_attempts = $1, updated_at = NOW()
			WHERE school_id = $2 AND id = $3`,
			newAttempts, schoolID, session.ID,
		)
		return contracts.Fail(http.StatusUnauthorized, "INVALID_OTP", "The verification code is invalid or has expired.")
	}

	verifiedAt := now
	expiresAt := now.Add(s.sessionTTL())
	if _, err := tx.Exec(ctx, `
		UPDATE chat_sessions
		SET state = $1,
		    otp_hash = NULL,
		    otp_expires_at = NULL,
		    otp_attempts = 0,
		    verified_at = $2,
		    expires_at = $3,
		    updated_at = NOW()
		WHERE school_id = $4 AND id = $5`,
		StateVerified, verifiedAt, expiresAt, schoolID, session.ID,
	); err != nil {
		return err
	}

	*resp = SessionResponse{
		SessionID:  session.ID,
		SchoolID:   schoolID,
		State:      StateVerified,
		StudentID:  session.StudentID,
		VerifiedAt: &verifiedAt,
		ExpiresAt:  &expiresAt,
		Reply:      "Verification complete. Your WhatsApp session is active for 6 hours. How can we help you today?",
	}
	return nil
}

// StartOrAdvanceSession handles POST /api/v1/chat/sessions using the same underlying state machine.
func (s *Service) StartOrAdvanceSession(ctx context.Context, in StartSessionInput) (SessionResponse, error) {
	inputText := strings.TrimSpace(in.AdmissionNumber)
	if in.GuardianChoiceIndex != nil {
		inputText = strconv.Itoa(*in.GuardianChoiceIndex)
	} else if inputText == "" {
		inputText = strings.TrimSpace(in.Message)
	}
	if strings.TrimSpace(in.AdmissionNumber) != "" && in.GuardianChoiceIndex != nil {
		// Two-step convenience for REST testing: first lookup admission number, then select guardian
		step1, err := s.HandleIncomingText(ctx, in.SchoolID, in.PhoneNumberID, in.DisplayPhoneNumber, in.SenderWANumber, in.AdmissionNumber)
		if err != nil || step1.State != StateAwaitingGuardianSelection {
			return step1, err
		}
		return s.HandleIncomingText(ctx, in.SchoolID, in.PhoneNumberID, in.DisplayPhoneNumber, in.SenderWANumber, strconv.Itoa(*in.GuardianChoiceIndex))
	}
	return s.HandleIncomingText(ctx, in.SchoolID, in.PhoneNumberID, in.DisplayPhoneNumber, in.SenderWANumber, inputText)
}

// HandleSessionOTP handles POST /api/v1/chat/sessions/{id}/otp using the same underlying state machine.
func (s *Service) HandleSessionOTP(ctx context.Context, schoolIDHint, sessionID string, in SessionOTPInput) (SessionResponse, error) {
	if !contracts.ValidID(sessionID) {
		return SessionResponse{}, contracts.Fail(http.StatusBadRequest, "INVALID_SESSION_ID", "A valid session ID is required.")
	}
	schoolID, err := s.resolveSchoolID(ctx, schoolIDHint, "", "")
	if err != nil {
		return SessionResponse{}, err
	}
	var senderWA string
	err = s.DB.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT sender_wa_number
			FROM chat_sessions
			WHERE school_id = $1 AND id = $2`,
			schoolID, sessionID,
		).Scan(&senderWA)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionResponse{}, contracts.Fail(http.StatusNotFound, "SESSION_NOT_FOUND", "Chat session not found.")
	}
	if err != nil {
		return SessionResponse{}, err
	}
	if strings.TrimSpace(in.SenderWANumber) != "" {
		norm, err := notifications.NormalizePhone(in.SenderWANumber)
		if err != nil || norm != senderWA {
			return SessionResponse{}, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Sender WhatsApp number does not match this session.")
		}
	}
	if in.GuardianChoiceIndex != nil && strings.TrimSpace(in.OTP) == "" {
		return s.HandleIncomingText(ctx, schoolID, "", "", senderWA, strconv.Itoa(*in.GuardianChoiceIndex))
	}
	return s.HandleIncomingText(ctx, schoolID, "", "", senderWA, in.OTP)
}

type MetaWebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Messages []struct {
					From string `json:"from"`
					ID   string `json:"id"`
					Type string `json:"type"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func verifyWebhookSignature(appSecret, signatureHeader string, body []byte) bool {
	if strings.TrimSpace(appSecret) == "" {
		return true
	}
	sig := strings.TrimPrefix(strings.TrimSpace(signatureHeader), "sha256=")
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(expected)) == 1
}

func RegisterRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	if service == nil {
		service = &Service{}
	}
	mux.HandleFunc("GET /api/v1/whatsapp/webhook", func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")
		if mode != "subscribe" || service.WebhookVerifyToken == "" || subtle.ConstantTimeCompare([]byte(token), []byte(service.WebhookVerifyToken)) != 1 {
			contracts.WriteError(w, contracts.Fail(http.StatusForbidden, "WEBHOOK_VERIFICATION_FAILED", "Webhook verification failed."))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(challenge))
	})

	mux.HandleFunc("POST /api/v1/whatsapp/webhook", func(w http.ResponseWriter, r *http.Request) {
		var payload MetaWebhookPayload
		if err := contracts.Decode(w, r, &payload); err != nil {
			contracts.WriteError(w, err)
			return
		}
		var responses []SessionResponse
		for _, entry := range payload.Entry {
			for _, change := range entry.Changes {
				meta := change.Value.Metadata
				for _, msg := range change.Value.Messages {
					if msg.From == "" {
						continue
					}
					res, err := service.HandleIncomingText(
						r.Context(),
						r.URL.Query().Get("school_id"),
						meta.PhoneNumberID,
						meta.DisplayPhoneNumber,
						msg.From,
						msg.Text.Body,
					)
					if err != nil {
						contracts.WriteError(w, err)
						return
					}
					responses = append(responses, res)
				}
			}
		}
		contracts.Success(w, http.StatusOK, map[string]any{"responses": responses}, nil)
	})

	mux.HandleFunc("POST /api/v1/chat/sessions", func(w http.ResponseWriter, r *http.Request) {
		var input StartSessionInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		if input.SchoolID == "" && identity != nil {
			if p, err := contracts.RequireIdentity(r, identity); err == nil {
				input.SchoolID = p.SchoolID
			}
		}
		res, err := service.StartOrAdvanceSession(r.Context(), input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})

	mux.HandleFunc("POST /api/v1/chat/sessions/{id}/otp", func(w http.ResponseWriter, r *http.Request) {
		var input SessionOTPInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		schoolID := r.URL.Query().Get("school_id")
		if schoolID == "" && identity != nil {
			if p, err := contracts.RequireIdentity(r, identity); err == nil {
				schoolID = p.SchoolID
			}
		}
		res, err := service.HandleSessionOTP(r.Context(), schoolID, r.PathValue("id"), input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})
}
