package internalapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
)

// Header Keys
const (
	HeaderInternalServiceToken = "X-Internal-Service-Token"
	HeaderUserContext          = "X-User-Context"
)

// Models for internal facts
type SchoolProfile struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	County  string `json:"county,omitempty"`
	LogoURL string `json:"logo_url,omitempty"`
}

type LinkedChild struct {
	StudentID       string `json:"student_id"`
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	ClassName       string `json:"class_name,omitempty"`
	Relationship    string `json:"relationship"`
}

type AttendanceSummary struct {
	StudentID      string  `json:"student_id"`
	TermID         string  `json:"term_id"`
	TotalDays      int     `json:"total_days"`
	DaysPresent    int     `json:"days_present"`
	DaysAbsent     int     `json:"days_absent"`
	AttendanceRate float64 `json:"attendance_rate"` // e.g. 96.5%
}

type SubjectMarkSummary struct {
	SubjectCode string  `json:"subject_code"`
	SubjectName string  `json:"subject_name"`
	Score       float64 `json:"score"`
	CBCLevel    string  `json:"cbc_level"` // EE, ME, AE, BE
}

type StudentMarksSummary struct {
	StudentID   string               `json:"student_id"`
	TermID      string               `json:"term_id"`
	Subjects    []SubjectMarkSummary `json:"subjects"`
	MeanGrade   string               `json:"mean_grade,omitempty"`
	PublishedAt time.Time            `json:"published_at"`
}

type AIAuditLogEntry struct {
	ID            string                 `json:"id,omitempty"`
	SchoolID      string                 `json:"school_id"`
	UserID        string                 `json:"user_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	Channel       string                 `json:"channel"`
	Feature       string                 `json:"feature"`
	Intent        string                 `json:"intent,omitempty"`
	EndpointsRead []string               `json:"endpoints_read,omitempty"`
	ActionTaken   string                 `json:"action_taken,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Store persistence interface for Internal API facts
type Store interface {
	GetSchoolProfile(ctx context.Context, schoolID string) (*SchoolProfile, error)
	ListChildrenForParent(ctx context.Context, schoolID, parentUserID string) ([]LinkedChild, error)
	IsGuardianOf(ctx context.Context, schoolID, studentID, userID string) (bool, error)
	GetAttendanceSummary(ctx context.Context, schoolID, studentID, termID string) (*AttendanceSummary, error)
	GetMarksSummary(ctx context.Context, schoolID, studentID, termID string) (*StudentMarksSummary, error)
	SaveAIAuditLog(ctx context.Context, entry *AIAuditLogEntry) error
}

// BillingBridge interface to billing service
type BillingBridge interface {
	GetStudentFeeBalance(ctx context.Context, schoolID, studentID, callerUserID, callerRole string) (*billing.StudentBalance, error)
	InitiateSTKPush(ctx context.Context, req billing.STKPushRequest) (*billing.STKPushResponse, error)
}

// Handler handles internal AI service requests and frontend chat proxies
type Handler struct {
	serviceToken string
	signer       *ContextSigner
	store        Store
	billing      BillingBridge
	aiServiceURL string
}

func NewHandler(serviceToken string, signer *ContextSigner, store Store, billing BillingBridge, aiServiceURL string) (*Handler, error) {
	if len(strings.TrimSpace(serviceToken)) < 16 {
		return nil, errors.New("internalapi: service token is required and must be at least 16 characters")
	}
	if signer == nil {
		return nil, errors.New("internalapi: context signer is required")
	}
	if store == nil {
		return nil, errors.New("internalapi: store is required")
	}
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8000"
	}
	return &Handler{
		serviceToken: strings.TrimSpace(serviceToken),
		signer:       signer,
		store:        store,
		billing:      billing,
		aiServiceURL: strings.TrimRight(aiServiceURL, "/"),
	}, nil
}

// MustNewHandler constructs a Handler or panics on configuration failure
func MustNewHandler(serviceToken string, signer *ContextSigner, store Store, billing BillingBridge, aiServiceURL string) *Handler {
	h, err := NewHandler(serviceToken, signer, store, billing, aiServiceURL)
	if err != nil {
		panic(err)
	}
	return h
}

// RegisterRoutes registers internal endpoints onto mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Internal AI Endpoints (enforcing X-Internal-Service-Token + signed X-User-Context)
	mux.HandleFunc("/api/v1/internal/school-profile", h.authenticateInternal(h.handleSchoolProfile))
	mux.HandleFunc("/api/v1/internal/parents/", h.authenticateInternal(h.handleParentChildren))
	mux.HandleFunc("/api/v1/internal/students/", h.authenticateInternal(h.handleStudentEndpoints))
	mux.HandleFunc("/api/v1/internal/payments/stk-push", h.authenticateInternal(h.handleSTKPush))
	mux.HandleFunc("/api/v1/internal/ai/audit-log", h.authenticateInternal(h.handleAIAuditLog))

	// Frontend Chat Proxy (accepts parent query, signs X-User-Context, forwards to AI)
	mux.HandleFunc("/api/v1/chat/message", h.handleFrontendChatMessage)
}

// authenticateInternal middleware enforces service token and verifies HMAC X-User-Context
func (h *Handler) authenticateInternal(next func(http.ResponseWriter, *http.Request, *UserContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Validate X-Internal-Service-Token
		token := r.Header.Get(HeaderInternalServiceToken)
		if token != h.serviceToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid internal service token"})
			return
		}

		// 2. Validate and verify HMAC X-User-Context
		signedContext := r.Header.Get(HeaderUserContext)
		if signedContext == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing X-User-Context header"})
			return
		}

		uContext, err := h.signer.Verify(signedContext)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("invalid user context: %v", err)})
			return
		}
		if uContext.UserID == "" || uContext.SchoolID == "" || uContext.Role == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "incomplete user context"})
			return
		}
		entry := &AIAuditLogEntry{
			SchoolID:      uContext.SchoolID,
			UserID:        uContext.UserID,
			SessionID:     uContext.SessionID,
			Channel:       uContext.Channel,
			Feature:       "internal_api",
			Intent:        r.Method,
			EndpointsRead: []string{r.URL.Path},
			ActionTaken:   "request_received",
			CreatedAt:     time.Now(),
		}
		if err := h.withTenantStore(r.Context(), uContext.SchoolID, func(store Store) error {
			return store.SaveAIAuditLog(r.Context(), entry)
		}); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to record AI request"})
			return
		}

		next(w, r, uContext)
	}
}

// 1. GET /api/v1/internal/school-profile
func (h *Handler) handleSchoolProfile(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var profile *SchoolProfile
	err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
		var err error
		profile, err = store.GetSchoolProfile(r.Context(), u.SchoolID)
		return err
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "school profile not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(profile)
}

// 2. GET /api/v1/internal/parents/{user_id}/children
func (h *Handler) handleParentChildren(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract parent user_id from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/internal/parents/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "children" {
		http.NotFound(w, r)
		return
	}
	parentUserID := parts[0]

	// Parent can only query their own children unless staff
	if u.Role == "parent" && parentUserID != u.UserID {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot query children of another guardian"})
		return
	}

	var children []LinkedChild
	err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
		var err error
		children, err = store.ListChildrenForParent(r.Context(), u.SchoolID, parentUserID)
		return err
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(children)
}

// 3, 4, 5. Student endpoints: fee-balance, attendance-summary, marks-summary
func (h *Handler) handleStudentEndpoints(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	// Extract student ID: /api/v1/internal/students/{id}/...
	parts := strings.Split(path, "/")
	studentID := ""
	action := ""
	for i, p := range parts {
		if p == "students" && i+1 < len(parts) {
			studentID = parts[i+1]
			if i+2 < len(parts) {
				action = parts[i+2]
			}
			break
		}
	}

	if studentID == "" || len(parts) != 7 || action == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "student ID required"})
		return
	}
	if action != "fee-balance" && action != "attendance-summary" && action != "marks-summary" {
		http.NotFound(w, r)
		return
	}

	// Enforce Guardian-Child ownership (Docs/API_CONTRACTS.md, SECURITY.md)
	// Parent can ONLY query their own children!
	if u.Role == "parent" {
		var isGuardian bool
		err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
			var err error
			isGuardian, err = store.IsGuardianOf(r.Context(), u.SchoolID, studentID, u.UserID)
			return err
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to verify guardian relationship"})
			return
		}
		if !isGuardian {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "parent is not a registered guardian of this student"})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	switch action {
	case "fee-balance":
		// Query billing service
		if h.billing == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bal, err := h.billing.GetStudentFeeBalance(r.Context(), u.SchoolID, studentID, u.UserID, u.Role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(bal)

	case "attendance-summary":
		termID := r.URL.Query().Get("term_id")
		var summary *AttendanceSummary
		err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
			var err error
			summary, err = store.GetAttendanceSummary(r.Context(), u.SchoolID, studentID, termID)
			return err
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(summary)

	case "marks-summary":
		termID := r.URL.Query().Get("term_id")
		var summary *StudentMarksSummary
		err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
			var err error
			summary, err = store.GetMarksSummary(r.Context(), u.SchoolID, studentID, termID)
			return err
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(summary)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// 6. POST /api/v1/internal/payments/stk-push
func (h *Handler) handleSTKPush(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req billing.STKPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.SchoolID = u.SchoolID
	req.Source = "chat"

	// Explicit parent confirmation check (AC-23)
	if !req.ParentConfirmed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "PARENT_CONFIRMATION_REQUIRED",
			"message": "explicit parent confirmation required before initiating payment",
		})
		return
	}

	// Enforce parent-child ownership
	if u.Role == "parent" {
		var isGuardian bool
		err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
			var err error
			isGuardian, err = store.IsGuardianOf(r.Context(), u.SchoolID, req.StudentID, u.UserID)
			return err
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to verify guardian relationship"})
			return
		}
		if !isGuardian {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "parent is not a registered guardian of this student"})
			return
		}
	}

	if h.billing == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := h.billing.InitiateSTKPush(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// 7. POST /api/v1/internal/ai/audit-log (C21, AC-28, SECURITY.md)
func (h *Handler) handleAIAuditLog(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var entry AIAuditLogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	entry.SchoolID = u.SchoolID
	if entry.UserID == "" {
		entry.UserID = u.UserID
	}
	if entry.Channel == "" {
		entry.Channel = u.Channel
	}
	if entry.SessionID == "" {
		entry.SessionID = u.SessionID
	}
	entry.CreatedAt = time.Now()

	if err := h.withTenantStore(r.Context(), u.SchoolID, func(store Store) error {
		return store.SaveAIAuditLog(r.Context(), &entry)
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to record AI action"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged"})
}

// Frontend Chat Proxy: POST /api/v1/chat/message (enforcing parent auth, child ownership, and AI forwarding)
func (h *Handler) handleFrontendChatMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := billing.IdentityFromContext(r.Context())
	if !ok || id.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "authentication required"})
		return
	}

	var req struct {
		Message   string `json:"message"`
		StudentID string `json:"student_id"`
		ChildID   string `json:"child_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}
	if req.StudentID == "" && req.ChildID != "" {
		req.StudentID = req.ChildID
	}
	if strings.TrimSpace(req.Message) == "" || req.StudentID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "message and student_id are required"})
		return
	}

	// Enforce parent-child ownership
	if id.Role == "parent" {
		var isGuardian bool
		err := h.withTenantStore(r.Context(), id.SchoolID, func(store Store) error {
			var err error
			isGuardian, err = store.IsGuardianOf(r.Context(), id.SchoolID, req.StudentID, id.UserID)
			return err
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to verify guardian relationship"})
			return
		}
		if !isGuardian {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "parent is not a registered guardian of this student"})
			return
		}
	}

	uCtx := UserContext{
		UserID:   id.UserID,
		SchoolID: id.SchoolID,
		Role:     id.Role,
		Channel:  "in_app",
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
	}
	signedCtx, err := h.signer.Sign(uCtx, 15*time.Minute)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to sign user context"})
		return
	}

	aiReqBody, _ := json.Marshal(map[string]interface{}{
		"school_id":  id.SchoolID,
		"user_id":    id.UserID,
		"student_id": req.StudentID,
		"query":      req.Message,
	})
	aiHTTPReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.aiServiceURL+"/api/v1/ai/parent-query", bytes.NewReader(aiReqBody))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	aiHTTPReq.Header.Set("Content-Type", "application/json")
	aiHTTPReq.Header.Set(HeaderInternalServiceToken, h.serviceToken)
	aiHTTPReq.Header.Set(HeaderUserContext, signedCtx)

	client := &http.Client{Timeout: 30 * time.Second}
	aiResp, err := client.Do(aiHTTPReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("ai service unavailable: %v", err)})
		return
	}
	defer aiResp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(aiResp.StatusCode)
	_, _ = io.Copy(w, aiResp.Body)
}

// MemoryStore implementation for internalapi
type MemoryStore struct {
	mu             sync.RWMutex
	SchoolProfiles map[string]*SchoolProfile
	Children       map[string][]LinkedChild // key: schoolID:parentUserID
	Guardians      map[string]bool          // key: schoolID:studentID:parentUserID
	Attendance     map[string]*AttendanceSummary
	Marks          map[string]*StudentMarksSummary
	AuditLogs      []*AIAuditLogEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		SchoolProfiles: make(map[string]*SchoolProfile),
		Children:       make(map[string][]LinkedChild),
		Guardians:      make(map[string]bool),
		Attendance:     make(map[string]*AttendanceSummary),
		Marks:          make(map[string]*StudentMarksSummary),
	}
}

func (m *MemoryStore) GetSchoolProfile(ctx context.Context, schoolID string) (*SchoolProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.SchoolProfiles[schoolID]
	if !ok {
		return nil, errors.New("school not found")
	}
	return p, nil
}

func (m *MemoryStore) ListChildrenForParent(ctx context.Context, schoolID, parentUserID string) ([]LinkedChild, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", schoolID, parentUserID)
	c := m.Children[key]
	res := make([]LinkedChild, len(c))
	copy(res, c)
	return res, nil
}

func (m *MemoryStore) IsGuardianOf(ctx context.Context, schoolID, studentID, userID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s:%s", schoolID, studentID, userID)
	return m.Guardians[key], nil
}

func (m *MemoryStore) GetAttendanceSummary(ctx context.Context, schoolID, studentID, termID string) (*AttendanceSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s:%s", schoolID, studentID, termID)
	a, ok := m.Attendance[key]
	if !ok {
		return nil, errors.New("attendance summary not found")
	}
	return a, nil
}

func (m *MemoryStore) GetMarksSummary(ctx context.Context, schoolID, studentID, termID string) (*StudentMarksSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s:%s", schoolID, studentID, termID)
	marks, ok := m.Marks[key]
	if !ok {
		return nil, errors.New("marks summary not found")
	}
	return marks, nil
}

type tenantStore interface {
	WithTx(ctx context.Context, schoolID string, fn func(Store) error) error
}

func (h *Handler) withTenantStore(ctx context.Context, schoolID string, fn func(Store) error) error {
	if store, ok := h.store.(tenantStore); ok {
		return store.WithTx(ctx, schoolID, fn)
	}
	return fn(h.store)
}

func (m *MemoryStore) SaveAIAuditLog(ctx context.Context, entry *AIAuditLogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry.ID == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		entry.ID = fmt.Sprintf("ai_log_%x", b)
	}
	m.AuditLogs = append(m.AuditLogs, entry)
	return nil
}
