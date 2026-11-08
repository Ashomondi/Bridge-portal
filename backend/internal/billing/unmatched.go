package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// MatchPaymentRequest payload for POST /api/v1/payments/{id}/match
type MatchPaymentRequest struct {
	StudentID      string `json:"student_id"`
	ResolutionNote string `json:"resolution_note"`
}

// ListUnmatchedPayments returns all unresolved payments for the school (C50, AC-25).
func (s *Service) ListUnmatchedPayments(ctx context.Context, schoolID string) ([]UnmatchedPayment, error) {
	if schoolID == "" {
		return nil, errValidation("school_id is required")
	}
	var items []UnmatchedPayment
	err := s.withTransaction(ctx, schoolID, func(store Store) error {
		var err error
		items, err = store.ListUnmatchedPayments(ctx, schoolID)
		return err
	})
	return items, err
}

// MatchUnmatchedPayment resolves an unmatched payment by assigning it to a student and crediting their ledger (AC-25).
func (s *Service) MatchUnmatchedPayment(ctx context.Context, schoolID, paymentID, studentID, resolutionNote, actorUserID string) (*Payment, error) {
	if schoolID == "" || paymentID == "" || studentID == "" {
		return nil, errValidation("school_id, payment_id, and student_id are required")
	}

	var payment *Payment
	err := s.withTransaction(ctx, schoolID, func(store Store) error {
		if _, _, _, err := store.GetStudent(ctx, schoolID, studentID); err != nil {
			return err
		}
		unmatched, err := store.GetUnmatchedPayment(ctx, schoolID, paymentID)
		if err != nil {
			return err
		}
		if unmatched.Status != "unmatched" {
			return &APIError{Status: http.StatusConflict, Code: "PAYMENT_ALREADY_RESOLVED", Message: "payment has already been resolved"}
		}
		payment, err = store.GetPayment(ctx, schoolID, unmatched.PaymentID)
		if err != nil {
			return err
		}
		applied, err := creditStudentPayment(ctx, store, schoolID, studentID, payment.ID, payment.TransactionRef, payment.Provider, payment.Amount, payment.PayerPhone, payment.PayerName)
		if err != nil {
			return err
		}
		if !applied {
			return &APIError{Status: http.StatusConflict, Code: "PAYMENT_ALREADY_RESOLVED", Message: "payment has already been matched"}
		}
		if err := store.ResolveUnmatchedPayment(ctx, schoolID, unmatched.ID, studentID, actorUserID, resolutionNote); err != nil {
			return err
		}
		if err := store.UpdatePaymentMatched(ctx, schoolID, payment.ID, studentID, actorUserID); err != nil {
			return err
		}
		if s.auditLogger != nil {
			if err := s.auditLogger.LogAction(ctx, schoolID, actorUserID, "payment.manual_matched", "payment", payment.ID, map[string]interface{}{
				"student_id": studentID, "amount": payment.Amount, "transaction_ref": payment.TransactionRef, "resolution_note": resolutionNote,
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
		if errors.Is(err, ErrNotFound) {
			return nil, errNotFound("student or unmatched payment not found in this school")
		}
		return nil, errInternal(err)
	}
	payment.MatchedStatus = "matched"
	payment.StudentID = &studentID
	payment.MatchedBy = &actorUserID
	s.notifyPaymentReceipt(ctx, schoolID, studentID, payment.TransactionRef, payment.Amount)
	return payment, nil
}

func (s *Service) handleListUnmatchedPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok || id.Role != RoleIkonexAdmin {
		respondError(w, errForbidden("only Ikonex admin can view unmatched payments in Release 1"))
		return
	}

	items, err := s.ListUnmatchedPayments(r.Context(), id.SchoolID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (s *Service) handleMatchPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok || id.Role != RoleIkonexAdmin {
		respondError(w, errForbidden("only Ikonex admin can match payments in Release 1"))
		return
	}

	paymentID := extractPathID(r.URL.Path, "/api/v1/payments/", "/match")
	if paymentID == "" {
		respondError(w, errValidation("payment ID required in path"))
		return
	}

	var req MatchPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errValidation("invalid request body"))
		return
	}

	payment, err := s.MatchUnmatchedPayment(r.Context(), id.SchoolID, paymentID, req.StudentID, req.ResolutionNote, id.UserID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, payment)
}
