package academics

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned by a Store when the requested record does not exist
// inside the current tenant. It is deliberately distinct from a permission
// error so the service layer can decide whether a 404 or 403 is appropriate.
var ErrNotFound = errors.New("academics: not found")

// Class is a stream within a grade level (for example "Grade 7 East").
type Class struct {
	ID           string `json:"id"`
	SchoolID     string `json:"school_id"`
	GradeLevelID string `json:"grade_level_id"`
	StreamName   string `json:"stream_name"`
}

// Term is one of the three CBC school terms in an academic year.
type Term struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	AcademicYearID string    `json:"academic_year_id"`
	Year           int       `json:"year"`
	TermNumber     int       `json:"term_number"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	IsCurrent      bool      `json:"is_current"`
}

// Subject is a CBC learning area for one grade level.
type Subject struct {
	ID           string `json:"id"`
	SchoolID     string `json:"school_id"`
	GradeLevelID string `json:"grade_level_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	IsCompulsory bool   `json:"is_compulsory"`
}

// EnrolledStudent is the minimal roster row exposed to an authorised class
// teacher. It never includes marks from another school.
type EnrolledStudent struct {
	StudentID       string `json:"student_id"`
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	ClassID         string `json:"class_id"`
}

// Student is the tenant-scoped student record.
type Student struct {
	ID              string `json:"id"`
	SchoolID        string `json:"school_id"`
	AdmissionNumber string `json:"admission_number"`
	FullName        string `json:"full_name"`
	CurrentClassID  string `json:"current_class_id"`
}

// Membership is a person's role at one school.
type Membership struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// SubjectTeacherAssignment links one teacher membership to a class/subject
// pair for one academic year (C59).
type SubjectTeacherAssignment struct {
	ID                  string    `json:"id"`
	SchoolID            string    `json:"school_id"`
	AcademicYearID      string    `json:"academic_year_id"`
	ClassID             string    `json:"class_id"`
	SubjectID           string    `json:"subject_id"`
	TeacherMembershipID string    `json:"teacher_membership_id"`
	AssignedBy          string    `json:"assigned_by"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GradingScale is one immutable version of a school's CBC score bands.
// Versions are never edited in place (C43).
type GradingScale struct {
	ID            string    `json:"id"`
	SchoolID      string    `json:"school_id"`
	VersionNumber int       `json:"version_number"`
	EEMin         float64   `json:"ee_min"`
	MEMin         float64   `json:"me_min"`
	AEMin         float64   `json:"ae_min"`
	BEMin         float64   `json:"be_min"`
	IsActive      bool      `json:"is_active"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// TimetableSlot is one weekly lesson slot for a class.
type TimetableSlot struct {
	ID                  string `json:"id"`
	SchoolID            string `json:"school_id"`
	AcademicYearID      string `json:"academic_year_id"`
	TermID              string `json:"term_id"`
	ClassID             string `json:"class_id"`
	SubjectID           string `json:"subject_id"`
	SubjectName         string `json:"subject_name,omitempty"`
	TeacherMembershipID string `json:"teacher_membership_id"`
	DayOfWeek           int    `json:"day_of_week"`
	StartTime           string `json:"start_time"`
	EndTime             string `json:"end_time"`
	Room                string `json:"room,omitempty"`
}

// CalendarEvent is a school event shown to staff, parents and learners.
type CalendarEvent struct {
	ID             string    `json:"id"`
	SchoolID       string    `json:"school_id"`
	AcademicYearID string    `json:"academic_year_id"`
	TermID         string    `json:"term_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	EventType      string    `json:"event_type"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

// AttendanceRow is one student's daily roll-call record. The source column is
// present so Release 2 biometric attendance writes to the same table (C35).
type AttendanceRow struct {
	SchoolID       string    `json:"school_id"`
	StudentID      string    `json:"student_id"`
	ClassID        string    `json:"class_id"`
	TermID         string    `json:"term_id"`
	AttendanceDate time.Time `json:"attendance_date"`
	Status         string    `json:"status"`
	Source         string    `json:"source"`
	RecordedBy     string    `json:"recorded_by"`
}

// AttendanceSummary is a non-clinical attendance aggregate for one learner.
type AttendanceSummary struct {
	TotalDays int `json:"total_days"`
	Present   int `json:"present"`
	Absent    int `json:"absent"`
	Late      int `json:"late"`
}

// Mark is the raw numeric score for one student, subject and term.
type Mark struct {
	SchoolID  string    `json:"school_id"`
	StudentID string    `json:"student_id"`
	ClassID   string    `json:"class_id"`
	TermID    string    `json:"term_id"`
	SubjectID string    `json:"subject_id"`
	Score     float64   `json:"score"`
	EnteredBy string    `json:"entered_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResultRelease is the per-class, per-term release state machine:
// draft -> released -> reopened -> rereleased (C31, C33, C34).
type ResultRelease struct {
	ID                    string     `json:"id"`
	SchoolID              string     `json:"school_id"`
	ClassID               string     `json:"class_id"`
	TermID                string     `json:"term_id"`
	Status                string     `json:"status"`
	GradingScaleVersionID string     `json:"grading_scale_version_id,omitempty"`
	ReleasedBy            string     `json:"released_by,omitempty"`
	ReleasedAt            *time.Time `json:"released_at,omitempty"`
	ReopenedBy            string     `json:"reopened_by,omitempty"`
	ReopenedAt            *time.Time `json:"reopened_at,omitempty"`
	RereleasedBy          string     `json:"rereleased_by,omitempty"`
	RereleasedAt          *time.Time `json:"rereleased_at,omitempty"`
}

// Snapshot is an immutable parent-facing result row. The report_comment is
// captured in the same snapshot so a reopen cannot change what a parent sees.
type Snapshot struct {
	ID              string    `json:"id"`
	SchoolID        string    `json:"school_id"`
	ResultReleaseID string    `json:"result_release_id"`
	StudentID       string    `json:"student_id"`
	TermID          string    `json:"term_id"`
	SubjectID       string    `json:"subject_id"`
	Score           float64   `json:"score"`
	CBCLevel        string    `json:"cbc_level"`
	IsAmended       bool      `json:"is_amended"`
	AmendmentNote   string    `json:"amendment_note,omitempty"`
	SnapshotVersion int       `json:"snapshot_version"`
	ReportComment   string    `json:"report_comment,omitempty"`
	PublishedAt     time.Time `json:"published_at"`
}

// ReportComment is a class teacher's approved comment for a learner/term.
type ReportComment struct {
	ID          string    `json:"id"`
	SchoolID    string    `json:"school_id"`
	StudentID   string    `json:"student_id"`
	ClassID     string    `json:"class_id"`
	TermID      string    `json:"term_id"`
	CommentText string    `json:"comment_text"`
	AIDraftUsed bool      `json:"ai_draft_used"`
	ApprovedBy  string    `json:"approved_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuditEvent is written for every sensitive academic action (AC-29).
type AuditEvent struct {
	SchoolID     string         `json:"school_id"`
	UserID       string         `json:"user_id"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Details      map[string]any `json:"details,omitempty"`
}

// Store is the persistence boundary for the academics package. It is defined
// by the consumer so the service can be unit-tested with an in-memory store and
// wired to PostgreSQL in production without either side knowing the other's
// concrete types.
type Store interface {
	// WithTx runs fn inside a tenant-scoped transaction where
	// app.school_id has been set. Every academic query goes through it.
	WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error

	CurrentAcademicYearID(ctx context.Context, schoolID string) (string, error)
	AcademicYearExists(ctx context.Context, schoolID, academicYearID string) (bool, error)
	ClassByID(ctx context.Context, schoolID, classID string) (Class, error)
	TermByID(ctx context.Context, schoolID, termID string) (Term, error)
	TermForDate(ctx context.Context, schoolID string, date time.Time) (Term, error)
	ListTerms(ctx context.Context, schoolID string) ([]Term, error)
	StudentByID(ctx context.Context, schoolID, studentID string) (Student, error)
	SubjectsForClass(ctx context.Context, schoolID, classID string) ([]Subject, error)
	SubjectByID(ctx context.Context, schoolID, subjectID string) (Subject, error)
	EnrolledStudents(ctx context.Context, schoolID, classID string) ([]EnrolledStudent, error)

	MembershipByID(ctx context.Context, schoolID, membershipID string) (Membership, error)
	ClassTeacherMembership(ctx context.Context, schoolID, classID string) (string, error)
	HasClassTeacherBadge(ctx context.Context, schoolID, membershipID, classID string) (bool, error)
	AssignClassTeacher(ctx context.Context, schoolID, classID, membershipID, actorUserID string) error
	HasAnySubjectAssignment(ctx context.Context, schoolID, membershipID, classID string) (bool, error)
	SubjectAssignment(ctx context.Context, schoolID, classID, subjectID, academicYearID string) (SubjectTeacherAssignment, error)
	UpsertSubjectAssignment(ctx context.Context, schoolID string, a SubjectTeacherAssignment) error
	ListSubjectAssignments(ctx context.Context, schoolID, classID string) ([]SubjectTeacherAssignment, error)

	ListGradingScales(ctx context.Context, schoolID string) ([]GradingScale, error)
	ActiveGradingScale(ctx context.Context, schoolID string) (GradingScale, error)
	GradingScaleByID(ctx context.Context, schoolID, id string) (GradingScale, error)
	CreateGradingScale(ctx context.Context, schoolID string, g GradingScale) (GradingScale, error)

	ListTimetableSlots(ctx context.Context, schoolID, termID, classID string) ([]TimetableSlot, error)
	ReplaceTimetableSlots(ctx context.Context, schoolID, termID, classID string, slots []TimetableSlot) error

	ListCalendarEvents(ctx context.Context, schoolID, termID string) ([]CalendarEvent, error)
	UpsertCalendarEvents(ctx context.Context, schoolID string, events []CalendarEvent) error
	UpdateTermDates(ctx context.Context, schoolID, termID string, start, end time.Time) error

	UpsertAttendance(ctx context.Context, schoolID string, rows []AttendanceRow) (int, error)
	AttendanceForStudent(ctx context.Context, schoolID, studentID string) (AttendanceSummary, error)

	UpsertMarks(ctx context.Context, schoolID string, rows []Mark, enteredBy string) (int, error)
	MarksForClassTerm(ctx context.Context, schoolID, classID, termID string) ([]Mark, error)
	MarksForStudentTerm(ctx context.Context, schoolID, studentID, termID string) ([]Mark, error)
	MarksForStudentAllTerms(ctx context.Context, schoolID, studentID string) ([]Mark, error)

	ResultRelease(ctx context.Context, schoolID, classID, termID string) (ResultRelease, error)
	SaveResultRelease(ctx context.Context, schoolID string, r ResultRelease) error
	ReplaceSnapshots(ctx context.Context, schoolID, releaseID string, snapshots []Snapshot) error
	SnapshotsForRelease(ctx context.Context, schoolID, releaseID string) ([]Snapshot, error)
	SnapshotsForStudentTerm(ctx context.Context, schoolID, studentID, termID string) ([]Snapshot, error)
	SnapshotsForStudent(ctx context.Context, schoolID, studentID string) ([]Snapshot, error)

	UpsertReportComments(ctx context.Context, schoolID string, comments []ReportComment) (int, error)
	ReportCommentForStudentTerm(ctx context.Context, schoolID, studentID, termID string) (ReportComment, error)
	ReportCommentsForClassTerm(ctx context.Context, schoolID, classID, termID string) ([]ReportComment, error)

	GuardianIDsForStudent(ctx context.Context, schoolID, studentID string) ([]string, error)
	GuardianIDsForClass(ctx context.Context, schoolID, classID string) ([]string, error)
	GuardianIDsForSchool(ctx context.Context, schoolID string) ([]string, error)
	StaffUserIDsForSchool(ctx context.Context, schoolID string) ([]string, error)
	IsGuardian(ctx context.Context, schoolID, userID, studentID string) (bool, error)
	TeacherCanViewStudent(ctx context.Context, schoolID, membershipID, studentID string) (bool, error)

	AppendAudit(ctx context.Context, schoolID string, event AuditEvent) error
}

// newUUID returns a random RFC-4122 version 4 identifier. It is used only by
// the in-memory store; PostgreSQL generates UUIDs with gen_random_uuid().
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("academics: cannot generate uuid: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
