package billing

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// STKInitiateParams holds arguments for initiating Daraja STK Push
type STKInitiateParams struct {
	Shortcode        string
	Amount           float64
	PhoneNumber      string // 2547XXXXXXXX
	AccountReference string
	TransactionDesc  string
	CallBackURL      string
}

// DarajaSTKResponse represents Safaricom's response to an STK push request
type DarajaSTKResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// DarajaQueryResponse represents Safaricom's response to STK query request
type DarajaQueryResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResultCode          string `json:"ResultCode"`
	ResultDesc          string `json:"ResultDesc"`
}

// DarajaClient abstracts communications with Safaricom Daraja API
type DarajaClient interface {
	InitiateSTKPush(ctx context.Context, cfg *SchoolMpesaConfig, params STKInitiateParams) (*DarajaSTKResponse, error)
	QuerySTKStatus(ctx context.Context, cfg *SchoolMpesaConfig, checkoutRequestID string) (*DarajaQueryResponse, error)
}

// HTTPDarajaClient communicates with Safaricom Daraja endpoints over HTTP
type HTTPDarajaClient struct {
	httpClient *http.Client
}

// NewHTTPDarajaClient creates a new HTTP Daraja client
func NewHTTPDarajaClient(httpClient *http.Client) *HTTPDarajaClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &HTTPDarajaClient{httpClient: httpClient}
}

func (c *HTTPDarajaClient) baseURL(env string) string {
	if strings.ToLower(env) == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (c *HTTPDarajaClient) getAccessToken(ctx context.Context, baseURL, consumerKey, consumerSecret string) (string, error) {
	authURL := baseURL + "/oauth/v1/generate?grant_type=client_credentials"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(consumerKey, consumerSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("daraja oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("daraja oauth returned status %d: %s", resp.StatusCode, string(b))
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode daraja oauth response: %w", err)
	}
	if authResp.AccessToken == "" {
		return "", errors.New("daraja oauth returned empty access token")
	}
	return authResp.AccessToken, nil
}

func (c *HTTPDarajaClient) InitiateSTKPush(ctx context.Context, cfg *SchoolMpesaConfig, params STKInitiateParams) (*DarajaSTKResponse, error) {
	if cfg == nil {
		return nil, errors.New("school mpesa config is required")
	}
	base := c.baseURL(cfg.Environment)

	// If consumer keys are missing, reject
	if cfg.ConsumerKeyEncrypted == "" || cfg.ConsumerSecretEncrypted == "" {
		return nil, errors.New("daraja credentials (consumer key and secret) are not configured for this school")
	}

	token, err := c.getAccessToken(ctx, base, cfg.ConsumerKeyEncrypted, cfg.ConsumerSecretEncrypted)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	passkey := cfg.PasskeyEncrypted
	if passkey == "" {
		passkey = "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919" // Standard Daraja sandbox test passkey
	}
	password := base64.StdEncoding.EncodeToString([]byte(params.Shortcode + passkey + timestamp))

	payload := map[string]interface{}{
		"BusinessShortCode": params.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            int(params.Amount),
		"PartyA":            params.PhoneNumber,
		"PartyB":            params.Shortcode,
		"PhoneNumber":       params.PhoneNumber,
		"CallBackURL":       params.CallBackURL,
		"AccountReference":  params.AccountReference,
		"TransactionDesc":   params.TransactionDesc,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	stkURL := base + "/mpesa/stkpush/v1/processrequest"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, stkURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daraja stk push request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read daraja stk push response: %w", err)
	}

	var stkResp DarajaSTKResponse
	if err := json.Unmarshal(respBody, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse daraja stk push response: %s: %w", string(respBody), err)
	}

	if stkResp.CheckoutRequestID == "" {
		return nil, fmt.Errorf("daraja stk push rejected: %s", string(respBody))
	}

	return &stkResp, nil
}

func (c *HTTPDarajaClient) QuerySTKStatus(ctx context.Context, cfg *SchoolMpesaConfig, checkoutRequestID string) (*DarajaQueryResponse, error) {
	if cfg == nil {
		return nil, errors.New("school mpesa config is required")
	}
	base := c.baseURL(cfg.Environment)

	token, err := c.getAccessToken(ctx, base, cfg.ConsumerKeyEncrypted, cfg.ConsumerSecretEncrypted)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	passkey := cfg.PasskeyEncrypted
	if passkey == "" {
		passkey = "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919"
	}
	password := base64.StdEncoding.EncodeToString([]byte(cfg.Shortcode + passkey + timestamp))

	payload := map[string]interface{}{
		"BusinessShortCode": cfg.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": checkoutRequestID,
	}

	bodyBytes, _ := json.Marshal(payload)
	queryURL := base + "/mpesa/stkpushquery/v1/query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daraja stk query request failed: %w", err)
	}
	defer resp.Body.Close()

	var queryResp DarajaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, err
	}
	return &queryResp, nil
}
