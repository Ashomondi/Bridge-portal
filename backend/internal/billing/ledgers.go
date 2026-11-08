package billing

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

// GetStudentFeeBalance returns the student's balance sheet, enforcing guardian ownership (C10, AC-24).
func (s *Service) GetStudentFeeBalance(ctx context.Context, schoolID, studentID, callerUserID, callerRole string) (*StudentBalance, error) {
	var balance *StudentBalance
	err := s.withTransaction(ctx, schoolID, func(store Store) error {
		var err error
		balance, err = getStudentFeeBalance(ctx, store, schoolID, studentID, callerUserID, callerRole)
		return err
	})
	return balance, err
}

func getStudentFeeBalance(ctx context.Context, store Store, schoolID, studentID, callerUserID, callerRole string) (*StudentBalance, error) {
	if schoolID == "" || studentID == "" {
		return nil, errValidation("school_id and student_id are required")
	}

	// Permission check (C10): Linked guardian OR authorized school/ikonex staff
	if callerRole == RoleParent {
		isGuardian, err := store.IsGuardianOf(ctx, schoolID, studentID, callerUserID)
		if err != nil {
			return nil, errInternal(err)
		}
		if !isGuardian {
			return nil, errForbidden("parent is not a registered guardian of this student")
		}
	} else if callerRole != RoleIkonexAdmin {
		return nil, errForbidden("caller is not authorized to view student fee balance")
	}

	_, adminNum, fullName, err := store.GetStudent(ctx, schoolID, studentID)
	if err != nil {
		return nil, errNotFound("student not found in this school")
	}

	ledger, err := store.GetLatestFeeLedger(ctx, schoolID, studentID)
	if errors.Is(err, ErrNotFound) {
		credit, err := prepaidCredit(ctx, store, schoolID, studentID)
		if err != nil {
			return nil, errInternal(err)
		}
		return &StudentBalance{
			StudentID:       studentID,
			AdmissionNumber: adminNum,
			FullName:        fullName,
			OpeningBalance:  0,
			BilledAmount:    0,
			PaidAmount:      credit,
			CurrentBalance:  0,
			CreditBalance:   credit,
			DueDate:         time.Time{},
			Items:           []FeeStructureItem{},
			Notes:           []FeeChangeNote{},
		}, nil
	}
	if err != nil {
		return nil, errInternal(err)
	}

	// Fetch fee structure items for current term
	var items []FeeStructureItem
	charges, err := store.ListFeeChargesForStudent(ctx, schoolID, studentID)
	if err != nil {
		return nil, errInternal(err)
	}
	if len(charges) > 0 {
		latestCharge := charges[len(charges)-1]
		items, err = store.ListFeeStructureItems(ctx, latestCharge.FeeStructureID)
		if err != nil {
			return nil, errInternal(err)
		}
	}

	// Fetch fee change notes for school/term
	notes, err := store.ListFeeChangeNotes(ctx, schoolID, ledger.TermID)
	if err != nil {
		return nil, errInternal(err)
	}

	return &StudentBalance{
		StudentID:       studentID,
		AdmissionNumber: adminNum,
		FullName:        fullName,
		OpeningBalance:  ledger.OpeningBalance,
		BilledAmount:    ledger.BilledAmount,
		PaidAmount:      ledger.PaidAmount,
		CurrentBalance:  ledger.CurrentBalance,
		CreditBalance:   ledger.CreditBalance,
		DueDate:         ledger.DueDate,
		Items:           items,
		Notes:           notes,
	}, nil
}

func prepaidCredit(ctx context.Context, store Store, schoolID, studentID string) (float64, error) {
	entries, err := store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
	if err != nil {
		return 0, err
	}
	var cents int64
	for _, entry := range entries {
		if entry.EntryType == "payment" {
			amount, err := paymentAmountCents(entry.Amount)
			if err != nil {
				return 0, err
			}
			cents += amount
		}
	}
	return float64(cents) / 100, nil
}

// GetStudentLedger returns double-entry ledger entries for student (C10, AC-24).
func (s *Service) GetStudentLedger(ctx context.Context, schoolID, studentID, callerUserID, callerRole string) ([]LedgerEntry, error) {
	var entries []LedgerEntry
	err := s.withTransaction(ctx, schoolID, func(store Store) error {
		var err error
		entries, err = getStudentLedger(ctx, store, schoolID, studentID, callerUserID, callerRole)
		return err
	})
	return entries, err
}

func getStudentLedger(ctx context.Context, store Store, schoolID, studentID, callerUserID, callerRole string) ([]LedgerEntry, error) {
	if schoolID == "" || studentID == "" {
		return nil, errValidation("school_id and student_id are required")
	}

	if callerRole == RoleParent {
		isGuardian, err := store.IsGuardianOf(ctx, schoolID, studentID, callerUserID)
		if err != nil {
			return nil, errInternal(err)
		}
		if !isGuardian {
			return nil, errForbidden("parent is not a registered guardian of this student")
		}
	} else if callerRole != RoleIkonexAdmin {
		return nil, errForbidden("caller is not authorized to view student ledger")
	}

	_, _, _, err := store.GetStudent(ctx, schoolID, studentID)
	if err != nil {
		return nil, errNotFound("student not found in this school")
	}

	return store.ListLedgerEntriesForStudent(ctx, schoolID, studentID)
}

// HTTP Handlers

func (s *Service) handleStudentFeeBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok {
		respondError(w, &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: "authentication required"})
		return
	}

	studentID := extractPathID(r.URL.Path, "/api/v1/students/", "/fee-balance")
	if studentID == "" {
		respondError(w, errValidation("student ID required in path"))
		return
	}

	balance, err := s.GetStudentFeeBalance(r.Context(), id.SchoolID, studentID, id.UserID, id.Role)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, balance)
}

func (s *Service) handleStudentLedger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok {
		respondError(w, &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: "authentication required"})
		return
	}

	studentID := extractPathID(r.URL.Path, "/api/v1/students/", "/ledger")
	if studentID == "" {
		respondError(w, errValidation("student ID required in path"))
		return
	}

	entries, err := s.GetStudentLedger(r.Context(), id.SchoolID, studentID, id.UserID, id.Role)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, entries)
}

func extractPathID(path, prefix, suffix string) string {
	clean := strings.TrimPrefix(path, prefix)
	clean = strings.TrimSuffix(clean, suffix)
	clean = strings.Trim(clean, "/")
	return clean
}
