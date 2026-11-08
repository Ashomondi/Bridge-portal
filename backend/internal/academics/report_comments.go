package academics

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ReportCommentInput is one approved class-teacher comment.
type ReportCommentInput struct {
	StudentID   string `json:"student_id"`
	CommentText string `json:"comment_text"`
	AIDraftUsed bool   `json:"ai_draft_used"`
}

// PutReportCommentsRequest saves the class teacher's approved comments.
type PutReportCommentsRequest struct {
	TermID   string               `json:"term_id"`
	Comments []ReportCommentInput `json:"comments"`
}

// PutReportCommentsResult reports how many comments were saved.
type PutReportCommentsResult struct {
	ClassID    string `json:"class_id"`
	TermID     string `json:"term_id"`
	SavedCount int    `json:"saved_count"`
}

// SaveReportComments implements PUT /api/v1/classes/{id}/report-comments.
// Comments may be edited only while results are draft/reopened.
func (s *Service) SaveReportComments(ctx context.Context, classID string, req PutReportCommentsRequest) (PutReportCommentsResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return PutReportCommentsResult{}, err
	}
	classID = strings.TrimSpace(classID)
	req.TermID = strings.TrimSpace(req.TermID)
	if classID == "" || req.TermID == "" {
		return PutReportCommentsResult{}, errValidation("class_id and term_id are required.")
	}
	if len(req.Comments) == 0 {
		return PutReportCommentsResult{}, errValidation("At least one report comment is required.")
	}

	var out PutReportCommentsResult
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		allowed, err := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeClassTeacherScope, "Only the class teacher for this class may edit report comments.")
		}
		if _, err := tx.TermByID(ctx, actor.SchoolID, req.TermID); err != nil {
			return err
		}
		release, releaseErr := tx.ResultRelease(ctx, actor.SchoolID, classID, req.TermID)
		if releaseErr == nil && (release.Status == "released" || release.Status == "rereleased") {
			return errConflict(CodeResultsLocked, "Results are released; reopen them before editing report comments.")
		}

		students, err := tx.EnrolledStudents(ctx, actor.SchoolID, classID)
		if err != nil {
			return err
		}
		enrolled := map[string]bool{}
		for _, student := range students {
			enrolled[student.StudentID] = true
		}
		seen := map[string]bool{}
		comments := make([]ReportComment, 0, len(req.Comments))
		for _, in := range req.Comments {
			studentID := strings.TrimSpace(in.StudentID)
			text := strings.TrimSpace(in.CommentText)
			if studentID == "" || !enrolled[studentID] {
				return errValidation("A report comment references a student not enrolled in this class.")
			}
			if seen[studentID] {
				return errValidation("A student may only have one report comment per term.")
			}
			seen[studentID] = true
			if text == "" {
				return errValidation("Report comment text cannot be empty.")
			}
			comments = append(comments, ReportComment{
				StudentID:   studentID,
				ClassID:     classID,
				TermID:      req.TermID,
				CommentText: text,
				AIDraftUsed: in.AIDraftUsed,
				ApprovedBy:  actor.UserID,
			})
		}
		count, err := tx.UpsertReportComments(ctx, actor.SchoolID, comments)
		if err != nil {
			return err
		}
		out = PutReportCommentsResult{ClassID: classID, TermID: req.TermID, SavedCount: count}
		return s.audit(ctx, tx, actor, "report_comments_saved", "class", classID, map[string]any{
			"term_id":     req.TermID,
			"saved_count": count,
		})
	})
	return out, err
}

// DraftCommentRequest is the frontend request to POST /report-comments/draft.
type DraftCommentRequest struct {
	Mode          string `json:"mode"`
	StudentID     string `json:"student_id"`
	TermID        string `json:"term_id"`
	TeacherReason string `json:"teacher_reason"`
}

// DraftCommentResponse returns a draft for teacher approval.
type DraftCommentResponse struct {
	DraftComment            string `json:"draft_comment"`
	RequiresTeacherApproval bool   `json:"requires_teacher_approval"`
}

// DraftReportComment gathers permission-scoped facts and asks the AI service
// for a draft. The draft is never saved automatically (human approval rule).
func (s *Service) DraftReportComment(ctx context.Context, req DraftCommentRequest) (DraftCommentResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return DraftCommentResponse{}, err
	}
	if actor.Role == RoleParent {
		return DraftCommentResponse{}, errForbidden(CodeForbidden, "Parents cannot draft teacher report comments.")
	}
	req.StudentID = strings.TrimSpace(req.StudentID)
	req.TermID = strings.TrimSpace(req.TermID)
	req.Mode = strings.TrimSpace(req.Mode)
	req.TeacherReason = strings.TrimSpace(req.TeacherReason)
	if req.Mode == "" {
		req.Mode = "term_comment"
	}
	if req.Mode != "term_comment" && req.Mode != "rerelease_note" {
		return DraftCommentResponse{}, errValidation("mode must be term_comment or rerelease_note.")
	}
	if req.StudentID == "" || req.TermID == "" {
		return DraftCommentResponse{}, errValidation("student_id and term_id are required.")
	}
	if req.Mode == "rerelease_note" && req.TeacherReason == "" {
		return DraftCommentResponse{}, errValidation("teacher_reason is required when drafting a re-release note.")
	}

	var draftContext DraftContext
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		student, err := tx.StudentByID(ctx, actor.SchoolID, req.StudentID)
		if err != nil {
			return err
		}
		allowed, err := tx.TeacherCanViewStudent(ctx, actor.SchoolID, actor.MembershipID, req.StudentID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeForbidden, "You are not a teacher of this student.")
		}
		term, err := tx.TermByID(ctx, actor.SchoolID, req.TermID)
		if err != nil {
			return err
		}
		subjects, _ := tx.SubjectsForClass(ctx, actor.SchoolID, student.CurrentClassID)
		subjectNames := map[string]string{}
		for _, subject := range subjects {
			subjectNames[subject.ID] = subject.Name
		}
		scale, _ := tx.ActiveGradingScale(ctx, actor.SchoolID)

		currentMarks, _ := tx.MarksForStudentTerm(ctx, actor.SchoolID, req.StudentID, req.TermID)
		draftContext = DraftContext{
			Mode:             req.Mode,
			StudentID:        req.StudentID,
			TermID:           req.TermID,
			TeacherReason:    req.TeacherReason,
			Identity:         actor,
			CurrentTermMarks: make([]DraftCommentMark, 0, len(currentMarks)),
		}
		for _, mark := range currentMarks {
			level := ""
			if scale.ID != "" {
				level = resolveCBCLevel(scale, mark.Score)
			}
			draftContext.CurrentTermMarks = append(draftContext.CurrentTermMarks, DraftCommentMark{
				SubjectID:   mark.SubjectID,
				SubjectName: subjectNames[mark.SubjectID],
				Score:       mark.Score,
				CBCLevel:    level,
			})
		}

		// Prior terms use released snapshots where available so they retain the
		// grading scale version in force at the time.
		snapshots, _ := tx.SnapshotsForStudent(ctx, actor.SchoolID, req.StudentID)
		priorByTerm := map[string][]DraftCommentMark{}
		for _, snapshot := range snapshots {
			if snapshot.TermID == req.TermID {
				continue
			}
			priorByTerm[snapshot.TermID] = append(priorByTerm[snapshot.TermID], DraftCommentMark{
				SubjectID:   snapshot.SubjectID,
				SubjectName: subjectNames[snapshot.SubjectID],
				Score:       snapshot.Score,
				CBCLevel:    snapshot.CBCLevel,
				IsAmended:   snapshot.IsAmended,
			})
		}
		for termID, subjects := range priorByTerm {
			priorTerm, err := tx.TermByID(ctx, actor.SchoolID, termID)
			if err != nil {
				continue
			}
			draftContext.PriorTerms = append(draftContext.PriorTerms, DraftPriorTerm{
				TermID:     termID,
				TermNumber: priorTerm.TermNumber,
				Year:       priorTerm.Year,
				Subjects:   subjects,
			})
		}
		draftContext.Attendance, _ = tx.AttendanceForStudent(ctx, actor.SchoolID, req.StudentID)
		_ = term
		return nil
	})
	if err != nil {
		return DraftCommentResponse{}, err
	}
	if s.drafter == nil {
		return DraftCommentResponse{}, &APIError{
			Status:  http.StatusServiceUnavailable,
			Code:    CodeAIDrafterUnavailable,
			Message: "The AI report-comment drafter is not configured.",
		}
	}
	result, err := s.drafter.DraftReportComment(ctx, draftContext)
	if err != nil {
		return DraftCommentResponse{}, &APIError{
			Status:  http.StatusBadGateway,
			Code:    "AI_DRAFTER_FAILED",
			Message: "The AI report-comment drafter could not produce a draft.",
			Details: map[string]any{"cause": err.Error()},
		}
	}
	return DraftCommentResponse{
		DraftComment:            result.DraftComment,
		RequiresTeacherApproval: true,
	}, nil
}

// HTTPCommentDrafter calls Sos's POST /api/v1/ai/report-comments/draft with the
// service token and HMAC-signed X-User-Context.
type HTTPCommentDrafter struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewHTTPCommentDrafter builds a drafter with a sensible timeout.
func NewHTTPCommentDrafter(baseURL, token string) *HTTPCommentDrafter {
	return &HTTPCommentDrafter{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		Client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// DraftReportComment implements CommentDrafter.
func (d *HTTPCommentDrafter) DraftReportComment(ctx context.Context, req DraftContext) (DraftResult, error) {
	if d.Token == "" {
		return DraftResult{}, fmt.Errorf("academics: INTERNAL_SERVICE_TOKEN is not configured")
	}
	signed, err := signUserContext(req.Identity, d.Token, 5*time.Minute)
	if err != nil {
		return DraftResult{}, err
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return DraftResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.BaseURL+"/api/v1/ai/report-comments/draft", bytes.NewReader(payload))
	if err != nil {
		return DraftResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Service-Token", d.Token)
	httpReq.Header.Set("X-User-Context", signed)
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return DraftResult{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DraftResult{}, fmt.Errorf("academics: AI drafter returned %d: %s", resp.StatusCode, string(body))
	}
	var envelope struct {
		Status string      `json:"status"`
		Data   DraftResult `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Data.DraftComment != "" {
		envelope.Data.RequiresTeacherApproval = true
		return envelope.Data, nil
	}
	var direct DraftResult
	if err := json.Unmarshal(body, &direct); err != nil {
		return DraftResult{}, fmt.Errorf("academics: cannot parse AI drafter response: %w", err)
	}
	if strings.TrimSpace(direct.DraftComment) == "" {
		return DraftResult{}, fmt.Errorf("academics: AI drafter returned an empty draft")
	}
	direct.RequiresTeacherApproval = true
	return direct, nil
}

// signUserContext builds the shared base64url(json).hex(hmac) contract.
func signUserContext(identity Identity, secret string, ttl time.Duration) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"user_id":       identity.UserID,
		"school_id":     identity.SchoolID,
		"membership_id": identity.MembershipID,
		"role":          identity.Role,
		"channel":       "in_app",
		"session_id":    "",
		"exp":           time.Now().Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *Service) handlePutReportComments(w http.ResponseWriter, r *http.Request) {
	var req PutReportCommentsRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.SaveReportComments(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleDraftReportComment(w http.ResponseWriter, r *http.Request) {
	var req DraftCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.DraftReportComment(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodPut, "/api/v1/classes/{id}/report-comments", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handlePutReportComments(w, r)
	})
	registerRoute(http.MethodPost, "/api/v1/report-comments/draft", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleDraftReportComment(w, r)
	})
}
