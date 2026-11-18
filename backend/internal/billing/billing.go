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
// so incoming Daraja callbacks and C2B webhooks can open a tenant-scoped transaction
// without bypassing PostgreSQL RLS.
type ShortcodeResolver func(ctx context.Context, shortcode string) (string, error)

// Service handles fee structures, student ledgers, M-Pesa STK push payments, and reconciliation.
type Service struct {
	DB                contracts.TenantDB
	Audit             contracts.AuditLogger
	Notifier          InAppNotifier
	Daraja            DarajaProvider
	PollQueue         STKPollQueue
	ResolveShortcode  ShortcodeResolver
	CallbackAuthToken string
	VerifyCallback    func(r *http.Request) bool
	Now               func() time.Time
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
