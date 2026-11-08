package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// STKPushRequest payload
type STKPushRequest struct {
	SchoolID        string  `json:"school_id"`
	StudentID       string  `json:"student_id"`
	Amount          float64 `json:"amount"`
	PhoneNumber     string  `json:"phone_number"`
	Source          string  `json:"source"` // "portal" or "chat"
	ParentConfirmed bool    `json:"parent_confirmed"`
}

type STKPushResponse struct {
	CheckoutRequestID string `json:"checkout_request_id"`
	ResponseCode      string `json:"response_code"`
	CustomerMessage   string `json:"customer_message"`
}

// STKCallbackPayload represents Daraja M-Pesa STK callback structure
type STKCallbackPayload struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  *struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// InitiateSTKPush handles initiating an M-Pesa STK push for school fees (C23, C24, C27, AC-23, AC-24, AC-26).
func (s *Service) InitiateSTKPush(ctx context.Context, req STKPushRequest) (*STKPushResponse, error) {
	if req.SchoolID == "" || req.StudentID == "" || req.PhoneNumber == "" {
		return nil, errValidation("school_id, student_id, and phone_number are required")
	}

	// 1. Any positive amount accepted (C24, AC-23). Never reject for being less or greater than current balance.
	if req.Amount <= 0 {
		return nil, errValidation("payment amount must be greater than zero")
	}

	callerID := ""
	callerRole := ""
	if id, ok := IdentityFromContext(ctx); ok {
		callerID = id.UserID
		callerRole = id.Role
	}

	// 2. Permission check: Linked guardian (A2) or authorized staff
	if callerRole == RoleParent {
		isGuardian, err := s.store.IsGuardianOf(ctx, req.SchoolID, req.StudentID, callerID)
		if err != nil || !isGuardian {
			return nil, errForbidden("parent is not a registered guardian of this student")
		}
	}

	// 3. 24-Hour Phone-Change Security Freeze (C27, AC-26)
	// If source is chat, check if the guardian's chat payments are paused
	if req.Source == "chat" {
		targetUserID := callerID
		if targetUserID != "" {
			pausedUntil, err := s.store.GetUserChatPaymentPausedUntil(ctx, targetUserID)
			if err == nil && pausedUntil != nil && pausedUntil.After(time.Now()) {
				return nil, &APIError{
					Status:  http.StatusForbidden,
					Code:    CodeChatPaymentPaused,
					Message: fmt.Sprintf("Chat-initiated payments are locked until %s following a phone number change. Please pay via the parent portal.", pausedUntil.Format(time.RFC3339)),
				}
			}
		}

		// 4. Explicit parent confirmation check for chat/AI service (AC-23)
		if !req.ParentConfirmed {
			return nil, errUnprocessable(CodeParentConfirmationReq, "explicit parent confirmation required before initiating chat payment")
		}
	}

	// 5. Look up student to obtain admission number for Daraja AccountReference
	_, adminNum, studentName, err := s.store.GetStudent(ctx, req.SchoolID, req.StudentID)
	if err != nil {
		return nil, errNotFound("student not found in this school")
	}

	// 6. Look up School M-Pesa configuration
	mpesaConfig, err := s.store.GetSchoolMpesaConfig(ctx, req.SchoolID)
	shortcode := "174379" // Default Daraja sandbox business shortcode
	if err == nil && mpesaConfig != nil && mpesaConfig.Shortcode != "" {
		shortcode = mpesaConfig.Shortcode
	}
	if mpesaConfig == nil {
		mpesaConfig = &SchoolMpesaConfig{
			SchoolID:    req.SchoolID,
			Shortcode:   shortcode,
			Environment: "sandbox",
		}
	}

	// Format Safaricom phone: 2547XXXXXXXX
	phone := strings.TrimPrefix(req.PhoneNumber, "+")
	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	if s.darajaClient == nil {
		s.darajaClient = NewHTTPDarajaClient(nil)
	}

	// Real Daraja STK Initiation: send request to Daraja, do not generate fake ID
	params := STKInitiateParams{
		Shortcode:        shortcode,
		Amount:           req.Amount,
		PhoneNumber:      phone,
		AccountReference: adminNum,
		TransactionDesc:  fmt.Sprintf("Fee payment for %s (%s)", studentName, adminNum),
		CallBackURL:      s.callbackURL,
	}

	darajaResp, err := s.darajaClient.InitiateSTKPush(ctx, mpesaConfig, params)
	if err != nil {
		return nil, errInternal(fmt.Errorf("failed to initiate Daraja STK push: %w", err))
	}
	if darajaResp.ResponseCode != "0" {
		return nil, &APIError{
			Status:  http.StatusBadRequest,
			Code:    "DARAJA_ERROR",
			Message: darajaResp.ResponseDescription,
		}
	}

	checkoutID := darajaResp.CheckoutRequestID

	// Pre-record pending payment with actual CheckoutRequestID from Daraja
	pendingPayment := &Payment{
		SchoolID:       req.SchoolID,
		StudentID:      &req.StudentID,
		TransactionRef: checkoutID,
		Provider:       "mpesa_stk",
		Amount:         req.Amount,
		PayerPhone:     phone,
		MatchedStatus:  "pending",
		PaymentDate:    time.Now(),
		RawPayload:     fmt.Sprintf(`{"admission_number": "%s", "shortcode": "%s", "merchant_request_id": "%s", "checkout_request_id": "%s"}`, adminNum, shortcode, darajaResp.MerchantRequestID, checkoutID),
	}
	if err := s.store.CreatePayment(ctx, pendingPayment); err != nil {
		return nil, errInternal(err)
	}

	// 7. Enqueue fallback Asynq job to poll Daraja after 45s (C23, AC-23)
	if s.enqueuer != nil {
		_ = s.enqueuer.EnqueueSTKPoll(ctx, req.SchoolID, checkoutID, 45*time.Second)
	}

	if s.auditLogger != nil {
		_ = s.auditLogger.LogAction(ctx, req.SchoolID, callerID, "payment.stk_push.initiated", "payment", checkoutID, map[string]interface{}{
			"student_id":       req.StudentID,
			"admission_number": adminNum,
			"amount":           req.Amount,
			"phone":            phone,
			"source":           req.Source,
		})
	}

	customerMsg := darajaResp.CustomerMessage
	if customerMsg == "" {
		customerMsg = fmt.Sprintf("Success. Please enter your M-Pesa PIN on %s to pay KES %.2f for %s (%s).", phone, req.Amount, studentName, adminNum)
	}

	return &STKPushResponse{
		CheckoutRequestID: checkoutID,
		ResponseCode:      "0",
		CustomerMessage:   customerMsg,
	}, nil
}

// HandleSTKCallback processes asynchronous Daraja STK Push callbacks (AC-24).
func (s *Service) HandleSTKCallback(ctx context.Context, payload STKCallbackPayload) error {
	cb := payload.Body.StkCallback
	if cb.CheckoutRequestID == "" {
		return errValidation("missing CheckoutRequestID in STK callback")
	}

	// 1. Authenticate callback by matching to a known pending transaction and school
	pendingPayment, err := s.store.GetPaymentByCheckoutID(ctx, cb.CheckoutRequestID)
	if err != nil || pendingPayment == nil {
		return errNotFound("no pending payment found for checkout request ID")
	}

	// 2. Idempotency: repeated callback must not create another payment, ledger entry, notification, or credit
	if pendingPayment.MatchedStatus == "matched" {
		return nil
	}

	// 3. If transaction was cancelled or failed on Daraja
	if cb.ResultCode != 0 {
		_ = s.store.UpdatePaymentStatus(ctx, pendingPayment.SchoolID, pendingPayment.ID, "failed")
		return nil
	}

	var mpesaReceipt string
	var amount float64
	var phone string

	if cb.CallbackMetadata != nil {
		for _, it := range cb.CallbackMetadata.Item {
			switch it.Name {
			case "MpesaReceiptNumber":
				mpesaReceipt = fmt.Sprintf("%v", it.Value)
			case "Amount":
				switch v := it.Value.(type) {
				case float64:
					amount = v
				case int:
					amount = float64(v)
				case string:
					amount, _ = strconv.ParseFloat(v, 64)
				}
			case "PhoneNumber":
				phone = fmt.Sprintf("%v", it.Value)
			}
		}
	}

	if strings.TrimSpace(mpesaReceipt) == "" {
		return errValidation("missing or invalid MpesaReceiptNumber in callback metadata")
	}
	if amount <= 0 {
		amount = pendingPayment.Amount
	}
	if phone == "" {
		phone = pendingPayment.PayerPhone
	}

	// Additional idempotency check on MpesaReceiptNumber
	if existing, err := s.store.GetPaymentByTransactionRef(ctx, mpesaReceipt); err == nil && existing != nil && existing.MatchedStatus == "matched" {
		return nil
	}

	studentID := ""
	if pendingPayment.StudentID != nil {
		studentID = *pendingPayment.StudentID
	}

	// 4. Update the pending payment atomically and credit student's ledger
	return s.creditStudentPayment(ctx, pendingPayment.SchoolID, studentID, pendingPayment.ID, mpesaReceipt, "mpesa_stk", amount, phone, "")
}

// creditStudentPayment handles ledger crediting, double entry, overpayment credit_balance, and guardian notification (AC-24, AC-25).
func (s *Service) creditStudentPayment(ctx context.Context, schoolID, studentID, origRef, transactionRef, provider string, amount float64, payerPhone, payerName string) error {
	// If studentID or schoolID is empty, try to resolve from original payment reference
	var pendingPayment *Payment
	if origRef != "" {
		p, err := s.store.GetPayment(ctx, schoolID, origRef)
		if err == nil && p != nil {
			pendingPayment = p
			if schoolID == "" {
				schoolID = p.SchoolID
			}
			if studentID == "" && p.StudentID != nil {
				studentID = *p.StudentID
			}
			if amount <= 0 {
				amount = p.Amount
			}
			if payerPhone == "" {
				payerPhone = p.PayerPhone
			}
		}
	}

	if studentID == "" || schoolID == "" {
		return errValidation("school_id and student_id are required to credit payment")
	}

	// Idempotency check: has this transactionRef already been matched?
	if existing, err := s.store.GetPaymentByTransactionRef(ctx, transactionRef); err == nil && existing != nil && existing.MatchedStatus == "matched" && existing.ID != origRef {
		return nil
	}

	// 1. Record / Update payment record
	if pendingPayment != nil {
		pendingPayment.TransactionRef = transactionRef
		pendingPayment.Amount = amount
		pendingPayment.PayerPhone = payerPhone
		pendingPayment.PayerName = payerName
		pendingPayment.MatchedStatus = "matched"
		pendingPayment.PaymentDate = time.Now()
		if err := s.store.UpdatePaymentMatched(ctx, schoolID, pendingPayment.ID, studentID, provider); err != nil {
			_ = s.store.CreatePayment(ctx, pendingPayment)
		}
	} else {
		payment := &Payment{
			SchoolID:       schoolID,
			StudentID:      &studentID,
			TransactionRef: transactionRef,
			Provider:       provider,
			Amount:         amount,
			PayerPhone:     payerPhone,
			PayerName:      payerName,
			MatchedStatus:  "matched",
			PaymentDate:    time.Now(),
		}
		if err := s.store.CreatePayment(ctx, payment); err != nil {
			return errInternal(err)
		}
	}

	// 2. Record ledger_entries double-entry (entry_type = 'payment')
	ledgerEntry := &LedgerEntry{
		SchoolID:    schoolID,
		StudentID:   studentID,
		EntryType:   "payment",
		Amount:      amount,
		Reference:   transactionRef,
		Description: fmt.Sprintf("Fee payment via %s (Ref: %s)", provider, transactionRef),
	}
	if err := s.store.CreateLedgerEntry(ctx, ledgerEntry); err != nil {
		return errInternal(err)
	}

	// 3. Update FeeLedger with Overpayment Credit Rule (C24, AC-24):
	// If paid_amount > opening_balance + billed_amount:
	// current_balance = 0.00, credit_balance = paid_amount - (opening_balance + billed_amount)
	ledger, err := s.store.GetLatestFeeLedger(ctx, schoolID, studentID)
	if err == nil && ledger != nil {
		ledger.PaidAmount += amount
		totalBilled := ledger.OpeningBalance + ledger.BilledAmount
		if ledger.PaidAmount >= totalBilled {
			ledger.CurrentBalance = 0.00
			ledger.CreditBalance = ledger.PaidAmount - totalBilled
		} else {
			ledger.CurrentBalance = totalBilled - ledger.PaidAmount
			ledger.CreditBalance = 0.00
		}
		if err := s.store.SaveFeeLedger(ctx, ledger); err != nil {
			return errInternal(err)
		}
	}

	// 4. Send in-app payment receipt notification to student's guardians
	if s.notifier != nil {
		guardians, _ := s.store.ListGuardiansForStudent(ctx, schoolID, studentID)
		for _, guardianUserID := range guardians {
			_ = s.notifier.CreateInAppNotification(ctx, schoolID, guardianUserID, "payment_receipt", "Fee Payment Received", fmt.Sprintf("M-Pesa payment of KES %.2f (Receipt: %s) has been successfully credited to fees.", amount, transactionRef), map[string]interface{}{
				"student_id":      studentID,
				"amount":          amount,
				"transaction_ref": transactionRef,
			})
		}
	}

	return nil
}

func (s *Service) handleSTKPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	id, ok := IdentityFromContext(r.Context())
	if !ok {
		respondError(w, &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: "authentication required"})
		return
	}

	var req STKPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errValidation("invalid request body"))
		return
	}

	if req.SchoolID == "" {
		req.SchoolID = id.SchoolID
	}
	if req.Source == "" {
		req.Source = "portal"
	}

	resp, err := s.InitiateSTKPush(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (s *Service) handleSTKCallbackHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, &APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var payload STKCallbackPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, errValidation("invalid callback payload"))
		return
	}

	if err := s.HandleSTKCallback(r.Context(), payload); err != nil {
		respondError(w, err)
		return
	}

	// Daraja expects standard JSON acknowledgement
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"ResultCode": "0",
		"ResultDesc": "Accepted",
	})
}
