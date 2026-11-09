package onboarding

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

// TestProvisionSchoolAndCompleteWizard exercises the school bootstrap and setup
// wizard against a real database. It runs only when TEST_DATABASE_DSN is set
// and the migrations have been applied (AC-BE-A02).
func TestProvisionSchoolAndCompleteWizard(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the onboarding integration test")
	}

	db, err := database.Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	service := NewService(db.DB, database.NewAuditLogger(db.DB))
	suffix := time.Now().UnixNano()

	createBody := validCreateSchoolRequest()
	createBody.Code = fmt.Sprintf("IT-%d", suffix)
	createBody.ContactEmail = fmt.Sprintf("office%d@school.ac.ke", suffix)

	createRecorder := httptest.NewRecorder()
	service.handleCreateSchool(createRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/admin/schools", jsonBody(t, createBody)))
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create school status = %d, body = %s", createRecorder.Code, createRecorder.Body.String())
	}

	var created struct {
		Data struct {
			SchoolID        string `json:"school_id"`
			PrincipalUserID string `json:"principal_user_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	schoolID := created.Data.SchoolID
	principalID := created.Data.PrincipalUserID
	if schoolID == "" || principalID == "" {
		t.Fatalf("create response missing ids: %s", createRecorder.Body.String())
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM audit_log WHERE school_id = $1`, schoolID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM schools WHERE id = $1`, schoolID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, principalID)
	})

	// The school must be provisioned with a profile and the default badges.
	var profileCount, badgeCount int
	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT COUNT(*) FROM school_profile WHERE school_id = $1`, schoolID).Scan(&profileCount); err != nil {
			return err
		}
		return tx.QueryRow(`SELECT COUNT(*) FROM badges WHERE school_id = $1`, schoolID).Scan(&badgeCount)
	}); err != nil {
		t.Fatalf("verify provision: %v", err)
	}
	if profileCount != 1 {
		t.Fatalf("school_profile count = %d, want 1", profileCount)
	}
	if badgeCount != len(defaultBadges) {
		t.Fatalf("badge count = %d, want %d", badgeCount, len(defaultBadges))
	}

	identity := middleware.Identity{UserID: principalID, SchoolID: schoolID, Role: "principal"}
	wizardRecorder := httptest.NewRecorder()
	wizardRequest := httptest.NewRequest(http.MethodPost, "/api/v1/schools/setup-wizard", jsonBody(t, validSetupWizardRequest()))
	wizardRequest = wizardRequest.WithContext(middleware.WithIdentity(context.Background(), identity))
	service.handleSetupWizard(wizardRecorder, wizardRequest)
	if wizardRecorder.Code != http.StatusOK {
		t.Fatalf("setup wizard status = %d, body = %s", wizardRecorder.Code, wizardRecorder.Body.String())
	}

	var years, terms, grades, classes, dormitories int
	var wizardCompleted bool
	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT wizard_completed FROM school_profile WHERE school_id = $1`, schoolID).Scan(&wizardCompleted); err != nil {
			return err
		}
		checks := []struct {
			query string
			dest  *int
		}{
			{`SELECT COUNT(*) FROM academic_years WHERE school_id = $1`, &years},
			{`SELECT COUNT(*) FROM terms WHERE school_id = $1`, &terms},
			{`SELECT COUNT(*) FROM grade_levels WHERE school_id = $1`, &grades},
			{`SELECT COUNT(*) FROM classes WHERE school_id = $1`, &classes},
			{`SELECT COUNT(*) FROM dormitories WHERE school_id = $1`, &dormitories},
		}
		for _, check := range checks {
			if err := tx.QueryRow(check.query, schoolID).Scan(check.dest); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("verify wizard: %v", err)
	}

	if !wizardCompleted {
		t.Fatalf("wizard_completed = false, want true")
	}
	if years != 1 || terms != 3 || grades != 4 || classes != 1 || dormitories != 1 {
		t.Fatalf("unexpected counts: years=%d terms=%d grades=%d classes=%d dorms=%d",
			years, terms, grades, classes, dormitories)
	}

	var auditCount int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM audit_log WHERE school_id = $1 AND action IN ('school.created', 'school.setup_wizard.completed')`,
		schoolID).Scan(&auditCount); err != nil {
		t.Fatalf("query audit log: %v", err)
	}
	if auditCount != 2 {
		t.Fatalf("audit rows = %d, want 2", auditCount)
	}
}

func jsonBody(t *testing.T, value any) *bytes.Reader {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	return bytes.NewReader(encoded)
}
