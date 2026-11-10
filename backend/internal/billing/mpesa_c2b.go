package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

type DarajaC2BConfirmation struct {
	SchoolID          string          `json:"school_id"`
	TransactionType   string          `json:"TransactionType"`
	TransID           string          `json:"TransID"`
	TransTime         string          `json:"TransTime"`
	TransAmount       json.RawMessage `json:"TransAmount"`
	BusinessShortCode string          `json:"BusinessShortCode"`
	BillRefNumber     string          `json:"BillRefNumber"`
	InvoiceNumber     string          `json:"InvoiceNumber"`
	OrgAccountBalance string          `json:"OrgAccountBalance"`
	ThirdPartyTransID string          `json:"ThirdPartyTransID"`
	MSISDN            string          `json:"MSISDN"`
	FirstName         string          `json:"FirstName"`
	MiddleName        string          `json:"MiddleName"`
	LastName          string          `json:"LastName"`
}

type C2BProcessingResult struct {
	TransactionRef string  `json:"transaction_ref"`
	MatchedStatus  string  `json:"matched_status"`
	StudentID      *string `json:"student_id,omitempty"`
}

// ProcessC2BConfirmation resolves the receiving school from BusinessShortCode,
// matches UPPER(TRIM(admission_number)) = UPPER(TRIM(BillRefNumber)) strictly
// within that school, and either credits the matched student's ledger or stores
// an unmatched payment record without ever discarding incoming money.
func (s *Service) ProcessC2BConfirmation(ctx context.Context, schoolIDHint string, in DarajaC2BConfirmation) (C2BProcessingResult, error) {
	receipt := strings.TrimSpace(in.TransID)
	shortcode := strings.TrimSpace(in.BusinessShortCode)
	if receipt == "" || shortcode == "" || len(in.TransAmount) == 0 {
		return C2BProcessingResult{}, contracts.Fail(http.StatusBadRequest, "INVALID_C2B_PAYLOAD", "TransID, BusinessShortCode and TransAmount are required.")
	}
	var amount Money
	if err := json.Unmarshal(in.TransAmount, &amount); err != nil || amount <= 0 {
		return C2BProcessingResult{}, contracts.Fail(http.StatusUnprocessableEntity, "INVALID_PAYMENT_AMOUNT", "C2B TransAmount must be a positive decimal within NUMERIC(12,2).")
	}

	schoolID, err := s.resolveWebhookSchoolID(ctx, schoolIDHint, in.SchoolID, shortcode)
	if err != nil {
		return C2BProcessingResult{}, err
	}

	payerPhone := strings.TrimSpace(in.MSISDN)
	if norm, err := notifications.NormalizePhone(payerPhone); err == nil {
		payerPhone = norm
	}
	payerName := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(in.FirstName),
		strings.TrimSpace(in.MiddleName),
		strings.TrimSpace(in.LastName),
	}, " "))
	payerName = strings.Join(strings.Fields(payerName), " ")
	rawBillRef := strings.TrimSpace(in.BillRefNumber)

	var out C2BProcessingResult
	err = s.DB.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		// Verify shortcode belongs to this school in school_mpesa_configs
		var configuredShortcode string
		if err := tx.QueryRow(ctx, `
			SELECT shortcode FROM school_mpesa_configs WHERE school_id = $1`,
			schoolID,
		).Scan(&configuredShortcode); errors.Is(err, pgx.ErrNoRows) || !strings.EqualFold(strings.TrimSpace(configuredShortcode), shortcode) {
			return contracts.Fail(http.StatusBadRequest, "SHORTCODE_SCHOOL_MISMATCH", "BusinessShortCode does not match the school's configured M-Pesa shortcode.")
		} else if err != nil {
			return err
		}

		// Idempotency check if payment receipt already exists
		var existingStatus string
		var existingStudentID *string
		err := tx.QueryRow(ctx, `
			SELECT matched_status, student_id
			FROM payments
			WHERE school_id = $1 AND transaction_ref = $2`,
			schoolID, receipt,
		).Scan(&existingStatus, &existingStudentID)
		if err == nil {
			out = C2BProcessingResult{
				TransactionRef: receipt,
				MatchedStatus:  existingStatus,
				StudentID:      existingStudentID,
			}
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		var matchedStudentID string
		if rawBillRef != "" {
			err := tx.QueryRow(ctx, `
				SELECT id
				FROM students
				WHERE school_id = $1
				  AND status = 'active'
				  AND UPPER(TRIM(admission_number)) = UPPER(TRIM($2))`,
				schoolID, rawBillRef,
			).Scan(&matchedStudentID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		if matchedStudentID != "" {
			if err := s.creditMatchedStudentPaymentTx(ctx, tx, schoolID, matchedStudentID, receipt, "mpesa_c2b", payerPhone, payerName, amount, in); err != nil {
				return err
			}
			out = C2BProcessingResult{
				TransactionRef: receipt,
				MatchedStatus:  "matched",
				StudentID:      &matchedStudentID,
			}
			return nil
		}

		amtNum, err := amount.NumericValue()
		if err != nil {
			return err
		}
		rawBytes, err := json.Marshal(in)
		if err != nil {
			rawBytes = []byte(`{}`)
		}
		var paymentID string
		err = tx.QueryRow(ctx, `
			INSERT INTO payments (school_id, student_id, transaction_ref, provider, amount, payer_phone, payer_name, matched_status, payment_date, raw_payload)
			VALUES ($1, NULL, $2, 'mpesa_c2b', $3, NULLIF($4, ''), NULLIF($5, ''), 'unmatched', NOW(), $6)
			ON CONFLICT (transaction_ref) DO NOTHING
			RETURNING id`,
			schoolID, receipt, amtNum, payerPhone, payerName, rawBytes,
		).Scan(&paymentID)
		if errors.Is(err, pgx.ErrNoRows) {
			out = C2BProcessingResult{
				TransactionRef: receipt,
				MatchedStatus:  "unmatched",
			}
			return nil
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO unmatched_payments (
				school_id, payment_id, transaction_ref, amount, payer_phone, payer_name, raw_account_ref, status
			)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), 'unmatched')
			ON CONFLICT (school_id, payment_id) DO NOTHING`,
			schoolID, paymentID, receipt, amtNum, payerPhone, payerName, rawBillRef,
		); err != nil {
			return err
		}

		out = C2BProcessingResult{
			TransactionRef: receipt,
			MatchedStatus:  "unmatched",
		}
		return nil
	})
	return out, err
}
