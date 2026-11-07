package billing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBillingMigrationCoversAllTables verifies that migration 000004
// creates all Release 1 billing, WhatsApp, notification, and AI audit tables
// with Row-Level Security (RLS) enabled and forced.
func TestBillingMigrationCoversAllTables(t *testing.T) {
	upPath := filepath.Join("..", "..", "..", "migrations", "000004_r1_billing_whatsapp_notifications.up.sql")
	downPath := filepath.Join("..", "..", "..", "migrations", "000004_r1_billing_whatsapp_notifications.down.sql")

	upBytes, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}
	downBytes, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}
	up := string(upBytes)
	down := string(downBytes)

	requiredTables := []string{
		"school_mpesa_configs",
		"fee_periods",
		"fee_charges",
		"fee_change_notes",
		"ledger_entries",
		"unmatched_payments",
		"school_whatsapp_bots",
		"chat_sessions",
		"chat_otp_requests",
		"sender_rate_blocks",
		"in_app_notifications",
		"ai_action_logs",
	}

	for _, table := range requiredTables {
		if !strings.Contains(up, "CREATE TABLE "+table+" (") {
			t.Errorf("up migration is missing CREATE TABLE %s", table)
		}
		if !strings.Contains(down, "DROP TABLE IF EXISTS "+table+" CASCADE;") {
			t.Errorf("down migration is missing DROP TABLE IF EXISTS %s", table)
		}
	}

	requiredFragments := []string{
		"ALTER TABLE fee_ledgers ADD COLUMN IF NOT EXISTS credit_balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00;",
		"ALTER TABLE fee_ledgers ADD COLUMN IF NOT EXISTS due_date DATE;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_payment_paused_until TIMESTAMPTZ;",
		"ENABLE ROW LEVEL SECURITY",
		"FORCE ROW LEVEL SECURITY",
		"current_school_id()",
		"tenant_isolation",
	}

	for _, frag := range requiredFragments {
		if !strings.Contains(up, frag) {
			t.Errorf("up migration is missing required fragment: %s", frag)
		}
	}
}
