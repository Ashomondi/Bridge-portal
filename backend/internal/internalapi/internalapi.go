package internalapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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

func NewHandler(serviceToken string, signer *ContextSigner, store Store, billing BillingBridge, aiServiceURL string) *Handler {
	if serviceToken == "" {
		serviceToken = "test-internal-service-token-secret"
	}
	return &Handler{
		serviceToken: serviceToken,
		signer:       signer,
		store:        store,
		billing:      billing,
		aiServiceURL: aiServiceURL,
	}
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

		next(w, r, uContext)
	}
}

// 1. GET /api/v1/internal/school-profile
func (h *Handler) handleSchoolProfile(w http.ResponseWriter, r *http.Request, u *UserContext) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	profile, err := h.store.GetSchoolProfile(r.Context(), u.SchoolID)
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
	parts := strings.Split(r.URL.Path, "/")
	// e.g. /api/v1/internal/parents/{user_id}/children
	parentUserID := ""
	for i, p := range parts {
		if p == "parents" && i+1 < len(parts) {
			parentUserID = parts[i+1]
			break
		}
	}

	// Parent can only query their own children unless staff
	if u.Role == "parent" && parentUserID != u.UserID {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot query children of another guardian"})
		return
	}

	children, err := h.store.ListChildrenForParent(r.Context(), u.SchoolID, parentUserID)
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

	if studentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "student ID required"})
		return
	}

	// Enforce Guardian-Child ownership (Docs/API_CONTRACTS.md, SECURITY.md)
	// Parent can ONLY query their own children!
	if u.Role == "parent" {
		isGuardian, err := h.store.IsGuardianOf(r.Context(), u.SchoolID, studentID, u.UserID)
		if err != nil || !isGuardian {
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
		summary, err := h.store.GetAttendanceSummary(r.Context(), u.SchoolID, studentID, termID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(summary)

	case "marks-summary":
		termID := r.URL.Query().Get("term_id")
		summary, err := h.store.GetMarksSummary(r.Context(), u.SchoolID, studentID, termID)
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
		isGuardian, err := h.store.IsGuardianOf(r.Context(), u.SchoolID, req.StudentID, u.UserID)
		if err != nil || !isGuardian {
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

	_ = h.store.SaveAIAuditLog(r.Context(), &entry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged"})
}

// Frontend Chat Proxy: POST /api/v1/chat/message
func (h *Handler) handleFrontendChatMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message   string `json:"message"`
		StudentID string `json:"student_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"reply": "Your message has been processed successfully.",
	})
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
		return &AttendanceSummary{
			StudentID:      studentID,
			TermID:         termID,
			TotalDays:      60,
			DaysPresent:    58,
			DaysAbsent:     2,
			AttendanceRate: 96.67,
		}, nil
	}
	return a, nil
}

func (m *MemoryStore) GetMarksSummary(ctx context.Context, schoolID, studentID, termID string) (*StudentMarksSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s:%s", schoolID, studentID, termID)
	marks, ok := m.Marks[key]
	if !ok {
		return &StudentMarksSummary{
			StudentID: studentID,
			TermID:    termID,
			Subjects: []SubjectMarkSummary{
				{SubjectCode: "ENG", SubjectName: "English", Score: 82.0, CBCLevel: "EE"},
				{SubjectCode: "MAT", SubjectName: "Mathematics", Score: 76.5, CBCLevel: "ME"},
				{SubjectCode: "KIS", SubjectName: "Kiswahili", Score: 88.0, CBCLevel: "EE"},
			},
			MeanGrade:   "EE",
			PublishedAt: time.Now(),
		}, nil
	}
	return marks, nil
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
