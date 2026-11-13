package billing

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

type DarajaSchoolConfig struct {
	SchoolID       string
	Shortcode      string
	Environment    string
	Passkey        string
	ConsumerKey    string
	ConsumerSecret string
	CallbackURL    string
}

type DarajaSTKPushRequest struct {
	Config           DarajaSchoolConfig
	PhoneNumber      string
	Amount           Money
	AccountReference string
	Description      string
}

type DarajaSTKPushResponse struct {
	MerchantRequestID string
	CheckoutRequestID string
	ResponseCode      string
	ResponseDesc      string
	CustomerMessage   string
}

type DarajaSTKQueryRequest struct {
	Config            DarajaSchoolConfig
	CheckoutRequestID string
}

type DarajaSTKQueryResponse struct {
	MerchantRequestID string
	CheckoutRequestID string
	ResultCode        int
	ResultDesc        string
	ReceiptNumber     string
}

type DarajaProvider interface {
	InitiateSTKPush(ctx context.Context, req DarajaSTKPushRequest) (DarajaSTKPushResponse, error)
	QuerySTKStatus(ctx context.Context, req DarajaSTKQueryRequest) (DarajaSTKQueryResponse, error)
}

type STKPollTask struct {
	SchoolID          string    `json:"school_id"`
	STKRequestID      string    `json:"stk_request_id"`
	CheckoutRequestID string    `json:"checkout_request_id"`
	EnqueuedAt        time.Time `json:"enqueued_at"`
}

type STKPollQueue interface {
	EnqueueSTKPoll(ctx context.Context, task STKPollTask, delay time.Duration) error
}

// HTTPDarajaClient implements DarajaProvider against Safaricom Daraja OAuth,
// STK Push, and STK Query endpoints without exposing secrets or following redirects.
type HTTPDarajaClient struct {
	BaseURL        string
	ConsumerKey    string
	ConsumerSecret string
	CallbackURL    string
	DecryptPasskey func(ciphertext string) (string, error)
	HTTPClient     *http.Client
	Now            func() time.Time
}

func (c *HTTPDarajaClient) httpClient() *http.Client {
	base := http.DefaultClient
	if c != nil && c.HTTPClient != nil {
		base = c.HTTPClient
	}
	copied := *base
	if copied.Timeout <= 0 {
		copied.Timeout = 10 * time.Second
	}
	copied.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &copied
}

func (c *HTTPDarajaClient) baseURL(env string) string {
	if c != nil && strings.TrimSpace(c.BaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	}
	if strings.EqualFold(strings.TrimSpace(env), "production") {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (c *HTTPDarajaClient) now() time.Time {
	if c != nil && c.Now != nil {
		return c.Now().UTC()
	}
	return time.Now().UTC()
}

func (c *HTTPDarajaClient) FetchOAuthToken(ctx context.Context, cfg DarajaSchoolConfig) (string, error) {
	key := strings.TrimSpace(cfg.ConsumerKey)
	secret := strings.TrimSpace(cfg.ConsumerSecret)
	if key == "" && c != nil {
		key = strings.TrimSpace(c.ConsumerKey)
	}
	if secret == "" && c != nil {
		secret = strings.TrimSpace(c.ConsumerSecret)
	}
	if key == "" || secret == "" {
		return "", contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa credentials are not configured for this school.")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL(cfg.Environment)+"/oauth/v1/generate?grant_type=client_credentials", nil)
	if err != nil {
		return "", contracts.Fail(http.StatusBadGateway, "MPESA_OAUTH_FAILED", "Could not initialize M-Pesa authorization.")
	}
	req.SetBasicAuth(key, secret)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", contracts.Fail(http.StatusBadGateway, "MPESA_OAUTH_FAILED", "M-Pesa authorization service is unavailable.")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", contracts.Fail(http.StatusBadGateway, "MPESA_OAUTH_FAILED", "M-Pesa authorization was rejected.")
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || strings.TrimSpace(parsed.AccessToken) == "" {
		return "", contracts.Fail(http.StatusBadGateway, "MPESA_OAUTH_FAILED", "Invalid M-Pesa authorization response.")
	}
	return strings.TrimSpace(parsed.AccessToken), nil
}

func (c *HTTPDarajaClient) resolvePasskey(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa passkey is not configured for this school.")
	}
	if c != nil && c.DecryptPasskey != nil {
		dec, err := c.DecryptPasskey(trimmed)
		if err != nil || strings.TrimSpace(dec) == "" {
			return "", contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa passkey could not be decrypted.")
		}
		return strings.TrimSpace(dec), nil
	}
	return trimmed, nil
}

func (c *HTTPDarajaClient) InitiateSTKPush(ctx context.Context, in DarajaSTKPushRequest) (DarajaSTKPushResponse, error) {
	token, err := c.FetchOAuthToken(ctx, in.Config)
	if err != nil {
		return DarajaSTKPushResponse{}, err
	}
	passkey, err := c.resolvePasskey(in.Config.Passkey)
	if err != nil {
		return DarajaSTKPushResponse{}, err
	}
	callbackURL := strings.TrimSpace(in.Config.CallbackURL)
	if callbackURL == "" && c != nil {
		callbackURL = strings.TrimSpace(c.CallbackURL)
	}
	if callbackURL == "" {
		callbackURL = "https://example.invalid/api/v1/billing/mpesa/callback"
	}
	timestamp := c.now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(in.Config.Shortcode + passkey + timestamp))
	darajaDigits := strings.TrimPrefix(in.PhoneNumber, "+")
	wholeKES := int64(math.Ceil(float64(in.Amount) / 100.0))
	if wholeKES < 1 {
		wholeKES = 1
	}

	payload := map[string]any{
		"BusinessShortCode": in.Config.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            wholeKES,
		"PartyA":            darajaDigits,
		"PartyB":            in.Config.Shortcode,
		"PhoneNumber":       darajaDigits,
		"CallBackURL":       callbackURL,
		"AccountReference":  in.AccountReference,
		"TransactionDesc":   in.Description,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "Could not encode STK request.")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL(in.Config.Environment)+"/mpesa/stkpush/v1/processrequest", bytes.NewReader(encoded))
	if err != nil {
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_STK_FAILED", "Could not build STK push request.")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_STK_FAILED", "M-Pesa STK service is unavailable.")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil || resp.StatusCode != http.StatusOK {
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_STK_FAILED", "M-Pesa STK push request was rejected.")
	}
	var out struct {
		MerchantRequestID   string `json:"MerchantRequestID"`
		CheckoutRequestID   string `json:"CheckoutRequestID"`
		ResponseCode        string `json:"ResponseCode"`
		ResponseDescription string `json:"ResponseDescription"`
		CustomerMessage     string `json:"CustomerMessage"`
	}
	if err := json.Unmarshal(body, &out); err != nil || strings.TrimSpace(out.CheckoutRequestID) == "" || out.ResponseCode != "0" {
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_STK_FAILED", "M-Pesa did not accept the STK push request.")
	}
	return DarajaSTKPushResponse{
		MerchantRequestID: strings.TrimSpace(out.MerchantRequestID),
		CheckoutRequestID: strings.TrimSpace(out.CheckoutRequestID),
		ResponseCode:      out.ResponseCode,
		ResponseDesc:      out.ResponseDescription,
		CustomerMessage:   out.CustomerMessage,
	}, nil
}

func (c *HTTPDarajaClient) QuerySTKStatus(ctx context.Context, in DarajaSTKQueryRequest) (DarajaSTKQueryResponse, error) {
	token, err := c.FetchOAuthToken(ctx, in.Config)
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	passkey, err := c.resolvePasskey(in.Config.Passkey)
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	timestamp := c.now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(in.Config.Shortcode + passkey + timestamp))
	payload := map[string]any{
		"BusinessShortCode": in.Config.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": in.CheckoutRequestID,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL(in.Config.Environment)+"/mpesa/stkpushquery/v1/query", bytes.NewReader(encoded))
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return DarajaSTKQueryResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_QUERY_FAILED", "M-Pesa status query service is unavailable.")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil || resp.StatusCode != http.StatusOK {
		return DarajaSTKQueryResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_QUERY_FAILED", "M-Pesa status query failed.")
	}
	var raw struct {
		MerchantRequestID string          `json:"MerchantRequestID"`
		CheckoutRequestID string          `json:"CheckoutRequestID"`
		ResultCode        json.RawMessage `json:"ResultCode"`
		ResultDesc        string          `json:"ResultDesc"`
		MpesaReceipt      string          `json:"MpesaReceiptNumber"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return DarajaSTKQueryResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_QUERY_FAILED", "Invalid M-Pesa query response.")
	}
	code, err := parseFlexibleInt(raw.ResultCode)
	if err != nil {
		return DarajaSTKQueryResponse{}, contracts.Fail(http.StatusBadGateway, "MPESA_QUERY_FAILED", "Invalid M-Pesa result code.")
	}
	return DarajaSTKQueryResponse{
		MerchantRequestID: strings.TrimSpace(raw.MerchantRequestID),
		CheckoutRequestID: strings.TrimSpace(raw.CheckoutRequestID),
		ResultCode:        code,
		ResultDesc:        strings.TrimSpace(raw.ResultDesc),
		ReceiptNumber:     strings.TrimSpace(raw.MpesaReceipt),
	}, nil
}

func parseFlexibleInt(raw json.RawMessage) (int, error) {
	s := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if s == "" {
		return 0, errors.New("empty result code")
	}
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}

type InitiateSTKPushInput struct {
	StudentID       string          `json:"student_id"`
	Amount          json.RawMessage `json:"amount"`
	PhoneNumber     string          `json:"phone_number"`
	Source          string          `json:"source"`
	ParentConfirmed *bool           `json:"parent_confirmed"`
}

type STKPushInitiated struct {
	STKRequestID      string `json:"stk_request_id"`
	StudentID         string `json:"student_id"`
	AdmissionNumber   string `json:"admission_number"`
	AccountReference  string `json:"account_reference"`
	Amount            Money  `json:"amount"`
	PhoneNumber       string `json:"phone_number"`
	CheckoutRequestID string `json:"checkout_request_id"`
	MerchantRequestID string `json:"merchant_request_id"`
	Status            string `json:"status"`
	CustomerMessage   string `json:"customer_message"`
}

// InitiateSTKPush validates guardian authorization, any-positive-amount rule,
// chat 24-hour phone-change freeze, and explicit parent confirmation (when from AI/internal),
// persists the durable STK attempt, calls Daraja with AccountReference = admission_number,
// and schedules a 45-second STK status poll task.
func (s *Service) InitiateSTKPush(ctx context.Context, p contracts.Principal, in InitiateSTKPushInput, requireParentConfirmed bool) (STKPushInitiated, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return STKPushInitiated{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !contracts.ValidID(in.StudentID) {
		return STKPushInitiated{}, validationFail("student_id", "INVALID_STUDENT_ID", "A valid student_id is required.")
	}
	source := strings.ToLower(strings.TrimSpace(in.Source))
	if source == "" {
		if requireParentConfirmed {
			source = "chat"
		} else {
			source = "web"
		}
	}
	if source != "web" && source != "chat" && source != "whatsapp" && source != "in_app" {
		return STKPushInitiated{}, validationFail("source", "INVALID_PAYMENT_SOURCE", "source must be 'web' or 'chat'.")
	}
	isChatSource := source == "chat" || source == "whatsapp" || source == "in_app" || requireParentConfirmed
	if requireParentConfirmed || (in.ParentConfirmed != nil && isChatSource) {
		if in.ParentConfirmed == nil || !*in.ParentConfirmed {
			return STKPushInitiated{}, validationFail("parent_confirmed", "PARENT_CONFIRMATION_REQUIRED", "Explicit parental confirmation (parent_confirmed=true) is required before initiating an AI chat payment.")
		}
	}
	if len(in.Amount) == 0 {
		return STKPushInitiated{}, validationFail("amount", "INVALID_PAYMENT_AMOUNT", "A positive payment amount is required.")
	}
	var amount Money
	if err := json.Unmarshal(in.Amount, &amount); err != nil || amount <= 0 {
		return STKPushInitiated{}, validationFail("amount", "INVALID_PAYMENT_AMOUNT", "Payment amount must be strictly positive (> 0).")
	}

	if s == nil || s.DB == nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if s.Daraja == nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa provider is not configured.")
	}

	var stkRequestID, admissionNumber, normalizedPhone string
	var darajaCfg DarajaSchoolConfig

	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT admission_number
			FROM students
			WHERE school_id = $1 AND id = $2 AND status = 'active'`,
			p.SchoolID, in.StudentID,
		).Scan(&admissionNumber); errors.Is(err, pgx.ErrNoRows) {
			return contracts.Fail(http.StatusNotFound, "STUDENT_NOT_FOUND", "Student not found.")
		} else if err != nil {
			return err
		}

		var linked bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM student_guardians
				WHERE school_id = $1 AND student_id = $2 AND user_id = $3
			)`,
			p.SchoolID, in.StudentID, p.UserID,
		).Scan(&linked); err != nil {
			return err
		}
		if !linked {
			return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Only a linked guardian may initiate an M-Pesa payment for this student.")
		}

		var registeredPhone string
		if err := tx.QueryRow(ctx, `SELECT phone_number FROM users WHERE id = $1 AND is_active`, p.UserID).Scan(&registeredPhone); err != nil {
			return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Guardian account is not active.")
		}

		if isChatSource {
			paused, err := isUserChatPaymentPaused(ctx, tx, p.UserID, s.now())
			if err != nil {
				return err
			}
			if paused {
				return contracts.Fail(http.StatusForbidden, "CHAT_PAYMENT_PAUSED_PHONE_CHANGE", "Chat-initiated payments are paused for 24 hours after a phone number change.")
			}

			// Security Remediation 6.7: Chat-initiated payments must be sent to the guardian's registered phone
			phoneCandidate := strings.TrimSpace(in.PhoneNumber)
			if phoneCandidate != "" {
				normCandidate, errCandidate := notifications.NormalizePhone(phoneCandidate)
				normRegistered, errReg := notifications.NormalizePhone(registeredPhone)
				if errCandidate == nil && errReg == nil && normCandidate != normRegistered {
					return validationFail("phone_number", "CHAT_PHONE_OVERRIDE_FORBIDDEN", "Chat-initiated payments must be sent to the guardian's registered phone number.")
				}
			}
		}

		phoneCandidate := strings.TrimSpace(in.PhoneNumber)
		if phoneCandidate == "" {
			phoneCandidate = registeredPhone
		}
		norm, err := notifications.NormalizePhone(phoneCandidate)
		if err != nil {
			return validationFail("phone_number", "INVALID_PHONE_NUMBER", "A valid E.164 or Kenyan M-Pesa phone number is required.")
		}
		normalizedPhone = norm

		if err := tx.QueryRow(ctx, `
			SELECT school_id, shortcode, environment, COALESCE(passkey_encrypted, '')
			FROM school_mpesa_configs
			WHERE school_id = $1`,
			p.SchoolID,
		).Scan(&darajaCfg.SchoolID, &darajaCfg.Shortcode, &darajaCfg.Environment, &darajaCfg.Passkey); errors.Is(err, pgx.ErrNoRows) {
			return contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa Paybill is not configured for this school.")
		} else if err != nil {
			return err
		}

		amtNum, err := amount.NumericValue()
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `
			INSERT INTO mpesa_stk_requests (school_id, student_id, guardian_user_id, amount, phone_number, status)
			VALUES ($1, $2, $3, $4, $5, 'initiating')
			RETURNING id`,
			p.SchoolID, in.StudentID, p.UserID, amtNum, normalizedPhone,
		).Scan(&stkRequestID)
	})
	if err != nil {
		return STKPushInitiated{}, err
	}

	darajaResp, darajaErr := s.Daraja.InitiateSTKPush(ctx, DarajaSTKPushRequest{
		Config:           darajaCfg,
		PhoneNumber:      normalizedPhone,
		Amount:           amount,
		AccountReference: admissionNumber,
		Description:      "School Fee Payment " + admissionNumber,
	})
	if darajaErr != nil {
		_ = s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				UPDATE mpesa_stk_requests
				SET status = 'failed', updated_at = NOW()
				WHERE school_id = $1 AND id = $2 AND status = 'initiating'`,
				p.SchoolID, stkRequestID,
			)
			return err
		})
		return STKPushInitiated{}, darajaErr
	}

	if err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE mpesa_stk_requests
			SET status = 'pending',
			    checkout_request_id = $1,
			    merchant_request_id = $2,
			    updated_at = NOW()
			WHERE school_id = $3 AND id = $4`,
			darajaResp.CheckoutRequestID, darajaResp.MerchantRequestID, p.SchoolID, stkRequestID,
		); err != nil {
			return err
		}
		if s.Audit != nil {
			if err := s.Audit.Log(ctx, tx, contracts.AuditEvent{
				SchoolID:     p.SchoolID,
				UserID:       p.UserID,
				Action:       "billing.mpesa.stk_initiate",
				ResourceType: "mpesa_stk_request",
				ResourceID:   stkRequestID,
				Details: map[string]any{
					"student_id":          in.StudentID,
					"admission_number":    admissionNumber,
					"amount":              amount.String(),
					"checkout_request_id": darajaResp.CheckoutRequestID,
					"source":              source,
				},
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return STKPushInitiated{}, err
	}

	if s.PollQueue != nil {
		if err := s.PollQueue.EnqueueSTKPoll(ctx, STKPollTask{
			SchoolID:          p.SchoolID,
			STKRequestID:      stkRequestID,
			CheckoutRequestID: darajaResp.CheckoutRequestID,
			EnqueuedAt:        s.now(),
		}, 45*time.Second); err != nil && s.Audit != nil {
			_ = s.Audit.Log(ctx, nil, contracts.AuditEvent{
				SchoolID:     p.SchoolID,
				UserID:       p.UserID,
				Action:       "billing.mpesa.stk_poll_enqueue_failed",
				ResourceType: "mpesa_stk_request",
				ResourceID:   stkRequestID,
				Details: map[string]any{
					"checkout_request_id": darajaResp.CheckoutRequestID,
					"error":               err.Error(),
				},
			})
		}
	}

	return STKPushInitiated{
		STKRequestID:      stkRequestID,
		StudentID:         in.StudentID,
		AdmissionNumber:   admissionNumber,
		AccountReference:  admissionNumber,
		Amount:            amount,
		PhoneNumber:       normalizedPhone,
		CheckoutRequestID: darajaResp.CheckoutRequestID,
		MerchantRequestID: darajaResp.MerchantRequestID,
		Status:            "pending",
		CustomerMessage:   darajaResp.CustomerMessage,
	}, nil
}

func isUserChatPaymentPaused(ctx context.Context, tx pgx.Tx, userID string, now time.Time) (bool, error) {
	var hasColumn bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'chat_payment_paused_until'
		)`).Scan(&hasColumn); err != nil {
		return false, err
	}
	if !hasColumn {
		return false, nil
	}
	var paused bool
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(chat_payment_paused_until > $2, false)
		FROM users
		WHERE id = $1`,
		userID, now,
	).Scan(&paused)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return paused, err
}

type DarajaSTKCallbackEnvelope struct {
	SchoolID          string `json:"school_id"`
	BusinessShortCode string `json:"BusinessShortCode"`
	Body              struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        *int   `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  *struct {
				Item []struct {
					Name  string          `json:"Name"`
					Value json.RawMessage `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// ProcessSTKCallback processes a Daraja STK Push callback idempotently.
// Repeated callbacks with the same CheckoutRequestID or MpesaReceiptNumber
// never credit the student's ledger twice.
func (s *Service) ProcessSTKCallback(ctx context.Context, schoolIDHint string, payload DarajaSTKCallbackEnvelope) error {
	cb := payload.Body.StkCallback
	checkoutID := strings.TrimSpace(cb.CheckoutRequestID)
	if checkoutID == "" || cb.ResultCode == nil {
		return contracts.Fail(http.StatusBadRequest, "INVALID_MPESA_CALLBACK", "CheckoutRequestID and ResultCode are required.")
	}
	schoolID, err := s.resolveWebhookSchoolID(ctx, schoolIDHint, payload.SchoolID, payload.BusinessShortCode)
	if err != nil {
		return err
	}

	var receiptNumber, phoneStr string
	var callbackAmount Money
	if cb.CallbackMetadata != nil {
		for _, item := range cb.CallbackMetadata.Item {
			valStr := strings.Trim(strings.TrimSpace(string(item.Value)), `"`)
			switch item.Name {
			case "MpesaReceiptNumber":
				receiptNumber = valStr
			case "PhoneNumber":
				phoneStr = valStr
			case "Amount":
				if parsed, err := ParseMoney(valStr); err == nil && parsed > 0 {
					callbackAmount = parsed
				}
			}
		}
	}

	return s.DB.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		return s.settleSTKResultTx(ctx, tx, schoolID, checkoutID, *cb.ResultCode, receiptNumber, phoneStr, callbackAmount, payload)
	})
}

func (s *Service) settleSTKResultTx(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, checkoutID string,
	resultCode int,
	receiptNumber, phoneStr string,
	callbackAmount Money,
	rawPayload any,
) error {
	var stkID, studentID, guardianUserID, reqPhone, currentStatus string
	var reqAmount Money
	err := tx.QueryRow(ctx, `
		SELECT id, student_id, guardian_user_id, amount, phone_number, status
		FROM mpesa_stk_requests
		WHERE school_id = $1 AND checkout_request_id = $2
		FOR UPDATE`,
		schoolID, checkoutID,
	).Scan(&stkID, &studentID, &guardianUserID, &reqAmount, &reqPhone, &currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return contracts.Fail(http.StatusNotFound, "STK_REQUEST_NOT_FOUND", "STK request not found for this school.")
	}
	if err != nil {
		return err
	}

	if currentStatus == "succeeded" || currentStatus == "failed" {
		return nil
	}

	if resultCode != 0 {
		_, err := tx.Exec(ctx, `
			UPDATE mpesa_stk_requests
			SET status = 'failed', result_code = $1, updated_at = NOW()
			WHERE school_id = $2 AND id = $3`,
			resultCode, schoolID, stkID,
		)
		return err
	}

	receipt := strings.TrimSpace(receiptNumber)
	if receipt == "" {
		return contracts.Fail(http.StatusUnprocessableEntity, "MISSING_MPESA_RECEIPT", "Successful STK callback requires an MpesaReceiptNumber.")
	}
	// Security Remediation 6.4: Validate that callback amount matches initiated STK amount exactly
	if callbackAmount > 0 && callbackAmount != reqAmount {
		return contracts.Fail(http.StatusUnprocessableEntity, "MPESA_AMOUNT_MISMATCH", "Callback amount does not match initiated STK amount.")
	}
	settledAmount := reqAmount
	settledPhone := reqPhone
	if norm, err := notifications.NormalizePhone(phoneStr); err == nil {
		settledPhone = norm
	}

	if _, err := tx.Exec(ctx, `
		UPDATE mpesa_stk_requests
		SET status = 'succeeded', result_code = 0, transaction_ref = $1, updated_at = NOW()
		WHERE school_id = $2 AND id = $3`,
		receipt, schoolID, stkID,
	); err != nil {
		return err
	}

	return s.creditMatchedStudentPaymentTx(ctx, tx, schoolID, studentID, receipt, "mpesa_stk", settledPhone, "", settledAmount, rawPayload)
}

// PollSTKRequest queries Daraja for a pending STK push request and updates its status
// idempotently without ever fabricating a receipt or double-crediting.
func (s *Service) PollSTKRequest(ctx context.Context, schoolID, stkRequestID string) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(stkRequestID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_STK_POLL", "Valid school_id and stk_request_id are required.")
	}
	if s == nil || s.DB == nil || s.Daraja == nil {
		return contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa polling dependencies are not configured.")
	}

	var checkoutID, status string
	var cfg DarajaSchoolConfig
	err := s.DB.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(checkout_request_id, ''), status
			FROM mpesa_stk_requests
			WHERE school_id = $1 AND id = $2`,
			schoolID, stkRequestID,
		).Scan(&checkoutID, &status); err != nil {
			return err
		}
		if status != "pending" || checkoutID == "" {
			return nil
		}
		return tx.QueryRow(ctx, `
			SELECT school_id, shortcode, environment, COALESCE(passkey_encrypted, '')
			FROM school_mpesa_configs
			WHERE school_id = $1`,
			schoolID,
		).Scan(&cfg.SchoolID, &cfg.Shortcode, &cfg.Environment, &cfg.Passkey)
	})
	if err != nil || status != "pending" || checkoutID == "" {
		return err
	}

	queryResp, err := s.Daraja.QuerySTKStatus(ctx, DarajaSTKQueryRequest{
		Config:            cfg,
		CheckoutRequestID: checkoutID,
	})
	if err != nil {
		return err
	}

	return s.DB.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		if queryResp.ResultCode == 0 {
			if strings.TrimSpace(queryResp.ReceiptNumber) == "" {
				// Never fabricate a receipt if Daraja query does not include one; mark uncertain and await callback/C2B.
				_, err := tx.Exec(ctx, `
					UPDATE mpesa_stk_requests
					SET status = 'uncertain', result_code = 0, updated_at = NOW()
					WHERE school_id = $1 AND id = $2 AND status = 'pending'`,
					schoolID, stkRequestID,
				)
				return err
			}
			return s.settleSTKResultTx(ctx, tx, schoolID, checkoutID, 0, queryResp.ReceiptNumber, "", 0, queryResp)
		}
		newStatus := "failed"
		if queryResp.ResultCode == 1037 {
			newStatus = "timeout"
		}
		_, err := tx.Exec(ctx, `
			UPDATE mpesa_stk_requests
			SET status = $1, result_code = $2, updated_at = NOW()
			WHERE school_id = $3 AND id = $4 AND status = 'pending'`,
			newStatus, queryResp.ResultCode, schoolID, stkRequestID,
		)
		return err
	})
}

func (s *Service) resolveWebhookSchoolID(ctx context.Context, candidates ...string) (string, error) {
	for _, c := range candidates {
		trimmed := strings.TrimSpace(c)
		if contracts.ValidID(trimmed) {
			return trimmed, nil
		}
	}
	for _, c := range candidates {
		trimmed := strings.TrimSpace(c)
		if trimmed != "" && !contracts.ValidID(trimmed) && s != nil && s.ResolveShortcode != nil {
			schoolID, err := s.ResolveShortcode(ctx, trimmed)
			if err == nil && contracts.ValidID(schoolID) {
				return schoolID, nil
			}
		}
	}
	return "", contracts.Fail(http.StatusBadRequest, "UNKNOWN_MPESA_SHORTCODE", "Could not resolve school for the M-Pesa shortcode.")
}

func (s *Service) creditMatchedStudentPaymentTx(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, studentID, receipt, provider, payerPhone, payerName string,
	amount Money,
	rawPayload any,
) error {
	amtNum, err := amount.NumericValue()
	if err != nil {
		return err
	}
	rawBytes, err := json.Marshal(rawPayload)
	if err != nil {
		rawBytes = []byte(`{}`)
	}

	var paymentID string
	err = tx.QueryRow(ctx, `
		INSERT INTO payments (school_id, student_id, transaction_ref, provider, amount, payer_phone, payer_name, matched_status, payment_date, raw_payload)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), 'matched', NOW(), $8)
		ON CONFLICT (transaction_ref) DO NOTHING
		RETURNING id`,
		schoolID, studentID, receipt, provider, amtNum, payerPhone, payerName, rawBytes,
	).Scan(&paymentID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Duplicate callback/confirmation with the same M-Pesa receipt number: idempotently ignore.
		return nil
	}
	if err != nil {
		return err
	}

	if err := s.RecordStudentPaymentTx(ctx, tx, schoolID, studentID, receipt, "M-Pesa payment ("+receipt+")", amount); err != nil {
		return err
	}

	return s.notifyAffectedGuardians(
		ctx, tx, schoolID, []string{studentID},
		"payment_receipt",
		"M-Pesa Payment Received",
		fmt.Sprintf("Payment of KES %s (Receipt %s) has been credited to the student fee ledger.", amount.String(), receipt),
		map[string]any{
			"payment_id":      paymentID,
			"student_id":      studentID,
			"transaction_ref": receipt,
			"amount":          amount.String(),
			"provider":        provider,
		},
	)
}
