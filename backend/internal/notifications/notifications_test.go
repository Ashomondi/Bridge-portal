package notifications

import (
	"context"
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
