package academics

import "context"

// DraftCommentMark is one subject observation sent to the AI drafter.
type DraftCommentMark struct {
	SubjectID   string  `json:"subject_id"`
	SubjectName string  `json:"subject_name"`
	Score       float64 `json:"score"`
	CBCLevel    string  `json:"cbc_level"`
	IsAmended   bool    `json:"is_amended,omitempty"`
}

// DraftPriorTerm is a prior term trend for the drafter.
type DraftPriorTerm struct {
	TermID     string             `json:"term_id"`
	TermNumber int                `json:"term_number"`
	Year       int                `json:"year"`
	Subjects   []DraftCommentMark `json:"subjects"`
}

// DraftContext is the fact pack sent to Sos's AI service. The AI only phrases
// these facts; it never invents a mark (Docs/API_CONTRACTS.md).
type DraftContext struct {
	Mode             string             `json:"mode"`
	StudentID        string             `json:"student_id"`
	TermID           string             `json:"term_id"`
	TeacherReason    string             `json:"teacher_reason,omitempty"`
	Identity         Identity           `json:"identity"`
	CurrentTermMarks []DraftCommentMark `json:"current_term_marks"`
	PriorTerms       []DraftPriorTerm   `json:"prior_terms"`
	Attendance       AttendanceSummary  `json:"attendance"`
}

// DraftResult is the AI service's response.
type DraftResult struct {
	DraftComment            string `json:"draft_comment"`
	RequiresTeacherApproval bool   `json:"requires_teacher_approval"`
}

// CommentDrafter is implemented by the HTTP proxy to Sos's FastAPI service.
// It lives in the core contract file so the Dean/attendance feature commits can
// compile before the AI proxy file is added.
type CommentDrafter interface {
	DraftReportComment(ctx context.Context, req DraftContext) (DraftResult, error)
}
