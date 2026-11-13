package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ErrMissingSchoolID is returned when a tenant transaction is requested without
// a school id. Running without it would silently return zero rows under RLS.
var ErrMissingSchoolID = errors.New("database: school id is required for a tenant transaction")

// DB wraps the PostgreSQL connection pool.
type DB struct {
	*sql.DB
}

// Open creates a connection pool and verifies it can reach PostgreSQL.
func Open(dsn string, maxOpen, maxIdle int) (*DB, error) {
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}
	pool.SetMaxOpenConns(maxOpen)
	pool.SetMaxIdleConns(maxIdle)
	pool.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return &DB{DB: pool}, nil
}

// WithUserTx runs fn inside a transaction with app.user_id set. It is used only
// for self-scoped reads (such as discovering a user's own memberships at
// login) where no school context exists yet.
func WithUserTx(ctx context.Context, db *sql.DB, userID string, fn func(tx *sql.Tx) error) error {
	if db == nil {
		return errors.New("database: nil connection")
	}
	if userID == "" {
		return errors.New("database: user id is required for a self transaction")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database: begin user tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, userID); err != nil {
		return fmt.Errorf("database: set user context: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: commit user tx: %w", err)
	}
	return nil
}

// WithTenantTx runs fn inside a transaction with app.school_id set so that
// PostgreSQL Row-Level Security isolates every statement to one school.
func WithTenantTx(ctx context.Context, db *sql.DB, schoolID string, fn func(tx *sql.Tx) error) error {
	if db == nil {
		return errors.New("database: nil connection")
	}
	if schoolID == "" {
		return ErrMissingSchoolID
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database: begin tenant tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
		return fmt.Errorf("database: set tenant context: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: commit tenant tx: %w", err)
	}
	return nil
}
