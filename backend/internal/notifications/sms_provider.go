package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SMSConfig struct {
	Username string
	APIKey   string
	SenderID string
}

// SMSDeliveryError exposes retry policy without provider text, credentials,
// message bodies, or recipient numbers. Retries belong to the queue worker.
type SMSDeliveryError struct {
	Code      string
	Retryable bool
}

func (e *SMSDeliveryError) Error() string { return "SMS delivery failed: " + e.Code }

type africaTalkingDispatcher struct {
	config   SMSConfig
	client   *http.Client
	endpoint string
}

// NewSMSDispatcher selects the local logger for username "sandbox". The HTTP
// client is copied so timeouts and redirect policy do not mutate a shared client.
func NewSMSDispatcher(config SMSConfig, client *http.Client, logger *slog.Logger) (SMSDispatcher, error) {
	config.Username = strings.TrimSpace(config.Username)
	if config.Username == "sandbox" {
		return NewSandboxSMSDispatcher(logger), nil
	}
	if config.Username == "" || strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New("Africa's Talking username and API key are required")
	}
	httpClient := http.Client{Timeout: 15 * time.Second}
	if client != nil {
		httpClient = *client
		if httpClient.Timeout <= 0 {
			httpClient.Timeout = 15 * time.Second
		}
	}
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &africaTalkingDispatcher{
		config: config, client: &httpClient,
		endpoint: "https://api.africastalking.com/version1/messaging",
	}, nil
}

func (d *africaTalkingDispatcher) DispatchSMS(ctx context.Context, message SMSMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message, err := message.normalized()
	if err != nil {
		return err
	}
	form := url.Values{"username": {d.config.Username}, "to": {message.To}, "message": {message.Body}}
	if d.config.SenderID != "" {
		form.Set("from", d.config.SenderID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return &SMSDeliveryError{Code: "INVALID_REQUEST"}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apiKey", d.config.APIKey)
	resp, err := d.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &SMSDeliveryError{Code: "TRANSPORT_ERROR", Retryable: true}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &SMSDeliveryError{Code: "HTTP_ERROR", Retryable: resp.StatusCode == 429 || resp.StatusCode == 408 || resp.StatusCode >= 500}
	}
	const maxResponseBytes = 64 << 10
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return &SMSDeliveryError{Code: "INVALID_RESPONSE"}
	}
	var response struct {
		Data struct {
			Recipients []struct {
				Number     string `json:"number"`
				StatusCode int    `json:"statusCode"`
			} `json:"Recipients"`
		} `json:"SMSMessageData"`
	}
	if err := json.Unmarshal(body, &response); err != nil || len(response.Data.Recipients) != 1 {
		return &SMSDeliveryError{Code: "INVALID_RESPONSE"}
	}
	recipient := response.Data.Recipients[0]
	phone, err := NormalizePhone(recipient.Number)
	if err != nil || phone != message.To {
		return &SMSDeliveryError{Code: "RECIPIENT_MISMATCH"}
	}
	switch recipient.StatusCode {
	case 100, 101, 102: // Processed, sent, or queued by the provider.
		return nil
	default:
		return &SMSDeliveryError{Code: "PROVIDER_REJECTED", Retryable: recipient.StatusCode >= 500 && recipient.StatusCode <= 502}
	}
}
