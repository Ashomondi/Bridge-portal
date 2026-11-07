package academics

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStore is a deterministic in-memory Store used by the unit tests and by
// local demos before Evans's PostgreSQL layer is wired in. It enforces the same
// tenant keys as the SQL implementation: every lookup is keyed by school_id.
type MemoryStore struct {
	mu sync.RWMutex

	classes      map[string]Class
	terms        map[string]Term
	students     map[string]Student
	subjects     map[string]Subject
	enrolments   map[string]string // studentID -> classID
	memberships  map[string]Membership
	classTeacher map[string]string // classID -> membershipID
	assignments  map[string]SubjectTeacherAssignment

	scales      map[string]GradingScale
	scaleOrder  map[string][]string // schoolID -> scale IDs
	timetables  map[string][]TimetableSlot
	events      []CalendarEvent
	attendance  map[string]AttendanceRow
	marks       map[string]Mark
	releases    map[string]ResultRelease
	snapshots   map[string]Snapshot
	comments    map[string]ReportComment
	guardians   map[string][]string // schoolID|studentID -> user IDs
	audits      []AuditEvent
	currentYear string
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		classes:      map[string]Class{},
		terms:        map[string]Term{},
		students:     map[string]Student{},
		subjects:     map[string]Subject{},
		enrolments:   map[string]string{},
		memberships:  map[string]Membership{},
		classTeacher: map[string]string{},
		assignments:  map[string]SubjectTeacherAssignment{},
		scales:       map[string]GradingScale{},
		scaleOrder:   map[string][]string{},
		timetables:   map[string][]TimetableSlot{},
		attendance:   map[string]AttendanceRow{},
		marks:        map[string]Mark{},
		releases:     map[string]ResultRelease{},
		snapshots:    map[string]Snapshot{},
		comments:     map[string]ReportComment{},
		guardians:    map[string][]string{},
	}
}

func key(parts ...string) string { return strings.Join(parts, "|") }

// WithTx executes fn. The memory store has no real transaction to commit; the
// service layer validates before writing, and tests exercise the same service
// paths as production.
func (m *MemoryStore) WithTx(_ context.Context, _ string, fn func(tx Store) error) error {
	return fn(m)
}

// ---- Seeding helpers used by tests, demos and the synthetic-data mandate. ----

func (m *MemoryStore) AddClass(c Class) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.classes[c.ID] = c
}
func (m *MemoryStore) AddTerm(t Term) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terms[t.ID] = t
}
func (m *MemoryStore) AddSubject(s Subject) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subjects[s.ID] = s
}
func (m *MemoryStore) AddStudent(s Student) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.students[s.ID] = s
}
func (m *MemoryStore) Enrol(schoolID, studentID, classID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enrolments[key(schoolID, studentID)] = classID
}
func (m *MemoryStore) AddMembership(mem Membership) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memberships[mem.ID] = mem
}
func (m *MemoryStore) AddGuardian(schoolID, studentID, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(schoolID, studentID)
	m.guardians[k] = append(m.guardians[k], userID)
}
func (m *MemoryStore) SetCurrentAcademicYear(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentYear = id
}
func (m *MemoryStore) AuditEvents() []AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]AuditEvent, len(m.audits))
	copy(out, m.audits)
	return out
}

// ---- Reads ----

func (m *MemoryStore) CurrentAcademicYearID(_ context.Context, schoolID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentYear != "" {
		return m.currentYear, nil
	}
	for _, t := range m.terms {
		if t.SchoolID == schoolID && t.IsCurrent {
			return t.AcademicYearID, nil
		}
	}
	for _, t := range m.terms {
		if t.SchoolID == schoolID {
			return t.AcademicYearID, nil
		}
	}
	return "", ErrNotFound
}

func (m *MemoryStore) AcademicYearExists(_ context.Context, schoolID, academicYearID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentYear == academicYearID {
		return true, nil
	}
	for _, t := range m.terms {
		if t.SchoolID == schoolID && t.AcademicYearID == academicYearID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) ClassByID(_ context.Context, schoolID, classID string) (Class, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.classes[classID]
	if !ok || c.SchoolID != schoolID {
		return Class{}, ErrNotFound
	}
	return c, nil
}

func (m *MemoryStore) TermByID(_ context.Context, schoolID, termID string) (Term, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.terms[termID]
	if !ok || t.SchoolID != schoolID {
		return Term{}, ErrNotFound
	}
	return t, nil
}

func (m *MemoryStore) TermForDate(_ context.Context, schoolID string, date time.Time) (Term, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	day := dateOnly(date)
	for _, t := range m.terms {
		if t.SchoolID != schoolID {
			continue
		}
		if !day.Before(dateOnly(t.StartDate)) && !day.After(dateOnly(t.EndDate)) {
			return t, nil
		}
	}
	for _, t := range m.terms {
		if t.SchoolID == schoolID && t.IsCurrent {
			return t, nil
		}
	}
	return Term{}, ErrNotFound
}

func (m *MemoryStore) ListTerms(_ context.Context, schoolID string) ([]Term, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Term
	for _, t := range m.terms {
		if t.SchoolID == schoolID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Year != out[j].Year {
			return out[i].Year < out[j].Year
		}
		return out[i].TermNumber < out[j].TermNumber
	})
	return out, nil
}

func (m *MemoryStore) StudentByID(_ context.Context, schoolID, studentID string) (Student, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.students[studentID]
	if !ok || s.SchoolID != schoolID {
		return Student{}, ErrNotFound
	}
	return s, nil
}

func (m *MemoryStore) SubjectsForClass(_ context.Context, schoolID, classID string) ([]Subject, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.classes[classID]
	if !ok || c.SchoolID != schoolID {
		return nil, ErrNotFound
	}
	var out []Subject
	for _, s := range m.subjects {
		if s.SchoolID == schoolID && s.GradeLevelID == c.GradeLevelID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (m *MemoryStore) SubjectByID(_ context.Context, schoolID, subjectID string) (Subject, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.subjects[subjectID]
	if !ok || s.SchoolID != schoolID {
		return Subject{}, ErrNotFound
	}
	return s, nil
}

func (m *MemoryStore) EnrolledStudents(_ context.Context, schoolID, classID string) ([]EnrolledStudent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []EnrolledStudent
	for k, cid := range m.enrolments {
		parts := strings.SplitN(k, "|", 2)
		if len(parts) != 2 || parts[0] != schoolID || cid != classID {
			continue
		}
		s, ok := m.students[parts[1]]
		if !ok {
			continue
		}
		out = append(out, EnrolledStudent{
			StudentID:       s.ID,
			AdmissionNumber: s.AdmissionNumber,
			FullName:        s.FullName,
			ClassID:         classID,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AdmissionNumber < out[j].AdmissionNumber })
	return out, nil
}

func (m *MemoryStore) MembershipByID(_ context.Context, schoolID, membershipID string) (Membership, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mem, ok := m.memberships[membershipID]
	if !ok || mem.SchoolID != schoolID {
		return Membership{}, ErrNotFound
	}
	return mem, nil
}

func (m *MemoryStore) ClassTeacherMembership(_ context.Context, schoolID, classID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.classes[classID]
	if !ok || c.SchoolID != schoolID {
		return "", ErrNotFound
	}
	id, ok := m.classTeacher[classID]
	if !ok || id == "" {
		return "", ErrNotFound
	}
	return id, nil
}

func (m *MemoryStore) HasClassTeacherBadge(_ context.Context, schoolID, membershipID, classID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.classes[classID]
	if !ok || c.SchoolID != schoolID {
		return false, nil
	}
	return m.classTeacher[classID] == membershipID, nil
}

func (m *MemoryStore) AssignClassTeacher(_ context.Context, schoolID, classID, membershipID, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.classes[classID]
	if !ok || c.SchoolID != schoolID {
		return ErrNotFound
	}
	if _, ok := m.memberships[membershipID]; !ok {
		return ErrNotFound
	}
	m.classTeacher[classID] = membershipID
	return nil
}

func (m *MemoryStore) HasAnySubjectAssignment(_ context.Context, schoolID, membershipID, classID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.assignments {
		if a.SchoolID == schoolID && a.ClassID == classID && a.TeacherMembershipID == membershipID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) SubjectAssignment(_ context.Context, schoolID, classID, subjectID, academicYearID string) (SubjectTeacherAssignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.assignments[key(schoolID, academicYearID, classID, subjectID)]
	if !ok {
		return SubjectTeacherAssignment{}, ErrNotFound
	}
	return a, nil
}

func (m *MemoryStore) UpsertSubjectAssignment(_ context.Context, schoolID string, a SubjectTeacherAssignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = newUUID()
	}
	a.SchoolID = schoolID
	a.UpdatedAt = time.Now()
	m.assignments[key(schoolID, a.AcademicYearID, a.ClassID, a.SubjectID)] = a
	return nil
}

func (m *MemoryStore) ListSubjectAssignments(_ context.Context, schoolID, classID string) ([]SubjectTeacherAssignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []SubjectTeacherAssignment
	for _, a := range m.assignments {
		if a.SchoolID == schoolID && a.ClassID == classID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (m *MemoryStore) ListGradingScales(_ context.Context, schoolID string) ([]GradingScale, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []GradingScale
	for _, id := range m.scaleOrder[schoolID] {
		out = append(out, m.scales[id])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VersionNumber < out[j].VersionNumber })
	return out, nil
}

func (m *MemoryStore) ActiveGradingScale(_ context.Context, schoolID string) (GradingScale, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var best GradingScale
	found := false
	for _, id := range m.scaleOrder[schoolID] {
		g := m.scales[id]
		if g.IsActive && (!found || g.VersionNumber > best.VersionNumber) {
			best = g
			found = true
		}
	}
	if !found {
		return GradingScale{}, ErrNotFound
	}
	return best, nil
}

func (m *MemoryStore) GradingScaleByID(_ context.Context, schoolID, id string) (GradingScale, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	g, ok := m.scales[id]
	if !ok || g.SchoolID != schoolID {
		return GradingScale{}, ErrNotFound
	}
	return g, nil
}

func (m *MemoryStore) CreateGradingScale(_ context.Context, schoolID string, g GradingScale) (GradingScale, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, old := range m.scales {
		if old.SchoolID == schoolID && old.IsActive {
			old.IsActive = false
			m.scales[id] = old
		}
	}
	if g.ID == "" {
		g.ID = newUUID()
	}
	g.SchoolID = schoolID
	g.CreatedAt = time.Now()
	m.scales[g.ID] = g
	m.scaleOrder[schoolID] = append(m.scaleOrder[schoolID], g.ID)
	return g, nil
}

func (m *MemoryStore) ListTimetableSlots(_ context.Context, schoolID, termID, classID string) ([]TimetableSlot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if classID != "" {
		return append([]TimetableSlot(nil), m.timetables[key(schoolID, termID, classID)]...), nil
	}
	var out []TimetableSlot
	for k, slots := range m.timetables {
		if strings.HasPrefix(k, schoolID+"|"+termID+"|") {
			out = append(out, slots...)
		}
	}
	return out, nil
}

func (m *MemoryStore) ReplaceTimetableSlots(_ context.Context, schoolID, termID, classID string, slots []TimetableSlot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range slots {
		if slots[i].ID == "" {
			slots[i].ID = newUUID()
		}
		slots[i].SchoolID = schoolID
		slots[i].TermID = termID
		slots[i].ClassID = classID
	}
	m.timetables[key(schoolID, termID, classID)] = append([]TimetableSlot(nil), slots...)
	return nil
}

func (m *MemoryStore) ListCalendarEvents(_ context.Context, schoolID, termID string) ([]CalendarEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []CalendarEvent
	for _, e := range m.events {
		if e.SchoolID == schoolID && (termID == "" || e.TermID == termID) {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartAt.Before(out[j].StartAt) })
	return out, nil
}

func (m *MemoryStore) UpsertCalendarEvents(_ context.Context, schoolID string, events []CalendarEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range events {
		if e.ID == "" {
			e.ID = newUUID()
			e.CreatedAt = time.Now()
			e.SchoolID = schoolID
			m.events = append(m.events, e)
			continue
		}
		replaced := false
		for i := range m.events {
			if m.events[i].ID == e.ID && m.events[i].SchoolID == schoolID {
				m.events[i] = e
				replaced = true
				break
			}
		}
		if !replaced {
			e.SchoolID = schoolID
			m.events = append(m.events, e)
		}
	}
	return nil
}

func (m *MemoryStore) UpdateTermDates(_ context.Context, schoolID, termID string, start, end time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.terms[termID]
	if !ok || t.SchoolID != schoolID {
		return ErrNotFound
	}
	t.StartDate = start
	t.EndDate = end
	m.terms[termID] = t
	return nil
}

func (m *MemoryStore) UpsertAttendance(_ context.Context, schoolID string, rows []AttendanceRow) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range rows {
		r.SchoolID = schoolID
		if r.Source == "" {
			r.Source = "manual_roll_call"
		}
		m.attendance[key(schoolID, r.StudentID, dateOnly(r.AttendanceDate).Format("2006-01-02"))] = r
	}
	return len(rows), nil
}

func (m *MemoryStore) AttendanceForStudent(_ context.Context, schoolID, studentID string) (AttendanceSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var s AttendanceSummary
	for k, r := range m.attendance {
		if !strings.HasPrefix(k, key(schoolID, studentID)+"|") {
			continue
		}
		s.TotalDays++
		switch r.Status {
		case "present":
			s.Present++
		case "absent":
			s.Absent++
		case "late":
			s.Late++
		}
	}
	return s, nil
}

func (m *MemoryStore) UpsertMarks(_ context.Context, schoolID string, rows []Mark, enteredBy string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range rows {
		r.SchoolID = schoolID
		r.EnteredBy = enteredBy
		r.UpdatedAt = time.Now()
		m.marks[key(schoolID, r.StudentID, r.TermID, r.SubjectID)] = r
	}
	return len(rows), nil
}

func (m *MemoryStore) MarksForClassTerm(_ context.Context, schoolID, classID, termID string) ([]Mark, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Mark
	for _, r := range m.marks {
		if r.SchoolID == schoolID && r.ClassID == classID && r.TermID == termID {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].StudentID != out[j].StudentID {
			return out[i].StudentID < out[j].StudentID
		}
		return out[i].SubjectID < out[j].SubjectID
	})
	return out, nil
}

func (m *MemoryStore) MarksForStudentTerm(_ context.Context, schoolID, studentID, termID string) ([]Mark, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Mark
	for _, r := range m.marks {
		if r.SchoolID == schoolID && r.StudentID == studentID && r.TermID == termID {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SubjectID < out[j].SubjectID })
	return out, nil
}

func (m *MemoryStore) MarksForStudentAllTerms(_ context.Context, schoolID, studentID string) ([]Mark, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Mark
	for _, r := range m.marks {
		if r.SchoolID == schoolID && r.StudentID == studentID {
			out = append(out, r)
		}
	}
	return out, nil
}

func releaseKey(schoolID, classID, termID string) string { return key(schoolID, classID, termID) }

func (m *MemoryStore) ResultRelease(_ context.Context, schoolID, classID, termID string) (ResultRelease, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.releases[releaseKey(schoolID, classID, termID)]
	if !ok {
		return ResultRelease{}, ErrNotFound
	}
	return r, nil
}

func (m *MemoryStore) SaveResultRelease(_ context.Context, schoolID string, r ResultRelease) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = newUUID()
	}
	r.SchoolID = schoolID
	m.releases[releaseKey(schoolID, r.ClassID, r.TermID)] = r
	return nil
}

func (m *MemoryStore) ReplaceSnapshots(_ context.Context, schoolID, releaseID string, snapshots []Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, s := range m.snapshots {
		if s.SchoolID == schoolID && s.ResultReleaseID == releaseID {
			delete(m.snapshots, k)
		}
	}
	for _, s := range snapshots {
		if s.ID == "" {
			s.ID = newUUID()
		}
		if s.PublishedAt.IsZero() {
			s.PublishedAt = time.Now()
		}
		s.SchoolID = schoolID
		s.ResultReleaseID = releaseID
		m.snapshots[key(releaseID, s.StudentID, s.TermID, s.SubjectID)] = s
	}
	return nil
}

func (m *MemoryStore) SnapshotsForRelease(_ context.Context, schoolID, releaseID string) ([]Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Snapshot
	for _, s := range m.snapshots {
		if s.SchoolID == schoolID && s.ResultReleaseID == releaseID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *MemoryStore) SnapshotsForStudentTerm(_ context.Context, schoolID, studentID, termID string) ([]Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Snapshot
	for _, s := range m.snapshots {
		if s.SchoolID == schoolID && s.StudentID == studentID && s.TermID == termID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SubjectID < out[j].SubjectID })
	return out, nil
}

func (m *MemoryStore) SnapshotsForStudent(_ context.Context, schoolID, studentID string) ([]Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Snapshot
	for _, s := range m.snapshots {
		if s.SchoolID == schoolID && s.StudentID == studentID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *MemoryStore) UpsertReportComments(_ context.Context, schoolID string, comments []ReportComment) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range comments {
		if c.ID == "" {
			c.ID = newUUID()
		}
		c.SchoolID = schoolID
		c.UpdatedAt = time.Now()
		m.comments[key(schoolID, c.StudentID, c.TermID)] = c
	}
	return len(comments), nil
}

func (m *MemoryStore) ReportCommentForStudentTerm(_ context.Context, schoolID, studentID, termID string) (ReportComment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.comments[key(schoolID, studentID, termID)]
	if !ok {
		return ReportComment{}, ErrNotFound
	}
	return c, nil
}

func (m *MemoryStore) ReportCommentsForClassTerm(_ context.Context, schoolID, classID, termID string) ([]ReportComment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []ReportComment
	for _, c := range m.comments {
		if c.SchoolID == schoolID && c.ClassID == classID && c.TermID == termID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *MemoryStore) GuardianIDsForStudent(_ context.Context, schoolID, studentID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string(nil), m.guardians[key(schoolID, studentID)]...), nil
}

func (m *MemoryStore) GuardianIDsForClass(_ context.Context, schoolID, classID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := map[string]bool{}
	var out []string
	for k, cid := range m.enrolments {
		if cid != classID {
			continue
		}
		parts := strings.SplitN(k, "|", 2)
		if len(parts) != 2 || parts[0] != schoolID {
			continue
		}
		for _, uid := range m.guardians[key(schoolID, parts[1])] {
			if !seen[uid] {
				seen[uid] = true
				out = append(out, uid)
			}
		}
	}
	return out, nil
}

func (m *MemoryStore) GuardianIDsForSchool(_ context.Context, schoolID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := map[string]bool{}
	var out []string
	prefix := schoolID + "|"
	for k, users := range m.guardians {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		for _, uid := range users {
			if !seen[uid] {
				seen[uid] = true
				out = append(out, uid)
			}
		}
	}
	return out, nil
}

func (m *MemoryStore) StaffUserIDsForSchool(_ context.Context, schoolID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for _, mem := range m.memberships {
		if mem.SchoolID == schoolID && mem.Role != RoleParent {
			out = append(out, mem.UserID)
		}
	}
	return out, nil
}

func (m *MemoryStore) IsGuardian(_ context.Context, schoolID, userID, studentID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, uid := range m.guardians[key(schoolID, studentID)] {
		if uid == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) TeacherCanViewStudent(_ context.Context, schoolID, membershipID, studentID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.students[studentID]
	if !ok || s.SchoolID != schoolID {
		return false, nil
	}
	classID := m.enrolments[key(schoolID, studentID)]
	if classID == "" {
		classID = s.CurrentClassID
	}
	if m.classTeacher[classID] == membershipID {
		return true, nil
	}
	for _, a := range m.assignments {
		if a.SchoolID == schoolID && a.ClassID == classID && a.TeacherMembershipID == membershipID {
			return true, nil
		}
	}
	if mem, ok := m.memberships[membershipID]; ok && mem.SchoolID == schoolID {
		switch mem.Role {
		case "dean_of_studies", "principal", "deputy":
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) AppendAudit(_ context.Context, schoolID string, event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	event.SchoolID = schoolID
	m.audits = append(m.audits, event)
	return nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
