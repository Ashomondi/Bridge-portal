package notifications

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	DB contracts.TenantDB
}

type CreateNotification struct {
	RecipientUserID string         `json:"recipient_user_id"`
	Category        string         `json:"category"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Metadata        map[string]any `json:"metadata"`
}

// InAppNotifier participates in its caller's transaction: billing/calendar
// changes and their notices commit or roll back together. It never sends SMS.
type InAppNotifier interface {
	CreateInAppNotification(context.Context, pgx.Tx, string, CreateNotification) error
}

func (s *Service) CreateInAppNotification(ctx context.Context, tx pgx.Tx, schoolID string, input CreateNotification) error {
	if !contracts.ValidID(schoolID) || !contracts.ValidID(input.RecipientUserID) ||
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
	var recipientExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM memberships m JOIN users u ON u.id=m.user_id
		WHERE m.school_id=$1 AND m.user_id=$2 AND m.status='active' AND u.is_active)`, schoolID, input.RecipientUserID).Scan(&recipientExists)
	if err != nil {
		return err
	}
	if !recipientExists {
		return contracts.Fail(422, "INVALID_NOTIFICATION_RECIPIENT", "The recipient is not active in this school.")
	}
	_, err = tx.Exec(ctx, `INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body,metadata)
		VALUES($1,$2,$3,$4,$5,$6)`, schoolID, input.RecipientUserID, input.Category, input.Title, input.Body, metadata)
	return err
}

var _ InAppNotifier = (*Service)(nil)
