package onboarding

import (
	"database/sql"
	"net/http"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

// Service handles school provisioning, the principal setup wizard, staff QR
// invites, and (in later feature work) parent registration and student
// admission.
type Service struct {
	db    *sql.DB
	audit *database.AuditLogger
}

// NewService builds the onboarding service.
func NewService(db *sql.DB, audit *database.AuditLogger) *Service {
	return &Service{db: db, audit: audit}
}

// RegisterRoutes mounts the onboarding endpoints. requireInternal guards the
// Ikonex-team-only school provisioning route.
func (s *Service) RegisterRoutes(mux *http.ServeMux, auth *middleware.AuthMiddleware, requireInternal func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/admin/schools", requireInternal(http.HandlerFunc(s.handleCreateSchool)))
	mux.Handle("POST /api/v1/schools/setup-wizard", s.protect(auth, middleware.PermSchoolSetup, http.HandlerFunc(s.handleSetupWizard)))
	mux.Handle("POST /api/v1/staff-invites", s.protect(auth, middleware.PermStaffInvitesCreate, http.HandlerFunc(s.handleCreateStaffInvite)))
	mux.Handle("POST /api/v1/staff-invites/{id}/revoke", s.protect(auth, middleware.PermStaffInvitesCreate, http.HandlerFunc(s.handleRevokeStaffInvite)))
}

func (s *Service) protect(auth *middleware.AuthMiddleware, perm string, next http.Handler) http.Handler {
	handler := auth.RequirePermission(perm)(next)
	handler = auth.RequireTenant(handler)
	return auth.RequireAuth(handler)
}
