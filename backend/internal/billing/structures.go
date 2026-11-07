package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LoadFeeStructureRequest payload for POST /api/v1/billing/fee-structures
type LoadFeeStructureRequest struct {
	SchoolID       string                 `json:"school_id"`
	AcademicYearID string                 `json:"academic_year_id"`
	GradeLevelID   string                 `json:"grade_level_id"`
	TermID         string                 `json:"term_id"`
	TermStartDate  time.Time              `json:"term_start_date"`
	Items          []FeeStructureItemInput `json:"items"`
}

type FeeStructureItemInput struct {
	Title        string  `json:"title"`
	Amount       float64 `json:"amount"`
	IsCompulsory bool    `json:"is_compulsory"`
}

// LoadFeeStructure loads a team fee structure and automatically charges enrolled students for the term (AC-20).
func (s *Service) LoadFeeStructure(ctx context.Context, req LoadFeeStructureRequest) (*FeeStructure, error) {
	if req.SchoolID == "" || req.AcademicYearID == "" || req.GradeLevelID == "" || req.TermID == "" {
		return nil, errValidation("school_id, academic_year_id, grade_level_id, and term_id are required")
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

	dueDate := req.TermStartDate
	if dueDate.IsZero() {
		dueDate = time.Now()
	}

	// 1. Ensure FeePeriod exists for this term (C46, C61)
	period, err := s.store.GetFeePeriod(ctx, req.SchoolID, req.AcademicYearID, req.TermID)
	if err != nil {
		period = &FeePeriod{
			SchoolID:       req.SchoolID,
			AcademicYearID: req.AcademicYearID,
			TermID:         req.TermID,
			PeriodType:     "term",
			Name:           fmt.Sprintf("Term Billing %s", req.TermID),
			DueDate:        dueDate,
		}
		if err := s.store.CreateFeePeriod(ctx, period); err != nil {
			return nil, errInternal(err)
		}
	} else if !dueDate.IsZero() && !period.DueDate.Equal(dueDate) {
		_ = s.store.UpdateFeePeriodDueDate(ctx, req.SchoolID, req.TermID, dueDate)
		period.DueDate = dueDate
	}

	// 2. Persist fee structure and items
	structure := &FeeStructure{
		SchoolID:       req.SchoolID,
		AcademicYearID: req.AcademicYearID,
		GradeLevelID:   req.GradeLevelID,
		TermID:         req.TermID,
		TotalAmount:    totalAmount,
		Items:          items,
	}
	if err := s.store.CreateFeeStructure(ctx, structure, items); err != nil {
		return nil, errInternal(err)
	}

	// 3. Automated Start-of-Year / Term Charging for all enrolled students in grade (AC-20)
	studentIDs, err := s.store.GetEnrolledStudentsForGrade(ctx, req.SchoolID, req.GradeLevelID, req.AcademicYearID)
	if err != nil {
		return nil, errInternal(err)
	}

	for _, studentID := range studentIDs {
		if err := s.chargeStudentForStructure(ctx, req.SchoolID, studentID, req.AcademicYearID, req.TermID, structure.ID, period.ID, totalAmount, period.DueDate); err != nil {
			return nil, errInternal(err)
		}
	}

	if s.auditLogger != nil {
		callerID := ""
		if id, ok := IdentityFromContext(ctx); ok {
			callerID = id.UserID
		}
		_ = s.auditLogger.LogAction(ctx, req.SchoolID, callerID, "billing.fee_structure.loaded", "fee_structure", structure.ID, map[string]interface{}{
			"grade_level_id": req.GradeLevelID,
			"term_id":        req.TermID,
			"total_amount":   totalAmount,
			"charged_count":  len(studentIDs),
		})
	}

	return structure, nil
}

// chargeStudentForStructure applies term charge and automatically applies existing ledger credit (AC-20, A22)
func (s *Service) chargeStudentForStructure(ctx context.Context, schoolID, studentID, academicYearID, termID, structureID, periodID string, amount float64, dueDate time.Time) error {
	// 1. Record fee charge
	charge := &FeeCharge{
		SchoolID:       schoolID,
		StudentID:      studentID,
		FeePeriodID:    periodID,
		FeeStructureID: structureID,
		Amount:         amount,
		DueDate:        dueDate,
	}
	if err := s.store.CreateFeeCharge(ctx, charge); err != nil {
		return err
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
	if err := s.store.CreateLedgerEntry(ctx, chargeEntry); err != nil {
		return err
	}

	// 3. Fetch latest ledger to check for existing credit_balance (A22)
	latestLedger, _ := s.store.GetLatestFeeLedger(ctx, schoolID, studentID)
	creditAvail := 0.0
	openingBal := 0.0
	if latestLedger != nil {
		creditAvail = latestLedger.CreditBalance
		// If there is an outstanding current balance from earlier, roll it over as opening balance
		openingBal = latestLedger.CurrentBalance
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
			newLedger.PaidAmount = applied
			newLedger.CurrentBalance = 0.0
			newLedger.CreditBalance = creditAvail - applied

			_ = s.store.CreateLedgerEntry(ctx, &LedgerEntry{
				SchoolID:    schoolID,
				StudentID:   studentID,
				TermID:      termID,
				EntryType:   "credit_applied",
				Amount:      applied,
				Reference:   "auto-credit-rollover",
				Description: fmt.Sprintf("Applied %.2f from existing credit balance", applied),
			})
		} else {
			// Credit partially covers the charge
			applied := creditAvail
			newLedger.PaidAmount = applied
			newLedger.CurrentBalance = totalDue - applied
			newLedger.CreditBalance = 0.0

			_ = s.store.CreateLedgerEntry(ctx, &LedgerEntry{
				SchoolID:    schoolID,
				StudentID:   studentID,
				TermID:      termID,
				EntryType:   "credit_applied",
				Amount:      applied,
				Reference:   "auto-credit-rollover",
				Description: fmt.Sprintf("Applied %.2f from existing credit balance", applied),
			})
		}
	}

	return s.store.SaveFeeLedger(ctx, newLedger)
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
		dueDate := time.Now()
		periodID := ""
		if period != nil {
			dueDate = period.DueDate
			periodID = period.ID
		}
		if err := s.chargeStudentForStructure(ctx, schoolID, studentID, academicYearID, termID, structure.ID, periodID, structure.TotalAmount, dueDate); err != nil {
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
	if !ok || (id.Role != RoleIkonexAdmin && id.Role != RoleBursar) {
		respondError(w, errForbidden("only Ikonex team or school bursar can load fee structures"))
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
