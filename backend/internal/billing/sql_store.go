package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// SQLStore is the PostgreSQL implementation of billing.Store.
// It enforces Row-Level Security (RLS) via current_school_id() and tenant parameters.
type SQLStore struct {
	db *sql.DB
	tx *sql.Tx
}

// NewSQLStore wraps an existing *sql.DB.
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

// WithTx starts a transaction, sets the RLS tenant context, and commits on success.
func (s *SQLStore) WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.db == nil {
		return errors.New("billing: SQLStore has no database connection")
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

// ----------------------------------------------------------------------------
// Fee Periods
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateFeePeriod(ctx context.Context, period *FeePeriod) error {
	if period.ID == "" {
		period.ID = newUUID()
	}
	if period.PeriodType == "" {
		period.PeriodType = "term"
	}
	if period.CreatedAt.IsZero() {
		period.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO fee_periods (id, school_id, academic_year_id, term_id, period_type, name, due_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (school_id, academic_year_id, term_id) DO UPDATE SET
			name = EXCLUDED.name,
			due_date = EXCLUDED.due_date;
	`
	_, err := s.q().ExecContext(ctx, query,
		period.ID, period.SchoolID, period.AcademicYearID, period.TermID,
		period.PeriodType, period.Name, period.DueDate, period.CreatedAt,
	)
	return err
}

func (s *SQLStore) GetFeePeriod(ctx context.Context, schoolID, academicYearID, termID string) (*FeePeriod, error) {
	query := `
		SELECT id, school_id, academic_year_id, term_id, period_type, name, due_date, created_at
		FROM fee_periods
		WHERE school_id = $1 AND academic_year_id = $2 AND term_id = $3
		LIMIT 1;
	`
	var p FeePeriod
	err := s.q().QueryRowContext(ctx, query, schoolID, academicYearID, termID).Scan(
		&p.ID, &p.SchoolID, &p.AcademicYearID, &p.TermID, &p.PeriodType, &p.Name, &p.DueDate, &p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (s *SQLStore) UpdateFeePeriodDueDate(ctx context.Context, schoolID, termID string, newDueDate time.Time) error {
	query := `
		UPDATE fee_periods
		SET due_date = $1
		WHERE school_id = $2 AND term_id = $3;
	`
	res, err := s.q().ExecContext(ctx, query, newDueDate, schoolID, termID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

// ----------------------------------------------------------------------------
// Fee Structures
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateFeeStructure(ctx context.Context, structure *FeeStructure, items []FeeStructureItem) error {
	if structure.ID == "" {
		structure.ID = newUUID()
	}
	if structure.CreatedAt.IsZero() {
		structure.CreatedAt = time.Now()
	}

	insertStruct := `
		INSERT INTO fee_structures (id, school_id, academic_year_id, grade_level_id, term_id, total_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (school_id, academic_year_id, grade_level_id, term_id) DO UPDATE SET
			total_amount = EXCLUDED.total_amount;
	`
	if _, err := s.q().ExecContext(ctx, insertStruct,
		structure.ID, structure.SchoolID, structure.AcademicYearID,
		structure.GradeLevelID, structure.TermID, structure.TotalAmount, structure.CreatedAt,
	); err != nil {
		return err
	}

	// Remove old items on update
	_, _ = s.q().ExecContext(ctx, `DELETE FROM fee_structure_items WHERE fee_structure_id = $1`, structure.ID)

	for _, item := range items {
		itemID := item.ID
		if itemID == "" {
			itemID = newUUID()
		}
		insertItem := `
			INSERT INTO fee_structure_items (id, fee_structure_id, title, amount, is_compulsory)
			VALUES ($1, $2, $3, $4, $5);
		`
		if _, err := s.q().ExecContext(ctx, insertItem, itemID, structure.ID, item.Title, item.Amount, item.IsCompulsory); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) GetFeeStructure(ctx context.Context, schoolID, academicYearID, gradeLevelID, termID string) (*FeeStructure, error) {
	query := `
		SELECT id, school_id, academic_year_id, grade_level_id, term_id, total_amount, created_at
		FROM fee_structures
		WHERE school_id = $1 AND academic_year_id = $2 AND grade_level_id = $3 AND term_id = $4
		LIMIT 1;
	`
	var fs FeeStructure
	err := s.q().QueryRowContext(ctx, query, schoolID, academicYearID, gradeLevelID, termID).Scan(
		&fs.ID, &fs.SchoolID, &fs.AcademicYearID, &fs.GradeLevelID, &fs.TermID, &fs.TotalAmount, &fs.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &fs, err
}

func (s *SQLStore) ListFeeStructureItems(ctx context.Context, structureID string) ([]FeeStructureItem, error) {
	query := `
		SELECT id, fee_structure_id, title, amount, is_compulsory
		FROM fee_structure_items
		WHERE fee_structure_id = $1
		ORDER BY id ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, structureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeeStructureItem
	for rows.Next() {
		var it FeeStructureItem
		if err := rows.Scan(&it.ID, &it.FeeStructureID, &it.Title, &it.Amount, &it.IsCompulsory); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// ----------------------------------------------------------------------------
// Students & Enrollments
// ----------------------------------------------------------------------------

func (s *SQLStore) GetEnrolledStudentsForGrade(ctx context.Context, schoolID, gradeLevelID, academicYearID string) ([]string, error) {
	query := `
		SELECT student_id
		FROM enrolments
		WHERE school_id = $1 AND grade_level_id = $2 AND academic_year_id = $3 AND status = 'active';
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, gradeLevelID, academicYearID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id)
	}
	return studentIDs, rows.Err()
}

func (s *SQLStore) GetStudent(ctx context.Context, schoolID, studentID string) (id, admissionNumber, fullName string, err error) {
	query := `
		SELECT id, admission_number, full_name
		FROM students
		WHERE school_id = $1 AND id = $2
		LIMIT 1;
	`
	err = s.q().QueryRowContext(ctx, query, schoolID, studentID).Scan(&id, &admissionNumber, &fullName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", ErrNotFound
	}
	return id, admissionNumber, fullName, err
}

func (s *SQLStore) GetStudentByAdmissionNumber(ctx context.Context, schoolID, admissionNumber string) (id, fullName string, err error) {
	query := `
		SELECT id, full_name
		FROM students
		WHERE school_id = $1 AND UPPER(TRIM(admission_number)) = UPPER(TRIM($2))
		LIMIT 1;
	`
	err = s.q().QueryRowContext(ctx, query, schoolID, admissionNumber).Scan(&id, &fullName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return id, fullName, err
}

func (s *SQLStore) ListGuardiansForStudent(ctx context.Context, schoolID, studentID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM student_guardians
		WHERE school_id = $1 AND student_id = $2
		ORDER BY is_primary DESC, created_at ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guardianIDs []string
	for rows.Next() {
		var gID string
		if err := rows.Scan(&gID); err != nil {
			return nil, err
		}
		guardianIDs = append(guardianIDs, gID)
	}
	return guardianIDs, rows.Err()
}

func (s *SQLStore) IsGuardianOf(ctx context.Context, schoolID, studentID, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM student_guardians
			WHERE school_id = $1 AND student_id = $2 AND user_id = $3
		);
	`
	var exists bool
	err := s.q().QueryRowContext(ctx, query, schoolID, studentID, userID).Scan(&exists)
	return exists, err
}

// ----------------------------------------------------------------------------
// Fee Charges
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateFeeCharge(ctx context.Context, charge *FeeCharge) error {
	if charge.ID == "" {
		charge.ID = newUUID()
	}
	if charge.CreatedAt.IsZero() {
		charge.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO fee_charges (id, school_id, student_id, fee_period_id, fee_structure_id, amount, due_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (school_id, student_id, fee_period_id) DO UPDATE SET
			amount = EXCLUDED.amount,
			due_date = EXCLUDED.due_date;
	`
	_, err := s.q().ExecContext(ctx, query,
		charge.ID, charge.SchoolID, charge.StudentID, charge.FeePeriodID,
		charge.FeeStructureID, charge.Amount, charge.DueDate, charge.CreatedAt,
	)
	return err
}

func (s *SQLStore) ListFeeChargesForStudent(ctx context.Context, schoolID, studentID string) ([]FeeCharge, error) {
	query := `
		SELECT id, school_id, student_id, fee_period_id, fee_structure_id, amount, due_date, created_at
		FROM fee_charges
		WHERE school_id = $1 AND student_id = $2
		ORDER BY due_date ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charges []FeeCharge
	for rows.Next() {
		var fc FeeCharge
		if err := rows.Scan(
			&fc.ID, &fc.SchoolID, &fc.StudentID, &fc.FeePeriodID,
			&fc.FeeStructureID, &fc.Amount, &fc.DueDate, &fc.CreatedAt,
		); err != nil {
			return nil, err
		}
		charges = append(charges, fc)
	}
	return charges, rows.Err()
}

func (s *SQLStore) UpdateFeeChargesDueDate(ctx context.Context, schoolID, termID string, newDueDate time.Time) (int, error) {
	query := `
		UPDATE fee_charges fc
		SET due_date = $1
		FROM fee_periods fp
		WHERE fc.fee_period_id = fp.id
		  AND fc.school_id = $2
		  AND fp.term_id = $3;
	`
	res, err := s.q().ExecContext(ctx, query, newDueDate, schoolID, termID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

func (s *SQLStore) AdjustFeeChargesForGrade(ctx context.Context, schoolID, termID, gradeLevelID string, delta float64) ([]string, error) {
	query := `
		UPDATE fee_charges fc
		SET amount = amount + $1
		FROM fee_structures fs, fee_periods fp
		WHERE fc.fee_structure_id = fs.id
		  AND fc.fee_period_id = fp.id
		  AND fc.school_id = $2
		  AND fp.term_id = $3
		  AND fs.grade_level_id = $4
		RETURNING fc.student_id;
	`
	rows, err := s.q().QueryContext(ctx, query, delta, schoolID, termID, gradeLevelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var affectedStudentIDs []string
	seen := make(map[string]bool)
	for rows.Next() {
		var sID string
		if err := rows.Scan(&sID); err != nil {
			return nil, err
		}
		if !seen[sID] {
			seen[sID] = true
			affectedStudentIDs = append(affectedStudentIDs, sID)
		}
	}
	return affectedStudentIDs, rows.Err()
}

func (s *SQLStore) GetStudentsWithChargesForTerm(ctx context.Context, schoolID, termID string) ([]string, error) {
	query := `
		SELECT DISTINCT fc.student_id
		FROM fee_charges fc
		JOIN fee_periods fp ON fc.fee_period_id = fp.id AND fp.school_id = fc.school_id
		WHERE fc.school_id = $1 AND fp.term_id = $2;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var sID string
		if err := rows.Scan(&sID); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, sID)
	}
	return studentIDs, rows.Err()
}

// ----------------------------------------------------------------------------
// Ledger Entries & Balances
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateLedgerEntry(ctx context.Context, entry *LedgerEntry) error {
	if entry.ID == "" {
		entry.ID = newUUID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO ledger_entries (id, school_id, student_id, term_id, entry_type, amount, reference, description, created_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9);
	`
	_, err := s.q().ExecContext(ctx, query,
		entry.ID, entry.SchoolID, entry.StudentID, entry.TermID,
		entry.EntryType, entry.Amount, entry.Reference, entry.Description, entry.CreatedAt,
	)
	return err
}

func (s *SQLStore) ListLedgerEntriesForStudent(ctx context.Context, schoolID, studentID string) ([]LedgerEntry, error) {
	query := `
		SELECT id, school_id, student_id, COALESCE(term_id::text, ''), entry_type, amount,
		       COALESCE(reference, ''), COALESCE(description, ''), created_at
		FROM ledger_entries
		WHERE school_id = $1 AND student_id = $2
		ORDER BY created_at ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var le LedgerEntry
		if err := rows.Scan(
			&le.ID, &le.SchoolID, &le.StudentID, &le.TermID,
			&le.EntryType, &le.Amount, &le.Reference, &le.Description, &le.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, le)
	}
	return entries, rows.Err()
}

func (s *SQLStore) GetFeeLedger(ctx context.Context, schoolID, studentID, academicYearID, termID string) (*FeeLedger, error) {
	query := `
		SELECT id, school_id, student_id, academic_year_id, term_id,
		       opening_balance, billed_amount, paid_amount, current_balance,
		       COALESCE(credit_balance, 0.00), COALESCE(due_date, NOW()), updated_at
		FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2 AND academic_year_id = $3 AND term_id = $4
		LIMIT 1;
	`
	var fl FeeLedger
	err := s.q().QueryRowContext(ctx, query, schoolID, studentID, academicYearID, termID).Scan(
		&fl.ID, &fl.SchoolID, &fl.StudentID, &fl.AcademicYearID, &fl.TermID,
		&fl.OpeningBalance, &fl.BilledAmount, &fl.PaidAmount, &fl.CurrentBalance,
		&fl.CreditBalance, &fl.DueDate, &fl.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &fl, err
}

func (s *SQLStore) GetLatestFeeLedger(ctx context.Context, schoolID, studentID string) (*FeeLedger, error) {
	query := `
		SELECT id, school_id, student_id, academic_year_id, term_id,
		       opening_balance, billed_amount, paid_amount, current_balance,
		       COALESCE(credit_balance, 0.00), COALESCE(due_date, NOW()), updated_at
		FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2
		ORDER BY updated_at DESC
		LIMIT 1;
	`
	var fl FeeLedger
	err := s.q().QueryRowContext(ctx, query, schoolID, studentID).Scan(
		&fl.ID, &fl.SchoolID, &fl.StudentID, &fl.AcademicYearID, &fl.TermID,
		&fl.OpeningBalance, &fl.BilledAmount, &fl.PaidAmount, &fl.CurrentBalance,
		&fl.CreditBalance, &fl.DueDate, &fl.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &fl, err
}

func (s *SQLStore) SaveFeeLedger(ctx context.Context, ledger *FeeLedger) error {
	if ledger.ID == "" {
		ledger.ID = newUUID()
	}
	ledger.UpdatedAt = time.Now()
	query := `
		INSERT INTO fee_ledgers (
			id, school_id, student_id, academic_year_id, term_id,
			opening_balance, billed_amount, paid_amount, current_balance,
			credit_balance, due_date, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO UPDATE SET
			opening_balance = EXCLUDED.opening_balance,
			billed_amount = EXCLUDED.billed_amount,
			paid_amount = EXCLUDED.paid_amount,
			current_balance = EXCLUDED.current_balance,
			credit_balance = EXCLUDED.credit_balance,
			due_date = EXCLUDED.due_date,
			updated_at = EXCLUDED.updated_at;
	`
	_, err := s.q().ExecContext(ctx, query,
		ledger.ID, ledger.SchoolID, ledger.StudentID, ledger.AcademicYearID, ledger.TermID,
		ledger.OpeningBalance, ledger.BilledAmount, ledger.PaidAmount, ledger.CurrentBalance,
		ledger.CreditBalance, ledger.DueDate, ledger.UpdatedAt,
	)
	return err
}

// ----------------------------------------------------------------------------
// Fee Change Notes
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateFeeChangeNote(ctx context.Context, note *FeeChangeNote) error {
	if note.ID == "" {
		note.ID = newUUID()
	}
	if note.CreatedAt.IsZero() {
		note.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO fee_change_notes (
			id, school_id, term_id, fee_period_id, change_type, amount_delta,
			old_due_date, new_due_date, note, created_by, affected_students_count, created_at
		) VALUES (
			$1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6,
			$7, $8, $9, $10, $11, $12
		);
	`
	_, err := s.q().ExecContext(ctx, query,
		note.ID, note.SchoolID, note.TermID, note.FeePeriodID, note.ChangeType, note.AmountDelta,
		note.OldDueDate, note.NewDueDate, note.Note, note.CreatedBy, note.AffectedStudentsCount, note.CreatedAt,
	)
	return err
}

func (s *SQLStore) ListFeeChangeNotes(ctx context.Context, schoolID, termID string) ([]FeeChangeNote, error) {
	query := `
		SELECT id, school_id, COALESCE(term_id::text, ''), COALESCE(fee_period_id::text, ''),
		       change_type, amount_delta, old_due_date, new_due_date, note, created_by,
		       affected_students_count, created_at
		FROM fee_change_notes
		WHERE school_id = $1 AND (term_id = NULLIF($2, '')::uuid OR $2 = '')
		ORDER BY created_at DESC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []FeeChangeNote
	for rows.Next() {
		var n FeeChangeNote
		if err := rows.Scan(
			&n.ID, &n.SchoolID, &n.TermID, &n.FeePeriodID,
			&n.ChangeType, &n.AmountDelta, &n.OldDueDate, &n.NewDueDate, &n.Note, &n.CreatedBy,
			&n.AffectedStudentsCount, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// ----------------------------------------------------------------------------
// M-Pesa Configs
// ----------------------------------------------------------------------------

func (s *SQLStore) GetSchoolMpesaConfig(ctx context.Context, schoolID string) (*SchoolMpesaConfig, error) {
	query := `
		SELECT school_id, shortcode, environment,
		       COALESCE(passkey_encrypted, ''), COALESCE(consumer_key_encrypted, ''), COALESCE(consumer_secret_encrypted, ''),
		       test_payment_verified, updated_at
		FROM school_mpesa_configs
		WHERE school_id = $1
		LIMIT 1;
	`
	var cfg SchoolMpesaConfig
	err := s.q().QueryRowContext(ctx, query, schoolID).Scan(
		&cfg.SchoolID, &cfg.Shortcode, &cfg.Environment,
		&cfg.PasskeyEncrypted, &cfg.ConsumerKeyEncrypted, &cfg.ConsumerSecretEncrypted,
		&cfg.TestPaymentVerified, &cfg.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &cfg, err
}

func (s *SQLStore) GetSchoolMpesaConfigByShortcode(ctx context.Context, shortcode string) (*SchoolMpesaConfig, error) {
	query := `
		SELECT school_id, shortcode, environment,
		       COALESCE(passkey_encrypted, ''), COALESCE(consumer_key_encrypted, ''), COALESCE(consumer_secret_encrypted, ''),
		       test_payment_verified, updated_at
		FROM school_mpesa_configs
		WHERE shortcode = $1
		LIMIT 1;
	`
	var cfg SchoolMpesaConfig
	err := s.q().QueryRowContext(ctx, query, shortcode).Scan(
		&cfg.SchoolID, &cfg.Shortcode, &cfg.Environment,
		&cfg.PasskeyEncrypted, &cfg.ConsumerKeyEncrypted, &cfg.ConsumerSecretEncrypted,
		&cfg.TestPaymentVerified, &cfg.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &cfg, err
}

func (s *SQLStore) SaveSchoolMpesaConfig(ctx context.Context, config *SchoolMpesaConfig) error {
	config.UpdatedAt = time.Now()
	query := `
		INSERT INTO school_mpesa_configs (
			school_id, shortcode, environment,
			passkey_encrypted, consumer_key_encrypted, consumer_secret_encrypted,
			test_payment_verified, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (school_id) DO UPDATE SET
			shortcode = EXCLUDED.shortcode,
			environment = EXCLUDED.environment,
			passkey_encrypted = EXCLUDED.passkey_encrypted,
			consumer_key_encrypted = EXCLUDED.consumer_key_encrypted,
			consumer_secret_encrypted = EXCLUDED.consumer_secret_encrypted,
			test_payment_verified = EXCLUDED.test_payment_verified,
			updated_at = EXCLUDED.updated_at;
	`
	_, err := s.q().ExecContext(ctx, query,
		config.SchoolID, config.Shortcode, config.Environment,
		config.PasskeyEncrypted, config.ConsumerKeyEncrypted, config.ConsumerSecretEncrypted,
		config.TestPaymentVerified, config.UpdatedAt,
	)
	return err
}

// ----------------------------------------------------------------------------
// Payments
// ----------------------------------------------------------------------------

func (s *SQLStore) CreatePayment(ctx context.Context, payment *Payment) error {
	if payment.ID == "" {
		payment.ID = newUUID()
	}
	if payment.PaymentDate.IsZero() {
		payment.PaymentDate = time.Now()
	}

	rawPayload := payment.RawPayload
	if rawPayload == "" || !json.Valid([]byte(rawPayload)) {
		encoded, err := json.Marshal(map[string]string{"raw": rawPayload})
		if err == nil {
			rawPayload = string(encoded)
		} else {
			rawPayload = "{}"
		}
	}

	query := `
		INSERT INTO payments (
			id, school_id, student_id, transaction_ref, provider, amount,
			payer_phone, payer_name, matched_status, matched_by, payment_date, raw_payload
		) VALUES (
			$1, $2, NULLIF($3, '')::uuid, $4, $5, $6,
			$7, $8, $9, NULLIF($10, '')::uuid, $11, $12::jsonb
		)
		ON CONFLICT (transaction_ref) DO UPDATE SET
			matched_status = EXCLUDED.matched_status,
			student_id = EXCLUDED.student_id,
			matched_by = EXCLUDED.matched_by;
	`
	studentID := ""
	if payment.StudentID != nil {
		studentID = *payment.StudentID
	}
	matchedBy := ""
	if payment.MatchedBy != nil {
		matchedBy = *payment.MatchedBy
	}

	_, err := s.q().ExecContext(ctx, query,
		payment.ID, payment.SchoolID, studentID, payment.TransactionRef, payment.Provider, payment.Amount,
		payment.PayerPhone, payment.PayerName, payment.MatchedStatus, matchedBy, payment.PaymentDate, rawPayload,
	)
	return err
}

func (s *SQLStore) GetPayment(ctx context.Context, schoolID, paymentID string) (*Payment, error) {
	query := `
		SELECT id, school_id, student_id, transaction_ref, provider, amount,
		       COALESCE(payer_phone, ''), COALESCE(payer_name, ''), matched_status, matched_by,
		       payment_date, COALESCE(raw_payload::text, '')
		FROM payments
		WHERE school_id = $1 AND id = $2
		LIMIT 1;
	`
	return s.scanPaymentRow(s.q().QueryRowContext(ctx, query, schoolID, paymentID))
}

func (s *SQLStore) GetPaymentByCheckoutID(ctx context.Context, checkoutID string) (*Payment, error) {
	query := `
		SELECT id, school_id, student_id, transaction_ref, provider, amount,
		       COALESCE(payer_phone, ''), COALESCE(payer_name, ''), matched_status, matched_by,
		       payment_date, COALESCE(raw_payload::text, '')
		FROM payments
		WHERE transaction_ref = $1 OR raw_payload::text LIKE '%' || $1 || '%'
		LIMIT 1;
	`
	return s.scanPaymentRow(s.q().QueryRowContext(ctx, query, checkoutID))
}

func (s *SQLStore) GetPaymentByTransactionRef(ctx context.Context, transactionRef string) (*Payment, error) {
	query := `
		SELECT id, school_id, student_id, transaction_ref, provider, amount,
		       COALESCE(payer_phone, ''), COALESCE(payer_name, ''), matched_status, matched_by,
		       payment_date, COALESCE(raw_payload::text, '')
		FROM payments
		WHERE transaction_ref = $1
		LIMIT 1;
	`
	return s.scanPaymentRow(s.q().QueryRowContext(ctx, query, transactionRef))
}

func (s *SQLStore) scanPaymentRow(row *sql.Row) (*Payment, error) {
	var p Payment
	var studentID sql.NullString
	var matchedBy sql.NullString

	err := row.Scan(
		&p.ID, &p.SchoolID, &studentID, &p.TransactionRef, &p.Provider, &p.Amount,
		&p.PayerPhone, &p.PayerName, &p.MatchedStatus, &matchedBy,
		&p.PaymentDate, &p.RawPayload,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if studentID.Valid {
		val := studentID.String
		p.StudentID = &val
	}
	if matchedBy.Valid {
		val := matchedBy.String
		p.MatchedBy = &val
	}
	return &p, nil
}

func (s *SQLStore) UpdatePaymentStatus(ctx context.Context, schoolID, paymentID, status string) error {
	query := `
		UPDATE payments
		SET matched_status = $1
		WHERE school_id = $2 AND id = $3;
	`
	res, err := s.q().ExecContext(ctx, query, status, schoolID, paymentID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

func (s *SQLStore) UpdatePaymentMatched(ctx context.Context, schoolID, paymentID, studentID, matchedBy string) error {
	query := `
		UPDATE payments
		SET matched_status = 'matched', student_id = $1, matched_by = $2
		WHERE school_id = $3 AND id = $4;
	`
	res, err := s.q().ExecContext(ctx, query, studentID, matchedBy, schoolID, paymentID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

// ----------------------------------------------------------------------------
// Unmatched Payments
// ----------------------------------------------------------------------------

func (s *SQLStore) CreateUnmatchedPayment(ctx context.Context, unmatched *UnmatchedPayment) error {
	if unmatched.ID == "" {
		unmatched.ID = newUUID()
	}
	if unmatched.Status == "" {
		unmatched.Status = "unmatched"
	}
	query := `
		INSERT INTO unmatched_payments (
			id, school_id, payment_id, transaction_ref, amount,
			payer_phone, payer_name, raw_account_ref, status, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, NOW()
		);
	`
	_, err := s.q().ExecContext(ctx, query,
		unmatched.ID, unmatched.SchoolID, unmatched.PaymentID, unmatched.TransactionRef, unmatched.Amount,
		unmatched.PayerPhone, unmatched.PayerName, unmatched.RawAccountRef, unmatched.Status,
	)
	return err
}

func (s *SQLStore) ListUnmatchedPayments(ctx context.Context, schoolID string) ([]UnmatchedPayment, error) {
	query := `
		SELECT id, school_id, payment_id, transaction_ref, amount,
		       COALESCE(payer_phone, ''), COALESCE(payer_name, ''), COALESCE(raw_account_ref, ''),
		       status, resolved_student_id, resolved_by, COALESCE(resolution_note, ''), resolved_at, created_at
		FROM unmatched_payments
		WHERE school_id = $1 AND status = 'unmatched'
		ORDER BY created_at DESC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UnmatchedPayment
	for rows.Next() {
		var u UnmatchedPayment
		var resolvedStudentID, resolvedBy sql.NullString
		var resolvedAt sql.NullTime

		if err := rows.Scan(
			&u.ID, &u.SchoolID, &u.PaymentID, &u.TransactionRef, &u.Amount,
			&u.PayerPhone, &u.PayerName, &u.RawAccountRef,
			&u.Status, &resolvedStudentID, &resolvedBy, &u.ResolutionNote, &resolvedAt, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		if resolvedStudentID.Valid {
			val := resolvedStudentID.String
			u.ResolvedStudentID = &val
		}
		if resolvedBy.Valid {
			val := resolvedBy.String
			u.ResolvedBy = &val
		}
		if resolvedAt.Valid {
			val := resolvedAt.Time
			u.ResolvedAt = &val
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (s *SQLStore) GetUnmatchedPayment(ctx context.Context, schoolID, id string) (*UnmatchedPayment, error) {
	query := `
		SELECT id, school_id, payment_id, transaction_ref, amount,
		       COALESCE(payer_phone, ''), COALESCE(payer_name, ''), COALESCE(raw_account_ref, ''),
		       status, resolved_student_id, resolved_by, COALESCE(resolution_note, ''), resolved_at, created_at
		FROM unmatched_payments
		WHERE school_id = $1 AND (id = $2 OR payment_id = $2)
		LIMIT 1;
	`
	var u UnmatchedPayment
	var resolvedStudentID, resolvedBy sql.NullString
	var resolvedAt sql.NullTime

	err := s.q().QueryRowContext(ctx, query, schoolID, id).Scan(
		&u.ID, &u.SchoolID, &u.PaymentID, &u.TransactionRef, &u.Amount,
		&u.PayerPhone, &u.PayerName, &u.RawAccountRef,
		&u.Status, &resolvedStudentID, &resolvedBy, &u.ResolutionNote, &resolvedAt, &u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if resolvedStudentID.Valid {
		val := resolvedStudentID.String
		u.ResolvedStudentID = &val
	}
	if resolvedBy.Valid {
		val := resolvedBy.String
		u.ResolvedBy = &val
	}
	if resolvedAt.Valid {
		val := resolvedAt.Time
		u.ResolvedAt = &val
	}
	return &u, nil
}

func (s *SQLStore) ResolveUnmatchedPayment(ctx context.Context, schoolID, id, studentID, resolvedBy, note string) error {
	query := `
		UPDATE unmatched_payments
		SET status = 'resolved',
		    resolved_student_id = $1,
		    resolved_by = $2,
		    resolution_note = $3,
		    resolved_at = NOW()
		WHERE school_id = $4 AND (id = $5 OR payment_id = $5);
	`
	res, err := s.q().ExecContext(ctx, query, studentID, resolvedBy, note, schoolID, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}

// ----------------------------------------------------------------------------
// User Chat Payment Freeze
// ----------------------------------------------------------------------------

func (s *SQLStore) GetUserChatPaymentPausedUntil(ctx context.Context, userID string) (*time.Time, error) {
	query := `SELECT chat_payment_paused_until FROM users WHERE id = $1 LIMIT 1;`
	var t sql.NullTime
	err := s.q().QueryRowContext(ctx, query, userID).Scan(&t)
	if errors.Is(err, sql.ErrNoRows) || !t.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	val := t.Time
	return &val, nil
}

func (s *SQLStore) SetUserChatPaymentPausedUntil(ctx context.Context, userID string, until time.Time) error {
	query := `UPDATE users SET chat_payment_paused_until = $1 WHERE id = $2;`
	_, err := s.q().ExecContext(ctx, query, until, userID)
	return err
}
