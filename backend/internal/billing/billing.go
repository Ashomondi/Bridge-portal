package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Roles recognized by billing
const (
	RoleIkonexAdmin = "ikonex_admin"
	RoleBursar      = "bursar"
	RolePrincipal   = "principal"
	RoleDeputy      = "deputy"
	RoleParent      = "parent"
)

// Error codes
const (
	CodeUnauthenticated       = "UNAUTHENTICATED"
	CodeForbidden             = "FORBIDDEN"
	CodeNotFound              = "NOT_FOUND"
	CodeValidation            = "VALIDATION_ERROR"
	CodeFeeChangeNoteRequired = "FEE_CHANGE_NOTE_REQUIRED"
	CodeChatPaymentPaused     = "CHAT_PAYMENT_PAUSED_PHONE_CHANGE"
	CodeParentConfirmationReq = "PARENT_CONFIRMATION_REQUIRED"
	CodeInternal              = "INTERNAL_ERROR"
)

// APIError is a structured error that carries HTTP status and machine-readable code.
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func errNotFound(msg string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: CodeNotFound, Message: msg}
}

func errForbidden(msg string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: CodeForbidden, Message: msg}
}

func errValidation(msg string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: CodeValidation, Message: msg}
}

func errUnprocessable(code, msg string) *APIError {
	return &APIError{Status: http.StatusUnprocessableEntity, Code: code, Message: msg}
}

func errInternal(err error) *APIError {
	msg := "internal server error"
	if err != nil {
		msg = err.Error()
	}
	return &APIError{Status: http.StatusInternalServerError, Code: CodeInternal, Message: msg}
}

// Identity context key and helpers
type identityKey struct{}

// Identity represents the caller identity
type Identity struct {
	UserID       string   `json:"user_id"`
	SchoolID     string   `json:"school_id"`
	MembershipID string   `json:"membership_id"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions,omitempty"`
}

// WithIdentity injects Identity into Context
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// IdentityFromContext extracts Identity from Context
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// InAppNotifier interface for dispatching notifications
type InAppNotifier interface {
	CreateInAppNotification(ctx context.Context, schoolID, recipientUserID, category, title, body string, metadata map[string]interface{}) error
}

// AuditLogger interface for writing to audit_log
type AuditLogger interface {
	LogAction(ctx context.Context, schoolID, userID, action, resourceType, resourceID string, details map[string]interface{}) error
}

// AsynqEnqueuer queues background jobs
type AsynqEnqueuer interface {
	EnqueueSTKPoll(ctx context.Context, schoolID, checkoutRequestID string, delay time.Duration) error
}

// Models

type FeePeriod struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	AcademicYearID string    `json:"academic_year_id"`
	TermID         string    `json:"term_id"`
	PeriodType     string    `json:"period_type"`
	Name           string    `json:"name"`
	DueDate        time.Time `json:"due_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type FeeStructure struct {
	ID             string             `json:"id"`
	SchoolID       string             `json:"school_id"`
	AcademicYearID string             `json:"academic_year_id"`
	GradeLevelID   string             `json:"grade_level_id"`
	TermID         string             `json:"term_id"`
	TotalAmount    float64            `json:"total_amount"`
	Items          []FeeStructureItem `json:"items,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

type FeeStructureItem struct {
	ID             string  `json:"id"`
	FeeStructureID string  `json:"fee_structure_id"`
	Title          string  `json:"title"`
	Amount         float64 `json:"amount"`
	IsCompulsory   bool    `json:"is_compulsory"`
}

type FeeCharge struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	StudentID      string    `json:"student_id"`
	FeePeriodID    string    `json:"fee_period_id"`
	FeeStructureID string    `json:"fee_structure_id"`
	Amount         float64   `json:"amount"`
	DueDate        time.Time `json:"due_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type FeeChangeNote struct {
	ID                    string     `json:"id"`
	SchoolID              string     `json:"school_id"`
	TermID                string     `json:"term_id,omitempty"`
	FeePeriodID           string     `json:"fee_period_id,omitempty"`
	ChangeType            string     `json:"change_type"`
	AmountDelta           float64    `json:"amount_delta"`
	OldDueDate            *time.Time `json:"old_due_date,omitempty"`
	NewDueDate            *time.Time `json:"new_due_date,omitempty"`
	Note                  string     `json:"note"`
	CreatedBy             string     `json:"created_by"`
	AffectedStudentsCount int        `json:"affected_students_count"`
	CreatedAt             time.Time  `json:"created_at"`
}

type LedgerEntry struct {
	ID          string    `json:"id"`
	SchoolID    string    `json:"school_id"`
	StudentID   string    `json:"student_id"`
	TermID      string    `json:"term_id,omitempty"`
	EntryType   string    `json:"entry_type"` // charge, payment, credit_applied, adjustment
	Amount      float64   `json:"amount"`
	Reference   string    `json:"reference,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type FeeLedger struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	StudentID      string    `json:"student_id"`
	AcademicYearID string    `json:"academic_year_id"`
	TermID         string    `json:"term_id"`
	OpeningBalance float64   `json:"opening_balance"`
	BilledAmount   float64   `json:"billed_amount"`
	PaidAmount     float64   `json:"paid_amount"`
	CurrentBalance float64   `json:"current_balance"`
	CreditBalance  float64   `json:"credit_balance"`
	DueDate        time.Time `json:"due_date"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StudentBalance struct {
	StudentID       string             `json:"student_id"`
	AdmissionNumber string             `json:"admission_number"`
	FullName        string             `json:"full_name"`
	OpeningBalance  float64            `json:"opening_balance"`
	BilledAmount    float64            `json:"billed_amount"`
	PaidAmount      float64            `json:"paid_amount"`
	CurrentBalance  float64            `json:"current_balance"`
	CreditBalance   float64            `json:"credit_balance"`
	DueDate         time.Time          `json:"due_date"`
	Items           []FeeStructureItem `json:"items,omitempty"`
	Notes           []FeeChangeNote    `json:"notes,omitempty"`
}

type Payment struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	StudentID      *string   `json:"student_id,omitempty"`
	TransactionRef string    `json:"transaction_ref"`
	Provider       string    `json:"provider"` // mpesa_stk, mpesa_c2b, bank
	Amount         float64   `json:"amount"`
	PayerPhone     string    `json:"payer_phone,omitempty"`
	PayerName      string    `json:"payer_name,omitempty"`
	MatchedStatus  string    `json:"matched_status"` // matched, unmatched
	MatchedBy      *string   `json:"matched_by,omitempty"`
	PaymentDate    time.Time `json:"payment_date"`
	RawPayload     string    `json:"raw_payload,omitempty"`
}

type UnmatchedPayment struct {
	ID                string     `json:"id"`
	SchoolID          string     `json:"school_id"`
	PaymentID         string     `json:"payment_id"`
	TransactionRef    string     `json:"transaction_ref"`
	Amount            float64    `json:"amount"`
	PayerPhone        string     `json:"payer_phone,omitempty"`
	PayerName         string     `json:"payer_name,omitempty"`
	RawAccountRef     string     `json:"raw_account_ref,omitempty"`
	Status            string     `json:"status"` // unmatched, resolved
	ResolvedStudentID *string    `json:"resolved_student_id,omitempty"`
	ResolvedBy        *string    `json:"resolved_by,omitempty"`
	ResolutionNote    string     `json:"resolution_note,omitempty"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type SchoolMpesaConfig struct {
	SchoolID                string    `json:"school_id"`
	Shortcode               string    `json:"shortcode"`
	Environment             string    `json:"environment"`
	PasskeyEncrypted        string    `json:"-"`
	ConsumerKeyEncrypted    string    `json:"-"`
	ConsumerSecretEncrypted string    `json:"-"`
	TestPaymentVerified     bool      `json:"test_payment_verified"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Service manages all billing, M-Pesa, ledgers, and fee changes.
type Service struct {
	store        Store
	notifier     InAppNotifier
	auditLogger  AuditLogger
	enqueuer     AsynqEnqueuer
	darajaClient DarajaClient
	callbackURL  string
}

// NewService constructs a Billing Service.
func NewService(store Store, notifier InAppNotifier, auditLogger AuditLogger, enqueuer AsynqEnqueuer) *Service {
	return &Service{
		store:        store,
		notifier:     notifier,
		auditLogger:  auditLogger,
		enqueuer:     enqueuer,
		darajaClient: NewHTTPDarajaClient(nil),
		callbackURL:  "https://api.ikonex.academy/api/v1/billing/mpesa/callback",
	}
}

// SetDarajaClient sets a custom or mock DarajaClient
func (s *Service) SetDarajaClient(client DarajaClient) {
	s.darajaClient = client
}

// SetCallbackURL sets the Daraja callback URL
func (s *Service) SetCallbackURL(url string) {
	s.callbackURL = url
}

// GetStore returns the underlying Store.
func (s *Service) GetStore() Store {
	return s.store
}

// RegisterRoutes registers billing HTTP routes onto mux.
func (s *Service) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	wrap := func(h http.HandlerFunc) http.Handler {
		if authMiddleware != nil {
			return authMiddleware(h)
		}
		return h
	}

	// Fee Structures
	mux.Handle("/api/v1/billing/fee-structures", wrap(s.handleLoadFeeStructure))

	// Mid-Year Fee Changes
	mux.Handle("/api/v1/fee-changes", wrap(s.handleFeeChange))

	// Student Fee Balances & Ledgers
	mux.HandleFunc("/api/v1/students/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/fee-balance") {
			wrap(s.handleStudentFeeBalance).ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(path, "/ledger") {
			wrap(s.handleStudentLedger).ServeHTTP(w, r)
			return
		}
		respondError(w, errNotFound("endpoint not found"))
	})

	// M-Pesa endpoints will be wired in Task 3
	mux.Handle("/api/v1/payments/stk-push", wrap(s.handleSTKPush))
	mux.HandleFunc("/api/v1/billing/mpesa/callback", s.handleSTKCallbackHTTP)
	mux.HandleFunc("/api/v1/billing/mpesa/c2b-confirmation", s.handleC2BConfirmationHTTP)
	mux.Handle("/api/v1/payments/unmatched", wrap(s.handleListUnmatchedPayments))
	mux.HandleFunc("/api/v1/payments/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/match") {
			wrap(s.handleMatchPayment).ServeHTTP(w, r)
			return
		}
		respondError(w, errNotFound("endpoint not found"))
	})
}

// Helper JSON responses
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   data,
	})
}

func respondError(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = errInternal(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "error",
		"error": map[string]any{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		},
	})
}
