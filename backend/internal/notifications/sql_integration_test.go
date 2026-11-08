package notifications

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

func TestSQLNotificationsKeepTenantContext(t *testing.T) {
	url := os.Getenv("DEV5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("disposable PostgreSQL database and registered driver required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewSQLStore(db)
	ctx := context.Background()
	school := "00000000-0000-4000-8000-000000000001"
	guardian := "00000000-0000-4000-8000-000000000300"
	n := &InAppNotification{SchoolID: school, RecipientUserID: guardian, Category: "synthetic-test", Title: "Synthetic notice", Body: "Synthetic test content"}
	if err := store.CreateInAppNotification(ctx, n); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListInAppNotifications(ctx, school, guardian, false, 50)
	if err != nil || len(items) == 0 {
		t.Fatalf("notification missing: %+v %v", items, err)
	}
	items, err = store.ListInAppNotifications(ctx, "00000000-0000-4000-8000-000000000002", guardian, false, 50)
	if err != nil || len(items) != 0 {
		t.Fatalf("cross-school read: %+v %v", items, err)
	}
	if err := store.MarkNotificationRead(ctx, "00000000-0000-4000-8000-000000000002", n.ID, guardian); err == nil {
		t.Fatal("cross-school update accepted")
	}
	if err := store.MarkNotificationRead(ctx, school, n.ID, guardian); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM in_app_notifications`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("tenant context escaped transaction: %d %v", count, err)
	}
}
