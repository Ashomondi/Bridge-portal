package billing

import (
	"context"
	"encoding/json"
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
	return s.store.ListUnmatchedPayments(ctx, schoolID)
}

// MatchUnmatchedPayment resolves an unmatched payment by assigning it to a student and crediting their ledger (AC-25).
func (s *Service) MatchUnmatchedPayment(ctx context.Context, schoolID, paymentID, studentID, resolutionNote, actorUserID string) (*Payment, error) {
	if schoolID == "" || paymentID == "" || studentID == "" {
		return nil, errValidation("school_id, payment_id, and student_id are required")
	}

	// 1. Verify student exists in this school
	_, _, _, err := s.store.GetStudent(ctx, schoolID, studentID)
	if err != nil {
		return nil, errNotFound("student not found in this school")
	}

	// 2. Look up unmatched payment and verify it is not already resolved
	unmatched, err := s.store.GetUnmatchedPayment(ctx, schoolID, paymentID)
	if err != nil {
		return nil, errNotFound("unmatched payment record not found")
	}
	if unmatched.Status != "unmatched" {
		return nil, &APIError{
			Status:  http.StatusConflict,
			Code:    "PAYMENT_ALREADY_RESOLVED",
			Message: "unmatched payment has already been resolved",
		}
	}

	// 3. Look up payment and ensure it has not already been matched
	payment, err := s.store.GetPayment(ctx, schoolID, unmatched.PaymentID)
	if err != nil {
		payment, err = s.store.GetPayment(ctx, schoolID, paymentID)
	}
	if err != nil {
		return nil, errNotFound("payment record not found")
	}
	if payment.MatchedStatus == "matched" {
		return nil, &APIError{
			Status:  http.StatusConflict,
			Code:    "PAYMENT_ALREADY_RESOLVED",
			Message: "payment record has already been matched",
		}
	}

	// Match and credit the payment in one transaction before marking the queue item resolved.
	if err := s.creditStudentPayment(ctx, schoolID, studentID, payment.ID, payment.TransactionRef, payment.Provider, payment.Amount, payment.PayerPhone, payment.PayerName); err != nil {
		return nil, errInternal(err)
	}
	if err := s.store.ResolveUnmatchedPayment(ctx, schoolID, paymentID, studentID, actorUserID, resolutionNote); err != nil {
		return nil, errInternal(err)
	}

	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(ctx, schoolID, actorUserID, "payment.manual_matched", "payment", payment.ID, map[string]interface{}{
			"student_id":      studentID,
			"amount":          payment.Amount,
			"transaction_ref": payment.TransactionRef,
			"resolution_note": resolutionNote,
		})
	}

	payment.MatchedStatus = "matched"
	payment.StudentID = &studentID
	payment.MatchedBy = &actorUserID

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
