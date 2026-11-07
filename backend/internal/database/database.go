package database

import (
	"context"
	"database/sql"
	"fmt"
)

// DB handles PostgreSQL connection pooling and Row-Level Security (RLS) enforcement.
type DB struct {
	*sql.DB
}

// AuditLogger records platform actions into the audit_log table.
type AuditLogger interface {
	LogAction(ctx context.Context, schoolID, userID, action, resourceType, resourceID string, details map[string]interface{}) error
}

// WithTenantTx executes fn inside a transaction with app.school_id set for RLS.
func WithTenantTx(ctx context.Context, db *sql.DB, schoolID string, fn func(tx *sql.Tx) error) error {
	if db == nil {
		return fn(nil)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if schoolID != "" {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.school_id = '%s'", schoolID)); err != nil {
			return err
		}
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
