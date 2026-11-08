package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FeeChangeRequest payload for POST /api/v1/fee-changes
type FeeChangeRequest struct {
	SchoolID     string  `json:"school_id"`
	TermID       string  `json:"term_id"`
	GradeLevelID string  `json:"grade_level_id"`
	AmountDelta  float64 `json:"amount_delta"`
	Note         string  `json:"note"`
}

// ApplyMidYearFeeChange applies a mid-year fee adjustment to all affected students in a grade level (C45, AC-21).
// Mandatory Note Rule: If note is empty, rejects with 422 Unprocessable Entity (FEE_CHANGE_NOTE_REQUIRED).
func (s *Service) ApplyMidYearFeeChange(ctx context.Context, req FeeChangeRequest) (*FeeChangeNote, error) {
	if strings.TrimSpace(req.Note) == "" {
		return nil, errUnprocessable(CodeFeeChangeNoteRequired, "a written explanation note is mandatory for mid-year fee adjustments")
	}
	if req.SchoolID == "" || req.TermID == "" || req.GradeLevelID == "" {
		return nil, errValidation("school_id, term_id, and grade_level_id are required")
	}

	callerID := ""
	if id, ok := IdentityFromContext(ctx); ok {
		callerID = id.UserID
	}

	// 1. Adjust charges for students in this grade
	affectedStudentIDs, err := s.store.AdjustFeeChargesForGrade(ctx, req.SchoolID, req.TermID, req.GradeLevelID, req.AmountDelta)
	if err != nil {
		return nil, errInternal(err)
	}

	// 2. Adjust fee ledgers and record ledger_entries for each affected student
	for _, studentID := range affectedStudentIDs {
		ledger, err := s.store.GetFeeLedger(ctx, req.SchoolID, studentID, "", req.TermID)
		if err != nil || ledger == nil {
			ledger, err = s.store.GetLatestFeeLedger(ctx, req.SchoolID, studentID)
		}
		if err == nil && ledger != nil && (ledger.TermID == "" || ledger.TermID == req.TermID) {
			ledger.BilledAmount += req.AmountDelta
			totalBilled := ledger.OpeningBalance + ledger.BilledAmount
			if ledger.PaidAmount >= totalBilled {
				ledger.CurrentBalance = 0.00
				ledger.CreditBalance = ledger.PaidAmount - totalBilled
			} else {
				ledger.CurrentBalance = totalBilled - ledger.PaidAmount
				ledger.CreditBalance = 0.00
			}
			if err := s.store.SaveFeeLedger(ctx, ledger); err != nil {
				return nil, errInternal(err)
			}
		}

		if err := s.store.CreateLedgerEntry(ctx, &LedgerEntry{
			SchoolID:    req.SchoolID,
			StudentID:   studentID,
			TermID:      req.TermID,
			EntryType:   "adjustment",
			Amount:      req.AmountDelta,
			Reference:   "mid-year-fee-adjustment",
			Description: req.Note,
		}); err != nil {
			return nil, errInternal(err)
		}

		// 3. Notify ONLY guardians of the affected students (AC-21)
		if s.notifier != nil {
			guardians, err := s.store.ListGuardiansForStudent(ctx, req.SchoolID, studentID)
			if err == nil {
				for _, guardianUserID := range guardians {
					_ = s.notifier.CreateInAppNotification(ctx, req.SchoolID, guardianUserID, "fee_change", "Fee Structure Adjustment", fmt.Sprintf("A fee adjustment of KES %.2f has been applied. Reason: %s", req.AmountDelta, req.Note), map[string]interface{}{
						"student_id":   studentID,
						"term_id":      req.TermID,
						"amount_delta": req.AmountDelta,
						"note":         req.Note,
					})
				}
			}
		}
	}

	// 4. Record fee_change_notes entry
	noteRecord := &FeeChangeNote{
		SchoolID:              req.SchoolID,
		TermID:                req.TermID,
		ChangeType:            "mid_year_adjustment",
		AmountDelta:           req.AmountDelta,
		Note:                  req.Note,
		CreatedBy:             callerID,
		AffectedStudentsCount: len(affectedStudentIDs),
	}
	if err := s.store.CreateFeeChangeNote(ctx, noteRecord); err != nil {
		return nil, errInternal(err)
	}

	// 5. Audit log entry (AC-29)
	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(ctx, req.SchoolID, callerID, "billing.fee_change.applied", "fee_change_note", noteRecord.ID, map[string]interface{}{
			"term_id":                 req.TermID,
			"grade_level_id":          req.GradeLevelID,
			"amount_delta":            req.AmountDelta,
			"note":                    req.Note,
			"affected_students_count": len(affectedStudentIDs),
		})
	}

	return noteRecord, nil
}

// OnTermStartDateChanged handles calendar start-date updates from Ema's academics module (C61, AC-22).
// Shifting term start date shifts fee_periods.due_date and fee_charges.due_date to newStartDate,
// records a fee_change_notes entry, and sends in-app notifications to affected guardians.
func (s *Service) OnTermStartDateChanged(ctx context.Context, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error {
	if schoolID == "" || termID == "" || newStartDate.IsZero() {
		return errValidation("school_id, term_id, and newStartDate are required")
	}

	// 1. Fetch current fee period to get old due date
	var oldDate *time.Time
	// Look up through all periods for this school/term
	for _, yr := range []string{"2026", "2027"} {
		p, err := s.store.GetFeePeriod(ctx, schoolID, yr, termID)
		if err == nil && p != nil {
			d := p.DueDate
			oldDate = &d
			break
		}
	}

	// 2. Update fee_periods due_date
	if err := s.store.UpdateFeePeriodDueDate(ctx, schoolID, termID, newStartDate); err != nil {
		return errInternal(err)
	}

	// 3. Update fee_charges due_date
	count, err := s.store.UpdateFeeChargesDueDate(ctx, schoolID, termID, newStartDate)
	if err != nil {
		return errInternal(err)
	}

	// 4. Record fee_change_notes
	explanation := note
	if explanation == "" {
		explanation = fmt.Sprintf("Term start date changed to %s", newStartDate.Format("2006-01-02"))
	}

	noteRecord := &FeeChangeNote{
		SchoolID:              schoolID,
		TermID:                termID,
		ChangeType:            "term_date_shift",
		OldDueDate:            oldDate,
		NewDueDate:            &newStartDate,
		Note:                  explanation,
		CreatedBy:             actorUserID,
		AffectedStudentsCount: count,
	}
	if err := s.store.CreateFeeChangeNote(ctx, noteRecord); err != nil {
		return errInternal(err)
	}

	// 5. Notify affected parents (AC-22) - notify ONLY linked guardians of students with charges in this term, NOT actorUserID!
	if s.notifier != nil {
		affectedStudents, err := s.store.GetStudentsWithChargesForTerm(ctx, schoolID, termID)
		if err == nil {
			notified := make(map[string]bool)
			for _, stID := range affectedStudents {
				guardians, _ := s.store.ListGuardiansForStudent(ctx, schoolID, stID)
				for _, gID := range guardians {
					if !notified[gID] {
						notified[gID] = true
						_ = s.notifier.CreateInAppNotification(ctx, schoolID, gID, "term_date_shift", "Term Fee Due Date Updated", fmt.Sprintf("Term %s fee due date has been shifted to %s. Note: %s", termID, newStartDate.Format("2006-01-02"), explanation), map[string]interface{}{
							"term_id":      termID,
							"new_due_date": newStartDate.Format("2006-01-02"),
							"note":         explanation,
						})
					}
				}
			}
		}
	}

	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(ctx, schoolID, actorUserID, "billing.term_date_shifted", "term", termID, map[string]interface{}{
			"new_start_date": newStartDate.Format("2006-01-02"),
			"note":           explanation,
			"affected_count": count,
		})
	}

	return nil
}

// handleFeeChange HTTP handler for POST /api/v1/fee-changes
func (s *Service) handleFeeChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok || id.Role != RoleIkonexAdmin {
		respondError(w, errForbidden("only Ikonex admin can apply fee changes in Release 1"))
		return
	}

	var req FeeChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errValidation("invalid request body"))
		return
	}

	if req.SchoolID == "" {
		req.SchoolID = id.SchoolID
	}

	noteRecord, err := s.ApplyMidYearFeeChange(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, noteRecord)
}
