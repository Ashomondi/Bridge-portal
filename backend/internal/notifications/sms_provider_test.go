package notifications

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func providerDispatcher(t *testing.T, server *httptest.Server) *africaTalkingDispatcher {
	t.Helper()
	d, err := NewSMSDispatcher(SMSConfig{Username: "synthetic-user", APIKey: "synthetic-secret", SenderID: "IKONEX"}, server.Client(), nil)
	if err != nil {
		t.Fatal(err)
	}
	provider := d.(*africaTalkingDispatcher)
	provider.endpoint = server.URL
	return provider
}

func TestAfricaTalkingRequestAndResponses(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		wantCode   string
		retry      bool
	}{
		{"sent", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":101}]}}`, 201, "", false},
		{"queued", `{"SMSMessageData":{"Recipients":[{"number":"254712345678","statusCode":102}]}}`, 201, "", false},
		{"processed", `{"SMSMessageData":{"Recipients":[{"number":"254712345678","statusCode":100}]}}`, 201, "", false},
		{"invalid phone", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":403,"status":"private 654321"}]}}`, 201, "PROVIDER_REJECTED", false},
		{"gateway failure", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":501}]}}`, 201, "PROVIDER_REJECTED", true},
		{"wrong recipient", `{"SMSMessageData":{"Recipients":[{"number":"+254712345679","statusCode":101}]}}`, 201, "RECIPIENT_MISMATCH", false},
		{"no recipients", `{"SMSMessageData":{"Recipients":[]}}`, 201, "INVALID_RESPONSE", false},
		{"multiple recipients", `{"SMSMessageData":{"Recipients":[{},{}]}}`, 201, "INVALID_RESPONSE", false},
		{"malformed", `{"secret":"654321"`, 200, "INVALID_RESPONSE", false},
		{"oversized", strings.Repeat("x", (64<<10)+1), 200, "INVALID_RESPONSE", false},
		{"rate limited", "private 654321", 429, "HTTP_ERROR", true},
		{"request timeout", "private 654321", 408, "HTTP_ERROR", true},
		{"server failure", "private 654321", 503, "HTTP_ERROR", true},
		{"unauthorized", "synthetic-secret", 401, "HTTP_ERROR", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.Header.Get("apiKey") != "synthetic-secret" || r.Header.Get("Accept") != "application/json" {
					t.Error("incorrect provider request headers or method")
				}
				if err := r.ParseForm(); err != nil {
					t.Error(err)
				}
				for key, want := range map[string]string{"username": "synthetic-user", "to": "+254712345678", "from": "IKONEX", "message": "Synthetic OTP 654321 & = +"} {
					if r.Form.Get(key) != want {
						t.Errorf("incorrect form field %s", key)
					}
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			d := providerDispatcher(t, server)
			err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP 654321 & = +", Purpose: AppLoginOTP})
			if tc.wantCode == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var deliveryErr *SMSDeliveryError
			if !errors.As(err, &deliveryErr) || deliveryErr.Code != tc.wantCode || deliveryErr.Retryable != tc.retry {
				t.Fatalf("got %v; want %s retry=%v", err, tc.wantCode, tc.retry)
			}
			for _, secret := range []string{"654321", "synthetic-secret", "712345678"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatal("provider error exposed private data")
				}
			}
		})
	}
}

func TestSMSFactorySandboxDoesNotUseHTTP(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	var log bytes.Buffer
	d, err := NewSMSDispatcher(SMSConfig{Username: "sandbox"}, server.Client(), slog.New(slog.NewJSONHandler(&log, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.(*sandboxSMSDispatcher); !ok {
		t.Fatal("sandbox selected a provider client")
	}
	if err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP}); err != nil || calls.Load() != 0 || log.Len() == 0 {
		t.Fatalf("sandbox did not log locally: %v", err)
	}
	for _, config := range []SMSConfig{{}, {Username: "synthetic-user"}, {APIKey: "secret"}} {
		if _, err := NewSMSDispatcher(config, nil, nil); err == nil {
			t.Fatal("incomplete provider configuration accepted")
		}
	}
}

func TestAfricaTalkingRejectsRedirects(t *testing.T) {
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected.Add(1) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	d := providerDispatcher(t, server)
	err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP})
	var deliveryErr *SMSDeliveryError
	if !errors.As(err, &deliveryErr) || deliveryErr.Retryable || redirected.Load() != 0 {
		t.Fatalf("redirect was followed or classified incorrectly: %v", err)
	}
}

func TestAfricaTalkingCancellationAndTimeout(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		<-release
	}))
	defer server.Close()
	defer close(release)
	d := providerDispatcher(t, server)
	msg := SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := d.DispatchSMS(ctx, msg); !errors.Is(err, context.Canceled) || calls.Load() != 0 {
		t.Fatalf("canceled dispatch reached provider: %v", err)
	}
	msg.Purpose = "absence"
	if err := d.DispatchSMS(context.Background(), msg); !errors.Is(err, ErrInvalidSMS) || calls.Load() != 0 {
		t.Fatalf("invalid dispatch reached provider: %v", err)
	}
	msg.Purpose = AppLoginOTP
	d.client.Timeout = 30 * time.Millisecond
	var deliveryErr *SMSDeliveryError
	if err := d.DispatchSMS(context.Background(), msg); !errors.As(err, &deliveryErr) || !deliveryErr.Retryable {
		t.Fatalf("timeout retry classification: %v", err)
	}
}

func TestSMSFactoryDoesNotMutateHTTPClient(t *testing.T) {
	client := &http.Client{}
	dispatcher, err := NewSMSDispatcher(SMSConfig{Username: "synthetic-user", APIKey: "synthetic-secret"}, client, nil)
	if err != nil {
		t.Fatal(err)
	}
	d := dispatcher.(*africaTalkingDispatcher)
	if client.Timeout != 0 || client.CheckRedirect != nil || d.client == client || d.client.Timeout <= 0 {
		t.Fatal("constructor mutated shared client or left timeout unset")
	}
	if got := fmt.Sprint(d.client.CheckRedirect(nil, nil)); got != http.ErrUseLastResponse.Error() {
		t.Fatalf("redirect policy: %s", got)
	}
}
