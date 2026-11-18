package billing

import (
	"context"
	"net/http"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

const (
	// PermFeeStructuresManage authorizes loading and triggering fee structures
	// within a school via an active, school-scoped membership badge.
	PermFeeStructuresManage = "billing.fee_structures.manage"

	// PermLedgersView authorizes read-only staff/bursar visibility into student
	// fee balances and ledger entries within the school.
	PermLedgersView = "billing.ledgers.view"
)

func hasActiveSchoolPermissionTx(ctx context.Context, tx pgx.Tx, p contracts.Principal, permissionCode string) (bool, error) {
	var allowed bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM memberships m
			JOIN users u ON u.id = m.user_id
			JOIN membership_badges mb ON mb.membership_id = m.id
			JOIN badges b ON b.id = mb.badge_id AND b.school_id = m.school_id
			JOIN badge_permissions bp ON bp.badge_id = b.id
			JOIN permissions perm ON perm.id = bp.permission_id
			WHERE m.id = $1
			  AND m.school_id = $2
			  AND m.user_id = $3
			  AND m.status = 'active'
			  AND m.role <> 'parent'
			  AND u.is_active = TRUE
			  AND perm.code = $4
			  AND (mb.scope_id IS NULL OR b.scope_type = 'none')
			  AND (mb.valid_from IS NULL OR mb.valid_from <= NOW())
			  AND (mb.valid_until IS NULL OR mb.valid_until > NOW())
		)`,
		p.MembershipID, p.SchoolID, p.UserID, permissionCode,
	).Scan(&allowed)
	return allowed, err
}

func requireSchoolPermissionTx(ctx context.Context, tx pgx.Tx, p contracts.Principal, permissionCode, forbiddenMsg string) error {
	allowed, err := hasActiveSchoolPermissionTx(ctx, tx, p, permissionCode)
	if err != nil {
		return err
	}
	if !allowed {
		return contracts.Fail(http.StatusForbidden, "FORBIDDEN", forbiddenMsg)
	}
	return nil
}

func isActiveFeePayingGuardianTx(ctx context.Context, tx pgx.Tx, p contracts.Principal, studentID string) (bool, error) {
	var allowed bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM student_guardians sg
			JOIN memberships m ON m.school_id = sg.school_id AND m.user_id = sg.user_id
			JOIN users u ON u.id = sg.user_id
			WHERE sg.school_id = $1
			  AND sg.student_id = $2
			  AND sg.user_id = $3
			  AND sg.is_fee_payer = TRUE
			  AND m.id = $4
			  AND m.status = 'active'
			  AND u.is_active = TRUE
		)`,
		p.SchoolID, studentID, p.UserID, p.MembershipID,
	).Scan(&allowed)
	return allowed, err
}

// authorizeStudentFinancialReadTx enforces least-privilege access to student
// fee balances and ledgers: either a staff/bursar membership holding an active
// billing.ledgers.view badge permission, or an active guardian designated as
// is_fee_payer = TRUE for that student in the same school.
func authorizeStudentFinancialReadTx(ctx context.Context, tx pgx.Tx, p contracts.Principal, studentID string) error {
	if allowed, err := hasActiveSchoolPermissionTx(ctx, tx, p, PermLedgersView); err != nil {
		return err
	} else if allowed {
		return nil
	}
	if feePayer, err := isActiveFeePayingGuardianTx(ctx, tx, p, studentID); err != nil {
		return err
	} else if feePayer {
		return nil
	}
	return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "You are not authorized to view this student's financial records.")
}
