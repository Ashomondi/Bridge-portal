package onboarding

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

const defaultRetentionDays = 7

// SMSDispatcher sends transactional SMS messages.
type SMSDispatcher interface {
	SendTransactionalSMS(ctx context.Context, toPhoneNumber, message string) error
}

// StudentChargeHook is implemented by internal/billing. It creates the initial
// term fee charges for a newly admitted or verified student.
type StudentChargeHook interface {
	StudentChargeHook(ctx context.Context, schoolID, studentID, gradeLevelID, academicYearID string) error
}

// Service handles school provisioning, the principal setup wizard, staff and
// parent QR onboarding, registration verification, student admission and
// guardian phone changes.
type Service struct {
	db            *sql.DB
	audit         *database.AuditLogger
	sms           SMSDispatcher
	charges       StudentChargeHook
	retentionDays int
}

// NewService builds the onboarding service.
func NewService(db *sql.DB, audit *database.AuditLogger, sms SMSDispatcher, charges StudentChargeHook, retentionDays int) *Service {
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}
	return &Service{db: db, audit: audit, sms: sms, charges: charges, retentionDays: retentionDays}
}

// RegisterRoutes mounts the onboarding endpoints. requireInternal guards the
// Ikonex-team-only school provisioning route.
func (s *Service) RegisterRoutes(mux *http.ServeMux, auth *middleware.AuthMiddleware, requireInternal func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/admin/schools", requireInternal(http.HandlerFunc(s.handleCreateSchool)))
	mux.Handle("POST /api/v1/schools/setup-wizard", s.protect(auth, middleware.PermSchoolSetup, http.HandlerFunc(s.handleSetupWizard)))
	mux.Handle("POST /api/v1/staff-invites", s.protect(auth, middleware.PermStaffInvitesCreate, http.HandlerFunc(s.handleCreateStaffInvite)))
	mux.Handle("POST /api/v1/staff-invites/{id}/revoke", s.protect(auth, middleware.PermStaffInvitesCreate, http.HandlerFunc(s.handleRevokeStaffInvite)))

	mux.Handle("POST /api/v1/parent-qr", s.protect(auth, middleware.PermParentsQRManage, http.HandlerFunc(s.handleCreateParentQR)))
	mux.Handle("POST /api/v1/parent-qr/{id}/revoke", s.protect(auth, middleware.PermParentsQRManage, http.HandlerFunc(s.handleRevokeParentQR)))
	mux.Handle("POST /api/v1/verifier-grants", s.protect(auth, middleware.PermVerifierGrant, http.HandlerFunc(s.handleVerifierGrant)))
	mux.Handle("POST /api/v1/registrations", http.HandlerFunc(s.handleRegistrationSubmission))
	mux.Handle("POST /api/v1/registrations/{id}/verify", s.protect(auth, middleware.PermRegistrationsVerify, http.HandlerFunc(s.handleVerifyRegistration)))
	mux.Handle("POST /api/v1/students/admit", s.protect(auth, middleware.PermStudentsAdmit, http.HandlerFunc(s.handleAdmitStudent)))
	mux.Handle("POST /api/v1/students/re-enrol", s.protect(auth, middleware.PermStudentsAdmit, http.HandlerFunc(s.handleReEnrolStudents)))
	mux.Handle("POST /api/v1/guardians/me/phone-change/request-otp", s.protect(auth, middleware.PermPhoneChange, http.HandlerFunc(s.handlePhoneChangeRequestOTP)))
	mux.Handle("POST /api/v1/guardians/me/phone-change/confirm", s.protect(auth, middleware.PermPhoneChange, http.HandlerFunc(s.handlePhoneChangeConfirm)))
}

func (s *Service) protect(auth *middleware.AuthMiddleware, perm string, next http.Handler) http.Handler {
	handler := auth.RequirePermission(perm)(next)
	handler = auth.RequireTenant(handler)
	return auth.RequireAuth(handler)
}
