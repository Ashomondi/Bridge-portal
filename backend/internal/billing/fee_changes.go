package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type ApplyFeeChangeInput struct {
	FeePeriodID    string          `json:"fee_period_id"`
	TermID         string          `json:"term_id"`
	FeeStructureID string          `json:"fee_structure_id"`
	GradeLevelID   string          `json:"grade_level_id"`
	StudentIDs     []string        `json:"student_ids"`
	AmountDelta    json.RawMessage `json:"amount_delta"`
	Note           string          `json:"note"`
}

// ApplyFeeChange validates and transactionally applies a mid-year fee adjustment
// across affected students in the school, writing a fee_change_notes record,
// an audit log entry, and deduplicated in-app notifications to linked guardians.
func (s *Service) ApplyFeeChange(ctx context.Context, p contracts.Principal, in ApplyFeeChangeInput) (FeeChangeNoteView, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return FeeChangeNoteView{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !isIkonexTeamRole(p.Role) {
		return FeeChangeNoteView{}, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Only Ikonex administrators may apply mid-year fee changes.")
	}
	if s == nil || s.DB == nil {
		return FeeChangeNoteView{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	var out FeeChangeNoteView
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		var err error
		out, err = s.ApplyFeeChangeTx(ctx, tx, p.SchoolID, p.UserID, in)
		return err
	})
	return out, err
}

func (s *Service) ApplyFeeChangeTx(ctx context.Context, tx pgx.Tx, schoolID, actorUserID string, in ApplyFeeChangeInput) (FeeChangeNoteView, error) {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(actorUserID) {
		return FeeChangeNoteView{}, contracts.Fail(http.StatusBadRequest, "INVALID_ACTOR", "Valid school and actor IDs are required.")
	}
	note := strings.TrimSpace(in.Note)
	if note == "" {
		return FeeChangeNoteView{}, validationFail("note", "FEE_CHANGE_NOTE_REQUIRED", "An explanation note is required for every fee change.")
	}
	if len(in.AmountDelta) == 0 {
		return FeeChangeNoteView{}, validationFail("amount_delta", "INVALID_FEE_CHANGE_AMOUNT", "A non-zero amount_delta is required.")
	}
	var delta Money
	if err := json.Unmarshal(in.AmountDelta, &delta); err != nil || delta == 0 {
		return FeeChangeNoteView{}, validationFail("amount_delta", "INVALID_FEE_CHANGE_AMOUNT", "amount_delta must be a non-zero decimal within NUMERIC(12,2).")
	}

	var feePeriodID, termID, yearID string
	switch {
	case in.FeePeriodID != "":
		if !contracts.ValidID(in.FeePeriodID) {
			return FeeChangeNoteView{}, validationFail("fee_period_id", "INVALID_FEE_PERIOD", "A valid fee_period_id is required.")
		}
		err := tx.QueryRow(ctx, `
			SELECT id, COALESCE(term_id::text, ''), academic_year_id
			FROM fee_periods
			WHERE school_id = $1 AND id = $2`,
			schoolID, in.FeePeriodID,
		).Scan(&feePeriodID, &termID, &yearID)
		if errors.Is(err, pgx.ErrNoRows) {
			return FeeChangeNoteView{}, validationFail("fee_period_id", "INVALID_FEE_PERIOD", "Fee period does not exist in this school.")
		}
		if err != nil {
			return FeeChangeNoteView{}, err
		}
	case in.TermID != "":
		if !contracts.ValidID(in.TermID) {
			return FeeChangeNoteView{}, validationFail("term_id", "INVALID_TERM", "A valid term_id is required.")
		}
		err := tx.QueryRow(ctx, `
			SELECT id, term_id, academic_year_id
			FROM fee_periods
			WHERE school_id = $1 AND term_id = $2`,
			schoolID, in.TermID,
		).Scan(&feePeriodID, &termID, &yearID)
		if errors.Is(err, pgx.ErrNoRows) {
			return FeeChangeNoteView{}, validationFail("term_id", "INVALID_TERM", "No fee period exists for the specified term.")
		}
		if err != nil {
			return FeeChangeNoteView{}, err
		}
	case in.FeeStructureID != "":
		if !contracts.ValidID(in.FeeStructureID) {
			return FeeChangeNoteView{}, validationFail("fee_structure_id", "INVALID_FEE_STRUCTURE", "A valid fee_structure_id is required.")
		}
		err := tx.QueryRow(ctx, `
			SELECT fp.id, fs.term_id, fs.academic_year_id
			FROM fee_structures fs
			JOIN fee_periods fp ON fp.school_id = fs.school_id AND fp.academic_year_id = fs.academic_year_id AND fp.term_id = fs.term_id
			WHERE fs.school_id = $1 AND fs.id = $2`,
			schoolID, in.FeeStructureID,
		).Scan(&feePeriodID, &termID, &yearID)
		if errors.Is(err, pgx.ErrNoRows) {
			return FeeChangeNoteView{}, validationFail("fee_structure_id", "INVALID_FEE_STRUCTURE", "Fee structure does not exist in this school.")
		}
		if err != nil {
			return FeeChangeNoteView{}, err
		}
	default:
		return FeeChangeNoteView{}, validationFail("fee_period_id", "INVALID_FEE_PERIOD", "One of fee_period_id, term_id, or fee_structure_id is required.")
	}

	if in.GradeLevelID != "" && !contracts.ValidID(in.GradeLevelID) {
		return FeeChangeNoteView{}, validationFail("grade_level_id", "INVALID_GRADE_LEVEL", "A valid grade_level_id is required.")
	}
	if in.FeeStructureID != "" && !contracts.ValidID(in.FeeStructureID) {
		return FeeChangeNoteView{}, validationFail("fee_structure_id", "INVALID_FEE_STRUCTURE", "A valid fee_structure_id is required.")
	}
	for idx, sid := range in.StudentIDs {
		if !contracts.ValidID(sid) {
			return FeeChangeNoteView{}, validationFail(fmt.Sprintf("student_ids[%d]", idx), "INVALID_STUDENT_ID", "Every student_id must be a valid UUID.")
		}
	}

	type targetCharge struct {
		ChargeID  string
		StudentID string
		Amount    Money
	}
	studentFilter := make(map[string]struct{}, len(in.StudentIDs))
	for _, sid := range in.StudentIDs {
		studentFilter[sid] = struct{}{}
	}

	rows, err := tx.Query(ctx, `
		SELECT fc.id, fc.student_id, fc.amount
		FROM fee_charges fc
		JOIN fee_structures fs ON fs.school_id = fc.school_id AND fs.id = fc.fee_structure_id
		WHERE fc.school_id = $1
		  AND fc.fee_period_id = $2
		  AND (NULLIF($3, '')::uuid IS NULL OR fc.fee_structure_id = NULLIF($3, '')::uuid)
		  AND (NULLIF($4, '')::uuid IS NULL OR fs.grade_level_id = NULLIF($4, '')::uuid)
		ORDER BY fc.student_id ASC, fc.id ASC
		FOR UPDATE OF fc`,
		schoolID, feePeriodID, in.FeeStructureID, in.GradeLevelID,
	)
	if err != nil {
		return FeeChangeNoteView{}, err
	}
	var targets []targetCharge
	for rows.Next() {
		var tc targetCharge
		if err := rows.Scan(&tc.ChargeID, &tc.StudentID, &tc.Amount); err != nil {
			rows.Close()
			return FeeChangeNoteView{}, err
		}
		if len(studentFilter) > 0 {
			if _, ok := studentFilter[tc.StudentID]; !ok {
				continue
			}
		}
		targets = append(targets, tc)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return FeeChangeNoteView{}, err
	}
	if len(targets) == 0 {
		return FeeChangeNoteView{}, validationFail("fee_period_id", "NO_AFFECTED_CHARGES", "No student fee charges matched the adjustment criteria.")
	}

	deltaNum, err := delta.NumericValue()
	if err != nil {
		return FeeChangeNoteView{}, err
	}

	var noteID string
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO fee_change_notes (
			school_id, term_id, fee_period_id, change_type, amount_delta, note, created_by, affected_students_count
		)
		VALUES ($1, NULLIF($2, '')::uuid, $3, 'amount', $4, $5, $6, $7)
		RETURNING id, created_at`,
		schoolID, termID, feePeriodID, deltaNum, note, actorUserID, len(targets),
	).Scan(&noteID, &createdAt)
	if err != nil {
		return FeeChangeNoteView{}, err
	}

	affectedStudents := make([]string, 0, len(targets))
	for _, tc := range targets {
		newChargeAmt, err := addMoney(tc.Amount, delta)
		if err != nil || newChargeAmt < 0 {
			return FeeChangeNoteView{}, validationFail("amount_delta", "NEGATIVE_CHARGE_AMOUNT", "Fee adjustment cannot reduce a student's term charge below zero.")
		}
		newChargeNum, err := newChargeAmt.NumericValue()
		if err != nil {
			return FeeChangeNoteView{}, err
		}

		// Lock student and student ledgers for balance update
		var lockedStudent string
		if err := tx.QueryRow(ctx, `SELECT id FROM students WHERE school_id = $1 AND id = $2 FOR UPDATE`, schoolID, tc.StudentID).Scan(&lockedStudent); err != nil {
			return FeeChangeNoteView{}, err
		}
		if _, err := tx.Exec(ctx, `SELECT id FROM fee_ledgers WHERE school_id = $1 AND student_id = $2 ORDER BY id ASC FOR UPDATE`, schoolID, tc.StudentID); err != nil {
			return FeeChangeNoteView{}, err
		}

		var creditAppliedDelta Money
		if delta > 0 {
			_, _, _, _, creditBefore, err := computeStudentTotals(ctx, tx, schoolID, tc.StudentID)
			if err != nil {
				return FeeChangeNoteView{}, err
			}
			if creditBefore > 0 {
				creditAppliedDelta = creditBefore
				if delta < creditAppliedDelta {
					creditAppliedDelta = delta
				}
			}
		}

		if _, err := tx.Exec(ctx, `
			UPDATE fee_charges
			SET amount = $1
			WHERE school_id = $2 AND id = $3`,
			newChargeNum, schoolID, tc.ChargeID,
		); err != nil {
			return FeeChangeNoteView{}, err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
			VALUES ($1, $2, $3, 'adjustment', $4, $5, $6)`,
			schoolID, tc.StudentID, termID, deltaNum, "ADJ:"+noteID, note,
		); err != nil {
			return FeeChangeNoteView{}, err
		}

		if creditAppliedDelta > 0 {
			capNum, err := creditAppliedDelta.NumericValue()
			if err != nil {
				return FeeChangeNoteView{}, err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO ledger_entries (school_id, student_id, term_id, entry_type, amount, reference, description)
				VALUES ($1, $2, $3, 'credit_applied', $4, $5, $6)`,
				schoolID, tc.StudentID, termID, capNum, "CRD-ADJ:"+noteID, "Credit applied to mid-year fee adjustment",
			); err != nil {
				return FeeChangeNoteView{}, err
			}
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO fee_change_note_students (school_id, note_id, student_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING`,
			schoolID, noteID, tc.StudentID,
		); err != nil {
			return FeeChangeNoteView{}, err
		}

		if err := syncStudentLedgerRow(ctx, tx, schoolID, tc.StudentID, yearID, termID, delta, 0, creditAppliedDelta, nil); err != nil {
			return FeeChangeNoteView{}, err
		}
		affectedStudents = append(affectedStudents, tc.StudentID)
	}

	if err := s.notifyAffectedGuardians(
		ctx, tx, schoolID, affectedStudents,
		"fee_change",
		"Fee Adjustment Notice",
		fmt.Sprintf("A fee adjustment of KES %s was recorded: %s", delta.String(), note),
		map[string]any{
			"fee_change_note_id": noteID,
			"fee_period_id":      feePeriodID,
			"term_id":            termID,
			"change_type":        "amount",
			"amount_delta":       delta.String(),
		},
	); err != nil {
		return FeeChangeNoteView{}, err
	}

	if s.Audit != nil {
		if err := s.Audit.Log(ctx, tx, contracts.AuditEvent{
			SchoolID:     schoolID,
			UserID:       actorUserID,
			Action:       "billing.fee_change.apply",
			ResourceType: "fee_change_note",
			ResourceID:   noteID,
			Details: map[string]any{
				"fee_period_id":           feePeriodID,
				"term_id":                 termID,
				"amount_delta":            delta.String(),
				"affected_students_count": len(affectedStudents),
				"note":                    note,
			},
		}); err != nil {
			return FeeChangeNoteView{}, err
		}
	}

	var termIDPtr *string
	if termID != "" {
		termIDPtr = &termID
	}
	return FeeChangeNoteView{
		ID:                    noteID,
		TermID:                termIDPtr,
		FeePeriodID:           &feePeriodID,
		ChangeType:            "amount",
		AmountDelta:           delta,
		Note:                  note,
		AffectedStudentsCount: len(affectedStudents),
		CreatedAt:             createdAt.UTC(),
	}, nil
}

// OnTermStartDateChanged implements TermDateFeeSyncer for Ema's academics calendar handler.
// It updates fee_periods.due_date, fee_charges.due_date, and fee_ledgers.due_date to
// newStartDate, records a fee_change_notes entry, writes an audit log entry, and sends
// deduplicated in-app notifications only to guardians of affected students.
func (s *Service) OnTermStartDateChanged(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, termID string,
	newStartDate time.Time,
	actorUserID, note string,
) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(termID) || !contracts.ValidID(actorUserID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_TERM_DATE_SYNC", "Valid school_id, term_id, and actorUserID are required.")
	}
	trimmedNote := strings.TrimSpace(note)
	if trimmedNote == "" {
		return validationFail("note", "FEE_CHANGE_NOTE_REQUIRED", "An explanation note is required when changing a term start date.")
	}
	if newStartDate.IsZero() {
		return validationFail("new_start_date", "INVALID_DUE_DATE", "A valid new term start date is required.")
	}
	newDateStr := newStartDate.Format("2006-01-02")

	var feePeriodID string
	var oldDueDate time.Time
	err := tx.QueryRow(ctx, `
		SELECT id, due_date
		FROM fee_periods
		WHERE school_id = $1 AND term_id = $2
		FOR UPDATE`,
		schoolID, termID,
	).Scan(&feePeriodID, &oldDueDate)
	if errors.Is(err, pgx.ErrNoRows) {
		var yearID string
		var termNumber int
		err := tx.QueryRow(ctx, `
			SELECT academic_year_id, term_number, start_date
			FROM terms
			WHERE school_id = $1 AND id = $2`,
			schoolID, termID,
		).Scan(&yearID, &termNumber, &oldDueDate)
		if errors.Is(err, pgx.ErrNoRows) {
			return validationFail("term_id", "INVALID_TERM", "Term does not exist in this school.")
		}
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `
			INSERT INTO fee_periods (school_id, academic_year_id, term_id, period_type, name, due_date)
			VALUES ($1, $2, $3, 'term', $4, $5::date)
			RETURNING id`,
			schoolID, yearID, termID, fmt.Sprintf("Term %d", termNumber), newDateStr,
		).Scan(&feePeriodID); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE fee_periods
			SET due_date = $1::date
			WHERE school_id = $2 AND id = $3`,
			newDateStr, schoolID, feePeriodID,
		); err != nil {
			return err
		}
	}

	oldDateStr := oldDueDate.Format("2006-01-02")

	rows, err := tx.Query(ctx, `
		UPDATE fee_charges
		SET due_date = $1::date
		WHERE school_id = $2 AND fee_period_id = $3
		RETURNING student_id`,
		newDateStr, schoolID, feePeriodID,
	)
	if err != nil {
		return err
	}
	studentSet := make(map[string]struct{})
	var affectedStudents []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return err
		}
		if _, seen := studentSet[sid]; !seen {
			studentSet[sid] = struct{}{}
			affectedStudents = append(affectedStudents, sid)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	sort.Strings(affectedStudents)

	if _, err := tx.Exec(ctx, `
		UPDATE fee_ledgers
		SET due_date = $1::date, updated_at = NOW()
		WHERE school_id = $2 AND term_id = $3`,
		newDateStr, schoolID, termID,
	); err != nil {
		return err
	}

	var noteID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO fee_change_notes (
			school_id, term_id, fee_period_id, change_type, amount_delta,
			old_due_date, new_due_date, note, created_by, affected_students_count
		)
		VALUES ($1, $2, $3, 'due_date', 0.00, $4::date, $5::date, $6, $7, $8)
		RETURNING id`,
		schoolID, termID, feePeriodID, oldDateStr, newDateStr, trimmedNote, actorUserID, len(affectedStudents),
	).Scan(&noteID); err != nil {
		return err
	}

	for _, sid := range affectedStudents {
		if _, err := tx.Exec(ctx, `
			INSERT INTO fee_change_note_students (school_id, note_id, student_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING`,
			schoolID, noteID, sid,
		); err != nil {
			return err
		}
	}

	if err := s.notifyAffectedGuardians(
		ctx, tx, schoolID, affectedStudents,
		"fee_due_date_change",
		"Fee Due Date Updated",
		fmt.Sprintf("The fee due date has been updated from %s to %s: %s", oldDateStr, newDateStr, trimmedNote),
		map[string]any{
			"fee_change_note_id": noteID,
			"fee_period_id":      feePeriodID,
			"term_id":            termID,
			"change_type":        "due_date",
			"old_due_date":       oldDateStr,
			"new_due_date":       newDateStr,
		},
	); err != nil {
		return err
	}

	if s != nil && s.Audit != nil {
		if err := s.Audit.Log(ctx, tx, contracts.AuditEvent{
			SchoolID:     schoolID,
			UserID:       actorUserID,
			Action:       "billing.term_due_date.sync",
			ResourceType: "fee_change_note",
			ResourceID:   noteID,
			Details: map[string]any{
				"term_id":                 termID,
				"fee_period_id":           feePeriodID,
				"old_due_date":            oldDateStr,
				"new_due_date":            newDateStr,
				"affected_students_count": len(affectedStudents),
				"note":                    trimmedNote,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) notifyAffectedGuardians(
	ctx context.Context,
	tx pgx.Tx,
	schoolID string,
	studentIDs []string,
	category, title, body string,
	metadata map[string]any,
) error {
	if s == nil || s.Notifier == nil || len(studentIDs) == 0 {
		return nil
	}
	rows, err := tx.Query(ctx, `
		SELECT DISTINCT sg.user_id
		FROM student_guardians sg
		JOIN memberships m ON m.school_id = sg.school_id AND m.user_id = sg.user_id AND m.status = 'active'
		JOIN users u ON u.id = sg.user_id AND u.is_active
		WHERE sg.school_id = $1 AND sg.student_id = ANY($2::uuid[])
		ORDER BY sg.user_id ASC`,
		schoolID, studentIDs,
	)
	if err != nil {
		return err
	}
	var guardianUserIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			rows.Close()
			return err
		}
		guardianUserIDs = append(guardianUserIDs, uid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, uid := range guardianUserIDs {
		if err := s.Notifier.CreateInAppNotification(ctx, tx, schoolID, CreateNotification{
			RecipientUserID: uid,
			Category:        category,
			Title:           title,
			Body:            body,
			Metadata:        metadata,
		}); err != nil {
			return err
		}
	}
	return nil
}

func registerFeeChangeRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	mux.HandleFunc("POST /api/v1/fee-changes", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input ApplyFeeChangeInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		result, err := service.ApplyFeeChange(r.Context(), p, input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusCreated, result, nil)
	})
}
