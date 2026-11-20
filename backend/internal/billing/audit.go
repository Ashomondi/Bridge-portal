package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

// SQLAuditLogger is optionally implemented by AuditLogger implementations that
// can persist audit events inside a database/sql transaction (e.g. when called
// from academics.PutCalendar via TermDateFeeSyncer).
type SQLAuditLogger interface {
	contracts.AuditLogger
	LogSQL(ctx context.Context, tx *sql.Tx, event contracts.AuditEvent) error
}

// PostgresAuditLogger persists billing audit events into the audit_log table
// inside the caller's active database transaction so business mutations and
// audit records commit or roll back atomically.
type PostgresAuditLogger struct{}

// NewPostgresAuditLogger returns a transaction-bound PostgreSQL AuditLogger.
func NewPostgresAuditLogger() *PostgresAuditLogger {
	return &PostgresAuditLogger{}
}

func (l *PostgresAuditLogger) Log(ctx context.Context, tx pgx.Tx, event contracts.AuditEvent) error {
	if tx == nil {
		return contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "An active transaction is required for audit logging.")
	}
	details, err := marshalAuditEvent(event)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_log (school_id, user_id, action, resource_type, resource_id, details)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, ''), $6::jsonb)`,
		event.SchoolID, event.UserID, event.Action, event.ResourceType, event.ResourceID, string(details),
	)
	return err
}

func (l *PostgresAuditLogger) LogSQL(ctx context.Context, tx *sql.Tx, event contracts.AuditEvent) error {
	if tx == nil {
		return contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "An active transaction is required for audit logging.")
	}
	details, err := marshalAuditEvent(event)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_log (school_id, user_id, action, resource_type, resource_id, details)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, ''), $6::jsonb)`,
		event.SchoolID, event.UserID, event.Action, event.ResourceType, event.ResourceID, string(details),
	)
	return err
}

func marshalAuditEvent(event contracts.AuditEvent) ([]byte, error) {
	if !contracts.ValidID(event.SchoolID) {
		return nil, contracts.Fail(http.StatusBadRequest, "INVALID_AUDIT_SCHOOL", "Audit event requires a valid school_id.")
	}
	if event.UserID != "" && !contracts.ValidID(event.UserID) {
		return nil, contracts.Fail(http.StatusBadRequest, "INVALID_AUDIT_ACTOR", "Audit event user_id must be empty or a valid UUID.")
	}
	if event.Action == "" || event.ResourceType == "" {
		return nil, contracts.Fail(http.StatusInternalServerError, "INVALID_AUDIT_EVENT", "Audit event action and resource_type are required.")
	}
	if len(event.Details) == 0 {
		return []byte("{}"), nil
	}
	encoded, err := json.Marshal(event.Details)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (s *Service) requireAuditLogger() error {
	if s == nil || s.Audit == nil {
		return contracts.Fail(http.StatusInternalServerError, "AUDIT_LOGGER_REQUIRED", "A transactional audit logger is required for financial operations.")
	}
	return nil
}

func (s *Service) logAuditPgx(ctx context.Context, tx pgx.Tx, event contracts.AuditEvent) error {
	if err := s.requireAuditLogger(); err != nil {
		return err
	}
	return s.Audit.Log(ctx, tx, event)
}

func (s *Service) logAuditSQL(ctx context.Context, tx *sql.Tx, event contracts.AuditEvent) error {
	if err := s.requireAuditLogger(); err != nil {
		return err
	}
	sqlLogger, ok := s.Audit.(SQLAuditLogger)
	if !ok {
		return contracts.Fail(http.StatusInternalServerError, "AUDIT_LOGGER_INCOMPATIBLE", "Configured audit logger does not support SQL transactions.")
	}
	return sqlLogger.LogSQL(ctx, tx, event)
}

var (
	_ contracts.AuditLogger = (*PostgresAuditLogger)(nil)
	_ SQLAuditLogger        = (*PostgresAuditLogger)(nil)
)
