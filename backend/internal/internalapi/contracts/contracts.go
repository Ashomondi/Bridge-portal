// Package contracts contains Developer 05's consumer boundaries. The application
// supplies authenticated identities and tenant transactions. AuditLogger is an
// integration interface only; it does not implement or satisfy AC-29 logging.
package contracts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SQLTx is the minimal query/exec subset needed by shared transactional
// contracts so callers using pgx.Tx (via PgxTx) or database/sql *sql.Tx (via
// StdSQLTx) can participate in the same database transaction.
type SQLTx interface {
	QueryRowExists(ctx context.Context, query string, args ...any) (bool, error)
	Exec(ctx context.Context, query string, args ...any) error
}

type pgxSQLTx struct{ tx pgx.Tx }

// PgxTx adapts a pgx.Tx to SQLTx, returning nil when tx is nil.
func PgxTx(tx pgx.Tx) SQLTx {
	if tx == nil {
		return nil
	}
	return pgxSQLTx{tx: tx}
}

func (p pgxSQLTx) QueryRowExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := p.tx.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (p pgxSQLTx) Exec(ctx context.Context, query string, args ...any) error {
	_, err := p.tx.Exec(ctx, query, args...)
	return err
}

type stdSQLTx struct{ tx *sql.Tx }

// StdSQLTx adapts a database/sql *sql.Tx to SQLTx, returning nil when tx is nil.
func StdSQLTx(tx *sql.Tx) SQLTx {
	if tx == nil {
		return nil
	}
	return stdSQLTx{tx: tx}
}

func (s stdSQLTx) QueryRowExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := s.tx.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (s stdSQLTx) Exec(ctx context.Context, query string, args ...any) error {
	_, err := s.tx.ExecContext(ctx, query, args...)
	return err
}

// InAppNotification is a single transaction-bound in-app notification request.
type InAppNotification struct {
	SchoolID        string         `json:"school_id"`
	RecipientUserID string         `json:"recipient_user_id"`
	Category        string         `json:"category"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// InAppNotifier persists an in-app notification inside the caller's active
// tenant-scoped database transaction so business writes and notices commit or
// roll back together.
type InAppNotifier interface {
	CreateInAppNotification(context.Context, SQLTx, InAppNotification) error
}

type TenantDB interface {
	WithTenantTx(context.Context, Principal, func(pgx.Tx) error) error
}

// TenantTxFunc adapts the application's WithTenantTx without creating another
// connection pool or taking ownership of transaction commit/rollback. The
// principal must come from RequireIdentity with trusted authentication middleware;
// callers cannot select a separate school from a URL, header, or request body.
type TenantTxFunc func(context.Context, string, func(pgx.Tx) error) error

func (f TenantTxFunc) WithTenantTx(ctx context.Context, principal Principal, fn func(pgx.Tx) error) error {
	if !principal.valid() {
		return Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	return f(ctx, principal.SchoolID, fn)
}

type AuditEvent struct {
	SchoolID, UserID, Action, ResourceType, ResourceID string
	Details                                            map[string]any
}

type AuditLogger interface {
	Log(context.Context, pgx.Tx, AuditEvent) error
}

type Principal struct {
	UserID       string `json:"user_id"`
	SchoolID     string `json:"school_id"`
	MembershipID string `json:"membership_id"`
	Role         string `json:"role"`
}

func (p Principal) valid() bool {
	return ValidID(p.UserID) && ValidID(p.SchoolID) && ValidID(p.MembershipID) && p.Role != ""
}

// Identity must be supplied from trusted authentication middleware, never from
// unchecked request headers or request JSON.
type Identity func(*http.Request) (Principal, error)

func RequireIdentity(r *http.Request, identity Identity) (Principal, error) {
	if identity == nil {
		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
	}
	p, err := identity(r)
	if err != nil {
		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
	}
	if !p.valid() {
		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
	}
	return p, nil
}

func ValidID(id string) bool {
	if len(id) != 36 || id == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for _, index := range []int{8, 13, 18, 23} {
		if id[index] != '-' {
			return false
		}
	}
	var value pgtype.UUID
	return value.Scan(id) == nil && value.Valid
}

type APIError struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

func (e *APIError) Error() string { return e.Message }

func Fail(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message, Details: map[string]any{}}
}

func WriteError(w http.ResponseWriter, err error) {
	var apiError *APIError
	if !errors.As(err, &apiError) {
		apiError = Fail(500, "INTERNAL_ERROR", "The request could not be completed.")
	}
	write(w, apiError.Status, struct {
		Status string    `json:"status"`
		Error  *APIError `json:"error"`
	}{"error", apiError})
}

type Page struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

func Success(w http.ResponseWriter, status int, data any, page *Page) {
	write(w, status, struct {
		Status string `json:"status"`
		Data   any    `json:"data"`
		Meta   *Page  `json:"meta,omitempty"`
	}{"success", data, page})
}

func write(w http.ResponseWriter, status int, body any) {
	encoded, err := json.Marshal(body)
	if err != nil {
		status = 500
		encoded = []byte(`{"status":"error","error":{"code":"INTERNAL_ERROR","message":"The request could not be completed.","details":{}}}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(encoded)
}

func Decode(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return Fail(400, "INVALID_JSON", "A valid JSON request is required.")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Fail(400, "INVALID_JSON", "Only one JSON value is allowed.")
	}
	return nil
}

func Pagination(r *http.Request) (Page, error) {
	p := Page{Page: 1, PerPage: 25}
	for key, field := range map[string]*int{"page": &p.Page, "per_page": &p.PerPage} {
		if value := r.URL.Query().Get(key); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return Page{}, Fail(422, "INVALID_PAGINATION", "Pagination values must be positive integers.")
			}
			*field = n
		}
	}
	if p.PerPage > 100 || p.Page > 1000000 {
		return Page{}, Fail(422, "INVALID_PAGINATION", "Pagination values exceed the supported limit.")
	}
	return p, nil
}
