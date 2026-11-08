package whatsapp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func maskPhoneNumber(phone string) string {
	clean := strings.TrimSpace(phone)
	if len(clean) < 6 {
		return clean
	}
	prefix := clean[:len(clean)-6]
	suffix := clean[len(clean)-2:]
	return prefix + "****" + suffix
}

// SQLStore is the PostgreSQL implementation of whatsapp.Store.
// It persists sessions, bots, rate limits, and hashed OTP requests to PostgreSQL.
type SQLStore struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

type sqlQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *SQLStore) q() sqlQuerier {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *SQLStore) WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.db == nil {
		return errors.New("whatsapp: SQLStore has no database connection")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if schoolID != "" {
		if _, err := tx.ExecContext(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	tenantStore := &SQLStore{db: s.db, tx: tx}
	if err := fn(tenantStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *SQLStore) GetSchoolByBotPhone(ctx context.Context, botPhone string) (*SchoolWhatsAppBot, error) {
	query := `
		SELECT id, school_id, phone_number_id, display_phone_number, is_active
		FROM school_whatsapp_bots
		WHERE display_phone_number = $1
		LIMIT 1;
	`
	var bot SchoolWhatsAppBot
	err := s.q().QueryRowContext(ctx, query, botPhone).Scan(
		&bot.ID, &bot.SchoolID, &bot.PhoneNumberID, &bot.DisplayPhoneNumber, &bot.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("school whatsapp bot not found")
	}
	return &bot, err
}

func (s *SQLStore) GetSchoolByBotPhoneID(ctx context.Context, phoneID string) (*SchoolWhatsAppBot, error) {
	query := `
		SELECT id, school_id, phone_number_id, display_phone_number, is_active
		FROM school_whatsapp_bots
		WHERE phone_number_id = $1
		LIMIT 1;
	`
	var bot SchoolWhatsAppBot
	err := s.q().QueryRowContext(ctx, query, phoneID).Scan(
		&bot.ID, &bot.SchoolID, &bot.PhoneNumberID, &bot.DisplayPhoneNumber, &bot.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("school whatsapp bot not found")
	}
	return &bot, err
}

func (s *SQLStore) RegisterSchoolBot(ctx context.Context, bot *SchoolWhatsAppBot) error {
	if bot.ID == "" {
		bot.ID = newUUID()
	}
	query := `
		INSERT INTO school_whatsapp_bots (id, school_id, phone_number_id, display_phone_number, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (phone_number_id) DO UPDATE SET
			display_phone_number = EXCLUDED.display_phone_number,
			is_active = EXCLUDED.is_active;
	`
	_, err := s.q().ExecContext(ctx, query, bot.ID, bot.SchoolID, bot.PhoneNumberID, bot.DisplayPhoneNumber, bot.IsActive)
	return err
}

func (s *SQLStore) GetActiveSession(ctx context.Context, schoolID, senderWANumber string) (*ChatSession, error) {
	query := `
		SELECT cs.id, cs.school_id, cs.channel, cs.sender_wa_number,
		       COALESCE(cs.student_id::text, ''), COALESCE(s.admission_number, ''),
		       COALESCE(cs.guardian_user_id::text, ''), cs.state,
		       cs.verified_at, cs.expires_at, cs.lookup_count, cs.payment_init_count,
		       cs.created_at, cs.updated_at
		FROM chat_sessions cs
		LEFT JOIN students s ON s.id = cs.student_id
		WHERE cs.school_id = $1 AND cs.sender_wa_number = $2 AND (cs.expires_at IS NULL OR cs.expires_at > NOW())
		ORDER BY cs.created_at DESC
		LIMIT 1;
	`
	return s.scanSession(s.q().QueryRowContext(ctx, query, schoolID, senderWANumber))
}

func (s *SQLStore) GetSessionByID(ctx context.Context, sessionID string) (*ChatSession, error) {
	query := `
		SELECT cs.id, cs.school_id, cs.channel, cs.sender_wa_number,
		       COALESCE(cs.student_id::text, ''), COALESCE(s.admission_number, ''),
		       COALESCE(cs.guardian_user_id::text, ''), cs.state,
		       cs.verified_at, cs.expires_at, cs.lookup_count, cs.payment_init_count,
		       cs.created_at, cs.updated_at
		FROM chat_sessions cs
		LEFT JOIN students s ON s.id = cs.student_id
		WHERE cs.id = $1
		LIMIT 1;
	`
	return s.scanSession(s.q().QueryRowContext(ctx, query, sessionID))
}

func (s *SQLStore) scanSession(row *sql.Row) (*ChatSession, error) {
	var cs ChatSession
	var studentID, admissionNum, guardianUserID string
	var verifiedAt, expiresAt sql.NullTime

	err := row.Scan(
		&cs.ID, &cs.SchoolID, &cs.Channel, &cs.SenderWANumber,
		&studentID, &admissionNum, &guardianUserID, &cs.State,
		&verifiedAt, &expiresAt, &cs.LookupCount, &cs.PaymentInitCount,
		&cs.CreatedAt, &cs.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("chat session not found")
	}
	if err != nil {
		return nil, err
	}
	cs.StudentID = studentID
	cs.AdmissionNumber = admissionNum
	cs.GuardianUserID = guardianUserID
	if verifiedAt.Valid {
		val := verifiedAt.Time
		cs.VerifiedAt = &val
	}
	if expiresAt.Valid {
		val := expiresAt.Time
		cs.ExpiresAt = &val
	}
	return &cs, nil
}

func (s *SQLStore) SaveSession(ctx context.Context, session *ChatSession) error {
	if session.ID == "" {
		session.ID = newUUID()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	session.UpdatedAt = time.Now()

	query := `
		INSERT INTO chat_sessions (
			id, school_id, channel, sender_wa_number, student_id, guardian_user_id,
			state, verified_at, expires_at, lookup_count, payment_init_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid,
			$7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			student_id = EXCLUDED.student_id,
			guardian_user_id = EXCLUDED.guardian_user_id,
			state = EXCLUDED.state,
			verified_at = EXCLUDED.verified_at,
			expires_at = EXCLUDED.expires_at,
			lookup_count = EXCLUDED.lookup_count,
			payment_init_count = EXCLUDED.payment_init_count,
			updated_at = EXCLUDED.updated_at;
	`
	_, err := s.q().ExecContext(ctx, query,
		session.ID, session.SchoolID, session.Channel, session.SenderWANumber,
		session.StudentID, session.GuardianUserID, session.State,
		session.VerifiedAt, session.ExpiresAt, session.LookupCount, session.PaymentInitCount,
		session.CreatedAt, session.UpdatedAt,
	)
	return err
}

func (s *SQLStore) GetStudentByAdmission(ctx context.Context, schoolID, admissionNumber string) (studentID, fullName string, err error) {
	query := `
		SELECT id, full_name
		FROM students
		WHERE school_id = $1 AND UPPER(TRIM(admission_number)) = UPPER(TRIM($2))
		LIMIT 1;
	`
	err = s.q().QueryRowContext(ctx, query, schoolID, admissionNumber).Scan(&studentID, &fullName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", errors.New("student not found")
	}
	return studentID, fullName, err
}

func (s *SQLStore) ListGuardiansForStudent(ctx context.Context, schoolID, studentID string) ([]GuardianCandidate, error) {
	query := `
		SELECT sg.user_id, sg.relationship, u.full_name, COALESCE(u.phone_number, '')
		FROM student_guardians sg
		JOIN users u ON u.id = sg.user_id
		WHERE sg.school_id = $1 AND sg.student_id = $2
		ORDER BY sg.is_primary DESC, sg.created_at ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GuardianCandidate
	for rows.Next() {
		var g GuardianCandidate
		if err := rows.Scan(&g.UserID, &g.Relationship, &g.FullName, &g.PhoneNumber); err != nil {
			return nil, err
		}
		g.MaskedPhone = maskPhoneNumber(g.PhoneNumber)
		result = append(result, g)
	}
	return result, rows.Err()
}

func (s *SQLStore) GetGuardianUser(ctx context.Context, userID string) (phone string, name string, err error) {
	query := `
		SELECT COALESCE(phone_number, ''), full_name
		FROM users
		WHERE id = $1
		LIMIT 1;
	`
	err = s.q().QueryRowContext(ctx, query, userID).Scan(&phone, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", errors.New("guardian user not found")
	}
	return phone, name, err
}

func (s *SQLStore) IsSenderBlocked(ctx context.Context, senderWANumber string) (bool, time.Time, error) {
	query := `
		SELECT blocked_until
		FROM sender_rate_blocks
		WHERE sender_wa_number = $1 AND blocked_until > NOW()
		ORDER BY blocked_until DESC
		LIMIT 1;
	`
	var until time.Time
	err := s.q().QueryRowContext(ctx, query, senderWANumber).Scan(&until)
	if errors.Is(err, sql.ErrNoRows) {
		return false, time.Time{}, nil
	}
	if err != nil {
		return false, time.Time{}, err
	}
	return true, until, nil
}

func (s *SQLStore) BlockSender(ctx context.Context, schoolID, senderWANumber, reason string, until time.Time) error {
	query := `
		INSERT INTO sender_rate_blocks (id, school_id, sender_wa_number, blocked_until, reason, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW());
	`
	_, err := s.q().ExecContext(ctx, query, schoolID, senderWANumber, until, reason)
	return err
}

func (s *SQLStore) GetRecentOTPRequestCount(ctx context.Context, senderWANumber string, window time.Duration) (int, error) {
	cutoff := time.Now().Add(-window)
	query := `
		SELECT COUNT(*)
		FROM chat_otp_requests
		WHERE sender_wa_number = $1 AND created_at > $2;
	`
	var count int
	err := s.q().QueryRowContext(ctx, query, senderWANumber, cutoff).Scan(&count)
	return count, err
}

func (s *SQLStore) HasRecentOTPForStudent(ctx context.Context, schoolID, admissionNumber string, window time.Duration) (bool, error) {
	cutoff := time.Now().Add(-window)
	query := `
		SELECT EXISTS (
			SELECT 1 FROM chat_otp_requests
			WHERE school_id = $1 AND UPPER(TRIM(admission_number)) = UPPER(TRIM($2)) AND created_at > $3
		);
	`
	var exists bool
	err := s.q().QueryRowContext(ctx, query, schoolID, admissionNumber, cutoff).Scan(&exists)
	return exists, err
}

func (s *SQLStore) HasRecentOTPForSender(ctx context.Context, senderWANumber string, window time.Duration) (bool, error) {
	cutoff := time.Now().Add(-window)
	query := `
		SELECT EXISTS (
			SELECT 1 FROM chat_otp_requests
			WHERE sender_wa_number = $1 AND created_at > $2
		);
	`
	var exists bool
	err := s.q().QueryRowContext(ctx, query, senderWANumber, cutoff).Scan(&exists)
	return exists, err
}

func (s *SQLStore) SaveOTP(ctx context.Context, sessionID, codeHash string, expiresAt time.Time) error {
	sess, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found for OTP save: %w", err)
	}

	query := `
		INSERT INTO chat_otp_requests (
			id, school_id, session_id, sender_wa_number, admission_number,
			guardian_user_id, otp_code, expires_at, attempts, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			NULLIF($5, '')::uuid, $6, $7, 0, NOW()
		);
	`
	_, err = s.q().ExecContext(ctx, query,
		sess.SchoolID, sessionID, sess.SenderWANumber, sess.AdmissionNumber,
		sess.GuardianUserID, codeHash, expiresAt,
	)
	return err
}

func (s *SQLStore) VerifyOTP(ctx context.Context, sessionID, code string) (bool, error) {
	query := `
		SELECT id, otp_code, expires_at, attempts
		FROM chat_otp_requests
		WHERE session_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1;
	`
	var id, storedHash string
	var expiresAt time.Time
	var attempts int

	err := s.q().QueryRowContext(ctx, query, sessionID).Scan(&id, &storedHash, &expiresAt, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return false, errors.New("expired or invalid OTP")
	}
	if err != nil {
		return false, err
	}

	attempts++
	if attempts > 3 {
		// Consume OTP on max attempts to prevent brute force
		_, _ = s.q().ExecContext(ctx, `DELETE FROM chat_otp_requests WHERE id = $1`, id)
		return false, errors.New("maximum verification attempts exceeded; OTP invalidated")
	}

	// Update attempt count
	_, _ = s.q().ExecContext(ctx, `UPDATE chat_otp_requests SET attempts = $1 WHERE id = $2`, attempts, id)

	hashed := hashOTP(code)
	if hashed != storedHash && code != storedHash {
		return false, nil
	}

	// Successfully verified: consume OTP so it cannot be replayed
	_, _ = s.q().ExecContext(ctx, `DELETE FROM chat_otp_requests WHERE id = $1`, id)
	return true, nil
}
