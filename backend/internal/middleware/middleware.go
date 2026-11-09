package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
)

// Claims is the verified payload of a bearer token.
type Claims struct {
	UserID      string
	PhoneNumber string
	ExpiresAt   int64
}

// TokenVerifier verifies signed bearer tokens.
type TokenVerifier interface {
	Verify(token string) (*Claims, error)
}

// BadgeGrant is an active (or time-bounded) badge held by a membership.
type BadgeGrant struct {
	ID          string
	Slug        string
	ScopeType   string
	ScopeID     string
	Permissions []string
	ValidFrom   *time.Time
	ValidUntil  *time.Time
}

// Identity is the authenticated caller resolved to one school membership.
type Identity struct {
	UserID       string
	PhoneNumber  string
	SchoolID     string
	MembershipID string
	Role         string
	Permissions  []string
	Badges       []BadgeGrant
}

type identityKey struct{}

// WithIdentity attaches an Identity to a context.
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, identity)
}

// IdentityFromContext extracts an Identity from a context.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey{}).(Identity)
	return identity, ok
}

// AuthMiddleware resolves bearer tokens to a tenant-scoped Identity.
type AuthMiddleware struct {
	db       *sql.DB
	verifier TokenVerifier
}

// NewAuthMiddleware builds the auth middleware.
func NewAuthMiddleware(db *sql.DB, verifier TokenVerifier) *AuthMiddleware {
	return &AuthMiddleware{db: db, verifier: verifier}
}

// RequireAuth rejects unauthenticated requests and places the resolved Identity
// on the request context. The active school comes from X-School-ID, falling
// back to the caller's only active membership.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
			return
		}
		claims, err := m.verifier.Verify(token)
		if err != nil || claims == nil {
			RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid or expired token.")
			return
		}

		identity, err := m.resolveIdentity(r, claims)
		if err != nil {
			if err == errSchoolRequired {
				RespondError(w, http.StatusBadRequest, "SCHOOL_REQUIRED", "X-School-ID is required when the caller belongs to multiple schools.")
				return
			}
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "No active membership for this user and school.")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), identity)))
	})
}

// RequireTenant asserts that a tenant-scoped identity is present.
func (m *AuthMiddleware) RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFromContext(r.Context())
		if !ok || identity.SchoolID == "" {
			RespondError(w, http.StatusForbidden, "TENANT_REQUIRED", "An active school membership is required.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireInternalService gates endpoints that only the AI service may call.
func RequireInternalService(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" || r.Header.Get("X-Internal-Service-Token") != expectedToken {
				RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "A valid internal service token is required.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var errSchoolRequired = &membershipError{"school required"}

type membershipError struct{ msg string }

func (e *membershipError) Error() string { return e.msg }

type membershipRow struct {
	id       string
	schoolID string
	role     string
}

func (m *AuthMiddleware) resolveIdentity(r *http.Request, claims *Claims) (Identity, error) {
	requestedSchool := strings.TrimSpace(r.Header.Get("X-School-ID"))
	if requestedSchool != "" {
		membership, err := m.membershipForSchool(r.Context(), claims.UserID, requestedSchool)
		if err != nil {
			return Identity{}, err
		}
		return m.buildIdentity(r.Context(), claims, membership)
	}

	memberships, err := m.membershipsForUser(r.Context(), claims.UserID)
	if err != nil {
		return Identity{}, err
	}

	switch len(memberships) {
	case 0:
		return Identity{}, &membershipError{"no memberships"}
	case 1:
		return m.buildIdentity(r.Context(), claims, memberships[0])
	default:
		return Identity{}, errSchoolRequired
	}
}

func (m *AuthMiddleware) membershipsForUser(ctx context.Context, userID string) ([]membershipRow, error) {
	var memberships []membershipRow
	err := database.WithUserTx(ctx, m.db, userID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT id::text, school_id::text, role
			   FROM memberships
			  WHERE user_id = $1 AND status = 'active'
			  ORDER BY created_at`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var membership membershipRow
			if err := rows.Scan(&membership.id, &membership.schoolID, &membership.role); err != nil {
				return err
			}
			memberships = append(memberships, membership)
		}
		return rows.Err()
	})
	return memberships, err
}

func (m *AuthMiddleware) membershipForSchool(ctx context.Context, userID, schoolID string) (membershipRow, error) {
	var membership membershipRow
	err := database.WithTenantTx(ctx, m.db, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`SELECT id::text, school_id::text, role
			   FROM memberships
			  WHERE user_id = $1 AND school_id = $2 AND status = 'active'`,
			userID, schoolID).Scan(&membership.id, &membership.schoolID, &membership.role)
	})
	return membership, err
}

func (m *AuthMiddleware) buildIdentity(ctx context.Context, claims *Claims, membership membershipRow) (Identity, error) {
	badges, err := m.badgesForMembership(ctx, membership.schoolID, membership.id)
	if err != nil {
		return Identity{}, err
	}
	return Identity{
		UserID:       claims.UserID,
		PhoneNumber:  claims.PhoneNumber,
		SchoolID:     membership.schoolID,
		MembershipID: membership.id,
		Role:         membership.role,
		Permissions:  rolePermissions[membership.role],
		Badges:       badges,
	}, nil
}

func (m *AuthMiddleware) badgesForMembership(ctx context.Context, schoolID, membershipID string) ([]BadgeGrant, error) {
	var badges []BadgeGrant
	err := database.WithTenantTx(ctx, m.db, schoolID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT mb.id::text, b.slug, b.scope_type, COALESCE(mb.scope_id::text, ''),
			        mb.valid_from, mb.valid_until, COALESCE(p.code, '')
			   FROM membership_badges mb
			   JOIN badges b ON b.id = mb.badge_id AND b.school_id = $2
			   LEFT JOIN badge_permissions bp ON bp.badge_id = b.id
			   LEFT JOIN permissions p ON p.id = bp.permission_id
			  WHERE mb.membership_id = $1
			  ORDER BY mb.created_at`, membershipID, schoolID)
		if err != nil {
			return err
		}
		defer rows.Close()

		byID := map[string]*BadgeGrant{}
		var order []string
		for rows.Next() {
			var (
				id, slug, scopeType, scopeID, permission string
				validFrom, validUntil                    sql.NullTime
			)
			if err := rows.Scan(&id, &slug, &scopeType, &scopeID, &validFrom, &validUntil, &permission); err != nil {
				return err
			}
			grant, ok := byID[id]
			if !ok {
				grant = &BadgeGrant{ID: id, Slug: slug, ScopeType: scopeType, ScopeID: scopeID}
				if validFrom.Valid {
					grant.ValidFrom = &validFrom.Time
				}
				if validUntil.Valid {
					grant.ValidUntil = &validUntil.Time
				}
				byID[id] = grant
				order = append(order, id)
			}
			if permission != "" {
				grant.Permissions = append(grant.Permissions, permission)
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}

		badges = make([]BadgeGrant, 0, len(order))
		for _, id := range order {
			badges = append(badges, *byID[id])
		}
		return nil
	})
	return badges, err
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
