package notifications

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
)

type Service struct {
	DB contracts.TenantDB
}

// CreateNotification aliases the shared contracts.InAppNotification payload.
type CreateNotification = contracts.InAppNotification

// CreateInAppNotification validates and persists an in-app notification inside
// the caller's active tenant-scoped database transaction. It never sends SMS.
func (s *Service) CreateInAppNotification(ctx context.Context, tx contracts.SQLTx, input contracts.InAppNotification) error {
	if !contracts.ValidID(input.SchoolID) || !contracts.ValidID(input.RecipientUserID) ||
		strings.TrimSpace(input.Category) == "" || utf8.RuneCountInString(input.Category) > 50 ||
		strings.TrimSpace(input.Title) == "" || utf8.RuneCountInString(input.Title) > 255 ||
		strings.TrimSpace(input.Body) == "" {
		return contracts.Fail(422, "INVALID_NOTIFICATION", "A recipient, category, title and body are required.")
	}
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return contracts.Fail(422, "INVALID_NOTIFICATION", "Notification metadata must be valid JSON.")
	}
	if tx == nil {
		return contracts.Fail(500, "INTERNAL_ERROR", "An active tenant transaction is required.")
	}
	recipientExists, err := tx.QueryRowExists(ctx, `SELECT EXISTS(SELECT 1 FROM memberships m JOIN users u ON u.id=m.user_id
		WHERE m.school_id=$1 AND m.user_id=$2 AND m.status='active' AND u.is_active)`, input.SchoolID, input.RecipientUserID)
	if err != nil {
		return err
	}
	if !recipientExists {
		return contracts.Fail(422, "INVALID_NOTIFICATION_RECIPIENT", "The recipient is not active in this school.")
	}
	return tx.Exec(ctx, `INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body,metadata)
		VALUES($1,$2,$3,$4,$5,$6)`, input.SchoolID, input.RecipientUserID, input.Category, input.Title, input.Body, string(metadata))
}

var _ contracts.InAppNotifier = (*Service)(nil)
