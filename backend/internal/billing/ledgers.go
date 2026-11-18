package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

type TermFeeLedger struct {
	AcademicYearID    string  `json:"academic_year_id"`
	Year              int     `json:"year"`
	TermID            string  `json:"term_id"`
	TermNumber        int     `json:"term_number"`
	DueDate           *string `json:"due_date,omitempty"`
	OpeningBalance    Money   `json:"opening_balance"`
	BilledAmount      Money   `json:"billed_amount"`
	PaidAmount        Money   `json:"paid_amount"`
	CreditApplied     Money   `json:"credit_applied"`
	CreditTransferred Money   `json:"credit_transferred"`
	CurrentBalance    Money   `json:"current_balance"`
	CreditBalance     Money   `json:"credit_balance"`
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
	Terms           []TermFeeLedger     `json:"terms"`
	Items           []FeeStructureItem  `json:"items"`
	FeeChangeNotes  []FeeChangeNoteView `json:"fee_change_notes"`
}

type StudentLedger struct {
	StudentFeeBalance
	Entries []LedgerEntry `json:"entries"`
}

type studentLedgerRow struct {
	ID                string
	SchoolID          string
	StudentID         string
	AcademicYearID    string
	Year              int
	TermID            string
	TermNumber        int
	TermStartDate     time.Time
	DueDate           *time.Time
	OpeningBalance    Money
	BilledAmount      Money
	PaidAmount        Money
	CreditApplied     Money
	CreditTransferred Money
	CurrentBalance    Money
	CreditBalance     Money
}

func (s *Service) chargeGradeForTerm(ctx context.Context, tx pgx.Tx, structureID, actorUserID string) error {
	_, err := s.chargeGradeForTermCount(ctx, tx, structureID, actorUserID)
	return err
}

func (s *Service) chargeGradeForTermCount(ctx context.Context, tx pgx.Tx, structureID, actorUserID string) (int, error) {
	if !contracts.ValidID(structureID) {
		return 0, contracts.Fail(http.StatusBadRequest, "INVALID_FEE_STRUCTURE", "A valid fee structure ID is required.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return 0, err
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
		return 0, contracts.Fail(http.StatusNotFound, "FEE_STRUCTURE_NOT_FOUND", "Fee structure not found.")
	}
	if err != nil {
		return 0, err
	}
	if boardingStatus != "all" {
		return 0, validationFail("boarding_status", "UNSUPPORTED_BOARDING_VARIANT", "Release 1 supports only the 'all' boarding_status fee structure variant.")
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
			return 0, err
		}
		if err := tx.QueryRow(ctx, `
			INSERT INTO fee_periods (school_id, academic_year_id, term_id, period_type, name, due_date)
			VALUES ($1, $2, $3, 'term', $4, $5)
			ON CONFLICT (school_id, academic_year_id, term_id)
			DO UPDATE SET name = fee_periods.name
			RETURNING id, due_date`,
			schoolID, yearID, termID, fmt.Sprintf("Term %d", termNumber), startDate,
		).Scan(&feePeriodID, &dueDate); err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
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
		ORDER BY s.id ASC`,
		schoolID, yearID, gradeID,
	)
	if err != nil {
		return 0, err
	}
	var studentIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return 0, err
		}
		studentIDs = append(studentIDs, sid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	createdCount := 0
	for _, studentID := range studentIDs {
		created, err := s.applyStudentCharge(ctx, tx, schoolID, actorUserID, studentID, yearID, termID, feePeriodID, structureID, totalAmount, dueDate)
		if err != nil {
			return 0, err
		}
		if created {
			createdCount++
		}
	}
	return createdCount, nil
}

func (s *Service) chargeStudent(ctx context.Context, tx pgx.Tx, studentID, feePeriodID, actorUserID string) error {
	if !contracts.ValidID(studentID) || !contracts.ValidID(feePeriodID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_CHARGE_TARGET", "Valid student_id and fee_period_id are required.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
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
	err = tx.QueryRow(ctx, `
		SELECT e.grade_level_id
		FROM enrolments e
		JOIN students s ON s.school_id = e.school_id AND s.id = e.student_id
		WHERE e.school_id = $1 AND e.student_id = $2 AND e.academic_year_id = $3
		  AND e.status = 'active' AND s.status = 'active'`,
		schoolID, studentID, yearID,
	).Scan(&gradeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	var structureID string
	var totalAmount Money
	err = tx.QueryRow(ctx, `
		SELECT id, total_amount
		FROM fee_structures
		WHERE school_id = $1 AND academic_year_id = $2 AND grade_level_id = $3 AND term_id = $4
		  AND boarding_status = 'all'
		LIMIT 1`,
		schoolID, yearID, gradeID, termID,
	).Scan(&structureID, &totalAmount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = s.applyStudentCharge(ctx, tx, schoolID, actorUserID, studentID, yearID, termID, feePeriodID, structureID, totalAmount, dueDate)
	return err
}

func (s *Service) chargeAdmittedStudent(ctx context.Context, tx pgx.Tx, schoolID, studentID, academicYearID string) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(studentID) || !contracts.ValidID(academicYearID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_ADMISSION_CHARGE", "Valid school_id, student_id and academic_year_id are required.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
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
		if err := s.chargeStudent(ctx, tx, studentID, pid, ""); err != nil {
			return err
		}
	}
	return nil
}

func lockAndLoadStudentLedgers(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]studentLedgerRow, error) {
	if _, err := tx.Exec(ctx, `
		SELECT id FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2
		ORDER BY id ASC
		FOR UPDATE`,
		schoolID, studentID,
	); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `
		SELECT fl.id, fl.school_id, fl.student_id, fl.academic_year_id, ay.year,
		       fl.term_id, t.term_number, t.start_date, fl.due_date,
		       fl.opening_balance, fl.billed_amount, fl.paid_amount,
		       fl.credit_applied, fl.credit_transferred,
		       fl.current_balance, fl.credit_balance
		FROM fee_ledgers fl
		JOIN academic_years ay ON ay.school_id = fl.school_id AND ay.id = fl.academic_year_id
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2
		ORDER BY ay.year ASC, t.term_number ASC, fl.id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []studentLedgerRow{}
	for rows.Next() {
		var r studentLedgerRow
		if err := rows.Scan(
			&r.ID, &r.SchoolID, &r.StudentID, &r.AcademicYearID, &r.Year,
			&r.TermID, &r.TermNumber, &r.TermStartDate, &r.DueDate,
			&r.OpeningBalance, &r.BilledAmount, &r.PaidAmount,
			&r.CreditApplied, &r.CreditTransferred,
			&r.CurrentBalance, &r.CreditBalance,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func updateLedgerRowInPlace(ctx context.Context, tx pgx.Tx, row studentLedgerRow, dueDateStr *string) error {
	billedNum, err := row.BilledAmount.NumericValue()
	if err != nil {
		return err
	}
	paidNum, err := row.PaidAmount.NumericValue()
	if err != nil {
		return err
	}
	creditAppNum, err := row.CreditApplied.NumericValue()
	if err != nil {
		return err
	}
	creditXferNum, err := row.CreditTransferred.NumericValue()
	if err != nil {
		return err
	}
	currNum, err := row.CurrentBalance.NumericValue()
	if err != nil {
		return err
	}
	credNum, err := row.CreditBalance.NumericValue()
	if err != nil {
		return err
	}
	if dueDateStr != nil {
		_, err = tx.Exec(ctx, `
			UPDATE fee_ledgers
			SET billed_amount = $1,
			    paid_amount = $2,
			    credit_applied = $3,
			    credit_transferred = $4,
			    current_balance = $5,
			    credit_balance = $6,
			    due_date = $7::date,
			    updated_at = NOW()
			WHERE school_id = $8 AND id = $9`,
			billedNum, paidNum, creditAppNum, creditXferNum, currNum, credNum, *dueDateStr, row.SchoolID, row.ID,
		)
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE fee_ledgers
		SET billed_amount = $1,
		    paid_amount = $2,
		    credit_applied = $3,
		    credit_transferred = $4,
		    current_balance = $5,
		    credit_balance = $6,
		    updated_at = NOW()
		WHERE school_id = $7 AND id = $8`,
		billedNum, paidNum, creditAppNum, creditXferNum, currNum, credNum, row.SchoolID, row.ID,
	)
	return err
}

// applyStudentCharge idempotently creates a fee_charges row for (student, fee_period),
// records the immutable 'charge' ledger_entry, carries forward any available
// credit (same-term first, then explicit paired transfers from prior-term
// ledgers), updates only the affected fee_ledgers rows while preserving
// historical per-term balances, and emits a durable audit event.
func (s *Service) applyStudentCharge(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, actorUserID, studentID, yearID, termID, feePeriodID, structureID string,
	chargeAmount Money,
	dueDate time.Time,
) (bool, error) {
	if err := s.requireAuditLogger(); err != nil {
		return false, err
	}
	var lockedID string
	if err := tx.QueryRow(ctx, `SELECT id FROM students WHERE school_id = $1 AND id = $2 FOR UPDATE`, schoolID, studentID).Scan(&lockedID); err != nil {
		return false, err
	}

	dueDateStr := dueDate.Format("2006-01-02")

	if _, err := tx.Exec(ctx, `
		INSERT INTO fee_ledgers (school_id, student_id, academic_year_id, term_id, due_date)
		VALUES ($1, $2, $3, $4, $5::date)
		ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO NOTHING`,
		schoolID, studentID, yearID, termID, dueDateStr,
	); err != nil {
		return false, err
	}

	ledgers, err := lockAndLoadStudentLedgers(ctx, tx, schoolID, studentID)
	if err != nil {
		return false, err
	}

	chargeNum, err := chargeAmount.NumericValue()
	if err != nil {
		return false, err
	}

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
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
		VALUES ($1, $2, $3, 'charge', $4, $5, $6)`,
		schoolID, studentID, termID, chargeNum, "CHG:"+chargeID, "Term fee charge",
	); err != nil {
		return false, err
	}

	dstIdx := -1
	for i := range ledgers {
		if ledgers[i].AcademicYearID == yearID && ledgers[i].TermID == termID {
			dstIdx = i
			break
		}
	}
	if dstIdx < 0 {
		return false, fmt.Errorf("billing: missing fee_ledgers row for student %s term %s", studentID, termID)
	}

	dst := ledgers[dstIdx]
	remaining := chargeAmount
	var totalCreditApplied Money

	// 1. Consume any available credit already held on the destination term first.
	if dst.CreditBalance > 0 && remaining > 0 {
		absorbed := dst.CreditBalance
		if remaining < absorbed {
			absorbed = remaining
		}
		dst.CreditBalance -= absorbed
		nextApp, err := addMoney(dst.CreditApplied, absorbed)
		if err != nil {
			return false, err
		}
		dst.CreditApplied = nextApp
		remaining -= absorbed
		totalCreditApplied, err = addMoney(totalCreditApplied, absorbed)
		if err != nil {
			return false, err
		}

		absorbedNum, err := absorbed.NumericValue()
		if err != nil {
			return false, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'credit_applied', $4, $5, $6)`,
			schoolID, studentID, termID, absorbedNum, "CRD:"+chargeID, "Credit applied to term charge",
		); err != nil {
			return false, err
		}
	}

	// 2. Transfer remaining needed credit explicitly from other term ledgers.
	for i := range ledgers {
		if i == dstIdx || remaining == 0 {
			continue
		}
		src := ledgers[i]
		if src.CreditBalance <= 0 {
			continue
		}
		xfer := src.CreditBalance
		if remaining < xfer {
			xfer = remaining
		}
		src.CreditBalance -= xfer
		nextXfer, err := addMoney(src.CreditTransferred, xfer)
		if err != nil {
			return false, err
		}
		src.CreditTransferred = nextXfer
		if err := updateLedgerRowInPlace(ctx, tx, src, nil); err != nil {
			return false, err
		}
		ledgers[i] = src

		xferNum, err := xfer.NumericValue()
		if err != nil {
			return false, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'adjustment', $4, $5, $6)`,
			schoolID, studentID, src.TermID, xferNum,
			"XFER:"+chargeID+":"+src.TermID,
			fmt.Sprintf("Credit transferred to Year %d Term %d", dst.Year, dst.TermNumber),
		); err != nil {
			return false, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'credit_applied', $4, $5, $6)`,
			schoolID, studentID, termID, xferNum,
			"CRD:"+chargeID+":"+src.TermID,
			fmt.Sprintf("Credit carried forward from Year %d Term %d", src.Year, src.TermNumber),
		); err != nil {
			return false, err
		}

		nextApp, err := addMoney(dst.CreditApplied, xfer)
		if err != nil {
			return false, err
		}
		dst.CreditApplied = nextApp
		remaining -= xfer
		totalCreditApplied, err = addMoney(totalCreditApplied, xfer)
		if err != nil {
			return false, err
		}
	}

	nextBilled, err := addMoney(dst.BilledAmount, chargeAmount)
	if err != nil {
		return false, err
	}
	nextCurrent, err := addMoney(dst.CurrentBalance, remaining)
	if err != nil {
		return false, err
	}
	dst.BilledAmount = nextBilled
	dst.CurrentBalance = nextCurrent
	if err := updateLedgerRowInPlace(ctx, tx, dst, &dueDateStr); err != nil {
		return false, err
	}

	if err := s.logAuditPgx(ctx, tx, contracts.AuditEvent{
		SchoolID:     schoolID,
		UserID:       actorUserID,
		Action:       "billing.student.charge",
		ResourceType: "fee_charge",
		ResourceID:   chargeID,
		Details: map[string]any{
			"student_id":       studentID,
			"fee_period_id":    feePeriodID,
			"fee_structure_id": structureID,
			"term_id":          termID,
			"amount":           chargeAmount.String(),
			"credit_applied":   totalCreditApplied.String(),
		},
	}); err != nil {
		return false, err
	}
	return true, nil
}

// RecordStudentPaymentTx records an unallocated student payment within an open
// tenant transaction using deterministic oldest-due-first allocation across
// outstanding term balances, storing any overpayment as credit_balance on the
// active/current term.
func (s *Service) RecordStudentPaymentTx(ctx context.Context, tx pgx.Tx, schoolID, studentID, reference, description string, amount Money) error {
	return s.recordStudentPayment(ctx, tx, schoolID, "", studentID, "", reference, description, amount)
}

// RecordStudentPaymentWithActorTx records an unallocated payment attributed to
// actorUserID.
func (s *Service) RecordStudentPaymentWithActorTx(ctx context.Context, tx pgx.Tx, schoolID, actorUserID, studentID, reference, description string, amount Money) error {
	return s.recordStudentPayment(ctx, tx, schoolID, actorUserID, studentID, "", reference, description, amount)
}

// RecordStudentPaymentForTermTx records a payment explicitly targeted at
// targetTermID (reducing that term's current_balance first, transferring any
// excess to remaining owing terms oldest-due-first, and holding any remaining
// surplus as credit_balance on targetTermID).
func (s *Service) RecordStudentPaymentForTermTx(ctx context.Context, tx pgx.Tx, schoolID, actorUserID, studentID, targetTermID, reference, description string, amount Money) error {
	if !contracts.ValidID(targetTermID) {
		return validationFail("term_id", "INVALID_TERM", "A valid term_id is required for explicit term payment allocation.")
	}
	return s.recordStudentPayment(ctx, tx, schoolID, actorUserID, studentID, targetTermID, reference, description, amount)
}

func (s *Service) recordStudentPayment(ctx context.Context, tx pgx.Tx, schoolID, actorUserID, studentID, targetTermID, reference, description string, amount Money) error {
	if !contracts.ValidID(schoolID) {
		return validationFail("school_id", "INVALID_SCHOOL", "A valid school_id is required.")
	}
	if !contracts.ValidID(studentID) {
		return validationFail("student_id", "INVALID_STUDENT_ID", "A valid student_id is required.")
	}
	if actorUserID != "" && !contracts.ValidID(actorUserID) {
		return validationFail("actor_user_id", "INVALID_ACTOR", "A valid actor_user_id is required.")
	}
	ref := strings.TrimSpace(reference)
	if ref == "" || len(ref) > 100 {
		return validationFail("reference", "INVALID_PAYMENT_REFERENCE", "A non-empty payment reference (<= 100 chars) is required.")
	}
	if amount <= 0 {
		return validationFail("amount", "INVALID_PAYMENT_AMOUNT", "Payment amount must be strictly positive.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
	}

	var lockedID string
	if err := tx.QueryRow(ctx, `SELECT id FROM students WHERE school_id = $1 AND id = $2 FOR UPDATE`, schoolID, studentID).Scan(&lockedID); err != nil {
		return err
	}

	// Lock all ledger rows for this student before checking idempotency so
	// concurrent replays of the same payment reference serialize cleanly.
	if _, err := tx.Exec(ctx, `
		SELECT id FROM fee_ledgers
		WHERE school_id = $1 AND student_id = $2
		ORDER BY id ASC
		FOR UPDATE`,
		schoolID, studentID,
	); err != nil {
		return err
	}

	var alreadyRecorded bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM ledger_entries
			WHERE school_id = $1 AND student_id = $2 AND entry_type = 'payment' AND reference = $3
		)`,
		schoolID, studentID, ref,
	).Scan(&alreadyRecorded); err != nil {
		return err
	}
	if alreadyRecorded {
		return nil
	}

	type allocationLog struct {
		TermID string `json:"term_id"`
		Amount string `json:"amount"`
	}
	var allocations []allocationLog

	if targetTermID != "" {
		var yearID string
		var startDate time.Time
		err := tx.QueryRow(ctx, `
			SELECT academic_year_id, start_date
			FROM terms
			WHERE school_id = $1 AND id = $2`,
			schoolID, targetTermID,
		).Scan(&yearID, &startDate)
		if errors.Is(err, pgx.ErrNoRows) {
			return validationFail("term_id", "INVALID_TERM", "Term does not exist in this school.")
		}
		if err != nil {
			return err
		}

		dueStr := startDate.Format("2006-01-02")
		if _, err := tx.Exec(ctx, `
			INSERT INTO fee_ledgers (school_id, student_id, academic_year_id, term_id, due_date)
			VALUES ($1, $2, $3, $4, $5::date)
			ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO NOTHING`,
			schoolID, studentID, yearID, targetTermID, dueStr,
		); err != nil {
			return err
		}

		ledgers, err := lockAndLoadStudentLedgers(ctx, tx, schoolID, studentID)
		if err != nil {
			return err
		}
		targetIdx := -1
		for i := range ledgers {
			if ledgers[i].TermID == targetTermID {
				targetIdx = i
				break
			}
		}
		if targetIdx < 0 {
			return fmt.Errorf("billing: missing fee_ledgers row for target term %s", targetTermID)
		}

		amtNum, err := amount.NumericValue()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'payment', $4, $5, $6)`,
			schoolID, studentID, targetTermID, amtNum, ref, description,
		); err != nil {
			return err
		}
		allocations = append(allocations, allocationLog{TermID: targetTermID, Amount: amount.String()})

		target := ledgers[targetIdx]
		nextPaid, err := addMoney(target.PaidAmount, amount)
		if err != nil {
			return err
		}
		target.PaidAmount = nextPaid

		var excess Money
		if amount <= target.CurrentBalance {
			target.CurrentBalance -= amount
		} else {
			excess = amount - target.CurrentBalance
			target.CurrentBalance = 0
		}

		// If explicit payment overpaid targetTermID while other terms have
		// outstanding balances, transfer excess credit oldest-due-first.
		if excess > 0 {
			owingIDs, err := loadOwingLedgerIDsOldestDueFirst(ctx, tx, schoolID, studentID)
			if err != nil {
				return err
			}
			byID := make(map[string]int, len(ledgers))
			for i := range ledgers {
				byID[ledgers[i].ID] = i
			}
			for _, id := range owingIDs {
				if excess == 0 {
					break
				}
				idx, ok := byID[id]
				if !ok || idx == targetIdx {
					continue
				}
				dst := ledgers[idx]
				if dst.CurrentBalance <= 0 {
					continue
				}
				xfer := dst.CurrentBalance
				if excess < xfer {
					xfer = excess
				}
				nextXfer, err := addMoney(target.CreditTransferred, xfer)
				if err != nil {
					return err
				}
				target.CreditTransferred = nextXfer
				nextApp, err := addMoney(dst.CreditApplied, xfer)
				if err != nil {
					return err
				}
				dst.CreditApplied = nextApp
				dst.CurrentBalance -= xfer
				excess -= xfer
				if err := updateLedgerRowInPlace(ctx, tx, dst, nil); err != nil {
					return err
				}
				ledgers[idx] = dst

				xferNum, err := xfer.NumericValue()
				if err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `
					INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
					VALUES ($1, $2, $3, 'adjustment', $4, $5, $6)`,
					schoolID, studentID, target.TermID, xferNum,
					"XFER:"+ref+":"+dst.TermID,
					fmt.Sprintf("Excess payment credit transferred to Year %d Term %d", dst.Year, dst.TermNumber),
				); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `
					INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
					VALUES ($1, $2, $3, 'credit_applied', $4, $5, $6)`,
					schoolID, studentID, dst.TermID, xferNum,
					"CRD:"+ref+":"+target.TermID,
					fmt.Sprintf("Excess payment credit from Year %d Term %d applied", target.Year, target.TermNumber),
				); err != nil {
					return err
				}
			}
			if excess > 0 {
				nextCred, err := addMoney(target.CreditBalance, excess)
				if err != nil {
					return err
				}
				target.CreditBalance = nextCred
			}
		}

		if err := updateLedgerRowInPlace(ctx, tx, target, nil); err != nil {
			return err
		}
	} else {
		// Automatic unallocated payment: oldest-due-first across owing terms.
		owingIDs, err := loadOwingLedgerIDsOldestDueFirst(ctx, tx, schoolID, studentID)
		if err != nil {
			return err
		}
		ledgers, err := lockAndLoadStudentLedgers(ctx, tx, schoolID, studentID)
		if err != nil {
			return err
		}
		byID := make(map[string]int, len(ledgers))
		for i := range ledgers {
			byID[ledgers[i].ID] = i
		}

		remaining := amount
		termOrder := make([]int, 0, len(owingIDs)+1)
		seenIdx := make(map[int]bool)
		paidForIdx := make(map[int]Money)

		for _, id := range owingIDs {
			if remaining == 0 {
				break
			}
			idx, ok := byID[id]
			if !ok {
				continue
			}
			row := &ledgers[idx]
			if row.CurrentBalance <= 0 {
				continue
			}
			alloc := row.CurrentBalance
			if remaining < alloc {
				alloc = remaining
			}
			nextPaid, err := addMoney(row.PaidAmount, alloc)
			if err != nil {
				return err
			}
			row.PaidAmount = nextPaid
			row.CurrentBalance -= alloc
			remaining -= alloc
			paidForIdx[idx] = alloc
			if !seenIdx[idx] {
				seenIdx[idx] = true
				termOrder = append(termOrder, idx)
			}
		}

		if remaining > 0 {
			creditYearID, creditTermID, creditDueStr, err := resolveCreditHoldingTerm(ctx, tx, schoolID, studentID)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO fee_ledgers (school_id, student_id, academic_year_id, term_id, due_date)
				VALUES ($1, $2, $3, $4, $5::date)
				ON CONFLICT (school_id, student_id, academic_year_id, term_id) DO NOTHING`,
				schoolID, studentID, creditYearID, creditTermID, creditDueStr,
			); err != nil {
				return err
			}
			// Reload only if the holding row wasn't already present in ledgers.
			creditIdx := -1
			for i := range ledgers {
				if ledgers[i].TermID == creditTermID {
					creditIdx = i
					break
				}
			}
			if creditIdx < 0 {
				reloaded, err := lockAndLoadStudentLedgers(ctx, tx, schoolID, studentID)
				if err != nil {
					return err
				}
				for i := range reloaded {
					if reloaded[i].TermID == creditTermID {
						ledgers = append(ledgers, reloaded[i])
						creditIdx = len(ledgers) - 1
						break
					}
				}
			}
			if creditIdx < 0 {
				return fmt.Errorf("billing: missing credit holding ledger for term %s", creditTermID)
			}
			row := &ledgers[creditIdx]
			nextPaid, err := addMoney(row.PaidAmount, remaining)
			if err != nil {
				return err
			}
			nextCred, err := addMoney(row.CreditBalance, remaining)
			if err != nil {
				return err
			}
			row.PaidAmount = nextPaid
			row.CreditBalance = nextCred
			prevAlloc := paidForIdx[creditIdx]
			combinedAlloc, err := addMoney(prevAlloc, remaining)
			if err != nil {
				return err
			}
			paidForIdx[creditIdx] = combinedAlloc
			if !seenIdx[creditIdx] {
				seenIdx[creditIdx] = true
				termOrder = append(termOrder, creditIdx)
			}
		}

		for _, idx := range termOrder {
			row := ledgers[idx]
			termAmt := paidForIdx[idx]
			if err := updateLedgerRowInPlace(ctx, tx, row, nil); err != nil {
				return err
			}
			termAmtNum, err := termAmt.NumericValue()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
				VALUES ($1, $2, $3, 'payment', $4, $5, $6)`,
				schoolID, studentID, row.TermID, termAmtNum, ref, description,
			); err != nil {
				return err
			}
			allocations = append(allocations, allocationLog{TermID: row.TermID, Amount: termAmt.String()})
		}
	}

	return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
		SchoolID:     schoolID,
		UserID:       actorUserID,
		Action:       "billing.payment.record",
		ResourceType: "student_ledger",
		ResourceID:   studentID,
		Details: map[string]any{
			"reference":      ref,
			"amount":         amount.String(),
			"target_term_id": targetTermID,
			"allocations":    allocations,
		},
	})
}

func loadOwingLedgerIDsOldestDueFirst(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT fl.id
		FROM fee_ledgers fl
		JOIN academic_years ay ON ay.school_id = fl.school_id AND ay.id = fl.academic_year_id
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2 AND fl.current_balance > 0
		ORDER BY COALESCE(fl.due_date, t.start_date) ASC, ay.year ASC, t.term_number ASC, fl.id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func resolveCreditHoldingTerm(ctx context.Context, tx pgx.Tx, schoolID, studentID string) (yearID, termID string, dueDateStr *string, err error) {
	var dueDate *time.Time
	err = tx.QueryRow(ctx, `
		SELECT fl.academic_year_id, fl.term_id, fl.due_date
		FROM fee_ledgers fl
		JOIN academic_years y ON y.school_id = fl.school_id AND y.id = fl.academic_year_id
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2
		ORDER BY t.is_current DESC, y.is_current DESC, COALESCE(fl.due_date, t.start_date) DESC, y.year DESC, t.term_number DESC, fl.id DESC
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
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) || p.Role == "" {
		return StudentFeeBalance{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !contracts.ValidID(studentID) {
		return StudentFeeBalance{}, contracts.Fail(http.StatusBadRequest, "INVALID_STUDENT_ID", "A valid student ID is required.")
	}
	if s == nil || s.DB == nil {
		return StudentFeeBalance{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	var out StudentFeeBalance
	err := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
		var err error
		out, err = s.GetStudentFeeBalanceTx(ctx, tx, p, studentID)
		return err
	})
	return out, err
}

func (s *Service) GetStudentLedger(ctx context.Context, p contracts.Principal, studentID string) (StudentLedger, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) || p.Role == "" {
		return StudentLedger{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !contracts.ValidID(studentID) {
		return StudentLedger{}, contracts.Fail(http.StatusBadRequest, "INVALID_STUDENT_ID", "A valid student ID is required.")
	}
	if s == nil || s.DB == nil {
		return StudentLedger{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	var out StudentLedger
	err := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
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

	if err := authorizeStudentFinancialReadTx(ctx, tx, p, studentID); err != nil {
		return StudentFeeBalance{}, err
	}

	ledgers, err := loadStudentLedgersReadOnly(ctx, tx, p.SchoolID, studentID)
	if err != nil {
		return StudentFeeBalance{}, err
	}

	var opening, billed, paid, current, credit Money
	terms := make([]TermFeeLedger, 0, len(ledgers))
	var earliestOwingDue, latestAnyDue *time.Time
	for _, r := range ledgers {
		var err error
		if opening, err = addMoney(opening, r.OpeningBalance); err != nil {
			return StudentFeeBalance{}, err
		}
		if billed, err = addMoney(billed, r.BilledAmount); err != nil {
			return StudentFeeBalance{}, err
		}
		if paid, err = addMoney(paid, r.PaidAmount); err != nil {
			return StudentFeeBalance{}, err
		}
		if current, err = addMoney(current, r.CurrentBalance); err != nil {
			return StudentFeeBalance{}, err
		}
		if credit, err = addMoney(credit, r.CreditBalance); err != nil {
			return StudentFeeBalance{}, err
		}
		var dueStr *string
		if r.DueDate != nil {
			formatted := r.DueDate.Format("2006-01-02")
			dueStr = &formatted
			if r.CurrentBalance > 0 && (earliestOwingDue == nil || r.DueDate.Before(*earliestOwingDue)) {
				d := *r.DueDate
				earliestOwingDue = &d
			}
			if latestAnyDue == nil || r.DueDate.After(*latestAnyDue) {
				d := *r.DueDate
				latestAnyDue = &d
			}
		}
		terms = append(terms, TermFeeLedger{
			AcademicYearID:    r.AcademicYearID,
			Year:              r.Year,
			TermID:            r.TermID,
			TermNumber:        r.TermNumber,
			DueDate:           dueStr,
			OpeningBalance:    r.OpeningBalance,
			BilledAmount:      r.BilledAmount,
			PaidAmount:        r.PaidAmount,
			CreditApplied:     r.CreditApplied,
			CreditTransferred: r.CreditTransferred,
			CurrentBalance:    r.CurrentBalance,
			CreditBalance:     r.CreditBalance,
		})
	}

	var dueDateStr *string
	if earliestOwingDue != nil {
		s := earliestOwingDue.Format("2006-01-02")
		dueDateStr = &s
	} else if latestAnyDue != nil {
		s := latestAnyDue.Format("2006-01-02")
		dueDateStr = &s
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
		Terms:           terms,
		Items:           items,
		FeeChangeNotes:  notes,
	}, nil
}

func loadStudentLedgersReadOnly(ctx context.Context, tx pgx.Tx, schoolID, studentID string) ([]studentLedgerRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT fl.id, fl.school_id, fl.student_id, fl.academic_year_id, ay.year,
		       fl.term_id, t.term_number, t.start_date, fl.due_date,
		       fl.opening_balance, fl.billed_amount, fl.paid_amount,
		       fl.credit_applied, fl.credit_transferred,
		       fl.current_balance, fl.credit_balance
		FROM fee_ledgers fl
		JOIN academic_years ay ON ay.school_id = fl.school_id AND ay.id = fl.academic_year_id
		JOIN terms t ON t.school_id = fl.school_id AND t.id = fl.term_id
		WHERE fl.school_id = $1 AND fl.student_id = $2
		ORDER BY ay.year ASC, t.term_number ASC, fl.id ASC`,
		schoolID, studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []studentLedgerRow{}
	for rows.Next() {
		var r studentLedgerRow
		if err := rows.Scan(
			&r.ID, &r.SchoolID, &r.StudentID, &r.AcademicYearID, &r.Year,
			&r.TermID, &r.TermNumber, &r.TermStartDate, &r.DueDate,
			&r.OpeningBalance, &r.BilledAmount, &r.PaidAmount,
			&r.CreditApplied, &r.CreditTransferred,
			&r.CurrentBalance, &r.CreditBalance,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
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
