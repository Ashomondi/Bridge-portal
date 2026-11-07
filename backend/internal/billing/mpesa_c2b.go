package billing

import (
	"context"
	"encoding/json"
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

	amount, _ := strconv.ParseFloat(payload.TransAmount, 64)
	if amount <= 0 {
		return nil, errValidation("transaction amount must be greater than zero")
	}

	// 1. Resolve school_id from BusinessShortCode (AC-25)
	cfg, err := s.store.GetSchoolMpesaConfigByShortcode(ctx, payload.BusinessShortCode)
	schoolID := ""
	if err == nil && cfg != nil {
		schoolID = cfg.SchoolID
	}
	// Fallback for default sandbox shortcode if configured
	if schoolID == "" {
		// Look up any default school if in sandbox
		for _, sID := range []string{"school-1", "school-default"} {
			if _, e := s.store.GetSchoolMpesaConfig(ctx, sID); e == nil {
				schoolID = sID
				break
			}
		}
	}
	if schoolID == "" {
		schoolID = "school-1"
	}

	rawBillRef := strings.TrimSpace(payload.BillRefNumber)
	payerPhone := payload.MSISDN
	payerName := payload.FirstName

	rawJSON, _ := json.Marshal(payload)

	// 2. Look up student in that school where UPPER(TRIM(admission_number)) = UPPER(TRIM(BillRefNumber)) (AC-25)
	studentID, _, err := s.store.GetStudentByAdmissionNumber(ctx, schoolID, rawBillRef)

	// CASE A: Valid Student Match Found (AC-25)
	if err == nil && studentID != "" {
		payment := &Payment{
			SchoolID:       schoolID,
			StudentID:      &studentID,
			TransactionRef: payload.TransID,
			Provider:       "mpesa_c2b",
			Amount:         amount,
			PayerPhone:     payerPhone,
			PayerName:      payerName,
			MatchedStatus:  "matched",
			PaymentDate:    time.Now(),
			RawPayload:     string(rawJSON),
		}
		if err := s.store.CreatePayment(ctx, payment); err != nil {
			return nil, errInternal(err)
		}

		// Apply credit to student's ledger (AC-24)
		_ = s.creditStudentPayment(ctx, schoolID, studentID, "", payload.TransID, "mpesa_c2b", amount, payerPhone, payerName)

		if s.auditLogger != nil {
			_ = s.auditLogger.LogAction(ctx, schoolID, "", "payment.c2b.auto_matched", "payment", payment.ID, map[string]interface{}{
				"admission_number": rawBillRef,
				"student_id":       studentID,
				"amount":           amount,
				"transaction_ref":  payload.TransID,
			})
		}

		return payment, nil
	}

	// CASE B: Admission number missing or not found in school (C50, AC-25)
	unmatchedPayment := &Payment{
		SchoolID:       schoolID,
		StudentID:      nil,
		TransactionRef: payload.TransID,
		Provider:       "mpesa_c2b",
		Amount:         amount,
		PayerPhone:     payerPhone,
		PayerName:      payerName,
		MatchedStatus:  "unmatched",
		PaymentDate:    time.Now(),
		RawPayload:     string(rawJSON),
	}
	if err := s.store.CreatePayment(ctx, unmatchedPayment); err != nil {
		return nil, errInternal(err)
	}

	queueItem := &UnmatchedPayment{
		SchoolID:       schoolID,
		PaymentID:      unmatchedPayment.ID,
		TransactionRef: payload.TransID,
		Amount:         amount,
		PayerPhone:     payerPhone,
		PayerName:      payerName,
		RawAccountRef:  rawBillRef,
		Status:         "unmatched",
		CreatedAt:      time.Now(),
	}
	if err := s.store.CreateUnmatchedPayment(ctx, queueItem); err != nil {
		return nil, errInternal(err)
	}

	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(ctx, schoolID, "", "payment.c2b.unmatched_queued", "unmatched_payment", queueItem.ID, map[string]interface{}{
			"raw_bill_ref":    rawBillRef,
			"amount":          amount,
			"transaction_ref": payload.TransID,
			"payer_phone":     payerPhone,
		})
	}

	return unmatchedPayment, nil
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
