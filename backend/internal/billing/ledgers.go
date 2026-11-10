package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type LedgerEntry struct {
	ID          string    `json:"id"`
	TermID      string    `json:"term_id"`
	EntryType   string    `json:"entry_type"`
	Amount      Money     `json:"amount"`
	Reference   string    `json:"reference"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FeeChangeNoteView struct {
	ID                    string    `json:"id"`
	TermID                *string   `json:"term_id,omitempty"`
	FeePeriodID           *string   `json:"fee_period_id,omitempty"`
	ChangeType            string    `json:"change_type"`
	AmountDelta           Money     `json:"amount_delta"`
	OldDueDate            *string   `json:"old_due_date,omitempty"`
	NewDueDate            *string   `json:"new_due_date,omitempty"`
	Note                  string    `json:"note"`
	AffectedStudentsCount int       `json:"affected_students_count"`
	CreatedAt             time.Time `json:"created_at"`
}

type StudentFeeBalance struct {
	StudentID       string              `json:"student_id"`
	AdmissionNumber string              `json:"admission_number"`
	StudentName     string              `json:"student_name"`
	OpeningBalance  Money               `json:"opening_balance"`
	BilledAmount    Money               `json:"billed_amount"`
	PaidAmount      Money               `json:"paid_amount"`
	CurrentBalance  Money               `json:"current_balance"`
	CreditBalance   Money               `json:"credit_balance"`
	DueDate         *string             `json:"due_date"`
	Items           []FeeStructureItem  `json:"items"`
	FeeChangeNotes  []FeeChangeNoteView `json:"fee_change_notes"`
}

type StudentLedger struct {
	StudentFeeBalance
	Entries []LedgerEntry `json:"entries"`
}

func chargeGradeForTerm(ctx context.Context, tx pgx.Tx, structureID string) error {
	if !contracts.ValidID(structureID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_FEE_STRUCTURE", "A valid fee structure ID is required.")
	}
	var schoolID, yearID, gradeID, termID, boardingStatus string
	var totalAmount Money
	err := tx.QueryRow(ctx, `
		SELECT school_id, academic_year_id, grade_level_id, term_id, boarding_status, total_amount
		FROM fee_structures
		WHERE id = $1`,
		structureID,
	).Scan(&schoolID, &yearID, &gradeID, &termID, &boardingStatus, &totalAmount)
	if errors.Is(err, pgx.ErrNoRows) {
		return contracts.Fail(http.StatusNotFound, "FEE_STRUCTURE_NOT_FOUND", "Fee structure not found.")
	}
	if err != nil {
		return err
	}

	var feePeriodID string
	var dueDate time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, due_date
		FROM fee_periods
		WHERE school_id = $1 AND academic_year_id = $2 AND term_id = $3`,
		schoolID, yearID, termID,
	).Scan(&feePeriodID, &dueDate)
	if errors.Is(err, pgx.ErrNoRows) {
		var termNumber int
		var startDate time.Time
		if err := tx.QueryRow(ctx, `
			SELECT term_number, start_date
			FROM terms
			WHERE school_id = $1 AND academic_year_id = $2 AND id = $3`,
			schoolID, yearID, termID,
		).Scan(&termNumber, &startDate); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `
			INSERT INTO fee_periods (school_id, academic_year_id, term_id, period_type, name, due_date)
			VALUES ($1, $2, $3, 'term', $4, $5)
			ON CONFLICT (school_id, academic_year_id, term_id)
			DO UPDATE SET name = fee_periods.name
			RETURNING id, due_date`,
			schoolID, yearID, termID, fmt.Sprintf("Term %d", termNumber), startDate,
		).Scan(&feePeriodID, &dueDate); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `
		SELECT s.id
		FROM enrolments e
		JOIN students s ON s.school_id = e.school_id AND s.id = e.student_id
		WHERE e.school_id = $1
		  AND e.academic_year_id = $2
		  AND e.grade_level_id = $3
		  AND e.status = 'active'
		  AND s.status = 'active'
		  AND (
			$4 = 'all'
			OR ($4 = 'boarder' AND s.current_dormitory_id IS NOT NULL)
			OR ($4 = 'day' AND s.current_dormitory_id IS NULL)
		  )
		ORDER BY s.id ASC`,
		schoolID, yearID, gradeID, boardingStatus,
	)
	if err != nil {
		return err
	}
	var studentIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return err
		}
		studentIDs = append(studentIDs, sid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, studentID := range studentIDs {
		if err := applyStudentCharge(ctx, tx, schoolID, studentID, yearID, termID, feePeriodID, structureID, totalAmount, dueDate); err != nil {
			return err
		}
	}
	return nil
}

func chargeStudent(ctx context.Context, tx pgx.Tx, studentID, feePeriodID string) error {
	if !contracts.ValidID(studentID) || !contracts.ValidID(feePeriodID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_CHARGE_TARGET", "Valid student_id and fee_period_id are required.")
	}
	var schoolID, yearID, termID string
	var dueDate time.Time
	err := tx.QueryRow(ctx, `
		SELECT school_id, academic_year_id, term_id, due_date
		FROM fee_periods
		WHERE id = $1`,
		feePeriodID,
	).Scan(&schoolID, &yearID, &termID, &dueDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return contracts.Fail(http.StatusNotFound, "FEE_PERIOD_NOT_FOUND", "Fee period not found.")
	}
	if err != nil {
		return err
	}

	var gradeID string
	var hasDorm bool
	err = tx.QueryRow(ctx, `
		SELECT e.grade_level_id, (s.current_dormitory_id IS NOT NULL)
		FROM enrolments e
		JOIN students s ON s.school_id = e.school_id AND s.id = e.student_id
		WHERE e.school_id = $1 AND e.student_id = $2 AND e.academic_year_id = $3
		  AND e.status = 'active' AND s.status = 'active'`,
		schoolID, studentID, yearID,
	).Scan(&gradeID, &hasDorm)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	preferredBoarding := "day"
	if hasDorm {
		preferredBoarding = "boarder"
	}
	var structureID string
	var totalAmount Money
	err = tx.QueryRow(ctx, `
		SELECT id, total_amount
		FROM fee_structures
		WHERE school_id = $1 AND academic_year_id = $2 AND grade_level_id = $3 AND term_id = $4
		  AND boarding_status IN ($5, 'all')
		ORDER BY CASE WHEN boarding_status = $5 THEN 0 ELSE 1 END
		LIMIT 1`,
		schoolID, yearID, gradeID, termID, preferredBoarding,
	).Scan(&structureID, &totalAmount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return applyStudentCharge(ctx, tx, schoolID, studentID, yearID, termID, feePeriodID, structureID, totalAmount, dueDate)
}

func chargeAdmittedStudent(ctx context.Context, tx pgx.Tx, schoolID, studentID, academicYearID string) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(studentID) || !contracts.ValidID(academicYearID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_ADMISSION_CHARGE", "Valid school_id, student_id and academic_year_id are required.")
	}
	rows, err := tx.Query(ctx, `
		SELECT fp.id
		FROM fee_periods fp
		JOIN terms t ON t.school_id = fp.school_id AND t.id = fp.term_id
		WHERE fp.school_id = $1 AND fp.academic_year_id = $2
		ORDER BY t.term_number ASC, fp.created_at ASC, fp.id ASC`,
		schoolID, academicYearID,
	)
	if err != nil {
		return err
	}
	var periodIDs []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			rows.Close()
			return err
		}
		periodIDs = append(periodIDs, pid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, pid := range periodIDs {
		if err := chargeStudent(ctx, tx, studentID, pid); err != nil {
			return err
		}
	}
	return nil
}

// applyStudentCharge executes the 4-step per-student charging sequence and
// applies the derived credit model.
//
// Financial Credit Model:
//
//	current_balance = max(0, opening_balance + billed_amount - paid_amount)
//	credit_balance  = max(0, paid_amount - opening_balance - billed_amount)
//
// Credit is DERIVED from accumulated financial totals. When a new charge is
// applied while prior credit_before > 0 exists, an INFORMATIONAL ledger entry
// (entry_type = 'credit_applied', amount = min(credit_before, new_charge)) is
// recorded to preserve the history of credit carry-forward. It does NOT increase
// paid_amount or independently subtract from current_balance, preventing double
// credit application.
func applyStudentCharge(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, studentID, yearID, termID, feePeriodID, structureID string,
	chargeAmount Money,
	dueDate time.Time,
) error {
	// Serialize per-student financial mutations across terms.
	var lockedID string
	if err := tx.QueryRow(ctx, `SELECT id FROM students WHERE school_id = $1 AND id = $2 FOR UPDATE`, schoolID, studentID).Scan(&lockedID); err != nil {
		return err
	}

	dueDateStr := dueDate.Format("2006-01-02")

	// Step 1: Create the fee ledger if it does not exist.
	if _, err := tx.Exec(ctx, `
		INSERT INTO fee_ledgers (school_id, student_id, academic_year_id, term_id, due_date)
		VALUES ($1, $2, $3, $4, $5::date)
		ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO NOTHING`,
		schoolID, studentID, yearID, termID, dueDateStr,
	); err != nil {
		return err
	}

	// Step 2: Lock the student's ledgers.
	if _, err := tx.Exec(ctx, `
		SELECT id FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2
		ORDER BY id ASC
		FOR UPDATE`,
		schoolID, studentID,
	); err != nil {
		return err
	}

	// Compute cumulative totals before this charge to determine credit_before.
	_, _, _, _, creditBefore, err := computeStudentTotals(ctx, tx, schoolID, studentID)
	if err != nil {
		return err
	}

	chargeNum, err := chargeAmount.NumericValue()
	if err != nil {
		return err
	}

	// Step 3: Insert the fee charge using the uniqueness constraint for idempotency.
	var chargeID string
	err = tx.QueryRow(ctx, `
		INSERT INTO fee_charges (school_id, student_id, fee_period_id, fee_structure_id, amount, due_date)
		VALUES ($1, $2, $3, $4, $5, $6::date)
		ON CONFLICT (school_id, student_id, fee_period_id)
		DO NOTHING
		RETURNING id`,
		schoolID, studentID, feePeriodID, structureID, chargeNum, dueDateStr,
	).Scan(&chargeID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Charge already exists; idempotently return without mutating balances or entries.
		return nil
	}
	if err != nil {
		return err
	}

	// Step 4: Only when insertion succeeds, create ledger entry, increase billed_amount,
	// set due_date, record informational credit_applied if applicable, and recompute balances.
	if _, err := tx.Exec(ctx, `
		INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
		VALUES ($1, $2, $3, 'charge', $4, $5, $6)`,
		schoolID, studentID, termID, chargeNum, "CHG:"+chargeID, "Term fee charge",
	); err != nil {
		return err
	}

	absorbedCredit := creditBefore
	if chargeAmount < absorbedCredit {
		absorbedCredit = chargeAmount
	}
	if absorbedCredit > 0 {
		absorbedNum, err := absorbedCredit.NumericValue()
		if err != nil {
			return err
		}
		// Informational credit_applied entry: records credit carry-forward history
		// without modifying paid_amount or independently subtracting from current_balance.
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'credit_applied', $4, $5, $6)`,
			schoolID, studentID, termID, absorbedNum, "CRD:"+chargeID, "Credit carried forward applied to term charge",
		); err != nil {
			return err
		}
	}

	return syncStudentLedgerRow(ctx, tx, schoolID, studentID, yearID, termID, chargeAmount, 0, absorbedCredit, &dueDateStr)
}

func computeStudentTotals(ctx context.Context, tx pgx.Tx, schoolID, studentID string) (opening, billed, paid, current, credit Money, err error) {
	err = tx.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(opening_balance), 0)::numeric(12,2),
			COALESCE(SUM(billed_amount), 0)::numeric(12,2),
			COALESCE(SUM(paid_amount), 0)::numeric(12,2)
		FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2`,
		schoolID, studentID,
	).Scan(&opening, &billed, &paid)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	current, credit, err = balances(opening, billed, paid)
	return opening, billed, paid, current, credit, err
}

// syncStudentLedgerRow updates a term's billed/paid amounts and synchronizes the
// student's derived current_balance and credit_balance so that:
//
//	current_balance = max(0, total_opening + total_billed - total_paid)
//	credit_balance  = max(0, total_paid - total_opening - total_billed)
func syncStudentLedgerRow(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, studentID, yearID, termID string,
	billedDelta, paidDelta, creditAppliedDelta Money,
	dueDateStr *string,
) error {
	var rowID string
	var rowOpening, rowBilled, rowPaid, rowCreditApplied Money
	err := tx.QueryRow(ctx, `
		SELECT id, opening_balance, billed_amount, paid_amount, credit_applied
		FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2 AND academic_year_id = $3 AND term_id = $4
		FOR UPDATE`,
		schoolID, studentID, yearID, termID,
	).Scan(&rowID, &rowOpening, &rowBilled, &rowPaid, &rowCreditApplied)
	if err != nil {
		return err
	}

	newBilled, err := addMoney(rowBilled, billedDelta)
	if err != nil {
		return err
	}
	if newBilled < 0 {
		newBilled = 0
	}
	newPaid, err := addMoney(rowPaid, paidDelta)
	if err != nil {
		return err
	}
	newCreditApplied, err := addMoney(rowCreditApplied, creditAppliedDelta)
	if err != nil {
		return err
	}

	billedNum, err := newBilled.NumericValue()
	if err != nil {
		return err
	}
	paidNum, err := newPaid.NumericValue()
	if err != nil {
		return err
	}
	creditAppNum, err := newCreditApplied.NumericValue()
	if err != nil {
		return err
	}

	if dueDateStr != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE fee_ledgers
			SET billed_amount = $1, paid_amount = $2, credit_applied = $3, due_date = $4::date, updated_at = NOW()
			WHERE school_id = $5 AND id = $6`,
			billedNum, paidNum, creditAppNum, *dueDateStr, schoolID, rowID,
		); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE fee_ledgers
			SET billed_amount = $1, paid_amount = $2, credit_applied = $3, updated_at = NOW()
			WHERE school_id = $4 AND id = $5`,
			billedNum, paidNum, creditAppNum, schoolID, rowID,
		); err != nil {
			return err
		}
	}

	return recomputeStudentLedgerBalances(ctx, tx, schoolID, studentID, rowID)
}

func recomputeStudentLedgerBalances(ctx context.Context, tx pgx.Tx, schoolID, studentID, activeLedgerID string) error {
	_, _, _, current, credit, err := computeStudentTotals(ctx, tx, schoolID, studentID)
	if err != nil {
		return err
	}
	currNum, err := current.NumericValue()
	if err != nil {
		return err
	}
	credNum, err := credit.NumericValue()
	if err != nil {
		return err
	}
	zeroNum, _ := Money(0).NumericValue()

	// Zero out previous term ledger rows' active current/credit snapshot so the active
	// ledger row (and SUM across fee_ledgers) both reflect the exact derived student balance.
	if _, err := tx.Exec(ctx, `
		UPDATE fee_ledgers
		SET current_balance = CASE WHEN id = $3::uuid THEN $4::numeric(12,2) ELSE $6::numeric(12,2) END,
		    credit_balance  = CASE WHEN id = $3::uuid THEN $5::numeric(12,2) ELSE $6::numeric(12,2) END,
		    updated_at      = NOW()
		WHERE school_id = $1 AND student_id = $2`,
		schoolID, studentID, activeLedgerID, currNum, credNum, zeroNum,
	); err != nil {
		return err
	}
	return nil
}

// RecordStudentPaymentTx credits a student's ledger for a payment within an open tenant transaction.
func (s *Service) RecordStudentPaymentTx(ctx context.Context, tx pgx.Tx, schoolID, studentID, reference, description string, amount Money) error {
	if amount <= 0 {
		return validationFail("amount", "INVALID_PAYMENT_AMOUNT", "Payment amount must be strictly positive.")
	}
	var lockedID string
	if err := tx.QueryRow(ctx, `SELECT id FROM students WHERE school_id = $1 AND id = $2 FOR UPDATE`, schoolID, studentID).Scan(&lockedID); err != nil {
		return err
	}

	yearID, termID, dueDateStr, err := resolveActiveLedgerTerm(ctx, tx, schoolID, studentID)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO fee_ledgers (school_id, student_id, academic_year_id, term_id, due_date)
		VALUES ($1, $2, $3, $4, $5::date)
		ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO NOTHING`,
		schoolID, studentID, yearID, termID, dueDateStr,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		SELECT id FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2
		ORDER BY id ASC
		FOR UPDATE`,
		schoolID, studentID,
	); err != nil {
		return err
	}

	amtNum, err := amount.NumericValue()
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
		VALUES ($1, $2, $3, 'payment', $4, $5, $6)
		ON CONFLICT (school_id, student_id, term_id, entry_type, reference) DO NOTHING`,
		schoolID, studentID, termID, amtNum, reference, description,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}

	return syncStudentLedgerRow(ctx, tx, schoolID, studentID, yearID, termID, 0, amount, 0, nil)
}

func resolveActiveLedgerTerm(ctx context.Context, tx pgx.Tx, schoolID, studentID string) (yearID, termID string, dueDateStr *string, err error) {
	var dueDate *time.Time
	err = tx.QueryRow(ctx, `
		SELECT fl.academic_year_id, fl.term_id, fl.due_date
		FROM fee_ledgers fl
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2
		ORDER BY t.start_date DESC, t.term_number DESC, fl.updated_at DESC
		LIMIT 1`,
		schoolID, studentID,
	).Scan(&yearID, &termID, &dueDate)
	if err == nil {
		if dueDate != nil {
			s := dueDate.Format("2006-01-02")
			dueDateStr = &s
		}
		return yearID, termID, dueDateStr, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, err
	}

	var startDate time.Time
	err = tx.QueryRow(ctx, `
		SELECT t.academic_year_id, t.id, t.start_date
		FROM terms t
		JOIN academic_years y ON y.school_id = t.school_id AND y.id = t.academic_year_id
		WHERE t.school_id = $1
		ORDER BY t.is_current DESC, y.is_current DESC, y.year DESC, t.term_number DESC
		LIMIT 1`,
		schoolID,
	).Scan(&yearID, &termID, &startDate)
	if err != nil {
		return "", "", nil, err
	}
	s := startDate.Format("2006-01-02")
	return yearID, termID, &s, nil
}

func (s *Service) GetStudentFeeBalance(ctx context.Context, p contracts.Principal, studentID string) (StudentFeeBalance, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return StudentFeeBalance{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !contracts.ValidID(studentID) {
		return StudentFeeBalance{}, contracts.Fail(http.StatusBadRequest, "INVALID_STUDENT_ID", "A valid student ID is required.")
	}
	if s == nil || s.DB == nil {
		return StudentFeeBalance{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	var out StudentFeeBalance
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		var err error
		out, err = s.GetStudentFeeBalanceTx(ctx, tx, p, studentID)
		return err
	})
	return out, err
}

func (s *Service) GetStudentLedger(ctx context.Context, p contracts.Principal, studentID string) (StudentLedger, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return StudentLedger{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !contracts.ValidID(studentID) {
		return StudentLedger{}, contracts.Fail(http.StatusBadRequest, "INVALID_STUDENT_ID", "A valid student ID is required.")
	}
	if s == nil || s.DB == nil {
		return StudentLedger{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	var out StudentLedger
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		balance, err := s.GetStudentFeeBalanceTx(ctx, tx, p, studentID)
		if err != nil {
			return err
		}
		entries, err := loadStudentLedgerEntries(ctx, tx, p.SchoolID, studentID)
		if err != nil {
			return err
		}
		out = StudentLedger{
			StudentFeeBalance: balance,
			Entries:           entries,
		}
		return nil
	})
	return out, err
}

func (s *Service) GetStudentFeeBalanceTx(ctx context.Context, tx pgx.Tx, p contracts.Principal, studentID string) (StudentFeeBalance, error) {
	var admissionNumber, fullName string
	err := tx.QueryRow(ctx, `
		SELECT admission_number, full_name
		FROM students
		WHERE school_id = $1 AND id = $2`,
		p.SchoolID, studentID,
	).Scan(&admissionNumber, &fullName)
	if errors.Is(err, pgx.ErrNoRows) {
		return StudentFeeBalance{}, contracts.Fail(http.StatusNotFound, "STUDENT_NOT_FOUND", "Student not found.")
	}
	if err != nil {
		return StudentFeeBalance{}, err
	}

	if !isIkonexTeamRole(p.Role) {
		var linked bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM student_guardians
				WHERE school_id = $1 AND student_id = $2 AND user_id = $3
			)`,
			p.SchoolID, studentID, p.UserID,
		).Scan(&linked)
		if err != nil {
			return StudentFeeBalance{}, err
		}
		if !linked {
			return StudentFeeBalance{}, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "You are not authorized to view this student's financial records.")
		}
	}

	opening, billed, paid, current, credit, err := computeStudentTotals(ctx, tx, p.SchoolID, studentID)
	if err != nil {
		return StudentFeeBalance{}, err
	}

	var dueDate *time.Time
	_ = tx.QueryRow(ctx, `
		SELECT fl.due_date
		FROM fee_ledgers fl
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2 AND fl.due_date IS NOT NULL
		ORDER BY t.start_date DESC, t.term_number DESC, fl.updated_at DESC
		LIMIT 1`,
		p.SchoolID, studentID,
	).Scan(&dueDate)
	var dueDateStr *string
	if dueDate != nil {
		formatted := dueDate.Format("2006-01-02")
		dueDateStr = &formatted
	}

	items, err := loadStudentFeeItems(ctx, tx, p.SchoolID, studentID)
	if err != nil {
		return StudentFeeBalance{}, err
	}
	notes, err := loadStudentFeeChangeNotes(ctx, tx, p.SchoolID, studentID)
	if err != nil {
		return StudentFeeBalance{}, err
	}

	return StudentFeeBalance{
		StudentID:       studentID,
		AdmissionNumber: admissionNumber,
		StudentName:     fullName,
		OpeningBalance:  opening,
		BilledAmount:    billed,
		PaidAmount:      paid,
		CurrentBalance:  current,
		CreditBalance:   credit,
		DueDate:         dueDateStr,
		Items:           items,
		FeeChangeNotes:  notes,
	}, nil
}

func loadStudentFeeItems(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]FeeStructureItem, error) {
	rows, err := tx.Query(ctx, `
		SELECT i.id, i.title, i.amount, i.is_compulsory
		FROM fee_charges fc
		JOIN fee_structure_items i ON i.fee_structure_id = fc.fee_structure_id
		WHERE fc.school_id = $1 AND fc.student_id = $2
		ORDER BY fc.due_date DESC, lower(i.title) ASC, i.title ASC, i.id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FeeStructureItem{}
	for rows.Next() {
		var it FeeStructureItem
		if err := rows.Scan(&it.ID, &it.Title, &it.Amount, &it.IsCompulsory); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func loadStudentFeeChangeNotes(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]FeeChangeNoteView, error) {
	rows, err := tx.Query(ctx, `
		SELECT n.id, n.term_id, n.fee_period_id, n.change_type, n.amount_delta,
		       n.old_due_date, n.new_due_date, n.note, n.affected_students_count, n.created_at
		FROM fee_change_notes n
		WHERE n.school_id = $1
		  AND (
			EXISTS (
				SELECT 1 FROM fee_change_note_students ns
				WHERE ns.school_id = n.school_id AND ns.note_id = n.id AND ns.student_id = $2
			)
			OR (
				NOT EXISTS (
					SELECT 1 FROM fee_change_note_students ns2
					WHERE ns2.school_id = n.school_id AND ns2.note_id = n.id
				)
				AND EXISTS (
					SELECT 1 FROM fee_charges fc
					WHERE fc.school_id = n.school_id AND fc.student_id = $2
					  AND (fc.fee_period_id = n.fee_period_id OR EXISTS (
						SELECT 1 FROM fee_periods fp
						WHERE fp.school_id = n.school_id AND fp.id = fc.fee_period_id AND fp.term_id = n.term_id
					  ))
				)
			)
		  )
		ORDER BY n.created_at DESC, n.id DESC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := []FeeChangeNoteView{}
	for rows.Next() {
		var n FeeChangeNoteView
		var oldDate, newDate *time.Time
		if err := rows.Scan(
			&n.ID, &n.TermID, &n.FeePeriodID, &n.ChangeType, &n.AmountDelta,
			&oldDate, &newDate, &n.Note, &n.AffectedStudentsCount, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		if oldDate != nil {
			s := oldDate.Format("2006-01-02")
			n.OldDueDate = &s
		}
		if newDate != nil {
			s := newDate.Format("2006-01-02")
			n.NewDueDate = &s
		}
		n.CreatedAt = n.CreatedAt.UTC()
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func loadStudentLedgerEntries(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]LedgerEntry, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, term_id, entry_type, amount, reference, COALESCE(description, ''), created_at
		FROM ledger_entries
		WHERE school_id = $1 AND student_id = $2
		ORDER BY created_at ASC, id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.TermID, &e.EntryType, &e.Amount, &e.Reference, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = e.CreatedAt.UTC()
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func registerLedgerRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	mux.HandleFunc("GET /api/v1/students/{id}/fee-balance", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		balance, err := service.GetStudentFeeBalance(r.Context(), p, r.PathValue("id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, balance, nil)
	})
	mux.HandleFunc("GET /api/v1/students/{id}/ledger", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		ledger, err := service.GetStudentLedger(r.Context(), p, r.PathValue("id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, ledger, nil)
	})
}

type DarajaProvider interface{}

type STKPollQueue interface{}

func registerMpesaRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {}
