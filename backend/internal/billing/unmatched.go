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

type UnmatchedPayment struct {
	ID             string     `json:"id"`
	PaymentID      string     `json:"payment_id"`
	TransactionRef string     `json:"transaction_ref"`
	Amount         Money      `json:"amount"`
	PayerPhone     *string    `json:"payer_phone"`
	PayerName      *string    `json:"payer_name"`
	RawAccountRef  *string    `json:"raw_account_ref"`
	Status         string     `json:"status"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type MatchPaymentInput struct {
	StudentID      string `json:"student_id"`
	ResolutionNote string `json:"resolution_note"`
}

type MatchedPaymentResult struct {
	PaymentID      string    `json:"payment_id"`
	UnmatchedID    string    `json:"unmatched_id"`
	StudentID      string    `json:"student_id"`
	TransactionRef string    `json:"transaction_ref"`
	Amount         Money     `json:"amount"`
	Status         string    `json:"status"`
	ResolvedAt     time.Time `json:"resolved_at"`
}

// ListUnmatchedPayments returns paginated unresolved payments for the authenticated Ikonex team member's school.
func (s *Service) ListUnmatchedPayments(ctx context.Context, p contracts.Principal, page *contracts.Page) ([]UnmatchedPayment, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return nil, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !isIkonexTeamRole(p.Role) {
		return nil, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Only Ikonex administrators may view unmatched payments.")
	}
	if s == nil || s.DB == nil {
		return nil, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}

	items := []UnmatchedPayment{}
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT count(*) FROM unmatched_payments
			WHERE school_id = $1 AND status = 'unmatched'`,
			p.SchoolID,
		).Scan(&page.Total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, `
			SELECT id, payment_id, transaction_ref, amount, payer_phone, payer_name, raw_account_ref, status, resolved_at, created_at
			FROM unmatched_payments
			WHERE school_id = $1 AND status = 'unmatched'
			ORDER BY created_at DESC, id DESC
			LIMIT $2 OFFSET $3`,
			p.SchoolID, page.PerPage, (page.Page-1)*page.PerPage,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var u UnmatchedPayment
			if err := rows.Scan(&u.ID, &u.PaymentID, &u.TransactionRef, &u.Amount, &u.PayerPhone, &u.PayerName, &u.RawAccountRef, &u.Status, &u.ResolvedAt, &u.CreatedAt); err != nil {
				return err
			}
			u.CreatedAt = u.CreatedAt.UTC()
			if u.ResolvedAt != nil {
				rt := u.ResolvedAt.UTC()
				u.ResolvedAt = &rt
			}
			items = append(items, u)
		}
		return rows.Err()
	})
	return items, err
}

// MatchUnmatchedPayment manually assigns an unmatched payment (by payment_id or unmatched_payments.id)
// to a valid student in the school, credits the student's ledger, marks the unmatched record resolved,
// notifies linked guardians, and writes an audit log entry.
func (s *Service) MatchUnmatchedPayment(ctx context.Context, p contracts.Principal, id string, in MatchPaymentInput) (MatchedPaymentResult, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return MatchedPaymentResult{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !isIkonexTeamRole(p.Role) {
		return MatchedPaymentResult{}, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Only Ikonex administrators may resolve unmatched payments.")
	}
	if !contracts.ValidID(id) {
		return MatchedPaymentResult{}, contracts.Fail(http.StatusBadRequest, "INVALID_PAYMENT_ID", "A valid payment ID is required.")
	}
	if !contracts.ValidID(in.StudentID) {
		return MatchedPaymentResult{}, validationFail("student_id", "INVALID_STUDENT_ID", "A valid student_id is required.")
	}
	if s == nil || s.DB == nil {
		return MatchedPaymentResult{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}

	var out MatchedPaymentResult
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		var unmatchedID, paymentID, receipt, status string
		var amount Money
		err := tx.QueryRow(ctx, `
			SELECT id, payment_id, transaction_ref, amount, status
			FROM unmatched_payments
			WHERE school_id = $1 AND (payment_id = $2 OR id = $2)
			FOR UPDATE`,
			p.SchoolID, id,
		).Scan(&unmatchedID, &paymentID, &receipt, &amount, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return contracts.Fail(http.StatusNotFound, "UNMATCHED_PAYMENT_NOT_FOUND", "Unmatched payment not found.")
		}
		if err != nil {
			return err
		}
		if status != "unmatched" {
			return contracts.Fail(http.StatusConflict, "PAYMENT_ALREADY_RESOLVED", "This payment has already been matched and resolved.")
		}

		var studentExists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM students WHERE school_id = $1 AND id = $2)`,
			p.SchoolID, in.StudentID,
		).Scan(&studentExists); err != nil {
			return err
		}
		if !studentExists {
			return validationFail("student_id", "STUDENT_NOT_FOUND", "Target student does not exist in this school.")
		}

		note := strings.TrimSpace(in.ResolutionNote)
		var resolvedAt time.Time
		if err := tx.QueryRow(ctx, `
			UPDATE unmatched_payments
			SET status = 'resolved',
			    resolved_student_id = $1,
			    resolved_by = $2,
			    resolution_note = NULLIF($3, ''),
			    resolved_at = NOW()
			WHERE school_id = $4 AND id = $5 AND status = 'unmatched'
			RETURNING resolved_at`,
			in.StudentID, p.UserID, note, p.SchoolID, unmatchedID,
		).Scan(&resolvedAt); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE payments
			SET student_id = $1,
			    matched_status = 'matched',
			    matched_by = $2
			WHERE school_id = $3 AND id = $4`,
			in.StudentID, p.UserID, p.SchoolID, paymentID,
		); err != nil {
			return err
		}

		if err := s.RecordStudentPaymentTx(ctx, tx, p.SchoolID, in.StudentID, receipt, "Manually matched payment ("+receipt+")", amount); err != nil {
			return err
		}

		if err := s.notifyAffectedGuardians(
			ctx, tx, p.SchoolID, []string{in.StudentID},
			"payment_receipt",
			"M-Pesa Payment Matched",
			fmt.Sprintf("Payment of KES %s (Receipt %s) has been matched and credited to the student fee ledger.", amount.String(), receipt),
			map[string]any{
				"payment_id":      paymentID,
				"student_id":      in.StudentID,
				"transaction_ref": receipt,
				"amount":          amount.String(),
			},
		); err != nil {
			return err
		}

		if s.Audit != nil {
			if err := s.Audit.Log(ctx, tx, contracts.AuditEvent{
				SchoolID:     p.SchoolID,
				UserID:       p.UserID,
				Action:       "billing.payment.manual_match",
				ResourceType: "payment",
				ResourceID:   paymentID,
				Details: map[string]any{
					"unmatched_id":    unmatchedID,
					"student_id":      in.StudentID,
					"transaction_ref": receipt,
					"amount":          amount.String(),
					"resolution_note": note,
				},
			}); err != nil {
				return err
			}
		}

		out = MatchedPaymentResult{
			PaymentID:      paymentID,
			UnmatchedID:    unmatchedID,
			StudentID:      in.StudentID,
			TransactionRef: receipt,
			Amount:         amount,
			Status:         "resolved",
			ResolvedAt:     resolvedAt.UTC(),
		}
		return nil
	})
	return out, err
}

func (s *Service) isCallbackAuthorized(r *http.Request) bool {
	if s == nil {
		return true
	}
	if s.VerifyCallback != nil {
		return s.VerifyCallback(r)
	}
	if token := strings.TrimSpace(s.CallbackAuthToken); token != "" {
		headerToken := strings.TrimSpace(r.Header.Get("X-Callback-Token"))
		queryToken := strings.TrimSpace(r.URL.Query().Get("token"))
		return headerToken == token || queryToken == token
	}
	return true
}

func registerMpesaRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	mux.HandleFunc("POST /api/v1/payments/stk-push", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input InitiateSTKPushInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		res, err := service.InitiateSTKPush(r.Context(), p, input, false)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusAccepted, res, nil)
	})
	mux.HandleFunc("POST /api/v1/billing/mpesa/callback", func(w http.ResponseWriter, r *http.Request) {
		if !service.isCallbackAuthorized(r) {
			contracts.WriteError(w, contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_CALLBACK", "Invalid or missing M-Pesa callback credentials."))
			return
		}
		var payload DarajaSTKCallbackEnvelope
		if err := contracts.Decode(w, r, &payload); err != nil {
			contracts.WriteError(w, err)
			return
		}
		schoolHint := r.URL.Query().Get("school_id")
		if err := service.ProcessSTKCallback(r.Context(), schoolHint, payload); err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, map[string]any{"ResultCode": 0, "ResultDesc": "Accepted"}, nil)
	})
	mux.HandleFunc("POST /api/v1/billing/mpesa/c2b-confirmation", func(w http.ResponseWriter, r *http.Request) {
		if !service.isCallbackAuthorized(r) {
			contracts.WriteError(w, contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_CALLBACK", "Invalid or missing M-Pesa callback credentials."))
			return
		}
		var payload DarajaC2BConfirmation
		if err := contracts.Decode(w, r, &payload); err != nil {
			contracts.WriteError(w, err)
			return
		}
		schoolHint := r.URL.Query().Get("school_id")
		res, err := service.ProcessC2BConfirmation(r.Context(), schoolHint, payload)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})
	mux.HandleFunc("GET /api/v1/payments/unmatched", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		page, err := contracts.Pagination(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		items, err := service.ListUnmatchedPayments(r.Context(), p, &page)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, items, &page)
	})
	mux.HandleFunc("POST /api/v1/payments/{id}/match", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input MatchPaymentInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		res, err := service.MatchUnmatchedPayment(r.Context(), p, r.PathValue("id"), input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})
}
