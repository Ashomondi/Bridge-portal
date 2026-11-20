package consent

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

func openConsentTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the consent integration tests")
	}
	db, err := database.Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func postConsent(t *testing.T, service *Service, identity middleware.Identity, studentID string, granted bool) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"student_id": studentID,
		"purpose":    "sms",
		"granted":    granted,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/consents", bytes.NewReader(payload))
	req = req.WithContext(middleware.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	service.handleUpsert(rec, req)
	return rec
}

func TestIntegrationConsentUpsertAndList(t *testing.T) {
	db := openConsentTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suffix := time.Now().UnixNano()
	var schoolID, userID, studentID string

	if err := db.QueryRowContext(ctx,
		`INSERT INTO schools (code, name) VALUES ($1, $2) RETURNING id::text`,
		fmt.Sprintf("CONSENT-%d", suffix), "Consent Test School").Scan(&schoolID); err != nil {
		t.Fatalf("seed school: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM schools WHERE id = $1`, schoolID)
	})

	if err := db.QueryRowContext(ctx,
		`INSERT INTO users (phone_number, full_name) VALUES ($1, $2) RETURNING id::text`,
		fmt.Sprintf("+2547%08d", suffix%100000000), "Consent Test Parent").Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	err := database.WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO students (school_id, admission_number, full_name, dob, gender)
			 VALUES ($1, $2, 'Consent Test Student', '2011-01-01', 'F')
			 RETURNING id::text`,
			schoolID, fmt.Sprintf("ADM-%d", suffix)).Scan(&studentID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO student_guardians (school_id, student_id, user_id, relationship)
			 VALUES ($1, $2, $3, 'parent')`,
			schoolID, studentID, userID)
		return err
	})
	if err != nil {
		t.Fatalf("seed student and guardian: %v", err)
	}

	service := NewService(db.DB, nil)
	identity := middleware.Identity{UserID: userID, SchoolID: schoolID}

	if rec := postConsent(t, service, identity, studentID, true); rec.Code != http.StatusOK {
		t.Fatalf("grant status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if rec := postConsent(t, service, identity, studentID, false); rec.Code != http.StatusOK {
		t.Fatalf("revoke status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var count int
	var granted, revoked bool
	if err := database.WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*), bool_and(granted), bool_or(revoked_at IS NOT NULL)
			   FROM consents
			  WHERE school_id = current_school_id() AND student_id = $1 AND guardian_user_id = $2 AND purpose = 'sms'`,
			studentID, userID).Scan(&count, &granted, &revoked)
	}); err != nil {
		t.Fatalf("read consent: %v", err)
	}
	if count != 1 {
		t.Fatalf("consent rows = %d, want 1", count)
	}
	if granted {
		t.Fatal("granted = true after revoke")
	}
	if !revoked {
		t.Fatal("revoked_at was not set after revoke")
	}

	if rec := postConsent(t, service, identity, studentID, true); rec.Code != http.StatusOK {
		t.Fatalf("re-grant status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if err := database.WithTenantTx(ctx, db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT COUNT(*), bool_and(granted), bool_or(revoked_at IS NOT NULL)
			   FROM consents
			  WHERE school_id = current_school_id() AND student_id = $1 AND guardian_user_id = $2 AND purpose = 'sms'`,
			studentID, userID).Scan(&count, &granted, &revoked)
	}); err != nil {
		t.Fatalf("read consent after re-grant: %v", err)
	}
	if count != 1 || !granted || revoked {
		t.Fatalf("after re-grant count = %d, granted = %v, revoked = %v; want 1, true, false", count, granted, revoked)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/consents", nil)
	req = req.WithContext(middleware.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	service.handleList(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Status string            `json:"status"`
		Data   []consentResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("list returned %d consents, want 1", len(envelope.Data))
	}
	if !envelope.Data[0].Granted {
		t.Fatal("list returned a revoked consent")
	}
	if envelope.Data[0].RevokedAt != "" {
		t.Fatalf("list returned revoked_at = %q, want empty", envelope.Data[0].RevokedAt)
	}
}
