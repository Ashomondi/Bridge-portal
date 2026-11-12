package onboarding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
)

// PurgeExpiredUnverifiedSubmissions deletes pending or rejected registration
// submissions whose retention window has elapsed, across every school. It is
// called by Ian's Asynq worker (AC-07).
func (s *Service) PurgeExpiredUnverifiedSubmissions(ctx context.Context, retentionDays int) (int, error) {
	if s.db == nil {
		return 0, errors.New("onboarding: database is not configured")
	}
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}

	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM schools`)
	if err != nil {
		return 0, fmt.Errorf("onboarding: list schools: %w", err)
	}
	var schoolIDs []string
	for rows.Next() {
		var schoolID string
		if err := rows.Scan(&schoolID); err != nil {
			rows.Close()
			return 0, err
		}
		schoolIDs = append(schoolIDs, schoolID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	total := 0
	for _, schoolID := range schoolIDs {
		var deleted int64
		err := database.WithTenantTx(ctx, s.db, schoolID, func(tx *sql.Tx) error {
			result, err := tx.ExecContext(ctx,
				`DELETE FROM registration_submissions
				  WHERE status IN ('pending', 'rejected')
				    AND (
				      (expires_for_purge_at IS NOT NULL AND expires_for_purge_at <= NOW())
				      OR (expires_for_purge_at IS NULL AND created_at <= NOW() - ($1::int * INTERVAL '1 day'))
				    )`,
				retentionDays)
			if err != nil {
				return err
			}
			deleted, _ = result.RowsAffected()
			return nil
		})
		if err != nil {
			return total, err
		}
		total += int(deleted)
	}
	return total, nil
}
