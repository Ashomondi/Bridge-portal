package middleware

import (
	"net/http"
	"time"
)

// Permission codes used across Release 1 endpoints.
const (
	PermStaffInvitesCreate  = "staff.invites.create"
	PermParentsQRManage     = "parents.qr.manage"
	PermVerifierGrant       = "verifier.grant"
	PermRegistrationsVerify = "registrations.verify"
	PermStudentsAdmit       = "students.admit"
	PermMarksReassign       = "marks.reassign"
	PermClassStudentsView   = "class.students.view"
	PermClassStatsView      = "class.stats.view"
	PermResultsRelease      = "results.release"
	PermConsentsManage      = "consents.manage"
	PermPhoneChange         = "phone.change"
	PermSchoolSetup         = "school.setup"
)

// rolePermissions are the permissions a role holds without a badge. Badges add
// permissions on top of these, optionally scoped to a class or other resource.
var rolePermissions = map[string][]string{
	"principal": {
		PermStaffInvitesCreate, PermParentsQRManage, PermVerifierGrant,
		PermRegistrationsVerify, PermStudentsAdmit, PermMarksReassign,
		PermClassStatsView, PermResultsRelease, PermConsentsManage,
		PermPhoneChange, PermSchoolSetup,
	},
	"deputy": {
		PermStaffInvitesCreate, PermParentsQRManage, PermVerifierGrant,
		PermRegistrationsVerify, PermStudentsAdmit, PermMarksReassign,
		PermClassStatsView, PermResultsRelease, PermConsentsManage,
		PermPhoneChange, PermSchoolSetup,
	},
	"dean_of_studies": {
		PermParentsQRManage, PermVerifierGrant, PermRegistrationsVerify,
		PermStudentsAdmit, PermMarksReassign, PermClassStatsView,
		PermResultsRelease, PermConsentsManage,
	},
	"teacher": {PermClassStatsView},
	"parent":  {PermConsentsManage, PermPhoneChange},
}

// HasPermission reports whether the identity holds perm, either through its
// role defaults or an active, unexpired badge. A scoped badge is only valid
// when the target scope matches.
func (id Identity) HasPermission(perm, scopeID string) bool {
	if contains(id.Permissions, perm) || contains(rolePermissions[id.Role], perm) {
		return true
	}

	now := time.Now()
	for _, badge := range id.Badges {
		if badge.ValidFrom != nil && now.Before(*badge.ValidFrom) {
			continue
		}
		if badge.ValidUntil != nil && !now.Before(*badge.ValidUntil) {
			continue
		}
		if !contains(badge.Permissions, perm) {
			continue
		}
		if badge.ScopeType == "" || badge.ScopeType == "none" {
			return true
		}
		if scopeID != "" && badge.ScopeID == scopeID {
			return true
		}
	}
	return false
}

// ScopeFunc extracts the target resource id (for example a class id) from a
// request so scoped badges can be checked.
type ScopeFunc func(*http.Request) string

// QueryScope reads a scope id from the query string, e.g. ?class_id=...
func QueryScope(key string) ScopeFunc {
	return func(r *http.Request) string {
		return r.URL.Query().Get(key)
	}
}

// RequirePermission rejects requests where the caller lacks perm before the
// handler runs.
func (m *AuthMiddleware) RequirePermission(perm string, scopeExtractors ...ScopeFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok {
				RespondError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
				return
			}

			var scopeID string
			for _, extract := range scopeExtractors {
				if value := extract(r); value != "" {
					scopeID = value
					break
				}
			}

			if !identity.HasPermission(perm, scopeID) {
				RespondError(w, http.StatusForbidden, "FORBIDDEN", "The caller lacks the required permission.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
