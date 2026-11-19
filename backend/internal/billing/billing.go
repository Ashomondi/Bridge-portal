package billing

import (
	"context"
	"net/http"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/academics"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

// StudentChargeHook is injected into onboarding so newly admitted/enrolled
// students receive applicable term fee charges within the caller's transaction.
type StudentChargeHook interface {
	ChargeAdmittedStudent(ctx context.Context, tx pgx.Tx, schoolID, studentID, academicYearID string) error
}

// StudentBillingHook is a compatible alias for StudentChargeHook.
type StudentBillingHook = StudentChargeHook

// TermDateFeeSyncer aliases academics.TermDateFeeSyncer so billing implements
// the exact academic term-date synchronization contract.
type TermDateFeeSyncer = academics.TermDateFeeSyncer

// InAppNotifier is the consumer-oriented notification interface used by billing
// and academics within an open tenant transaction.
type InAppNotifier = contracts.InAppNotifier

// CreateNotification is the notification payload alias for InAppNotifier.
type CreateNotification = contracts.InAppNotification

// ShortcodeResolver maps a receiving M-Pesa Paybill/BusinessShortCode to a school_id
// so incoming C2B webhooks can open a tenant-scoped transaction without bypassing
// PostgreSQL RLS or trusting caller-supplied school identifiers.
type ShortcodeResolver func(ctx context.Context, shortcode string) (string, error)

// CheckoutRequestResolver maps a globally unique Daraja CheckoutRequestID to the
// trusted school_id stored on mpesa_stk_requests.
type CheckoutRequestResolver func(ctx context.Context, checkoutRequestID string) (string, error)

// CallbackTokenResolver maps a SHA-256 hash of a server-generated per-request
// STK callback token to its trusted (school_id, checkout_request_id) binding.
type CallbackTokenResolver func(ctx context.Context, callbackTokenHash string) (schoolID string, checkoutRequestID string, err error)

// Service handles fee structures, student ledgers, M-Pesa STK push payments, and reconciliation.
type Service struct {
	DB                     contracts.TenantDB
	SystemDB               func(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error
	Audit                  contracts.AuditLogger
	Notifier               InAppNotifier
	Daraja                 DarajaProvider
	PollQueue              STKPollQueue
	ResolveShortcode       ShortcodeResolver
	ResolveCheckoutRequest CheckoutRequestResolver
	ResolveCallbackToken   CallbackTokenResolver
	CallbackAuthToken      string
	TrustedCallbackProxies []string
	VerifyCallback         func(r *http.Request) bool
	Now                    func() time.Time
}

func (s *Service) requireNotifier() error {
	if s == nil || s.Notifier == nil {
		return contracts.Fail(http.StatusInternalServerError, "NOTIFIER_REQUIRED", "A transactional in-app notifier is required for payment settlement.")
	}
	return nil
}

// withSystemTenantTx executes fn inside a tenant-scoped transaction for authenticated
// machine-to-machine callbacks or background jobs that lack an end-user Principal.
func (s *Service) withSystemTenantTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
	if !contracts.ValidID(schoolID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_SCHOOL", "A valid school_id is required.")
	}
	if s == nil {
		return contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if s.SystemDB != nil {
		return s.SystemDB(ctx, schoolID, fn)
	}
	if txFn, ok := s.DB.(contracts.TenantTxFunc); ok && txFn != nil {
		return txFn(ctx, schoolID, fn)
	}
	return contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// ChargeGradeForTerm charges all actively enrolled students in the structure's
// school, grade level, and academic year for the structure's term period.
func (s *Service) ChargeGradeForTerm(ctx context.Context, tx pgx.Tx, structureID string) error {
	return s.chargeGradeForTerm(ctx, tx, structureID, "")
}

// ChargeStudent charges a single student for a specific fee period idempotently.
func (s *Service) ChargeStudent(ctx context.Context, tx pgx.Tx, studentID, feePeriodID string) error {
	return s.chargeStudent(ctx, tx, studentID, feePeriodID, "")
}

// ChargeAdmittedStudent implements StudentChargeHook / StudentBillingHook.
func (s *Service) ChargeAdmittedStudent(ctx context.Context, tx pgx.Tx, schoolID, studentID, academicYearID string) error {
	return s.chargeAdmittedStudent(ctx, tx, schoolID, studentID, academicYearID)
}

func RegisterRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	if service == nil {
		service = &Service{}
	}
	mux.HandleFunc("POST /api/v1/billing/fee-structures", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input LoadFeeStructureInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		structure, err := service.LoadFeeStructure(r.Context(), p, input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusCreated, structure, nil)
	})
	mux.HandleFunc("POST /api/v1/billing/fee-structures/charge-current", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		res, err := service.ChargeCurrentTermStructures(r.Context(), p)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})
	registerLedgerRoutes(mux, service, identity)
	registerFeeChangeRoutes(mux, service, identity)
	registerMpesaRoutes(mux, service, identity)
}

var (
	_ StudentChargeHook           = (*Service)(nil)
	_ StudentBillingHook          = (*Service)(nil)
	_ academics.TermDateFeeSyncer = (*Service)(nil)
	_ TermDateFeeSyncer           = (*Service)(nil)
	_ contracts.InAppNotifier     = (*notifications.Service)(nil)
)
