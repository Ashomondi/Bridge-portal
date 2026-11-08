package notifications

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestInAppNotificationsLifecycle(t *testing.T) {
	store := NewMemoryStore()
	sms := NewAfricasTalkingSMSClient("sandbox", "test-key", "IKONEX")
	svc := NewService(store, sms)
	ctx := context.Background()

	schoolID := "school-1"
	userID := "user-parent-1"

	// 1. Create in-app notifications
	err := svc.CreateInAppNotification(ctx, schoolID, userID, "fee_change", "Fee Update", "Fee has been adjusted", map[string]interface{}{"delta": 2000})
	if err != nil {
		t.Fatalf("CreateInAppNotification failed: %v", err)
	}

	err = svc.CreateInAppNotification(ctx, schoolID, userID, "payment_receipt", "Receipt", "Payment received", nil)
	if err != nil {
		t.Fatalf("CreateInAppNotification failed: %v", err)
	}

	// 2. List unread notifications
	items, err := store.ListInAppNotifications(ctx, schoolID, userID, true, 10)
	if err != nil || len(items) != 2 {
		t.Fatalf("expected 2 unread notifications, got %d", len(items))
	}

	// 3. Mark first notification as read
	firstID := items[0].ID
	err = store.MarkNotificationRead(ctx, schoolID, firstID, userID)
	if err != nil {
		t.Fatalf("MarkNotificationRead failed: %v", err)
	}

	// 4. Verify unread count is now 1
	unread, err := store.ListInAppNotifications(ctx, schoolID, userID, true, 10)
	if err != nil || len(unread) != 1 {
		t.Fatalf("expected 1 unread notification, got %d", len(unread))
	}
	if unread[0].ID == firstID {
		t.Errorf("read notification was still returned in unread list")
	}

	// 5. Test SMS Dispatcher sandbox logging (C12, C58)
	err = sms.SendTransactionalSMS(ctx, "+254712345678", "Your login OTP is 654321")
	if err != nil {
		t.Fatalf("SendTransactionalSMS failed: %v", err)
	}
	if len(sms.SentSMS) != 1 || sms.SentSMS[0].To != "+254712345678" {
		t.Errorf("SMS not logged properly in sandbox client")
	}
}

type smsRoundTripper func(*http.Request) (*http.Response, error)

func (fn smsRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func TestAfricasTalkingProductionDelivery(t *testing.T) {
	for _, tc := range []struct {
		name           string
		status         int
		body           string
		transportError bool
		wantError      bool
	}{
		{"accepted", 201, `{"SMSMessageData":{"Recipients":[{"number":"+254700000001","statusCode":101}]}}`, false, false},
		{"recipient rejected", 201, `{"SMSMessageData":{"Recipients":[{"number":"+254700000001","statusCode":401}]}}`, false, true},
		{"wrong recipient", 201, `{"SMSMessageData":{"Recipients":[{"number":"+254700000002","statusCode":101}]}}`, false, true},
		{"empty recipients", 201, `{"SMSMessageData":{"Recipients":[]}}`, false, true},
		{"malformed", 201, `broken`, false, true},
		{"provider error", 401, `{"error":"provider-secret"}`, false, true},
		{"transport error", 0, "", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := NewAfricasTalkingSMSClient("synthetic-app", "synthetic-key", "IKONEX")
			client.httpClient = &http.Client{Transport: smsRoundTripper(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://api.africastalking.com/version1/messaging" || r.Method != http.MethodPost || r.Header.Get("apiKey") != "synthetic-key" || r.Header.Get("Accept") != "application/json" {
					t.Fatal("incorrect provider request")
				}
				if err := r.ParseForm(); err != nil {
					t.Fatal(err)
				}
				if r.Form.Get("to") != "+254700000001" || r.Form.Get("message") != "Synthetic OTP 123456" || r.Form.Get("username") != "synthetic-app" || r.Form.Get("from") != "IKONEX" {
					t.Fatal("incorrect SMS form")
				}
				if tc.transportError {
					return nil, errors.New("provider-secret")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			err := client.SendTransactionalSMS(context.Background(), "+254700000001", "Synthetic OTP 123456")
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v", err)
			}
			if err != nil && strings.Contains(err.Error(), "provider-secret") {
				t.Fatal("sensitive provider information exposed")
			}
			if len(client.SentSMS) != 0 {
				t.Fatal("production delivery recorded in sandbox history")
			}
		})
	}
}

func TestAfricasTalkingRejectsMissingConfigurationAndInput(t *testing.T) {
	for _, tc := range []struct{ key, to, message string }{
		{"", "+254700000001", "OTP"}, {"key", "", "OTP"}, {"key", "+254700000001", ""},
	} {
		client := NewAfricasTalkingSMSClient("synthetic-app", tc.key, "")
		if err := client.SendTransactionalSMS(context.Background(), tc.to, tc.message); err == nil {
			t.Fatal("expected validation failure")
		}
	}
}
