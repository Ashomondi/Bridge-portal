package internalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/whatsapp"
	"github.com/jackc/pgx/v5"
)

type AttendanceSummary struct {
	StudentID       string `json:"student_id"`
	TermID          string `json:"term_id,omitempty"`
	PresentDays     int    `json:"present_days"`
	AbsentDays      int    `json:"absent_days"`
	LateDays        int    `json:"late_days"`
	TotalRecorded   int    `json:"total_recorded"`
	LastRollCallDay string `json:"last_roll_call_day,omitempty"`
}

type SubjectMarkSummary struct {
	Subject          string `json:"subject"`
	PerformanceLevel string `json:"performance_level"`
	TeacherComments  string `json:"teacher_comments,omitempty"`
}

type MarksSummary struct {
	StudentID string               `json:"student_id"`
	TermID    string               `json:"term_id,omitempty"`
	Released  bool                 `json:"released"`
	PortalURL string               `json:"portal_url,omitempty"`
	Subjects  []SubjectMarkSummary `json:"subjects"`
}

// AcademicsReader is the consumer-side interface for Ema's academics service.
type AcademicsReader interface {
	GetStudentAttendanceSummary(ctx context.Context, schoolID, studentID string) (AttendanceSummary, error)
	GetStudentMarksSummary(ctx context.Context, schoolID, studentID string) (MarksSummary, error)
}

type SchoolProfile struct {
	SchoolID     string  `json:"school_id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	County       *string `json:"county,omitempty"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty"`
	OfficeHours  *string `json:"office_hours,omitempty"`
}

type ParentChild struct {
	StudentID       string `json:"student_id"`
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	GradeName       string `json:"grade_name,omitempty"`
	StreamName      string `json:"stream_name,omitempty"`
	Relationship    string `json:"relationship"`
}

type AIAuditLogInput struct {
	SessionID     string         `json:"session_id"`
	Feature       string         `json:"feature"`
	Intent        string         `json:"intent"`
	EndpointsRead []string       `json:"endpoints_read"`
	ActionTaken   string         `json:"action_taken"`
	Metadata      map[string]any `json:"metadata"`
}

type ChatMessageInput struct {
	StudentID string `json:"student_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type ChatMessageResponse struct {
	Reply     string `json:"reply"`
	Channel   string `json:"channel"`
	SessionID string `json:"session_id,omitempty"`
}

// Service implements the Internal AI API gateway and outbound AI service client.
type Service struct {
	DB                   contracts.TenantDB
	Billing              *billing.Service
	Academics            AcademicsReader
	InternalServiceToken string
	AIServiceBaseURL     string
	PortalBaseURL        string
	HTTPClient           *http.Client
	ContextTTL           time.Duration
	Now                  func() time.Time
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) contextTTL() time.Duration {
	if s != nil && s.ContextTTL > 0 {
		return s.ContextTTL
	}
	return 5 * time.Minute
}

func (s *Service) httpClient() *http.Client {
	base := http.DefaultClient
	if s != nil && s.HTTPClient != nil {
		base = s.HTTPClient
	}
	copied := *base
	if copied.Timeout <= 0 {
		copied.Timeout = 15 * time.Second
	}
	copied.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &copied
}

func (s *Service) aiBaseURL() string {
	if s != nil && strings.TrimSpace(s.AIServiceBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(s.AIServiceBaseURL), "/")
	}
	return "http://ai_service:8000/api/v1/ai"
}

func (s *Service) portalBaseURL() string {
	if s != nil && strings.TrimSpace(s.PortalBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(s.PortalBaseURL), "/")
	}
	return "https://portal.ikonex.ac.ke"
}

// AuthenticateInternalRequest validates X-Internal-Service-Token and HMAC-signed X-User-Context,
// and verifies active membership (and active WhatsApp session when channel=='whatsapp') inside the tenant DB.
func (s *Service) AuthenticateInternalRequest(r *http.Request) (UserContext, contracts.Principal, error) {
	if s == nil || s.DB == nil {
		return UserContext{}, contracts.Principal{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if err := VerifyServiceToken(r.Header.Get(HeaderInternalServiceToken), s.InternalServiceToken); err != nil {
		return UserContext{}, contracts.Principal{}, err
	}
	now := s.now()
	uctx, err := VerifyUserContext(r.Header.Get(HeaderUserContext), s.InternalServiceToken, now)
	if err != nil {
		return UserContext{}, contracts.Principal{}, err
	}

	err = s.DB.WithTenantTx(r.Context(), uctx.SchoolID, func(tx pgx.Tx) error {
		var activeMembership bool
		if err := tx.QueryRow(r.Context(), `
			SELECT EXISTS(
				SELECT 1 FROM memberships m
				JOIN users u ON u.id = m.user_id
				WHERE m.school_id = $1 AND m.id = $2 AND m.user_id = $3 AND m.status = 'active' AND u.is_active
			)`,
			uctx.SchoolID, uctx.MembershipID, uctx.UserID,
		).Scan(&activeMembership); err != nil {
			return err
		}
		if !activeMembership {
			return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "User does not have an active membership in this school.")
		}

		if uctx.Channel == "whatsapp" {
			var sessionValid bool
			if err := tx.QueryRow(r.Context(), `
				SELECT EXISTS(
					SELECT 1 FROM chat_sessions
					WHERE school_id = $1
					  AND id = $2
					  AND guardian_user_id = $3
					  AND state = 'VERIFIED'
					  AND expires_at > $4
				)`,
				uctx.SchoolID, uctx.SessionID, uctx.UserID, now,
			).Scan(&sessionValid); err != nil {
				return err
			}
			if !sessionValid {
				return contracts.Fail(http.StatusUnauthorized, "SESSION_EXPIRED", "WhatsApp session is not verified or has expired.")
			}
		}
		return nil
	})
	if err != nil {
		return UserContext{}, contracts.Principal{}, err
	}

	p := contracts.Principal{
		UserID:       uctx.UserID,
		SchoolID:     uctx.SchoolID,
		MembershipID: uctx.MembershipID,
		Role:         uctx.Role,
	}
	return uctx, p, nil
}

func (s *Service) requireGuardianOwnsStudentTx(ctx context.Context, tx pgx.Tx, schoolID, guardianUserID, studentID string) error {
	if !contracts.ValidID(studentID) {
		return contracts.Fail(http.StatusBadRequest, "INVALID_STUDENT_ID", "A valid student ID is required.")
	}
	var studentExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM students WHERE school_id = $1 AND id = $2)`, schoolID, studentID).Scan(&studentExists); err != nil {
		return err
	}
	if !studentExists {
		return contracts.Fail(http.StatusNotFound, "STUDENT_NOT_FOUND", "Student not found.")
	}
	var linked bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM student_guardians
			WHERE school_id = $1 AND student_id = $2 AND user_id = $3
		)`,
		schoolID, studentID, guardianUserID,
	).Scan(&linked); err != nil {
		return err
	}
	if !linked {
		return contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Parent is not linked to the requested student.")
	}
	return nil
}

func (s *Service) GetSchoolProfile(ctx context.Context, uctx UserContext) (SchoolProfile, error) {
	var profile SchoolProfile
	err := s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT id, code, name, county
			FROM schools
			WHERE id = $1`,
			uctx.SchoolID,
		).Scan(&profile.SchoolID, &profile.Code, &profile.Name, &profile.County); errors.Is(err, pgx.ErrNoRows) {
			return contracts.Fail(http.StatusNotFound, "SCHOOL_NOT_FOUND", "School not found.")
		} else if err != nil {
			return err
		}

		// Optional school_profile table from migration 000002
		var hasProfileTable bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'school_profile'
			)`).Scan(&hasProfileTable); err == nil && hasProfileTable {
			_ = tx.QueryRow(ctx, `
				SELECT contact_phone, contact_email, office_hours
				FROM school_profile
				WHERE school_id = $1`,
				uctx.SchoolID,
			).Scan(&profile.ContactPhone, &profile.ContactEmail, &profile.OfficeHours)
		}
		return nil
	})
	return profile, err
}

func (s *Service) GetParentChildren(ctx context.Context, uctx UserContext, pathUserID string) ([]ParentChild, error) {
	if !contracts.ValidID(pathUserID) {
		return nil, contracts.Fail(http.StatusBadRequest, "INVALID_USER_ID", "A valid user_id is required.")
	}
	if pathUserID != uctx.UserID {
		return nil, contracts.Fail(http.StatusForbidden, "FORBIDDEN", "Parents may only query their own linked children.")
	}
	children := []ParentChild{}
	err := s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT s.id, s.admission_number, s.full_name,
			       COALESCE(g.name, ''), COALESCE(c.stream_name, ''), sg.relationship
			FROM student_guardians sg
			JOIN students s ON s.school_id = sg.school_id AND s.id = sg.student_id
			LEFT JOIN classes c ON c.school_id = s.school_id AND c.id = s.current_class_id
			LEFT JOIN grade_levels g ON g.school_id = c.school_id AND g.id = c.grade_level_id
			WHERE sg.school_id = $1 AND sg.user_id = $2 AND s.status = 'active'
			ORDER BY s.full_name ASC, s.id ASC`,
			uctx.SchoolID, uctx.UserID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var child ParentChild
			if err := rows.Scan(&child.StudentID, &child.AdmissionNumber, &child.FullName, &child.GradeName, &child.StreamName, &child.Relationship); err != nil {
				return err
			}
			children = append(children, child)
		}
		return rows.Err()
	})
	return children, err
}

func (s *Service) GetInternalStudentFeeBalance(ctx context.Context, uctx UserContext, p contracts.Principal, studentID string) (billing.StudentFeeBalance, error) {
	if s.Billing == nil {
		return billing.StudentFeeBalance{}, contracts.Fail(http.StatusServiceUnavailable, "BILLING_UNAVAILABLE", "Billing service is not configured.")
	}
	var out billing.StudentFeeBalance
	err := s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		if err := s.requireGuardianOwnsStudentTx(ctx, tx, uctx.SchoolID, uctx.UserID, studentID); err != nil {
			return err
		}
		var err error
		out, err = s.Billing.GetStudentFeeBalanceTx(ctx, tx, p, studentID)
		return err
	})
	return out, err
}

func (s *Service) GetInternalAttendanceSummary(ctx context.Context, uctx UserContext, studentID string) (AttendanceSummary, error) {
	err := s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		return s.requireGuardianOwnsStudentTx(ctx, tx, uctx.SchoolID, uctx.UserID, studentID)
	})
	if err != nil {
		return AttendanceSummary{}, err
	}
	if s.Academics == nil {
		return AttendanceSummary{}, contracts.Fail(http.StatusServiceUnavailable, "ACADEMICS_UNAVAILABLE", "Academics service is not configured.")
	}
	return s.Academics.GetStudentAttendanceSummary(ctx, uctx.SchoolID, studentID)
}

func (s *Service) GetInternalMarksSummary(ctx context.Context, uctx UserContext, studentID string) (MarksSummary, error) {
	err := s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		return s.requireGuardianOwnsStudentTx(ctx, tx, uctx.SchoolID, uctx.UserID, studentID)
	})
	if err != nil {
		return MarksSummary{}, err
	}
	// Non-negotiable Rule 9: WhatsApp carries minimal facts only (never results/report cards).
	if uctx.Channel == "whatsapp" {
		return MarksSummary{
			StudentID: studentID,
			Released:  false,
			PortalURL: s.portalBaseURL() + "/parent/results",
			Subjects:  []SubjectMarkSummary{},
		}, nil
	}
	if s.Academics == nil {
		return MarksSummary{}, contracts.Fail(http.StatusServiceUnavailable, "ACADEMICS_UNAVAILABLE", "Academics service is not configured.")
	}
	return s.Academics.GetStudentMarksSummary(ctx, uctx.SchoolID, studentID)
}

func (s *Service) InitiateInternalSTKPush(ctx context.Context, uctx UserContext, p contracts.Principal, in billing.InitiateSTKPushInput) (billing.STKPushInitiated, error) {
	if s.Billing == nil {
		return billing.STKPushInitiated{}, contracts.Fail(http.StatusServiceUnavailable, "BILLING_UNAVAILABLE", "Billing service is not configured.")
	}
	in.Source = "chat"
	return s.Billing.InitiateSTKPush(ctx, p, in, true)
}

func sanitizeMetadata(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		lower := strings.ToLower(strings.TrimSpace(k))
		if strings.Contains(lower, "secret") ||
			strings.Contains(lower, "token") ||
			strings.Contains(lower, "password") ||
			strings.Contains(lower, "passkey") ||
			strings.Contains(lower, "otp") ||
			strings.Contains(lower, "authorization") ||
			strings.Contains(lower, "x-internal") ||
			strings.Contains(lower, "x-user-context") {
			continue
		}
		out[k] = v
	}
	return out
}

func (s *Service) RecordAIAuditLog(ctx context.Context, uctx UserContext, in AIAuditLogInput) (string, error) {
	feature := strings.TrimSpace(in.Feature)
	if feature == "" {
		return "", contracts.Fail(http.StatusUnprocessableEntity, "INVALID_AI_AUDIT_LOG", "feature is required.")
	}
	endpoints := in.EndpointsRead
	if endpoints == nil {
		endpoints = []string{}
	}
	endpointsJSON, err := json.Marshal(endpoints)
	if err != nil {
		return "", contracts.Fail(http.StatusUnprocessableEntity, "INVALID_AI_AUDIT_LOG", "endpoints_read must be valid JSON.")
	}
	safeMeta := sanitizeMetadata(in.Metadata)
	metaJSON, err := json.Marshal(safeMeta)
	if err != nil {
		return "", contracts.Fail(http.StatusUnprocessableEntity, "INVALID_AI_AUDIT_LOG", "metadata must be valid JSON.")
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		sessionID = uctx.SessionID
	}
	if sessionID != "" && !contracts.ValidID(sessionID) {
		return "", contracts.Fail(http.StatusUnprocessableEntity, "INVALID_AI_AUDIT_LOG", "session_id must be a valid UUID.")
	}

	var logID string
	err = s.DB.WithTenantTx(ctx, uctx.SchoolID, func(tx pgx.Tx) error {
		var validSessionID *string
		if sessionID != "" {
			var sessionExists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM chat_sessions WHERE school_id = $1 AND id = $2)`, uctx.SchoolID, sessionID).Scan(&sessionExists); err != nil {
				return err
			}
			if sessionExists {
				validSessionID = &sessionID
			}
		}
		return tx.QueryRow(ctx, `
			INSERT INTO ai_action_logs (
				school_id, user_id, session_id, channel, feature, intent, endpoints_read, action_taken, metadata
			)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, NULLIF($8, ''), $9)
			RETURNING id`,
			uctx.SchoolID, uctx.UserID, validSessionID, uctx.Channel,
			feature, strings.TrimSpace(in.Intent), endpointsJSON, strings.TrimSpace(in.ActionTaken), metaJSON,
		).Scan(&logID)
	})
	return logID, err
}

// ForwardParentQuery implements whatsapp.AIForwarder and also powers the in-app chat proxy.
func (s *Service) ForwardParentQuery(ctx context.Context, input whatsapp.AIParentQueryInput) (string, error) {
	now := s.now()
	exp := now.Add(s.contextTTL())
	if !input.ExpiresAt.IsZero() && input.ExpiresAt.Before(exp) {
		exp = input.ExpiresAt
	}
	signedCtx, err := SignUserContext(UserContext{
		UserID:       input.UserID,
		SchoolID:     input.SchoolID,
		MembershipID: input.MembershipID,
		Role:         input.Role,
		Channel:      input.Channel,
		SessionID:    input.SessionID,
		Exp:          exp.UTC().Unix(),
	}, s.InternalServiceToken)
	if err != nil {
		return "", err
	}

	bodyPayload := map[string]any{
		"message":    input.Message,
		"student_id": input.StudentID,
		"channel":    input.Channel,
		"session_id": input.SessionID,
	}
	encoded, err := json.Marshal(bodyPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.aiBaseURL()+"/parent-query", bytes.NewReader(encoded))
	if err != nil {
		return "", contracts.Fail(http.StatusBadGateway, "AI_SERVICE_UNAVAILABLE", "Could not build AI assistant request.")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderInternalServiceToken, s.InternalServiceToken)
	req.Header.Set(HeaderUserContext, signedCtx)

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return "", contracts.Fail(http.StatusBadGateway, "AI_SERVICE_UNAVAILABLE", "The AI assistant is temporarily unavailable.")
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", contracts.Fail(http.StatusBadGateway, "AI_SERVICE_ERROR", "The AI assistant could not process your message.")
	}

	var envelope struct {
		Reply string `json:"reply"`
		Data  struct {
			Reply string `json:"reply"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &envelope); err != nil {
		return "", contracts.Fail(http.StatusBadGateway, "AI_SERVICE_ERROR", "Invalid AI assistant response.")
	}
	reply := strings.TrimSpace(envelope.Reply)
	if reply == "" {
		reply = strings.TrimSpace(envelope.Data.Reply)
	}
	if reply == "" {
		return "", contracts.Fail(http.StatusBadGateway, "AI_SERVICE_ERROR", "Empty AI assistant response.")
	}
	return reply, nil
}

// HandleChatProxy serves POST /api/v1/chat/message for authenticated in-app parents.
func (s *Service) HandleChatProxy(ctx context.Context, p contracts.Principal, in ChatMessageInput) (ChatMessageResponse, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) {
		return ChatMessageResponse{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		return ChatMessageResponse{}, contracts.Fail(http.StatusUnprocessableEntity, "EMPTY_CHAT_MESSAGE", "A non-empty message is required.")
	}
	if in.StudentID != "" {
		if err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
			return s.requireGuardianOwnsStudentTx(ctx, tx, p.SchoolID, p.UserID, in.StudentID)
		}); err != nil {
			return ChatMessageResponse{}, err
		}
	}

	reply, err := s.ForwardParentQuery(ctx, whatsapp.AIParentQueryInput{
		UserID:       p.UserID,
		SchoolID:     p.SchoolID,
		MembershipID: p.MembershipID,
		Role:         p.Role,
		Channel:      "in_app",
		SessionID:    in.SessionID,
		StudentID:    in.StudentID,
		ExpiresAt:    s.now().Add(s.contextTTL()),
		Message:      msg,
	})
	if err != nil {
		return ChatMessageResponse{}, err
	}
	return ChatMessageResponse{
		Reply:     reply,
		Channel:   "in_app",
		SessionID: in.SessionID,
	}, nil
}

func RegisterRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	if service == nil {
		service = &Service{}
	}
	mux.HandleFunc("GET /api/v1/internal/school-profile", func(w http.ResponseWriter, r *http.Request) {
		uctx, _, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		profile, err := service.GetSchoolProfile(r.Context(), uctx)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, profile, nil)
	})

	mux.HandleFunc("GET /api/v1/internal/parents/{user_id}/children", func(w http.ResponseWriter, r *http.Request) {
		uctx, _, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		children, err := service.GetParentChildren(r.Context(), uctx, r.PathValue("user_id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, children, nil)
	})

	mux.HandleFunc("GET /api/v1/internal/students/{id}/fee-balance", func(w http.ResponseWriter, r *http.Request) {
		uctx, p, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		balance, err := service.GetInternalStudentFeeBalance(r.Context(), uctx, p, r.PathValue("id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, balance, nil)
	})

	mux.HandleFunc("GET /api/v1/internal/students/{id}/attendance-summary", func(w http.ResponseWriter, r *http.Request) {
		uctx, _, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		summary, err := service.GetInternalAttendanceSummary(r.Context(), uctx, r.PathValue("id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, summary, nil)
	})

	mux.HandleFunc("GET /api/v1/internal/students/{id}/marks-summary", func(w http.ResponseWriter, r *http.Request) {
		uctx, _, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		summary, err := service.GetInternalMarksSummary(r.Context(), uctx, r.PathValue("id"))
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, summary, nil)
	})

	mux.HandleFunc("POST /api/v1/internal/payments/stk-push", func(w http.ResponseWriter, r *http.Request) {
		uctx, p, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input billing.InitiateSTKPushInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		res, err := service.InitiateInternalSTKPush(r.Context(), uctx, p, input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusAccepted, res, nil)
	})

	mux.HandleFunc("POST /api/v1/internal/ai/audit-log", func(w http.ResponseWriter, r *http.Request) {
		uctx, _, err := service.AuthenticateInternalRequest(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input AIAuditLogInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		logID, err := service.RecordAIAuditLog(r.Context(), uctx, input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusCreated, map[string]any{"id": logID}, nil)
	})

	mux.HandleFunc("POST /api/v1/chat/message", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		var input ChatMessageInput
		if err := contracts.Decode(w, r, &input); err != nil {
			contracts.WriteError(w, err)
			return
		}
		res, err := service.HandleChatProxy(r.Context(), p, input)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, http.StatusOK, res, nil)
	})
}

var _ whatsapp.AIForwarder = (*Service)(nil)
