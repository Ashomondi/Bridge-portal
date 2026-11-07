package academics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Role names used by the platform (C59, Docs/Ikonex.md section 4).
const (
	RoleDean      = "dean_of_studies"
	RolePrincipal = "principal"
	RoleDeputy    = "deputy"
	RoleTeacher   = "teacher"
	RoleParent    = "parent"
)

// Permission codes owned by the Dean of Studies workspace (O16).
const (
	PermAssignmentsManage  = "academics.assignments.manage"
	PermTimetablesManage   = "academics.timetables.manage"
	PermCalendarManage     = "academics.calendar.manage"
	PermGradingScaleManage = "academics.grading_scale.manage"
	PermClassStudentsView  = "class.students.view"
	PermClassStatsView     = "class.stats.view"
	PermMarksReassign      = "marks.reassign"
)

// Machine-readable error codes shared with the frontend and Docs/API_CONTRACTS.md.
const (
	CodeUnauthenticated      = "UNAUTHENTICATED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeValidation           = "VALIDATION_ERROR"
	CodeClassTeacherScope    = "CLASS_TEACHER_SCOPE_REQUIRED"
	CodeSubjectTeacherScope  = "SUBJECT_TEACHER_NOT_ASSIGNED"
	CodeResultsLocked        = "RESULTS_LOCKED_RELEASED"
	CodeResultsIncomplete    = "RESULTS_INCOMPLETE_MARKS"
	CodeResultsNotReleased   = "RESULTS_NOT_RELEASED"
	CodeResultsNotReopened   = "RESULTS_NOT_REOPENED"
	CodeResultsReopened      = "RESULTS_REOPENED_USE_RERELEASE"
	CodeRereleaseNote        = "RERELEASE_NOTE_REQUIRED"
	CodeTermDateNote         = "TERM_DATE_CHANGE_NOTE_REQUIRED"
	CodeAIDrafterUnavailable = "AI_DRAFTER_UNAVAILABLE"
	CodeInternal             = "INTERNAL_ERROR"
)

// Identity is the authenticated caller context extracted by Evans's
// middleware. RegisterRoutes wraps handlers with the injected AuthMiddleware,
// which should place an Identity on the request context with WithIdentity.
type Identity struct {
	UserID       string   `json:"user_id"`
	SchoolID     string   `json:"school_id"`
	MembershipID string   `json:"membership_id"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions,omitempty"`
}

type identityContextKey struct{}

// WithIdentity attaches an Identity to a context. It is exported so the core
// middleware can populate the academics request context without importing
// unexported types.
func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

// IdentityFromContext returns the caller identity, if present.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityContextKey{}).(Identity)
	return id, ok
}

// Config holds the configurable parts of the academic rules.
type Config struct {
	// AllowPrincipalDeanOverride implements the O16 toggle. When true the
	// principal and deputy may perform Dean of Studies actions.
	AllowPrincipalDeanOverride bool
	// InternalServiceToken is used to sign X-User-Context for the AI drafter.
	InternalServiceToken string
	// AIServiceBaseURL is Sos's FastAPI base URL, e.g. http://ai_service:8000.
	AIServiceBaseURL string
}

// DefaultConfig returns the Release 1 default described by the integration
// plan: principal/deputy may stand in for the Dean unless switched off.
func DefaultConfig() Config {
	return Config{AllowPrincipalDeanOverride: true}
}

// Notification is a single in-app notification request. Ian's notifications
// package implements InAppNotifier at the consumer boundary.
type Notification struct {
	SchoolID        string         `json:"school_id"`
	RecipientUserID string         `json:"recipient_user_id"`
	Category        string         `json:"category"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// InAppNotifier is implemented by internal/notifications (Ian, Dev 5).
type InAppNotifier interface {
	CreateInAppNotification(ctx context.Context, n Notification) error
}

// TermDateFeeSyncer is implemented by internal/billing (Ian, Dev 5). It shifts
// the term's fee due dates, records the fee-change note and notifies parents.
type TermDateFeeSyncer interface {
	OnTermStartDateChanged(ctx context.Context, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error
}

type noopNotifier struct{}

func (noopNotifier) CreateInAppNotification(context.Context, Notification) error { return nil }

type noopFeeSyncer struct{}

func (noopFeeSyncer) OnTermStartDateChanged(context.Context, string, string, time.Time, string, string) error {
	return nil
}

// Deps is the complete set of collaborators required by academics.
type Deps struct {
	Store             Store
	Notifier          InAppNotifier
	FeeSyncer         TermDateFeeSyncer
	CommentDrafter    CommentDrafter
	AuthMiddleware    func(http.Handler) http.Handler
	IdentityExtractor func(*http.Request) (Identity, bool)
	Config            Config
}

// Service implements all CBC academic workflows for Release 1.
type Service struct {
	store     Store
	notifier  InAppNotifier
	feeSyncer TermDateFeeSyncer
	drafter   CommentDrafter
	cfg       Config
}

// NewService builds an academics service, filling in safe no-op collaborators
// when optional integrations have not been wired yet.
func NewService(deps Deps) *Service {
	s := &Service{
		store:     deps.Store,
		notifier:  deps.Notifier,
		feeSyncer: deps.FeeSyncer,
		drafter:   deps.CommentDrafter,
		cfg:       deps.Config,
	}
	if s.notifier == nil {
		s.notifier = noopNotifier{}
	}
	if s.feeSyncer == nil {
		s.feeSyncer = noopFeeSyncer{}
	}
	return s
}

// RegisterRoutes mounts every academics endpoint on mux. Evans's main.go only
// needs a single call: academics.RegisterRoutes(mux, academics.Deps{...}).
// routeHandler is a method on Service with the http.HandlerFunc shape. Feature
// files register their own routes in init() so each stacked feature commit can
// compile and be reviewed independently.
type routeHandler func(*Service, http.ResponseWriter, *http.Request)

type routeRegistration struct {
	method  string
	pattern string
	handler routeHandler
}

var routeRegistrations []routeRegistration

func registerRoute(method, pattern string, handler routeHandler) {
	routeRegistrations = append(routeRegistrations, routeRegistration{method: method, pattern: pattern, handler: handler})
}

// RegisterRoutes mounts every academics endpoint on mux. Evans's main.go only
// needs a single call: academics.RegisterRoutes(mux, academics.Deps{...}).
func RegisterRoutes(mux *http.ServeMux, deps Deps) {
	s := NewService(deps)
	wrap := func(h http.HandlerFunc) http.Handler {
		var handler http.Handler = s.requireIdentity(h)
		if deps.IdentityExtractor != nil {
			handler = extractIdentity(deps.IdentityExtractor, handler)
		}
		if deps.AuthMiddleware != nil {
			return deps.AuthMiddleware(handler)
		}
		return handler
	}
	for _, registration := range routeRegistrations {
		reg := registration
		mux.Handle(reg.method+" "+reg.pattern, wrap(func(w http.ResponseWriter, r *http.Request) {
			reg.handler(s, w, r)
		}))
	}
}

func init() {
	// Explicitly Release 2 (Docs/Ikonex.md Appendix C, REQUIREMENTS.md D).
	registerRoute(http.MethodPost, "/api/v1/syllabus/coverage", func(_ *Service, w http.ResponseWriter, _ *http.Request) {
		respondError(w, &APIError{
			Status:  http.StatusNotImplemented,
			Code:    "NOT_IMPLEMENTED",
			Message: "Syllabus coverage and AI strand mapping are scheduled for Release 2.",
		})
	})
}

// APIError is the typed error returned by the service and mapped to the
// standard error envelope by the HTTP layer.
type APIError struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *APIError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func errUnauthenticated() *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: "Authentication is required."}
}
func errForbidden(code, message string) *APIError {
	if code == "" {
		code = CodeForbidden
	}
	return &APIError{Status: http.StatusForbidden, Code: code, Message: message}
}
func errNotFound(message string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: CodeNotFound, Message: message}
}
func errValidation(message string) *APIError {
	return &APIError{Status: http.StatusUnprocessableEntity, Code: CodeValidation, Message: message}
}
func errConflict(code, message string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: code, Message: message}
}
func errInternal(err error) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: CodeInternal, Message: "An internal error occurred.", Details: map[string]any{"cause": err.Error()}}
}

func asAPIError(err error) *APIError {
	if err == nil {
		return nil
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	if errors.Is(err, ErrNotFound) {
		return errNotFound("The requested academic record was not found.")
	}
	return errInternal(err)
}

func (s *Service) requireIdentity(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := IdentityFromContext(r.Context())
		if !ok || identity.SchoolID == "" || identity.MembershipID == "" {
			respondError(w, errUnauthenticated())
			return
		}
		next(w, r)
	}
}

// extractIdentity lets the core middleware adapt its own request context to the
// academics Identity without this package importing middleware.
func extractIdentity(extractor func(*http.Request) (Identity, bool), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := IdentityFromContext(r.Context()); !ok {
			if identity, ok := extractor(r); ok {
				r = r.WithContext(WithIdentity(r.Context(), identity))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) actor(ctx context.Context) (Identity, error) {
	id, ok := IdentityFromContext(ctx)
	if !ok || id.SchoolID == "" || id.MembershipID == "" {
		return Identity{}, errUnauthenticated()
	}
	return id, nil
}

func (s *Service) tx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.store == nil {
		return errInternal(errors.New("academics: store is not configured"))
	}
	return s.store.WithTx(ctx, schoolID, fn)
}

// canManage reports whether the caller holds the Dean-of-Studies authority for
// an academic configuration action. The O16 flag controls the principal/deputy
// stand-in rather than hardcoding it.
func (s *Service) canManage(actor Identity, permission string) bool {
	switch actor.Role {
	case RoleDean:
		return true
	case RolePrincipal, RoleDeputy:
		if s.cfg.AllowPrincipalDeanOverride {
			return true
		}
	}
	return contains(actor.Permissions, permission)
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func (s *Service) notify(ctx context.Context, n Notification) error {
	if s.notifier == nil {
		return nil
	}
	return s.notifier.CreateInAppNotification(ctx, n)
}

func (s *Service) notifyMany(ctx context.Context, schoolID string, recipients []string, category, title, body string, metadata map[string]any) (int, error) {
	seen := map[string]bool{}
	queued := 0
	for _, uid := range recipients {
		if uid == "" || seen[uid] {
			continue
		}
		seen[uid] = true
		if err := s.notify(ctx, Notification{
			SchoolID:        schoolID,
			RecipientUserID: uid,
			Category:        category,
			Title:           title,
			Body:            body,
			Metadata:        metadata,
		}); err != nil {
			return queued, err
		}
		queued++
	}
	return queued, nil
}

func (s *Service) audit(ctx context.Context, tx Store, actor Identity, action, resourceType, resourceID string, details map[string]any) error {
	return tx.AppendAudit(ctx, actor.SchoolID, AuditEvent{
		SchoolID:     actor.SchoolID,
		UserID:       actor.UserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	})
}

// ---- JSON helpers shared by every handler. ----

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   data,
	})
}

func respondError(w http.ResponseWriter, err error) {
	apiErr := asAPIError(err)
	if apiErr == nil {
		apiErr = errInternal(errors.New("unknown error"))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "error",
		"error": map[string]any{
			"code":    apiErr.Code,
			"message": apiErr.Message,
			"details": apiErr.Details,
		},
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errValidation("Request body is not valid JSON: " + err.Error())
	}
	return nil
}

func normalizeRole(role string) string { return strings.ToLower(strings.TrimSpace(role)) }
