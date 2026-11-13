package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// AuditEvent is a single sensitive platform action recorded in audit_log.
type AuditEvent struct {
	SchoolID     string
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	IPAddress    string
	UserAgent    string
	Details      map[string]any
}

// AuditLogger writes sensitive platform actions to audit_log. When a
// transaction is supplied the event is written atomically with the change it
// describes; otherwise the pool is used.
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger builds an audit logger backed by the given connection pool.
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

const insertAuditSQL = `
INSERT INTO audit_log
    (school_id, user_id, action, resource_type, resource_id, ip_address, user_agent, details)
VALUES
    (NULLIF($1, '')::uuid, NULLIF($2, '')::uuid, $3, $4,
     NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, '')::jsonb)`

// Log records an audit event. Pass a nil tx to use the connection pool.
func (l *AuditLogger) Log(ctx context.Context, tx *sql.Tx, event AuditEvent) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("audit: logger is not configured")
	}

	var details any
	if len(event.Details) > 0 {
		encoded, err := json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("audit: encode details: %w", err)
		}
		details = string(encoded)
	}

	args := []any{
		event.SchoolID, event.UserID, event.Action, event.ResourceType,
		event.ResourceID, event.IPAddress, event.UserAgent, details,
	}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, insertAuditSQL, args...)
	} else {
		_, err = l.db.ExecContext(ctx, insertAuditSQL, args...)
	}
	if err != nil {
		return fmt.Errorf("audit: insert %q: %w", event.Action, err)
	}
	return nil
}

// LogAction satisfies the shared audit contract used by billing and other
// modules that only need the flat signature.
func (l *AuditLogger) LogAction(ctx context.Context, schoolID, userID, action, resourceType, resourceID string, details map[string]any) error {
	return l.Log(ctx, nil, AuditEvent{
		SchoolID:     schoolID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	})
}
