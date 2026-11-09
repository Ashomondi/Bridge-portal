package notifications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

func TestCreateNotificationCommitsOrRollsBackWithCaller(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	s := &notifications.Service{DB: db}
	ctx := context.Background()
	input := notifications.CreateNotification{RecipientUserID: a.UserID, Category: "fee_change", Title: "Fee change", Body: "Synthetic fee notice", Metadata: map[string]any{"student_id": a.StudentID}}
	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("caller transaction failed")
	err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
		if err := s.CreateInAppNotification(ctx, tx, a.ID, input); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("caller rollback not propagated: %v", err)
	}
	var count int
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", a.ID, a.UserID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("partial or duplicate notice persisted: %d %v", count, err)
	}
	input.RecipientUserID = b.UserID
	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err == nil {
		t.Fatal("cross-school recipient accepted")
	}
	input.RecipientUserID = a.UserID
	if err := db.WithTenantTx(ctx, b.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err == nil {
		t.Fatal("cross-school transaction accepted")
	}
	input.Title = " "
	if err := s.CreateInAppNotification(ctx, nil, a.ID, input); err == nil {
		t.Fatal("blank notification accepted")
	}
}

func TestNotificationRoutesScopePaginationAndReadOwnership(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	ctx := context.Background()
	var ownID, otherUserID, otherSchoolID string
	for _, row := range []struct {
		school, user string
		id           *string
	}{{a.ID, a.UserID, &ownID}, {a.ID, b.UserID, &otherUserID}, {b.ID, b.UserID, &otherSchoolID}} {
		if err := db.Admin.QueryRow(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Synthetic','Synthetic') RETURNING id", row.school, row.user).Scan(row.id); err != nil {
			t.Fatal(err)
		}
	}
	s := &notifications.Service{DB: db}
	p := contracts.Principal{UserID: a.UserID, SchoolID: a.ID, MembershipID: a.MembershipID, Role: "parent"}
	mux := http.NewServeMux()
	notifications.RegisterRoutes(mux, s, func(*http.Request) (contracts.Principal, error) { return p, nil })
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications?per_page=1", nil))
	var result struct {
		Data []notifications.Notification `json:"data"`
		Meta contracts.Page               `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != 200 || len(result.Data) != 1 || result.Meta.Total != 1 || result.Data[0].ID != ownID {
		t.Fatalf("list scoping/pagination: %d %s %v", w.Code, w.Body.String(), err)
	}
	for _, id := range []string{otherUserID, otherSchoolID} {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("PATCH", "/api/v1/notifications/"+id+"/read", nil))
		if w.Code != 404 {
			t.Fatalf("another user's notice changed: %d", w.Code)
		}
	}
	var firstRead time.Time
	for i := 0; i < 2; i++ {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("PATCH", "/api/v1/notifications/"+ownID+"/read", nil))
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		var readAt time.Time
		if err := db.Admin.QueryRow(ctx, "SELECT read_at FROM in_app_notifications WHERE school_id=$1 AND id=$2 AND is_read", a.ID, ownID).Scan(&readAt); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			firstRead = readAt
		} else if !readAt.Equal(firstRead) {
			t.Fatal("repeat read replaced original timestamp")
		}
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications?page=2&per_page=1", nil))
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Data) != 0 || result.Meta.Total != 1 {
		t.Fatal(w.Body.String())
	}
	unauthenticated := http.NewServeMux()
	notifications.RegisterRoutes(unauthenticated, s, nil)
	w = httptest.NewRecorder()
	unauthenticated.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications", nil))
	if w.Code != 401 {
		t.Fatal("unauthenticated list was allowed")
	}
}
