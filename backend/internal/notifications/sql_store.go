package notifications

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SQLStore is the PostgreSQL implementation of notifications.Store.
type SQLStore struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

type sqlQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *SQLStore) q() sqlQuerier {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *SQLStore) WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.db == nil {
		return errors.New("notifications: SQLStore has no database connection")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if schoolID != "" {
		if _, err := tx.ExecContext(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	tenantStore := &SQLStore{db: s.db, tx: tx}
	if err := fn(tenantStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (s *SQLStore) CreateInAppNotification(ctx context.Context, n *InAppNotification) error {
	if n.ID == "" {
		n.ID = newUUID()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}

	var metaJSON []byte
	if n.Metadata != nil {
		var err error
		metaJSON, err = json.Marshal(n.Metadata)
		if err != nil {
			metaJSON = []byte("{}")
		}
	}

	query := `
		INSERT INTO in_app_notifications (
			id, school_id, recipient_user_id, category, title, body, metadata, is_read, read_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		);
	`
	_, err := s.q().ExecContext(ctx, query,
		n.ID, n.SchoolID, n.RecipientUserID, n.Category, n.Title, n.Body,
		metaJSON, n.IsRead, n.ReadAt, n.CreatedAt,
	)
	return err
}

func (s *SQLStore) ListInAppNotifications(ctx context.Context, schoolID, recipientUserID string, unreadOnly bool, limit int) ([]InAppNotification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, school_id, recipient_user_id, category, title, body, metadata, is_read, read_at, created_at
		FROM in_app_notifications
		WHERE school_id = $1 AND recipient_user_id = $2
	`
	if unreadOnly {
		query += " AND is_read = FALSE"
	}
	query += " ORDER BY created_at DESC LIMIT $3;"

	rows, err := s.q().QueryContext(ctx, query, schoolID, recipientUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InAppNotification
	for rows.Next() {
		var item InAppNotification
		var metaBytes []byte
		var readAt sql.NullTime

		if err := rows.Scan(
			&item.ID, &item.SchoolID, &item.RecipientUserID, &item.Category,
			&item.Title, &item.Body, &metaBytes, &item.IsRead, &readAt, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &item.Metadata)
		}
		if readAt.Valid {
			val := readAt.Time
			item.ReadAt = &val
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (s *SQLStore) MarkNotificationRead(ctx context.Context, schoolID, notificationID, recipientUserID string) error {
	query := `
		UPDATE in_app_notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE school_id = $1 AND id = $2 AND recipient_user_id = $3;
	`
	res, err := s.q().ExecContext(ctx, query, schoolID, notificationID, recipientUserID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return errors.New("notification not found")
	}
	return err
}
