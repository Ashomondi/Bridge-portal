package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// LoadFeeStructureRequest payload for POST /api/v1/billing/fee-structures
type LoadFeeStructureRequest struct {
	SchoolID       string                  `json:"school_id"`
	AcademicYearID string                  `json:"academic_year_id"`
	GradeLevelID   string                  `json:"grade_level_id"`
	TermID         string                  `json:"term_id"`
	TermStartDate  time.Time               `json:"term_start_date"`
	Items          []FeeStructureItemInput `json:"items"`
}

type FeeStructureItemInput struct {
	Title        string  `json:"title"`
	Amount       float64 `json:"amount"`
	IsCompulsory bool    `json:"is_compulsory"`
}

type feeStructureItemKey struct {
	title        string
	amount       float64
	isCompulsory bool
}

func feeStructureItemsEqual(a, b []FeeStructureItem) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[feeStructureItemKey]int, len(a))
	for _, item := range a {
		counts[feeStructureItemKey{item.Title, item.Amount, item.IsCompulsory}]++
	}
	for _, item := range b {
		key := feeStructureItemKey{item.Title, item.Amount, item.IsCompulsory}
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}

// LoadFeeStructure loads a team fee structure and automatically charges enrolled students for the term (AC-20).
func (s *Service) LoadFeeStructure(ctx context.Context, req LoadFeeStructureRequest) (*FeeStructure, error) {
	if req.SchoolID == "" || req.AcademicYearID == "" || req.GradeLevelID == "" || req.TermID == "" {
		return nil, errValidation("school_id, academic_year_id, grade_level_id, and term_id are required")
	}
	if req.TermStartDate.IsZero() {
		return nil, errValidation("term_start_date is required")
	}
	if len(req.Items) == 0 {
		return nil, errValidation("fee structure items cannot be empty")
	}

	var totalAmount float64
	var items []FeeStructureItem
	for _, it := range req.Items {
		if it.Amount <= 0 {
			return nil, errValidation("fee structure item amount must be greater than zero")
		}
		totalAmount += it.Amount
		items = append(items, FeeStructureItem{
			Title:        it.Title,
			Amount:       it.Amount,
			IsCompulsory: it.IsCompulsory,
		})
	}

	structure := &FeeStructure{
		SchoolID:       req.SchoolID,
		AcademicYearID: req.AcademicYearID,
		GradeLevelID:   req.GradeLevelID,
		TermID:         req.TermID,
		TotalAmount:    totalAmount,
		Items:          items,
	}

	err := s.withTransaction(ctx, req.SchoolID, func(store Store) error {
		period, err := store.GetFeePeriod(ctx, req.SchoolID, req.AcademicYearID, req.TermID)
		if errors.Is(err, ErrNotFound) {
			period = &FeePeriod{
				SchoolID:       req.SchoolID,
				AcademicYearID: req.AcademicYearID,
				TermID:         req.TermID,
				PeriodType:     "term",
				Name:           fmt.Sprintf("Term Billing %s", req.TermID),
				DueDate:        req.TermStartDate,
			}
			if err := store.CreateFeePeriod(ctx, period); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !period.DueDate.Equal(req.TermStartDate) {
			return &APIError{
				Status:  http.StatusConflict,
				Code:    "TERM_START_DATE_MISMATCH",
				Message: "fee period due date differs from term_start_date; update it through the term-date change flow",
			}
		}

		existingStructure, err := store.GetFeeStructure(ctx, req.SchoolID, req.AcademicYearID, req.GradeLevelID, req.TermID)
		if err == nil {
			existingItems, err := store.ListFeeStructureItems(ctx, existingStructure.ID)
			if err != nil {
				return err
			}
			if existingStructure.TotalAmount != totalAmount || !feeStructureItemsEqual(existingItems, items) {
				return &APIError{
					Status:  http.StatusConflict,
					Code:    "FEE_STRUCTURE_ALREADY_LOADED",
					Message: "an existing fee structure cannot be replaced; use the fee-change flow",
				}
			}
			structure.ID = existingStructure.ID
			structure.CreatedAt = existingStructure.CreatedAt
		} else if errors.Is(err, ErrNotFound) {
			if err := store.CreateFeeStructure(ctx, structure, items); err != nil {
				return err
			}
		} else {
			return err
		}

		studentIDs, err := store.GetEnrolledStudentsForGrade(ctx, req.SchoolID, req.GradeLevelID, req.AcademicYearID)
		if err != nil {
			return err
		}
		for _, studentID := range studentIDs {
			if err := chargeStudentForStructure(ctx, store, req.SchoolID, studentID, req.AcademicYearID, req.TermID, structure.ID, period.ID, totalAmount, period.DueDate); err != nil {
				return err
			}
		}

		if s.auditLogger != nil {
			callerID := ""
			if id, ok := IdentityFromContext(ctx); ok {
				callerID = id.UserID
			}
			if err := s.auditLogger.LogAction(ctx, req.SchoolID, callerID, "billing.fee_structure.loaded", "fee_structure", structure.ID, map[string]interface{}{
				"grade_level_id": req.GradeLevelID,
				"term_id":        req.TermID,
				"total_amount":   totalAmount,
				"charged_count":  len(studentIDs),
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, errInternal(err)
	}

	return structure, nil
}

type transactionalStore interface {
	WithTx(ctx context.Context, schoolID string, fn func(Store) error) error
}

func (s *Service) withTransaction(ctx context.Context, schoolID string, fn func(Store) error) error {
	if store, ok := s.store.(transactionalStore); ok {
		return store.WithTx(ctx, schoolID, fn)
	}
	return fn(s.store)
}

// chargeStudentForStructure applies a term charge once and automatically applies existing ledger credit (AC-20, A22).
func (s *Service) chargeStudentForStructure(ctx context.Context, schoolID, studentID, academicYearID, termID, structureID, periodID string, amount float64, dueDate time.Time) error {
	return chargeStudentForStructure(ctx, s.store, schoolID, studentID, academicYearID, termID, structureID, periodID, amount, dueDate)
}

func chargeStudentForStructure(ctx context.Context, store Store, schoolID, studentID, academicYearID, termID, structureID, periodID string, amount float64, dueDate time.Time) error {
	// 1. Record fee charge
	charge := &FeeCharge{
		SchoolID:       schoolID,
		StudentID:      studentID,
		FeePeriodID:    periodID,
		FeeStructureID: structureID,
		Amount:         amount,
		DueDate:        dueDate,
	}
	inserted, err := store.CreateFeeChargeIfAbsent(ctx, charge)
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}

	// 2. Record double-entry charge in ledger_entries
	chargeEntry := &LedgerEntry{
		SchoolID:    schoolID,
		StudentID:   studentID,
		TermID:      termID,
		EntryType:   "charge",
		Amount:      amount,
		Reference:   fmt.Sprintf("term-charge-%s", termID),
		Description: fmt.Sprintf("Term charge for period %s", periodID),
	}
	if err := store.CreateLedgerEntry(ctx, chargeEntry); err != nil {
		return err
	}

	// 3. Fetch latest ledger to check for existing credit_balance (A22)
	latestLedger, err := store.GetLatestFeeLedger(ctx, schoolID, studentID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	creditAvail := 0.0
	openingBal := 0.0
	if err == nil && latestLedger != nil {
		creditAvail = latestLedger.CreditBalance
		// If there is an outstanding current balance from earlier, roll it over as opening balance
		openingBal = latestLedger.CurrentBalance
	} else if errors.Is(err, ErrNotFound) {
		creditAvail, err = prepaidCredit(ctx, store, schoolID, studentID)
		if err != nil {
			return err
		}
	}

	// 4. Create new term ledger
	newLedger := &FeeLedger{
		SchoolID:       schoolID,
		StudentID:      studentID,
		AcademicYearID: academicYearID,
		TermID:         termID,
		OpeningBalance: openingBal,
		BilledAmount:   amount,
		PaidAmount:     0.0,
		CurrentBalance: openingBal + amount,
		CreditBalance:  0.0,
		DueDate:        dueDate,
	}

	// Automatic credit roll-forward application (A22)
	if creditAvail > 0 {
		totalDue := openingBal + amount
		if creditAvail >= totalDue {
			// Credit covers the entire charge and opening balance!
			applied := totalDue
			newLedger.PaidAmount = creditAvail
			newLedger.CurrentBalance = 0.0
			newLedger.CreditBalance = creditAvail - applied

			if err := store.CreateLedgerEntry(ctx, &LedgerEntry{
				SchoolID:    schoolID,
				StudentID:   studentID,
				TermID:      termID,
				EntryType:   "credit_applied",
				Amount:      applied,
				Reference:   "auto-credit-rollover",
				Description: fmt.Sprintf("Applied %.2f from existing credit balance", applied),
			}); err != nil {
				return err
			}
		} else {
			// Credit partially covers the charge
			applied := creditAvail
			newLedger.PaidAmount = applied
			newLedger.CurrentBalance = totalDue - applied
			newLedger.CreditBalance = 0.0

			if err := store.CreateLedgerEntry(ctx, &LedgerEntry{
				SchoolID:    schoolID,
				StudentID:   studentID,
				TermID:      termID,
				EntryType:   "credit_applied",
				Amount:      applied,
				Reference:   "auto-credit-rollover",
				Description: fmt.Sprintf("Applied %.2f from existing credit balance", applied),
			}); err != nil {
				return err
			}
		}
	}

	return store.SaveFeeLedger(ctx, newLedger)
}

// StudentChargeHook implements cross-module hook for Evans's onboarding package (C46, AC-20).
// When a new student is admitted, this hook charges the student for active fee structures.
func (s *Service) StudentChargeHook(ctx context.Context, schoolID, studentID, gradeLevelID, academicYearID string) error {
	// Look up fee structures for this grade and year
	// Release 1 CBC terms: 1, 2, 3
	for termNum := 1; termNum <= 3; termNum++ {
		termID := fmt.Sprintf("term-%d", termNum)
		structure, err := s.store.GetFeeStructure(ctx, schoolID, academicYearID, gradeLevelID, termID)
		if err != nil || structure == nil {
			continue
		}
		period, err := s.store.GetFeePeriod(ctx, schoolID, academicYearID, termID)
		if err != nil {
			return err
		}
		if err := s.chargeStudentForStructure(ctx, schoolID, studentID, academicYearID, termID, structure.ID, period.ID, structure.TotalAmount, period.DueDate); err != nil {
			return err
		}
	}
	return nil
}

// handleLoadFeeStructure HTTP handler for POST /api/v1/billing/fee-structures
func (s *Service) handleLoadFeeStructure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok || id.Role != RoleIkonexAdmin {
		respondError(w, errForbidden("only Ikonex admin can load fee structures in Release 1"))
		return
	}

	var req LoadFeeStructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errValidation("invalid request body"))
		return
	}

	if req.SchoolID == "" {
		req.SchoolID = id.SchoolID
	}

	structure, err := s.LoadFeeStructure(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, structure)
}
