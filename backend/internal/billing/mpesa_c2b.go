package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type C2BConfirmationPayload struct {
	TransactionType   string `json:"TransactionType"`
	TransID           string `json:"TransID"`
	TransTime         string `json:"TransTime"`
	TransAmount       string `json:"TransAmount"`
	BusinessShortCode string `json:"BusinessShortCode"`
	BillRefNumber     string `json:"BillRefNumber"`
	MSISDN            string `json:"MSISDN"`
	FirstName         string `json:"FirstName"`
}

// HandleC2BConfirmation handles Safaricom M-Pesa Direct Paybill (C2B) payments (C24, C50, AC-24, AC-25).
func (s *Service) HandleC2BConfirmation(ctx context.Context, payload C2BConfirmationPayload) (*Payment, error) {
	if payload.TransID == "" || payload.BusinessShortCode == "" {
		return nil, errValidation("TransID and BusinessShortCode are required")
	}

	amount, err := strconv.ParseFloat(payload.TransAmount, 64)
	if err != nil {
		return nil, errValidation("invalid transaction amount")
	}
	if _, err := paymentAmountCents(amount); err != nil {
		return nil, errValidation(err.Error())
	}

	// 1. Resolve school_id strictly from BusinessShortCode; fail closed if unknown (AC-25)
	cfg, err := s.store.GetSchoolMpesaConfigByShortcode(ctx, strings.TrimSpace(payload.BusinessShortCode))
	if err != nil || cfg == nil || cfg.SchoolID == "" {
		return nil, errValidation("unrecognized BusinessShortCode: no school configured for shortcode")
	}
	var payment *Payment
	applied := false
	err = s.withTransaction(ctx, cfg.SchoolID, func(store Store) error {
		var err error
		payment, applied, err = recordC2BConfirmation(ctx, store, cfg.SchoolID, amount, payload)
		if err != nil {
			return err
		}
		if s.auditLogger != nil {
			action := "payment.c2b.unmatched_queued"
			if payment.MatchedStatus == "matched" {
				action = "payment.c2b.auto_matched"
			}
			return s.auditLogger.LogAction(ctx, cfg.SchoolID, "", action, "payment", payment.ID, map[string]interface{}{
				"transaction_ref": payment.TransactionRef, "amount": amount, "raw_bill_ref": payload.BillRefNumber,
			})
		}
		return nil
	})
	if err != nil {
		return nil, errInternal(err)
	}
	if applied && payment.StudentID != nil {
		s.notifyPaymentReceipt(ctx, cfg.SchoolID, *payment.StudentID, payment.TransactionRef, payment.Amount)
	}
	return payment, nil
}

func recordC2BConfirmation(ctx context.Context, store Store, schoolID string, amount float64, payload C2BConfirmationPayload) (*Payment, bool, error) {
	existing, err := store.GetPaymentByTransactionRef(ctx, payload.TransID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, false, err
	}
	if existing != nil {
		if existing.SchoolID != schoolID || existing.Amount != amount {
			return nil, false, errForbidden("C2B receipt does not match the recorded payment")
		}
		return existing, false, nil
	}
	rawBillRef := strings.TrimSpace(payload.BillRefNumber)
	studentID, _, err := store.GetStudentByAdmissionNumber(ctx, schoolID, rawBillRef)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, false, err
	}
	if studentID != "" {
		applied, err := creditStudentPayment(ctx, store, schoolID, studentID, "", payload.TransID, "mpesa_c2b", amount, payload.MSISDN, payload.FirstName)
		if err != nil {
			return nil, false, err
		}
		payment, err := store.GetPaymentByTransactionRef(ctx, payload.TransID)
		return payment, applied, err
	}
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, false, err
	}
	payment := &Payment{
		SchoolID: schoolID, TransactionRef: payload.TransID, Provider: "mpesa_c2b", Amount: amount,
		PayerPhone: payload.MSISDN, PayerName: payload.FirstName, MatchedStatus: "unmatched",
		PaymentDate: time.Now(), RawPayload: string(rawJSON),
	}
	if err := store.CreatePayment(ctx, payment); err != nil {
		return nil, false, err
	}
	if err := store.CreateUnmatchedPayment(ctx, &UnmatchedPayment{
		SchoolID: schoolID, PaymentID: payment.ID, TransactionRef: payload.TransID, Amount: amount,
		PayerPhone: payload.MSISDN, PayerName: payload.FirstName, RawAccountRef: rawBillRef, Status: "unmatched", CreatedAt: time.Now(),
	}); err != nil {
		return nil, false, err
	}
	return payment, false, nil
}

func (s *Service) handleC2BConfirmationHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var payload C2BConfirmationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, errValidation("invalid C2B confirmation payload"))
		return
	}

	_, err := s.HandleC2BConfirmation(r.Context(), payload)
	if err != nil {
		respondError(w, err)
		return
	}

	// Safaricom expects standard C2B response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"ResultCode": "0",
		"ResultDesc": "Success",
	})
}
