package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantDBAdapter executes pgx transactions with transaction-local PostgreSQL
// Row-Level Security tenant context (`set_config('app.school_id', $1, true)`).
type TenantDBAdapter struct {
	pool *pgxpool.Pool
}

// NewTenantDBAdapter wraps a pgxpool.Pool for RLS-enforced tenant transactions.
func NewTenantDBAdapter(pool *pgxpool.Pool) (*TenantDBAdapter, error) {
	if pool == nil {
		return nil, errors.New("jobs: database pool is required")
	}
	return &TenantDBAdapter{pool: pool}, nil
}

// WithSchoolTx acquires a pooled connection, starts a transaction, configures
// transaction-local app.school_id for RLS, executes fn, and commits or rolls back.
func (a *TenantDBAdapter) WithSchoolTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
	if a == nil || a.pool == nil {
		return errors.New("jobs: database pool is nil")
	}
	if !contracts.ValidID(schoolID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_SCHOOL_ID", "A valid school_id UUID is required for tenant transactions.")
	}
	if fn == nil {
		return errors.New("jobs: transaction callback is nil")
	}

	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("jobs: acquire db connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("jobs: begin tenant tx: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if _, err := tx.Exec(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
		return fmt.Errorf("jobs: set tenant context: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("jobs: commit tenant tx: %w", err)
	}
	return nil
}

// AsContractsTenantDB returns a contracts.TenantDB backed by this adapter.
func (a *TenantDBAdapter) AsContractsTenantDB() contracts.TenantDB {
	return contracts.TenantTxFunc(a.WithSchoolTx)
}
