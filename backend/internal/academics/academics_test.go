package academics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testSchoolID = "00000000-0000-4000-8000-000000000001"
	testYearID   = "00000000-0000-4000-8000-000000000010"
	testTermID   = "00000000-0000-4000-8000-000000000020"
	testClassID  = "00000000-0000-4000-8000-000000000030"
	testGradeID  = "00000000-0000-4000-8000-000000000040"
	testMathID   = "00000000-0000-4000-8000-000000000050"
	testEngID    = "00000000-0000-4000-8000-000000000051"
	testStudentA = "00000000-0000-4000-8000-0000000000a1"
	testStudentB = "00000000-0000-4000-8000-0000000000b1"

	testUserDean    = "00000000-0000-4000-8000-000000000101"
	testUserCT      = "00000000-0000-4000-8000-000000000102"
	testUserMath    = "00000000-0000-4000-8000-000000000103"
	testUserEng     = "00000000-0000-4000-8000-000000000104"
	testUserOther   = "00000000-0000-4000-8000-000000000105"
	testUserParentA = "00000000-0000-4000-8000-000000000106"
	testUserParentB = "00000000-0000-4000-8000-000000000107"

	testMemDean    = "00000000-0000-4000-8000-000000000201"
	testMemCT      = "00000000-0000-4000-8000-000000000202"
	testMemMath    = "00000000-0000-4000-8000-000000000203"
	testMemEng     = "00000000-0000-4000-8000-000000000204"
	testMemOther   = "00000000-0000-4000-8000-000000000205"
	testMemParentA = "00000000-0000-4000-8000-000000000206"
	testMemParentB = "00000000-0000-4000-8000-000000000207"
)

type fakeNotifier struct {
	mu       sync.Mutex
	notified []Notification
}

func (f *fakeNotifier) CreateInAppNotification(_ context.Context, n Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notified = append(f.notified, n)
	return nil
}

func (f *fakeNotifier) forCategory(category string) []Notification {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Notification
	for _, n := range f.notified {
		if n.Category == category {
			out = append(out, n)
		}
	}
	return out
}

type feeSyncCall struct {
	SchoolID     string
	TermID       string
	NewStartDate time.Time
	ActorUserID  string
	Note         string
}

type fakeFeeSyncer struct {
	mu    sync.Mutex
	calls []feeSyncCall
}

func (f *fakeFeeSyncer) OnTermStartDateChanged(_ context.Context, schoolID, termID string, newStart time.Time, actorUserID, note string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, feeSyncCall{schoolID, termID, newStart, actorUserID, note})
	return nil
}

type fakeDrafter struct {
	mu   sync.Mutex
	last DraftContext
}

func (f *fakeDrafter) DraftReportComment(_ context.Context, req DraftContext) (DraftResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.last = req
	return DraftResult{DraftComment: "Synthetic draft comment.", RequiresTeacherApproval: true}, nil
}

type testEnv struct {
	store    *MemoryStore
	svc      *Service
	notifier *fakeNotifier
	feeSync  *fakeFeeSyncer
	drafter  *fakeDrafter
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func identityCtx(userID, membershipID, role string) context.Context {
	return WithIdentity(context.Background(), Identity{
		UserID:       userID,
		SchoolID:     testSchoolID,
		MembershipID: membershipID,
		Role:         role,
	})
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	store := NewMemoryStore()
	store.SetCurrentAcademicYear(testYearID)
	store.AddClass(Class{ID: testClassID, SchoolID: testSchoolID, GradeLevelID: testGradeID, StreamName: "East"})
	store.AddTerm(Term{
		ID: testTermID, SchoolID: testSchoolID, AcademicYearID: testYearID,
		Year: 2026, TermNumber: 1, StartDate: date(2026, 1, 5), EndDate: date(2026, 4, 3), IsCurrent: true,
	})
	store.AddSubject(Subject{ID: testMathID, SchoolID: testSchoolID, GradeLevelID: testGradeID, Code: "MAT", Name: "Mathematics", IsCompulsory: true})
	store.AddSubject(Subject{ID: testEngID, SchoolID: testSchoolID, GradeLevelID: testGradeID, Code: "ENG", Name: "English", IsCompulsory: true})
	store.AddStudent(Student{ID: testStudentA, SchoolID: testSchoolID, AdmissionNumber: "ADM-001", FullName: "Synthetic Student A", CurrentClassID: testClassID})
	store.AddStudent(Student{ID: testStudentB, SchoolID: testSchoolID, AdmissionNumber: "ADM-002", FullName: "Synthetic Student B", CurrentClassID: testClassID})
	store.Enrol(testSchoolID, testStudentA, testClassID)
	store.Enrol(testSchoolID, testStudentB, testClassID)

	for _, mem := range []Membership{
		{ID: testMemDean, SchoolID: testSchoolID, UserID: testUserDean, Role: RoleDean, Status: "active"},
		{ID: testMemCT, SchoolID: testSchoolID, UserID: testUserCT, Role: RoleTeacher, Status: "active"},
		{ID: testMemMath, SchoolID: testSchoolID, UserID: testUserMath, Role: RoleTeacher, Status: "active"},
		{ID: testMemEng, SchoolID: testSchoolID, UserID: testUserEng, Role: RoleTeacher, Status: "active"},
		{ID: testMemOther, SchoolID: testSchoolID, UserID: testUserOther, Role: RoleTeacher, Status: "active"},
		{ID: testMemParentA, SchoolID: testSchoolID, UserID: testUserParentA, Role: RoleParent, Status: "active"},
		{ID: testMemParentB, SchoolID: testSchoolID, UserID: testUserParentB, Role: RoleParent, Status: "active"},
	} {
		store.AddMembership(mem)
	}
	store.AddGuardian(testSchoolID, testStudentA, testUserParentA)
	store.AddGuardian(testSchoolID, testStudentB, testUserParentB)

	notifier := &fakeNotifier{}
	syncer := &fakeFeeSyncer{}
	drafter := &fakeDrafter{}
	svc := NewService(Deps{
		Store:          store,
		Notifier:       notifier,
		FeeSyncer:      syncer,
		CommentDrafter: drafter,
		Config:         Config{AllowPrincipalDeanOverride: true},
	})

	// Seed the first grading scale and teacher assignments as the Dean.
	ctx := identityCtx(testUserDean, testMemDean, RoleDean)
	if _, err := svc.CreateGradingScaleVersion(ctx, GradingScaleRequest{EEMin: 80, MEMin: 60, AEMin: 40, BEMin: 0}); err != nil {
		t.Fatalf("seed grading scale: %v", err)
	}
	if _, err := svc.AssignTeachers(ctx, AssignmentRequest{
		ClassID:                  testClassID,
		ClassTeacherMembershipID: testMemCT,
		SubjectAssignments: []SubjectAssignmentInput{
			{SubjectID: testMathID, TeacherMembershipID: testMemMath},
			{SubjectID: testEngID, TeacherMembershipID: testMemEng},
		},
	}); err != nil {
		t.Fatalf("seed assignments: %v", err)
	}
	return &testEnv{store: store, svc: svc, notifier: notifier, feeSync: syncer, drafter: drafter}
}

func saveMarks(t *testing.T, env *testEnv, ctx context.Context, subjectID string, scores map[string]float64) MarksResult {
	t.Helper()
	req := MarksRequest{TermID: testTermID}
	for studentID, score := range scores {
		req.Marks = append(req.Marks, MarkInput{StudentID: studentID, Score: score})
	}
	result, err := env.svc.SaveMarks(ctx, testClassID, subjectID, req)
	if err != nil {
		t.Fatalf("save marks: %v", err)
	}
	return result
}

func requireAPIError(t *testing.T, err error, status int, code string) *APIError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %s, got nil", code)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != status || apiErr.Code != code {
		t.Fatalf("expected %d %s, got %d %s (%s)", status, code, apiErr.Status, apiErr.Code, apiErr.Message)
	}
	return apiErr
}

func releaseComplete(t *testing.T, env *testEnv, mathA, mathB, engA, engB float64) {
	t.Helper()
	saveMarks(t, env, identityCtx(testUserMath, testMemMath, RoleTeacher), testMathID, map[string]float64{
		testStudentA: mathA, testStudentB: mathB,
	})
	saveMarks(t, env, identityCtx(testUserEng, testMemEng, RoleTeacher), testEngID, map[string]float64{
		testStudentA: engA, testStudentB: engB,
	})
	if _, err := env.svc.ReleaseResults(identityCtx(testUserCT, testMemCT, RoleTeacher), testClassID, ReleaseRequest{TermID: testTermID}); err != nil {
		t.Fatalf("release results: %v", err)
	}
}

// TestAttendanceAuthorizationAndNotifications covers AC-14 / C35 / C36.
func TestAttendanceAuthorizationAndNotifications(t *testing.T) {
	env := newTestEnv(t)
	classTeacher := identityCtx(testUserCT, testMemCT, RoleTeacher)
	otherTeacher := identityCtx(testUserOther, testMemOther, RoleTeacher)

	result, err := env.svc.RecordAttendance(classTeacher, testClassID, AttendanceRequest{
		Date: "2026-01-10",
		Records: []AttendanceRecordInput{
			{StudentID: testStudentA, Status: "present"},
			{StudentID: testStudentB, Status: "absent"},
		},
	})
	if err != nil {
		t.Fatalf("class teacher roll-call: %v", err)
	}
	if result.RecordedCount != 2 {
		t.Fatalf("expected 2 attendance records, got %d", result.RecordedCount)
	}
	if result.AbsentNotificationsQueued != 1 {
		t.Fatalf("expected 1 absence notification, got %d", result.AbsentNotificationsQueued)
	}
	notes := env.notifier.forCategory("academics.attendance")
	if len(notes) != 1 || notes[0].RecipientUserID != testUserParentB {
		t.Fatalf("expected only Student B's guardian to be notified, got %+v", notes)
	}

	_, err = env.svc.RecordAttendance(otherTeacher, testClassID, AttendanceRequest{
		Date:    "2026-01-10",
		Records: []AttendanceRecordInput{{StudentID: testStudentA, Status: "present"}},
	})
	requireAPIError(t, err, http.StatusForbidden, CodeClassTeacherScope)
}

// TestMarksAssignmentHardGateAndReassignment covers AC-15, AC-16, AC-17.
func TestMarksAssignmentHardGateAndReassignment(t *testing.T) {
	env := newTestEnv(t)
	mathTeacher := identityCtx(testUserMath, testMemMath, RoleTeacher)
	engTeacher := identityCtx(testUserEng, testMemEng, RoleTeacher)
	otherTeacher := identityCtx(testUserOther, testMemOther, RoleTeacher)
	classTeacher := identityCtx(testUserCT, testMemCT, RoleTeacher)
	dean := identityCtx(testUserDean, testMemDean, RoleDean)

	// An unassigned teacher cannot write this subject.
	_, err := env.svc.SaveMarks(otherTeacher, testClassID, testMathID, MarksRequest{
		TermID: testTermID,
		Marks:  []MarkInput{{StudentID: testStudentA, Score: 70}},
	})
	requireAPIError(t, err, http.StatusForbidden, CodeSubjectTeacherScope)

	// Math is complete, English is missing Student B.
	saveMarks(t, env, mathTeacher, testMathID, map[string]float64{testStudentA: 70, testStudentB: 70})
	saveMarks(t, env, engTeacher, testEngID, map[string]float64{testStudentA: 60})

	_, err = env.svc.ReleaseResults(classTeacher, testClassID, ReleaseRequest{TermID: testTermID})
	apiErr := requireAPIError(t, err, http.StatusUnprocessableEntity, CodeResultsIncomplete)
	details, _ := apiErr.Details["missing_subjects"].([]map[string]any)
	_ = details
	if apiErr.Details["missing_count"] != 1 {
		t.Fatalf("expected missing_count=1, got %v", apiErr.Details["missing_count"])
	}

	// The Dean reassigns Mathematics. The old teacher immediately loses write
	// access and the new teacher gains it (AC-17).
	if _, err := env.svc.ReassignMarks(dean, testClassID, ReassignMarksRequest{
		SubjectID: testMathID, NewTeacherMembershipID: testMemOther, Reason: "Medical leave",
	}); err != nil {
		t.Fatalf("reassign marks: %v", err)
	}
	_, err = env.svc.SaveMarks(mathTeacher, testClassID, testMathID, MarksRequest{
		TermID: testTermID, Marks: []MarkInput{{StudentID: testStudentA, Score: 71}},
	})
	requireAPIError(t, err, http.StatusForbidden, CodeSubjectTeacherScope)
	if _, err := env.svc.SaveMarks(otherTeacher, testClassID, testMathID, MarksRequest{
		TermID: testTermID, Marks: []MarkInput{{StudentID: testStudentA, Score: 71}, {StudentID: testStudentB, Score: 71}},
	}); err != nil {
		t.Fatalf("new teacher should be able to save marks: %v", err)
	}

	// Complete the missing English mark and release.
	saveMarks(t, env, engTeacher, testEngID, map[string]float64{testStudentB: 60})
	release, err := env.svc.ReleaseResults(classTeacher, testClassID, ReleaseRequest{TermID: testTermID})
	if err != nil {
		t.Fatalf("release after completion: %v", err)
	}
	if release.Status != "released" || release.SnapshotCount != 4 {
		t.Fatalf("unexpected release result: %+v", release)
	}

	// Released marks are locked (409).
	_, err = env.svc.SaveMarks(otherTeacher, testClassID, testMathID, MarksRequest{
		TermID: testTermID, Marks: []MarkInput{{StudentID: testStudentA, Score: 99}},
	})
	requireAPIError(t, err, http.StatusConflict, CodeResultsLocked)

	// Sensitive actions are audited.
	actions := map[string]bool{}
	for _, event := range env.store.AuditEvents() {
		actions[event.Action] = true
	}
	for _, want := range []string{"marks_entry_reassigned", "results_released"} {
		if !actions[want] {
			t.Fatalf("expected audit action %q, got %+v", want, actions)
		}
	}
}

// TestReopenAndRereleaseSnapshotIsolation covers AC-18.
func TestReopenAndRereleaseSnapshotIsolation(t *testing.T) {
	env := newTestEnv(t)
	classTeacher := identityCtx(testUserCT, testMemCT, RoleTeacher)
	mathTeacher := identityCtx(testUserMath, testMemMath, RoleTeacher)
	parentA := identityCtx(testUserParentA, testMemParentA, RoleParent)
	parentB := identityCtx(testUserParentB, testMemParentB, RoleParent)

	releaseComplete(t, env, 70, 70, 60, 60)

	resultsA, err := env.svc.GetStudentResults(parentA, testStudentA, testTermID)
	if err != nil {
		t.Fatalf("parent results before reopen: %v", err)
	}
	if got := findSubject(t, resultsA, testMathID); got.Score != 70 || got.CBCLevel != LevelME {
		t.Fatalf("expected initial Math 70 ME, got %+v", got)
	}

	if _, err := env.svc.ReopenResults(classTeacher, testClassID, ReopenRequest{
		TermID: testTermID, Reason: "Correcting Mathematics paper 2",
	}); err != nil {
		t.Fatalf("reopen: %v", err)
	}

	// Reopened results must go through re-release, not the initial release path.
	_, err = env.svc.ReleaseResults(classTeacher, testClassID, ReleaseRequest{TermID: testTermID})
	requireAPIError(t, err, http.StatusConflict, CodeResultsReopened)

	// The subject teacher corrects the mark while the class is reopened.
	saveMarks(t, env, mathTeacher, testMathID, map[string]float64{testStudentA: 85})

	// Parents must still see the pre-reopen snapshot.
	whileReopened, err := env.svc.GetStudentResults(parentA, testStudentA, testTermID)
	if err != nil {
		t.Fatalf("parent results while reopened: %v", err)
	}
	if got := findSubject(t, whileReopened, testMathID); got.Score != 70 || got.CBCLevel != LevelME || got.IsAmended {
		t.Fatalf("expected the old snapshot while reopened, got %+v", got)
	}
	history, err := env.svc.GetPerformanceHistory(parentA, testStudentA)
	if err != nil {
		t.Fatalf("performance history while reopened: %v", err)
	}
	if len(history.Terms) != 1 || findPerformanceSubject(t, history.Terms[0], testMathID).CBCLevel != LevelME {
		t.Fatalf("performance history should still show the old level, got %+v", history)
	}

	// Re-release without the mandatory teacher note is rejected (422).
	_, err = env.svc.RereleaseResults(classTeacher, testClassID, RereleaseRequest{TermID: testTermID})
	requireAPIError(t, err, http.StatusUnprocessableEntity, CodeRereleaseNote)

	// With the note it succeeds and only Student A's guardian is notified.
	result, err := env.svc.RereleaseResults(classTeacher, testClassID, RereleaseRequest{
		TermID: testTermID,
		CorrectionNotes: []CorrectionNoteInput{{
			StudentID: testStudentA, SubjectID: testMathID, TeacherNote: "Paper 2 score addition fixed",
		}},
	})
	if err != nil {
		t.Fatalf("re-release: %v", err)
	}
	if result.AmendedMarks != 1 || result.AffectedStudentsNotified != 1 {
		t.Fatalf("expected one amended mark and one notification, got %+v", result)
	}
	amendedNotes := env.notifier.forCategory("academics.results.amended")
	if len(amendedNotes) != 1 || amendedNotes[0].RecipientUserID != testUserParentA {
		t.Fatalf("expected only Student A's guardian to be notified, got %+v", amendedNotes)
	}

	// Parent A sees the new mark with the amendment note.
	after, err := env.svc.GetStudentResults(parentA, testStudentA, testTermID)
	if err != nil {
		t.Fatalf("parent results after re-release: %v", err)
	}
	mathA := findSubject(t, after, testMathID)
	if mathA.Score != 85 || mathA.CBCLevel != LevelEE || !mathA.IsAmended || mathA.AmendmentNote == "" {
		t.Fatalf("expected amended Math 85 EE with note, got %+v", mathA)
	}

	// Parent B keeps the unaffected original mark and is not notified.
	resultsB, err := env.svc.GetStudentResults(parentB, testStudentB, testTermID)
	if err != nil {
		t.Fatalf("parent B results: %v", err)
	}
	mathB := findSubject(t, resultsB, testMathID)
	if mathB.Score != 70 || mathB.IsAmended {
		t.Fatalf("Student B should be unchanged, got %+v", mathB)
	}
}

// TestGradingScaleVersioningAndAnonymisedStats covers AC-19.
func TestGradingScaleVersioningAndAnonymisedStats(t *testing.T) {
	env := newTestEnv(t)
	otherTeacher := identityCtx(testUserOther, testMemOther, RoleTeacher)
	parentA := identityCtx(testUserParentA, testMemParentA, RoleParent)
	dean := identityCtx(testUserDean, testMemDean, RoleDean)

	releaseComplete(t, env, 82, 82, 70, 70)

	// Editing the scale creates a new version; Term 1 remains EE.
	newScale, err := env.svc.CreateGradingScaleVersion(dean, GradingScaleRequest{EEMin: 85, MEMin: 65, AEMin: 45, BEMin: 0})
	if err != nil {
		t.Fatalf("create new scale: %v", err)
	}
	if newScale.VersionNumber != 2 {
		t.Fatalf("expected version 2, got %d", newScale.VersionNumber)
	}
	if resolveCBCLevel(newScale, 82) != LevelME {
		t.Fatalf("new scale should map 82 to ME")
	}
	results, err := env.svc.GetStudentResults(parentA, testStudentA, testTermID)
	if err != nil {
		t.Fatalf("results after scale edit: %v", err)
	}
	if got := findSubject(t, results, testMathID); got.Score != 82 || got.CBCLevel != LevelEE {
		t.Fatalf("past released term must keep EE under the old scale, got %+v", got)
	}

	// Stats are available to any teacher but contain zero student identifiers.
	stats, err := env.svc.GetClassStats(otherTeacher, testClassID, testTermID)
	if err != nil {
		t.Fatalf("class stats: %v", err)
	}
	if stats.TotalStudents != 2 || stats.ClassMeanScore == 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	raw, _ := json.Marshal(stats)
	payload := string(raw)
	for _, forbidden := range []string{
		"Synthetic Student A", "Synthetic Student B", "ADM-001", "ADM-002",
		testStudentA, testStudentB, "rank", "position", "leaderboard",
	} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("stats payload leaked %q: %s", forbidden, payload)
		}
	}
}

// TestTermStartDateFeeSync covers AC-22 / C61.
func TestTermStartDateFeeSync(t *testing.T) {
	env := newTestEnv(t)
	dean := identityCtx(testUserDean, testMemDean, RoleDean)

	// Start-date change without a note is rejected before any writes.
	_, err := env.svc.PutCalendar(dean, CalendarRequest{TermUpdates: []TermUpdateInput{{
		TermID: testTermID, StartDate: "2026-01-12", EndDate: "2026-04-03",
	}}})
	requireAPIError(t, err, http.StatusUnprocessableEntity, CodeTermDateNote)
	term, err := env.store.TermByID(context.Background(), testSchoolID, testTermID)
	if err != nil {
		t.Fatalf("read term: %v", err)
	}
	if !dateOnly(term.StartDate).Equal(date(2026, 1, 5)) {
		t.Fatalf("term start date should not have changed, got %v", term.StartDate)
	}
	if len(env.feeSync.calls) != 0 {
		t.Fatalf("fee sync should not be called on rejected update")
	}

	// With the note the term moves and billing is called.
	calendar, err := env.svc.PutCalendar(dean, CalendarRequest{TermUpdates: []TermUpdateInput{{
		TermID: testTermID, StartDate: "2026-01-12", EndDate: "2026-04-03",
		FeeChangeNote: "Ministry delayed Term 1 opening by one week",
	}}})
	if err != nil {
		t.Fatalf("calendar update: %v", err)
	}
	if !calendar.FeeDueDatesShifted {
		t.Fatalf("expected fee due dates to shift")
	}
	if len(env.feeSync.calls) != 1 || env.feeSync.calls[0].Note != "Ministry delayed Term 1 opening by one week" {
		t.Fatalf("unexpected fee sync calls: %+v", env.feeSync.calls)
	}
	term, _ = env.store.TermByID(context.Background(), testSchoolID, testTermID)
	if !dateOnly(term.StartDate).Equal(date(2026, 1, 12)) {
		t.Fatalf("expected term start 2026-01-12, got %v", term.StartDate)
	}

	// Moving only the end date does not require a fee note or call billing.
	calendar, err = env.svc.PutCalendar(dean, CalendarRequest{TermUpdates: []TermUpdateInput{{
		TermID: testTermID, StartDate: "2026-01-12", EndDate: "2026-04-10",
	}}})
	if err != nil {
		t.Fatalf("end-date-only update: %v", err)
	}
	if calendar.FeeDueDatesShifted || len(env.feeSync.calls) != 1 {
		t.Fatalf("end-date-only change must not shift fees")
	}

	audited := false
	for _, event := range env.store.AuditEvents() {
		if event.Action == "term_dates_updated" {
			audited = true
		}
	}
	if !audited {
		t.Fatalf("term date changes must be audited")
	}
}

// TestRosterPrivacyScoping verifies the roster/name boundary.
func TestRosterPrivacyScoping(t *testing.T) {
	env := newTestEnv(t)
	classTeacher := identityCtx(testUserCT, testMemCT, RoleTeacher)
	mathTeacher := identityCtx(testUserMath, testMemMath, RoleTeacher)
	otherTeacher := identityCtx(testUserOther, testMemOther, RoleTeacher)

	roster, err := env.svc.GetRoster(classTeacher, testClassID)
	if err != nil || len(roster.Students) != 2 {
		t.Fatalf("class teacher roster: %v %+v", err, roster)
	}
	if _, err := env.svc.GetRoster(mathTeacher, testClassID); err != nil {
		t.Fatalf("assigned subject teacher should see roster: %v", err)
	}
	_, err = env.svc.GetRoster(otherTeacher, testClassID)
	requireAPIError(t, err, http.StatusForbidden, CodeForbidden)
}

// TestAIDrafterProxy verifies the draft proxy returns a draft for approval.
func TestAIDrafterProxy(t *testing.T) {
	env := newTestEnv(t)
	mathTeacher := identityCtx(testUserMath, testMemMath, RoleTeacher)
	saveMarks(t, env, mathTeacher, testMathID, map[string]float64{testStudentA: 82})

	result, err := env.svc.DraftReportComment(mathTeacher, DraftCommentRequest{
		Mode: "term_comment", StudentID: testStudentA, TermID: testTermID,
	})
	if err != nil {
		t.Fatalf("draft report comment: %v", err)
	}
	if result.DraftComment == "" || !result.RequiresTeacherApproval {
		t.Fatalf("unexpected draft: %+v", result)
	}
	if len(env.drafter.last.CurrentTermMarks) != 1 || env.drafter.last.CurrentTermMarks[0].CBCLevel != LevelEE {
		t.Fatalf("drafter did not receive the current marks/level: %+v", env.drafter.last)
	}
}

// TestInternalSummaries verifies the read methods Ian's internalapi calls for
// Sos's AI service. They are tenant-scoped but do not require a caller identity
// because internalapi has already verified guardian ownership.
func TestInternalSummaries(t *testing.T) {
	env := newTestEnv(t)
	ctx := identityCtx(testUserCT, testMemCT, RoleTeacher)
	if _, err := env.svc.RecordAttendance(ctx, testClassID, AttendanceRequest{
		Date: "2026-01-10", Records: []AttendanceRecordInput{
			{StudentID: testStudentA, Status: "present"},
			{StudentID: testStudentB, Status: "late"},
		},
	}); err != nil {
		t.Fatalf("attendance: %v", err)
	}
	saveMarks(t, env, identityCtx(testUserMath, testMemMath, RoleTeacher), testMathID, map[string]float64{testStudentA: 82})

	attendance, err := env.svc.GetStudentAttendanceSummary(context.Background(), testSchoolID, testStudentA)
	if err != nil {
		t.Fatalf("attendance summary: %v", err)
	}
	if attendance.TotalDays != 1 || attendance.Present != 1 {
		t.Fatalf("unexpected attendance summary: %+v", attendance)
	}

	marks, err := env.svc.GetStudentMarksSummary(context.Background(), testSchoolID, testStudentA, testTermID)
	if err != nil {
		t.Fatalf("marks summary: %v", err)
	}
	if len(marks.Marks) != 1 || marks.Marks[0].CBCLevel != LevelEE {
		t.Fatalf("unexpected marks summary: %+v", marks)
	}
}

// TestHTTPRouteEnvelopeAndSyllabusStub verifies route wiring and the standard
// JSON error envelope plus the Release 2 stub.
func TestHTTPRouteEnvelopeAndSyllabusStub(t *testing.T) {
	env := newTestEnv(t)
	mux := http.NewServeMux()
	RegisterRoutes(mux, Deps{
		Store:          env.store,
		Notifier:       env.notifier,
		FeeSyncer:      env.feeSync,
		CommentDrafter: env.drafter,
		Config:         Config{AllowPrincipalDeanOverride: true},
	})

	body := bytes.NewBufferString(`{"term_id":"` + testTermID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/classes/"+testClassID+"/results/release", body)
	req = req.WithContext(identityCtx(testUserOther, testMemOther, RoleTeacher))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-class teacher release, got %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Status string `json:"status"`
		Error  struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Status != "error" || envelope.Error.Code != CodeClassTeacherScope {
		t.Fatalf("unexpected error envelope: %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/syllabus/coverage", bytes.NewBufferString(`{}`))
	req = req.WithContext(identityCtx(testUserCT, testMemCT, RoleTeacher))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for Release 2 stub, got %d", rec.Code)
	}
}

func findSubject(t *testing.T, results StudentResultsResponse, subjectID string) StudentSubjectResult {
	t.Helper()
	for _, subject := range results.Subjects {
		if subject.SubjectID == subjectID {
			return subject
		}
	}
	t.Fatalf("subject %s not found in %+v", subjectID, results.Subjects)
	return StudentSubjectResult{}
}

func findPerformanceSubject(t *testing.T, term TermPerformance, subjectID string) PerformanceSubject {
	t.Helper()
	for _, subject := range term.Subjects {
		if subject.SubjectID == subjectID {
			return subject
		}
	}
	t.Fatalf("subject %s not found in %+v", subjectID, term.Subjects)
	return PerformanceSubject{}
}
