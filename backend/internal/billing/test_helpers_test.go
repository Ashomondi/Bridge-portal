package billing_test

import (
	"context"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
)

func newTestBillingService(db *testenv.DB) *billing.Service {
	return &billing.Service{
		DB:    contracts.TenantTxFunc(db.WithTenantTx),
		Audit: billing.NewPostgresAuditLogger(),
	}
}

func seedMemberWithPermission(t *testing.T, db *testenv.DB, schoolID, phone, fullName, role, permissionCode string) contracts.Principal {
	t.Helper()
	ctx := context.Background()
	var userID, membershipID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO users (phone_number, full_name)
		VALUES ($1, $2)
		RETURNING id`,
		phone, fullName,
	).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO memberships (school_id, user_id, role, status)
		VALUES ($1, $2, $3, 'active')
		RETURNING id`,
		schoolID, userID, role,
	).Scan(&membershipID); err != nil {
		t.Fatal(err)
	}
	if permissionCode != "" {
		grantBadgePermission(t, db, schoolID, membershipID, permissionCode)
	}
	return contracts.Principal{
		UserID:       userID,
		SchoolID:     schoolID,
		MembershipID: membershipID,
		Role:         role,
	}
}

func grantBadgePermission(t *testing.T, db *testenv.DB, schoolID, membershipID, permissionCode string) {
	t.Helper()
	ctx := context.Background()
	var permID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO permissions (code, name, category)
		VALUES ($1, $1, 'billing')
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		permissionCode,
	).Scan(&permID); err != nil {
		t.Fatal(err)
	}
	slug := "badge_" + permissionCode
	var badgeID string
	if err := db.Admin.QueryRow(ctx, `
		INSERT INTO badges (school_id, name, slug, scope_type)
		VALUES ($1, $2, $2, 'none')
		ON CONFLICT (school_id, slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		schoolID, slug,
	).Scan(&badgeID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO badge_permissions (badge_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
		badgeID, permID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, `
		INSERT INTO membership_badges (membership_id, badge_id)
		VALUES ($1, $2)`,
		membershipID, badgeID,
	); err != nil {
		t.Fatal(err)
	}
}
