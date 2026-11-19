package billing

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

var nairobiLocation = func() *time.Location {
	loc, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return time.FixedZone("Africa/Nairobi", 3*60*60)
	}
	return loc
}()

// DarajaTimestamp formats an instant in Africa/Nairobi (EAT, UTC+3) using the
// YYYYMMDDHHmmss layout required by Safaricom Daraja APIs.
func DarajaTimestamp(t time.Time) string {
	return t.In(nairobiLocation).Format("20060102150405")
}

// HashCallbackToken returns the lowercase hex SHA-256 digest of a per-request
// STK callback token so raw tokens are never stored in the database or logs.
func HashCallbackToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func generateCallbackToken() (raw string, hash string, err error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(buf[:])
	return raw, HashCallbackToken(raw), nil
}

func computeSTKRequestHash(schoolID, guardianUserID, studentID string, amount Money, normalizedPhone, source string) string {
	material := strings.Join([]string{
		schoolID,
		guardianUserID,
		studentID,
		amount.String(),
		normalizedPhone,
		source,
	}, "|")
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])
}

type DarajaSchoolConfig struct {
	SchoolID       string
	Shortcode      string
	ShortcodeType  string
	Environment    string
	Passkey        string
	ConsumerKey    string
	ConsumerSecret string
	CallbackURL    string
}

func validateShortcodeType(shortcodeType string) error {
	norm := strings.ToLower(strings.TrimSpace(shortcodeType))
	if norm == "" || norm == "paybill" {
		return nil
	}
	return contracts.Fail(http.StatusUnprocessableEntity, "UNSUPPORTED_MPESA_SHORTCODE_TYPE", "Release 1 supports Paybill shortcodes only; Till (Buy Goods) is not supported.")
}

type DarajaSTKPushRequest struct {
	Config           DarajaSchoolConfig
	PhoneNumber      string
	Amount           Money
	AccountReference string
	Description      string
	CallbackToken    string
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
		return c.Now()
	}
	return time.Now()
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
	if err := validateShortcodeType(in.Config.ShortcodeType); err != nil {
		return DarajaSTKPushResponse{}, err
	}
	if in.Amount < 100 || in.Amount%100 != 0 {
		return DarajaSTKPushResponse{}, validationFail("amount", "INVALID_PAYMENT_AMOUNT", "M-Pesa STK push amounts must be whole shillings >= 1.00 KES.")
	}
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
		return DarajaSTKPushResponse{}, contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa callback URL is not configured.")
	}
	if cbToken := strings.TrimSpace(in.CallbackToken); cbToken != "" {
		callbackURL = strings.TrimRight(callbackURL, "/") + "/" + cbToken
	}

	timestamp := DarajaTimestamp(c.now())
	password := base64.StdEncoding.EncodeToString([]byte(in.Config.Shortcode + passkey + timestamp))
	darajaDigits := strings.TrimPrefix(in.PhoneNumber, "+")
	wholeKES := int64(in.Amount / 100)

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
	if err := validateShortcodeType(in.Config.ShortcodeType); err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	token, err := c.FetchOAuthToken(ctx, in.Config)
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	passkey, err := c.resolvePasskey(in.Config.Passkey)
	if err != nil {
		return DarajaSTKQueryResponse{}, err
	}
	timestamp := DarajaTimestamp(c.now())
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
	code, err := ParseFlexibleInt(raw.ResultCode)
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

// ParseFlexibleInt strictly parses a JSON integer or quoted integer string
// (e.g. 0 or "0") using strconv.Atoi, rejecting empty inputs, decimals,
// trailing characters, malformed strings, and out-of-range values.
func ParseFlexibleInt(raw json.RawMessage) (int, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return 0, errors.New("empty integer value")
	}
	if strings.HasPrefix(trimmed, `"`) {
		var unquoted string
		if err := json.Unmarshal([]byte(trimmed), &unquoted); err != nil {
			return 0, errors.New("malformed quoted integer")
		}
		trimmed = strings.TrimSpace(unquoted)
	}
	if trimmed == "" {
		return 0, errors.New("empty integer value")
	}
	v, err := strconv.Atoi(trimmed)
	if err != nil {
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
	IdempotencyKey  string          `json:"idempotency_key,omitempty"`
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
	PollScheduled     bool   `json:"poll_scheduled"`
}

// InitiateSTKPush validates fee-paying guardian authorization, whole-KES amount
// rules, chat 24-hour phone-change freeze, mandatory explicit parent confirmation
// for all chat-initiated payments, and optional scoped idempotency keys; persists
// a durable STK attempt with a server-generated per-request callback token hash;
// calls Daraja with AccountReference = admission_number; and schedules a 45-second
// STK status poll task.
func (s *Service) InitiateSTKPush(ctx context.Context, p contracts.Principal, in InitiateSTKPushInput, requireParentConfirmed bool) (STKPushInitiated, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) || strings.TrimSpace(p.Role) == "" {
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
		return STKPushInitiated{}, validationFail("source", "INVALID_PAYMENT_SOURCE", "source must be 'web', 'chat', 'whatsapp', or 'in_app'.")
	}
	isChatSource := source == "chat" || source == "whatsapp" || source == "in_app" || requireParentConfirmed
	if isChatSource {
		if in.ParentConfirmed == nil || !*in.ParentConfirmed {
			return STKPushInitiated{}, validationFail("parent_confirmed", "PARENT_CONFIRMATION_REQUIRED", "Explicit parental confirmation (parent_confirmed=true) is required before initiating a chat payment.")
		}
	}
	if len(in.Amount) == 0 {
		return STKPushInitiated{}, validationFail("amount", "INVALID_PAYMENT_AMOUNT", "A positive payment amount is required.")
	}
	var amount Money
	if err := json.Unmarshal(in.Amount, &amount); err != nil || amount < 100 || amount%100 != 0 {
		return STKPushInitiated{}, validationFail("amount", "INVALID_PAYMENT_AMOUNT", "M-Pesa STK push amounts must be whole shillings >= 1.00 KES.")
	}
	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if len(idempotencyKey) > 128 {
		return STKPushInitiated{}, validationFail("idempotency_key", "INVALID_IDEMPOTENCY_KEY", "Idempotency key must be at most 128 characters.")
	}

	if s == nil || s.DB == nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return STKPushInitiated{}, err
	}
	if err := s.requireNotifier(); err != nil {
		return STKPushInitiated{}, err
	}
	if s.Daraja == nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa provider is not configured.")
	}
	if s.PollQueue == nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusServiceUnavailable, "MPESA_POLL_QUEUE_NOT_CONFIGURED", "M-Pesa polling queue is not configured.")
	}

	callbackToken, callbackTokenHash, err := generateCallbackToken()
	if err != nil {
		return STKPushInitiated{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "Could not generate callback correlation token.")
	}

	var stkRequestID, admissionNumber, normalizedPhone string
	var darajaCfg DarajaSchoolConfig
	var idempotentReplay *STKPushInitiated

	err = s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
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

		linked, err := isActiveFeePayingGuardianTx(ctx, tx, p, in.StudentID)
		if err != nil {
			return err
		}
		if !linked {
			return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Only an active fee-paying guardian may initiate an M-Pesa payment for this student.")
		}

		var registeredPhone string
		if err := tx.QueryRow(ctx, `SELECT COALESCE(phone_number, '') FROM users WHERE id = $1 AND is_active = TRUE`, p.UserID).Scan(&registeredPhone); err != nil {
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

			normRegistered, errReg := notifications.NormalizePhone(registeredPhone)
			if errReg != nil {
				return validationFail("phone_number", "INVALID_REGISTERED_PHONE", "Guardian does not have a valid registered phone number for chat-initiated payments.")
			}
			// Chat-initiated payments must strictly target the guardian's verified registered phone.
			if phoneCandidate := strings.TrimSpace(in.PhoneNumber); phoneCandidate != "" {
				normCandidate, errCandidate := notifications.NormalizePhone(phoneCandidate)
				if errCandidate != nil || normCandidate != normRegistered {
					return validationFail("phone_number", "CHAT_PHONE_OVERRIDE_FORBIDDEN", "Chat-initiated payments must be sent to the guardian's registered phone number.")
				}
			}
			normalizedPhone = normRegistered
		} else {
			phoneCandidate := strings.TrimSpace(in.PhoneNumber)
			if phoneCandidate == "" {
				phoneCandidate = registeredPhone
			}
			norm, err := notifications.NormalizePhone(phoneCandidate)
			if err != nil {
				return validationFail("phone_number", "INVALID_PHONE_NUMBER", "A valid E.164 or Kenyan M-Pesa phone number is required.")
			}
			normalizedPhone = norm
		}

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

		if idempotencyKey != "" {
			reqHash := computeSTKRequestHash(p.SchoolID, p.UserID, in.StudentID, amount, normalizedPhone, source)
			err = tx.QueryRow(ctx, `
				INSERT INTO mpesa_stk_requests (
					school_id, student_id, guardian_user_id, amount, phone_number, status,
					idempotency_key, request_hash, callback_token_hash
				)
				VALUES ($1, $2, $3, $4, $5, 'initiating', $6, $7, $8)
				ON CONFLICT (school_id, guardian_user_id, idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
				RETURNING id`,
				p.SchoolID, in.StudentID, p.UserID, amtNum, normalizedPhone,
				idempotencyKey, reqHash, callbackTokenHash,
			).Scan(&stkRequestID)
			if errors.Is(err, pgx.ErrNoRows) {
				var existingID, existingStudentID, existingPhone, existingStatus, existingHash string
				var existingCheckout, existingMerchant *string
				var existingAmount Money
				var existingPollScheduled bool
				lookupErr := tx.QueryRow(ctx, `
					SELECT id, student_id, amount, phone_number, status,
					       checkout_request_id, merchant_request_id,
					       COALESCE(request_hash, ''), poll_scheduled
					FROM mpesa_stk_requests
					WHERE school_id = $1 AND guardian_user_id = $2 AND idempotency_key = $3`,
					p.SchoolID, p.UserID, idempotencyKey,
				).Scan(
					&existingID, &existingStudentID, &existingAmount, &existingPhone, &existingStatus,
					&existingCheckout, &existingMerchant, &existingHash, &existingPollScheduled,
				)
				if lookupErr != nil {
					return lookupErr
				}
				if subtle.ConstantTimeCompare([]byte(existingHash), []byte(reqHash)) != 1 {
					return contracts.Fail(http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", "Idempotency key was already used with different payment parameters.")
				}
				if existingStatus == "initiating" {
					return contracts.Fail(http.StatusConflict, "STK_REQUEST_IN_PROGRESS", "An STK push request with this idempotency key is already in progress.")
				}
				replay := STKPushInitiated{
					STKRequestID:     existingID,
					StudentID:        existingStudentID,
					AdmissionNumber:  admissionNumber,
					AccountReference: admissionNumber,
					Amount:           existingAmount,
					PhoneNumber:      existingPhone,
					Status:           existingStatus,
					PollScheduled:    existingPollScheduled,
				}
				if existingCheckout != nil {
					replay.CheckoutRequestID = *existingCheckout
				}
				if existingMerchant != nil {
					replay.MerchantRequestID = *existingMerchant
				}
				idempotentReplay = &replay
				return nil
			}
			return err
		}

		return tx.QueryRow(ctx, `
			INSERT INTO mpesa_stk_requests (
				school_id, student_id, guardian_user_id, amount, phone_number, status, callback_token_hash
			)
			VALUES ($1, $2, $3, $4, $5, 'initiating', $6)
			RETURNING id`,
			p.SchoolID, in.StudentID, p.UserID, amtNum, normalizedPhone, callbackTokenHash,
		).Scan(&stkRequestID)
	})
	if err != nil {
		return STKPushInitiated{}, err
	}
	if idempotentReplay != nil {
		return *idempotentReplay, nil
	}

	darajaResp, darajaErr := s.Daraja.InitiateSTKPush(ctx, DarajaSTKPushRequest{
		Config:           darajaCfg,
		PhoneNumber:      normalizedPhone,
		Amount:           amount,
		AccountReference: admissionNumber,
		Description:      "School Fee Payment " + admissionNumber,
		CallbackToken:    callbackToken,
	})
	if darajaErr != nil {
		_ = s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				UPDATE mpesa_stk_requests
				SET status = 'failed', updated_at = NOW()
				WHERE school_id = $1 AND id = $2 AND status = 'initiating'`,
				p.SchoolID, stkRequestID,
			); err != nil {
				return err
			}
			return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
				SchoolID:     p.SchoolID,
				UserID:       p.UserID,
				Action:       "billing.mpesa.stk_initiate_failed",
				ResourceType: "mpesa_stk_request",
				ResourceID:   stkRequestID,
				Details: map[string]any{
					"student_id":       in.StudentID,
					"admission_number": admissionNumber,
					"amount":           amount.String(),
					"source":           source,
				},
			})
		})
		return STKPushInitiated{}, darajaErr
	}

	if err := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
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
		return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
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
		})
	}); err != nil {
		return STKPushInitiated{}, err
	}

	pollScheduled := false
	if err := s.PollQueue.EnqueueSTKPoll(ctx, STKPollTask{
		SchoolID:          p.SchoolID,
		STKRequestID:      stkRequestID,
		CheckoutRequestID: darajaResp.CheckoutRequestID,
		EnqueuedAt:        s.now(),
	}, 45*time.Second); err != nil {
		// Daraja has already accepted the prompt; keep poll_scheduled=FALSE so the
		// stranded-request recovery worker can safely re-enqueue, and record a
		// mandatory audit event.
		if auditErr := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
			return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
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
		}); auditErr != nil {
			return STKPushInitiated{}, auditErr
		}
	} else {
		pollScheduled = true
		if markErr := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
			_, execErr := tx.Exec(ctx, `
				UPDATE mpesa_stk_requests
				SET poll_scheduled = TRUE, updated_at = NOW()
				WHERE school_id = $1 AND id = $2`,
				p.SchoolID, stkRequestID,
			)
			return execErr
		}); markErr != nil {
			log.Printf("billing: failed to mark poll_scheduled=true for stk_request_id=%s: %v", stkRequestID, markErr)
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
		PollScheduled:     pollScheduled,
	}, nil
}

// RetryUnscheduledSTKPolls recovers pending STK requests whose initial poll-queue
// publication failed (poll_scheduled = FALSE), re-enqueuing them and marking
// poll_scheduled = TRUE with an audit record.
func (s *Service) RetryUnscheduledSTKPolls(ctx context.Context, schoolID string) (int, error) {
	if !contracts.ValidID(schoolID) {
		return 0, contracts.Fail(http.StatusBadRequest, "INVALID_SCHOOL", "A valid school_id is required.")
	}
	if s == nil || s.PollQueue == nil {
		return 0, contracts.Fail(http.StatusServiceUnavailable, "MPESA_POLL_QUEUE_NOT_CONFIGURED", "M-Pesa polling queue is not configured.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return 0, err
	}

	type stranded struct {
		id         string
		checkoutID string
	}
	var pending []stranded
	if err := s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, COALESCE(checkout_request_id, '')
			FROM mpesa_stk_requests
			WHERE school_id = $1
			  AND status = 'pending'
			  AND poll_scheduled = FALSE
			  AND checkout_request_id IS NOT NULL
			ORDER BY created_at ASC, id ASC
			LIMIT 100`,
			schoolID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item stranded
			if err := rows.Scan(&item.id, &item.checkoutID); err != nil {
				return err
			}
			if item.checkoutID != "" {
				pending = append(pending, item)
			}
		}
		return rows.Err()
	}); err != nil {
		return 0, err
	}

	recovered := 0
	for _, item := range pending {
		if err := s.PollQueue.EnqueueSTKPoll(ctx, STKPollTask{
			SchoolID:          schoolID,
			STKRequestID:      item.id,
			CheckoutRequestID: item.checkoutID,
			EnqueuedAt:        s.now(),
		}, 0); err != nil {
			return recovered, err
		}
		if err := s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, `
				UPDATE mpesa_stk_requests
				SET poll_scheduled = TRUE, updated_at = NOW()
				WHERE school_id = $1 AND id = $2`,
				schoolID, item.id,
			); err != nil {
				return err
			}
			return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
				SchoolID:     schoolID,
				Action:       "billing.mpesa.stk_poll_recovered",
				ResourceType: "mpesa_stk_request",
				ResourceID:   item.id,
				Details: map[string]any{
					"checkout_request_id": item.checkoutID,
				},
			})
		}); err != nil {
			return recovered, err
		}
		recovered++
	}
	return recovered, nil
}

// isUserChatPaymentPaused checks users.chat_payment_paused_until and fails closed
// if the schema column is missing or the freeze state cannot be read.
func isUserChatPaymentPaused(ctx context.Context, tx pgx.Tx, userID string, now time.Time) (bool, error) {
	var pausedUntil *time.Time
	err := tx.QueryRow(ctx, `
		SELECT chat_payment_paused_until
		FROM users
		WHERE id = $1`,
		userID,
	).Scan(&pausedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Guardian account not found.")
	}
	if err != nil {
		log.Printf("billing: failed to read users.chat_payment_paused_until (failing closed): %v", err)
		return true, contracts.Fail(http.StatusServiceUnavailable, "CHAT_PAYMENT_FREEZE_UNAVAILABLE", "Unable to verify chat payment security freeze; chat payment initiation is temporarily unavailable.")
	}
	if pausedUntil == nil {
		return false, nil
	}
	return pausedUntil.After(now), nil
}

type DarajaSTKCallbackEnvelope struct {
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
// Tenant resolution comes strictly from trusted server-side state (either the
// per-request callback token hash bound to mpesa_stk_requests or the globally
// unique CheckoutRequestID in mpesa_stk_requests) — never from caller-supplied
// school_id parameters or BusinessShortCode alone.
func (s *Service) ProcessSTKCallback(ctx context.Context, callbackToken string, payload DarajaSTKCallbackEnvelope) error {
	if s == nil {
		return contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
	}
	if err := s.requireNotifier(); err != nil {
		return err
	}
	cb := payload.Body.StkCallback
	checkoutID := strings.TrimSpace(cb.CheckoutRequestID)
	if checkoutID == "" || cb.ResultCode == nil {
		return contracts.Fail(http.StatusBadRequest, "INVALID_MPESA_CALLBACK", "CheckoutRequestID and ResultCode are required.")
	}

	var schoolID string
	var expectedTokenHash string
	if trimmedToken := strings.TrimSpace(callbackToken); trimmedToken != "" {
		if s.ResolveCallbackToken == nil {
			return contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_CALLBACK", "Invalid or missing M-Pesa callback credentials.")
		}
		expectedTokenHash = HashCallbackToken(trimmedToken)
		resolvedSchool, expectedCheckout, err := s.ResolveCallbackToken(ctx, expectedTokenHash)
		if err != nil || !contracts.ValidID(resolvedSchool) || strings.TrimSpace(expectedCheckout) != checkoutID {
			return contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_CALLBACK", "Callback token does not match the STK request.")
		}
		schoolID = resolvedSchool
	} else {
		if s.ResolveCheckoutRequest == nil {
			return contracts.Fail(http.StatusInternalServerError, "CHECKOUT_RESOLVER_NOT_CONFIGURED", "STK checkout request resolver is not configured.")
		}
		resolvedSchool, err := s.ResolveCheckoutRequest(ctx, checkoutID)
		if err != nil || !contracts.ValidID(resolvedSchool) {
			return contracts.Fail(http.StatusNotFound, "STK_REQUEST_NOT_FOUND", "STK request not found.")
		}
		schoolID = resolvedSchool
	}

	var receiptNumber, phoneStr string
	var callbackAmount Money
	var hasAmount bool
	var amountInvalid bool
	if cb.CallbackMetadata != nil {
		for _, item := range cb.CallbackMetadata.Item {
			rawVal := strings.TrimSpace(string(item.Value))
			valStr := rawVal
			if strings.HasPrefix(rawVal, `"`) {
				var unquoted string
				if err := json.Unmarshal([]byte(rawVal), &unquoted); err == nil {
					valStr = strings.TrimSpace(unquoted)
				}
			}
			switch item.Name {
			case "MpesaReceiptNumber":
				receiptNumber = valStr
			case "PhoneNumber":
				phoneStr = valStr
			case "Amount":
				hasAmount = true
				parsed, err := ParseMoney(valStr)
				if err != nil || parsed <= 0 {
					amountInvalid = true
				} else {
					callbackAmount = parsed
				}
			}
		}
	}

	shortcode := strings.TrimSpace(payload.BusinessShortCode)
	merchantID := strings.TrimSpace(cb.MerchantRequestID)

	var rejectionReason string
	err := s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		var configuredShortcode string
		if err := tx.QueryRow(ctx, `
			SELECT shortcode FROM school_mpesa_configs WHERE school_id = $1`,
			schoolID,
		).Scan(&configuredShortcode); errors.Is(err, pgx.ErrNoRows) {
			rejectionReason = "mpesa_not_configured"
			return contracts.Fail(http.StatusBadRequest, "SHORTCODE_SCHOOL_MISMATCH", "M-Pesa shortcode is not configured for the resolved school.")
		} else if err != nil {
			return err
		}
		if shortcode != "" && !strings.EqualFold(strings.TrimSpace(configuredShortcode), shortcode) {
			rejectionReason = "shortcode_mismatch"
			return contracts.Fail(http.StatusBadRequest, "SHORTCODE_SCHOOL_MISMATCH", "BusinessShortCode does not match the school's configured M-Pesa shortcode.")
		}

		if expectedTokenHash != "" {
			var storedHash string
			if err := tx.QueryRow(ctx, `
				SELECT COALESCE(callback_token_hash, '')
				FROM mpesa_stk_requests
				WHERE school_id = $1 AND checkout_request_id = $2`,
				schoolID, checkoutID,
			).Scan(&storedHash); errors.Is(err, pgx.ErrNoRows) || subtle.ConstantTimeCompare([]byte(storedHash), []byte(expectedTokenHash)) != 1 {
				rejectionReason = "callback_token_mismatch"
				return contracts.Fail(http.StatusUnauthorized, "UNAUTHORIZED_CALLBACK", "Callback token does not match the STK request.")
			} else if err != nil {
				return err
			}
		}

		if *cb.ResultCode == 0 {
			if !hasAmount || amountInvalid || callbackAmount <= 0 {
				rejectionReason = "invalid_or_missing_amount"
				return contracts.Fail(http.StatusUnprocessableEntity, "MPESA_AMOUNT_MISMATCH", "Successful STK callback requires a valid positive Amount matching the initiated STK request.")
			}
		}

		settleErr := s.settleSTKResultTx(ctx, tx, schoolID, checkoutID, merchantID, *cb.ResultCode, receiptNumber, phoneStr, callbackAmount, true, payload)
		if settleErr != nil {
			var apiErr *contracts.APIError
			if errors.As(settleErr, &apiErr) {
				rejectionReason = apiErr.Code
			}
		}
		return settleErr
	})

	if err != nil && rejectionReason != "" {
		_ = s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
			return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
				SchoolID:     schoolID,
				Action:       "billing.mpesa.stk_callback_rejected",
				ResourceType: "mpesa_stk_request",
				ResourceID:   checkoutID,
				Details: map[string]any{
					"checkout_request_id": checkoutID,
					"reason":              rejectionReason,
					"result_code":         *cb.ResultCode,
				},
			})
		})
	}
	return err
}

func (s *Service) settleSTKResultTx(
	ctx context.Context,
	tx pgx.Tx,
	schoolID, checkoutID, merchantID string,
	resultCode int,
	receiptNumber, phoneStr string,
	callbackAmount Money,
	requireAmount bool,
	rawPayload any,
) error {
	var stkID, studentID, guardianUserID, reqPhone, currentStatus string
	var storedMerchant *string
	var reqAmount Money
	err := tx.QueryRow(ctx, `
		SELECT id, student_id, guardian_user_id, amount, phone_number, status, merchant_request_id
		FROM mpesa_stk_requests
		WHERE school_id = $1 AND checkout_request_id = $2
		FOR UPDATE`,
		schoolID, checkoutID,
	).Scan(&stkID, &studentID, &guardianUserID, &reqAmount, &reqPhone, &currentStatus, &storedMerchant)
	if errors.Is(err, pgx.ErrNoRows) {
		return contracts.Fail(http.StatusNotFound, "STK_REQUEST_NOT_FOUND", "STK request not found for this school.")
	}
	if err != nil {
		return err
	}

	if storedMerchant != nil && strings.TrimSpace(*storedMerchant) != "" && merchantID != "" && strings.TrimSpace(*storedMerchant) != merchantID {
		return contracts.Fail(http.StatusUnprocessableEntity, "MERCHANT_REQUEST_MISMATCH", "MerchantRequestID does not match the initiated STK request.")
	}

	// Preserve terminal succeeded state: duplicate success is an idempotent no-op,
	// and a subsequent failed callback must never overwrite a settled payment.
	if currentStatus == "succeeded" {
		if resultCode != 0 {
			return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
				SchoolID:     schoolID,
				Action:       "billing.mpesa.stk_callback_ignored_after_success",
				ResourceType: "mpesa_stk_request",
				ResourceID:   stkID,
				Details: map[string]any{
					"checkout_request_id": checkoutID,
					"result_code":         resultCode,
				},
			})
		}
		return nil
	}

	if resultCode != 0 {
		if currentStatus == "failed" || currentStatus == "timeout" {
			return nil
		}
		newStatus := "failed"
		if resultCode == 1037 {
			newStatus = "timeout"
		}
		if _, err := tx.Exec(ctx, `
			UPDATE mpesa_stk_requests
			SET status = $1, result_code = $2, updated_at = NOW()
			WHERE school_id = $3 AND id = $4`,
			newStatus, resultCode, schoolID, stkID,
		); err != nil {
			return err
		}
		return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
			SchoolID:     schoolID,
			Action:       "billing.mpesa.stk_" + newStatus,
			ResourceType: "mpesa_stk_request",
			ResourceID:   stkID,
			Details: map[string]any{
				"checkout_request_id": checkoutID,
				"result_code":         resultCode,
				"previous_status":     currentStatus,
			},
		})
	}

	// Provider-confirmed success (resultCode == 0) may settle initiating, pending,
	// timeout, uncertain, or earlier failed attempts.
	receipt := strings.TrimSpace(receiptNumber)
	if receipt == "" {
		return contracts.Fail(http.StatusUnprocessableEntity, "MISSING_MPESA_RECEIPT", "Successful STK callback requires an MpesaReceiptNumber.")
	}
	if requireAmount && callbackAmount <= 0 {
		return contracts.Fail(http.StatusUnprocessableEntity, "MPESA_AMOUNT_MISMATCH", "Successful STK callback requires a valid positive Amount matching the initiated STK request.")
	}
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

// PollSTKRequest queries Daraja for a pending/timeout/uncertain STK push request
// and updates its status idempotently without ever fabricating a receipt,
// double-crediting, or overwriting a confirmed successful settlement.
func (s *Service) PollSTKRequest(ctx context.Context, schoolID, stkRequestID string) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(stkRequestID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_STK_POLL", "Valid school_id and stk_request_id are required.")
	}
	if s == nil || s.DB == nil || s.Daraja == nil {
		return contracts.Fail(http.StatusServiceUnavailable, "MPESA_NOT_CONFIGURED", "M-Pesa polling dependencies are not configured.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
	}
	if err := s.requireNotifier(); err != nil {
		return err
	}

	var checkoutID, status string
	var cfg DarajaSchoolConfig
	err := s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(checkout_request_id, ''), status
			FROM mpesa_stk_requests
			WHERE school_id = $1 AND id = $2`,
			schoolID, stkRequestID,
		).Scan(&checkoutID, &status); err != nil {
			return err
		}
		if (status != "pending" && status != "timeout" && status != "uncertain") || checkoutID == "" {
			return nil
		}
		return tx.QueryRow(ctx, `
			SELECT school_id, shortcode, environment, COALESCE(passkey_encrypted, '')
			FROM school_mpesa_configs
			WHERE school_id = $1`,
			schoolID,
		).Scan(&cfg.SchoolID, &cfg.Shortcode, &cfg.Environment, &cfg.Passkey)
	})
	if err != nil || (status != "pending" && status != "timeout" && status != "uncertain") || checkoutID == "" {
		return err
	}

	queryResp, err := s.Daraja.QuerySTKStatus(ctx, DarajaSTKQueryRequest{
		Config:            cfg,
		CheckoutRequestID: checkoutID,
	})
	if err != nil {
		return err
	}

	return s.withSystemTenantTx(ctx, schoolID, func(tx pgx.Tx) error {
		var currentStatus string
		if err := tx.QueryRow(ctx, `
			SELECT status
			FROM mpesa_stk_requests
			WHERE school_id = $1 AND id = $2
			FOR UPDATE`,
			schoolID, stkRequestID,
		).Scan(&currentStatus); err != nil {
			return err
		}
		if currentStatus == "succeeded" {
			return nil
		}

		if queryResp.ResultCode == 0 {
			if strings.TrimSpace(queryResp.ReceiptNumber) == "" {
				// Never fabricate a receipt if Daraja query does not include one;
				// mark uncertain and await callback or C2B reconciliation.
				if _, err := tx.Exec(ctx, `
					UPDATE mpesa_stk_requests
					SET status = 'uncertain', result_code = 0, updated_at = NOW()
					WHERE school_id = $1 AND id = $2 AND status <> 'succeeded'`,
					schoolID, stkRequestID,
				); err != nil {
					return err
				}
				return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
					SchoolID:     schoolID,
					Action:       "billing.mpesa.stk_poll_uncertain",
					ResourceType: "mpesa_stk_request",
					ResourceID:   stkRequestID,
					Details: map[string]any{
						"checkout_request_id": checkoutID,
						"result_code":         0,
					},
				})
			}
			// Daraja STK Query responses do not carry an Amount field, so
			// requireAmount is false here while exact requested amount is settled.
			return s.settleSTKResultTx(ctx, tx, schoolID, checkoutID, queryResp.MerchantRequestID, 0, queryResp.ReceiptNumber, "", 0, false, queryResp)
		}
		newStatus := "failed"
		if queryResp.ResultCode == 1037 {
			newStatus = "timeout"
		}
		if _, err := tx.Exec(ctx, `
			UPDATE mpesa_stk_requests
			SET status = $1, result_code = $2, updated_at = NOW()
			WHERE school_id = $3 AND id = $4 AND status <> 'succeeded'`,
			newStatus, queryResp.ResultCode, schoolID, stkRequestID,
		); err != nil {
			return err
		}
		return s.logAuditPgx(ctx, tx, contracts.AuditEvent{
			SchoolID:     schoolID,
			Action:       "billing.mpesa.stk_poll_" + newStatus,
			ResourceType: "mpesa_stk_request",
			ResourceID:   stkRequestID,
			Details: map[string]any{
				"checkout_request_id": checkoutID,
				"result_code":         queryResp.ResultCode,
				"previous_status":     currentStatus,
			},
		})
	})
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
		var existingStudentID *string
		var existingStatus string
		lookupErr := tx.QueryRow(ctx, `
			SELECT student_id, matched_status
			FROM payments
			WHERE school_id = $1 AND transaction_ref = $2`,
			schoolID, receipt,
		).Scan(&existingStudentID, &existingStatus)
		if lookupErr == nil && existingStatus == "matched" && existingStudentID != nil && *existingStudentID == studentID {
			return nil
		}
		return contracts.Fail(http.StatusConflict, "DUPLICATE_PAYMENT_REFERENCE", "Transaction reference already exists.")
	}
	if err != nil {
		return err
	}

	if err := s.RecordStudentPaymentTx(ctx, tx, schoolID, studentID, receipt, "M-Pesa payment ("+receipt+")", amount); err != nil {
		return err
	}

	if err := s.logAuditPgx(ctx, tx, contracts.AuditEvent{
		SchoolID:     schoolID,
		Action:       "billing.mpesa.payment_matched",
		ResourceType: "payment",
		ResourceID:   paymentID,
		Details: map[string]any{
			"student_id":      studentID,
			"transaction_ref": receipt,
			"provider":        provider,
			"amount":          amount.String(),
		},
	}); err != nil {
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
