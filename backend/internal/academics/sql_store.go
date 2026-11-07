package academics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SQLStore is the PostgreSQL implementation of Store. It uses database/sql so
// the caller owns the driver (pgx stdlib or lib/pq) and go.mod stays under
// Evans's control.
type SQLStore struct {
	db *sql.DB
	tx *sql.Tx
}

// NewSQLStore wraps an existing *sql.DB.
func NewSQLStore(db *sql.DB) *SQLStore { return &SQLStore{db: db} }

type sqlQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *SQLStore) q() sqlQuerier {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// WithTx starts a transaction, sets the RLS tenant context and commits only on
// success (AC-27).
func (s *SQLStore) WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.db == nil {
		return errors.New("academics: SQLStore has no database")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
		_ = tx.Rollback()
		return err
	}
	tenant := &SQLStore{db: s.db, tx: tx}
	if err := fn(tenant); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *SQLStore) CurrentAcademicYearID(ctx context.Context, schoolID string) (string, error) {
	var id string
	err := s.q().QueryRowContext(ctx, `
		SELECT id FROM academic_years
		WHERE school_id = $1
		ORDER BY is_current DESC, year DESC
		LIMIT 1`, schoolID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}

func (s *SQLStore) AcademicYearExists(ctx context.Context, schoolID, academicYearID string) (bool, error) {
	var exists bool
	err := s.q().QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM academic_years WHERE id = $1 AND school_id = $2)`,
		academicYearID, schoolID).Scan(&exists)
	return exists, err
}

func (s *SQLStore) ClassByID(ctx context.Context, schoolID, classID string) (Class, error) {
	var c Class
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, grade_level_id, stream_name
		FROM classes WHERE id = $1 AND school_id = $2`, classID, schoolID).
		Scan(&c.ID, &c.SchoolID, &c.GradeLevelID, &c.StreamName)
	if errors.Is(err, sql.ErrNoRows) {
		return Class{}, ErrNotFound
	}
	return c, err
}

func (s *SQLStore) TermByID(ctx context.Context, schoolID, termID string) (Term, error) {
	var t Term
	err := s.q().QueryRowContext(ctx, `
		SELECT t.id, t.school_id, t.academic_year_id, ay.year, t.term_number,
		       t.start_date, t.end_date, t.is_current
		FROM terms t
		JOIN academic_years ay ON ay.id = t.academic_year_id AND ay.school_id = t.school_id
		WHERE t.id = $1 AND t.school_id = $2`, termID, schoolID).
		Scan(&t.ID, &t.SchoolID, &t.AcademicYearID, &t.Year, &t.TermNumber, &t.StartDate, &t.EndDate, &t.IsCurrent)
	if errors.Is(err, sql.ErrNoRows) {
		return Term{}, ErrNotFound
	}
	return t, err
}

func (s *SQLStore) TermForDate(ctx context.Context, schoolID string, date time.Time) (Term, error) {
	var t Term
	err := s.q().QueryRowContext(ctx, `
		SELECT t.id, t.school_id, t.academic_year_id, ay.year, t.term_number,
		       t.start_date, t.end_date, t.is_current
		FROM terms t
		JOIN academic_years ay ON ay.id = t.academic_year_id AND ay.school_id = t.school_id
		WHERE t.school_id = $1 AND $2::date BETWEEN t.start_date AND t.end_date
		ORDER BY t.start_date DESC
		LIMIT 1`, schoolID, dateOnly(date)).
		Scan(&t.ID, &t.SchoolID, &t.AcademicYearID, &t.Year, &t.TermNumber, &t.StartDate, &t.EndDate, &t.IsCurrent)
	if errors.Is(err, sql.ErrNoRows) {
		err = s.q().QueryRowContext(ctx, `
			SELECT t.id, t.school_id, t.academic_year_id, ay.year, t.term_number,
			       t.start_date, t.end_date, t.is_current
			FROM terms t
			JOIN academic_years ay ON ay.id = t.academic_year_id AND ay.school_id = t.school_id
			WHERE t.school_id = $1 AND t.is_current = TRUE
			ORDER BY t.start_date DESC LIMIT 1`, schoolID).
			Scan(&t.ID, &t.SchoolID, &t.AcademicYearID, &t.Year, &t.TermNumber, &t.StartDate, &t.EndDate, &t.IsCurrent)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return Term{}, ErrNotFound
	}
	return t, err
}

func (s *SQLStore) ListTerms(ctx context.Context, schoolID string) ([]Term, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT t.id, t.school_id, t.academic_year_id, ay.year, t.term_number,
		       t.start_date, t.end_date, t.is_current
		FROM terms t
		JOIN academic_years ay ON ay.id = t.academic_year_id AND ay.school_id = t.school_id
		WHERE t.school_id = $1
		ORDER BY ay.year, t.term_number`, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Term
	for rows.Next() {
		var t Term
		if err := rows.Scan(&t.ID, &t.SchoolID, &t.AcademicYearID, &t.Year, &t.TermNumber, &t.StartDate, &t.EndDate, &t.IsCurrent); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *SQLStore) StudentByID(ctx context.Context, schoolID, studentID string) (Student, error) {
	var st Student
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, admission_number, full_name, COALESCE(current_class_id::text, '')
		FROM students WHERE id = $1 AND school_id = $2`, studentID, schoolID).
		Scan(&st.ID, &st.SchoolID, &st.AdmissionNumber, &st.FullName, &st.CurrentClassID)
	if errors.Is(err, sql.ErrNoRows) {
		return Student{}, ErrNotFound
	}
	return st, err
}

func (s *SQLStore) SubjectsForClass(ctx context.Context, schoolID, classID string) ([]Subject, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT s.id, s.school_id, s.grade_level_id, s.code, s.name, s.is_compulsory
		FROM subjects s
		JOIN classes c ON c.school_id = s.school_id AND c.grade_level_id = s.grade_level_id
		WHERE c.id = $1 AND c.school_id = $2
		ORDER BY s.name`, classID, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subject
	for rows.Next() {
		var sub Subject
		if err := rows.Scan(&sub.ID, &sub.SchoolID, &sub.GradeLevelID, &sub.Code, &sub.Name, &sub.IsCompulsory); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *SQLStore) SubjectByID(ctx context.Context, schoolID, subjectID string) (Subject, error) {
	var sub Subject
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, grade_level_id, code, name, is_compulsory
		FROM subjects WHERE id = $1 AND school_id = $2`, subjectID, schoolID).
		Scan(&sub.ID, &sub.SchoolID, &sub.GradeLevelID, &sub.Code, &sub.Name, &sub.IsCompulsory)
	if errors.Is(err, sql.ErrNoRows) {
		return Subject{}, ErrNotFound
	}
	return sub, err
}

const activeEnrolmentYear = `(
	SELECT id FROM academic_years
	WHERE school_id = $1 AND is_current = TRUE
	ORDER BY year DESC LIMIT 1
)`

func (s *SQLStore) EnrolledStudents(ctx context.Context, schoolID, classID string) ([]EnrolledStudent, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT e.student_id, st.admission_number, st.full_name, e.class_id
		FROM enrolments e
		JOIN students st ON st.id = e.student_id AND st.school_id = e.school_id
		WHERE e.school_id = $1 AND e.class_id = $2 AND e.status = 'active'
		  AND e.academic_year_id = `+activeEnrolmentYear+`
		ORDER BY st.admission_number`, schoolID, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EnrolledStudent
	for rows.Next() {
		var st EnrolledStudent
		if err := rows.Scan(&st.StudentID, &st.AdmissionNumber, &st.FullName, &st.ClassID); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *SQLStore) MembershipByID(ctx context.Context, schoolID, membershipID string) (Membership, error) {
	var m Membership
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, user_id, role, status
		FROM memberships WHERE id = $1 AND school_id = $2`, membershipID, schoolID).
		Scan(&m.ID, &m.SchoolID, &m.UserID, &m.Role, &m.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Membership{}, ErrNotFound
	}
	return m, err
}

const activeBadgePredicate = `(mb.valid_from IS NULL OR mb.valid_from <= NOW())
	AND (mb.valid_until IS NULL OR mb.valid_until > NOW())`

func (s *SQLStore) ClassTeacherMembership(ctx context.Context, schoolID, classID string) (string, error) {
	var membershipID string
	err := s.q().QueryRowContext(ctx, `
		SELECT mb.membership_id
		FROM membership_badges mb
		JOIN badges b ON b.id = mb.badge_id AND b.school_id = $1
		JOIN memberships m ON m.id = mb.membership_id AND m.school_id = b.school_id
		WHERE b.school_id = $1 AND b.slug = 'class_teacher' AND mb.scope_id = $2
		  AND `+activeBadgePredicate+`
		ORDER BY mb.created_at DESC LIMIT 1`, schoolID, classID).Scan(&membershipID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return membershipID, err
}

func (s *SQLStore) HasClassTeacherBadge(ctx context.Context, schoolID, membershipID, classID string) (bool, error) {
	var exists bool
	err := s.q().QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM membership_badges mb
			JOIN badges b ON b.id = mb.badge_id AND b.school_id = $1
			WHERE b.school_id = $1 AND b.slug = 'class_teacher'
			  AND mb.scope_id = $2 AND mb.membership_id = $3
			  AND `+activeBadgePredicate+`
		)`, schoolID, classID, membershipID).Scan(&exists)
	return exists, err
}

func (s *SQLStore) AssignClassTeacher(ctx context.Context, schoolID, classID, membershipID, _ string) error {
	if _, err := s.q().ExecContext(ctx, `
		UPDATE membership_badges mb
		SET valid_until = NOW()
		FROM badges b
		WHERE mb.badge_id = b.id AND b.school_id = $1 AND b.slug = 'class_teacher'
		  AND mb.scope_id = $2 AND `+activeBadgePredicate, schoolID, classID); err != nil {
		return err
	}
	if _, err := s.q().ExecContext(ctx, `
		INSERT INTO badges (school_id, name, slug, scope_type, description)
		VALUES ($1, 'Class Teacher', 'class_teacher', 'class', 'Scoped to one class')
		ON CONFLICT (school_id, slug) DO NOTHING`, schoolID); err != nil {
		return err
	}
	_, err := s.q().ExecContext(ctx, `
		INSERT INTO membership_badges (membership_id, badge_id, scope_id, valid_from)
		SELECT $3, b.id, $2, NOW()
		FROM badges b WHERE b.school_id = $1 AND b.slug = 'class_teacher'`, schoolID, classID, membershipID)
	return err
}

func (s *SQLStore) HasAnySubjectAssignment(ctx context.Context, schoolID, membershipID, classID string) (bool, error) {
	var exists bool
	err := s.q().QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM subject_teacher_assignments
			WHERE school_id = $1 AND class_id = $2 AND teacher_membership_id = $3
		)`, schoolID, classID, membershipID).Scan(&exists)
	return exists, err
}

func (s *SQLStore) SubjectAssignment(ctx context.Context, schoolID, classID, subjectID, academicYearID string) (SubjectTeacherAssignment, error) {
	var a SubjectTeacherAssignment
	var assignedBy sql.NullString
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, academic_year_id, class_id, subject_id, teacher_membership_id, assigned_by, updated_at
		FROM subject_teacher_assignments
		WHERE school_id = $1 AND academic_year_id = $2 AND class_id = $3 AND subject_id = $4`,
		schoolID, academicYearID, classID, subjectID).
		Scan(&a.ID, &a.SchoolID, &a.AcademicYearID, &a.ClassID, &a.SubjectID, &a.TeacherMembershipID, &assignedBy, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return SubjectTeacherAssignment{}, ErrNotFound
	}
	a.AssignedBy = assignedBy.String
	return a, err
}

func (s *SQLStore) UpsertSubjectAssignment(ctx context.Context, schoolID string, a SubjectTeacherAssignment) error {
	_, err := s.q().ExecContext(ctx, `
		INSERT INTO subject_teacher_assignments
			(id, school_id, academic_year_id, class_id, subject_id, teacher_membership_id, assigned_by, updated_at)
		VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (school_id, academic_year_id, class_id, subject_id)
		DO UPDATE SET teacher_membership_id = EXCLUDED.teacher_membership_id,
		              assigned_by = EXCLUDED.assigned_by,
		              updated_at = NOW()`,
		a.ID, schoolID, a.AcademicYearID, a.ClassID, a.SubjectID, a.TeacherMembershipID, nullUUID(a.AssignedBy))
	return err
}

func (s *SQLStore) ListSubjectAssignments(ctx context.Context, schoolID, classID string) ([]SubjectTeacherAssignment, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT id, school_id, academic_year_id, class_id, subject_id, teacher_membership_id, assigned_by, updated_at
		FROM subject_teacher_assignments
		WHERE school_id = $1 AND class_id = $2
		ORDER BY subject_id`, schoolID, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SubjectTeacherAssignment
	for rows.Next() {
		var a SubjectTeacherAssignment
		var assignedBy sql.NullString
		if err := rows.Scan(&a.ID, &a.SchoolID, &a.AcademicYearID, &a.ClassID, &a.SubjectID, &a.TeacherMembershipID, &assignedBy, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.AssignedBy = assignedBy.String
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLStore) ListGradingScales(ctx context.Context, schoolID string) ([]GradingScale, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT id, school_id, version_number, ee_min, me_min, ae_min, be_min, is_active, created_by, created_at
		FROM grading_scale_versions WHERE school_id = $1
		ORDER BY version_number`, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GradingScale
	for rows.Next() {
		var g GradingScale
		var createdBy sql.NullString
		if err := rows.Scan(&g.ID, &g.SchoolID, &g.VersionNumber, &g.EEMin, &g.MEMin, &g.AEMin, &g.BEMin, &g.IsActive, &createdBy, &g.CreatedAt); err != nil {
			return nil, err
		}
		g.CreatedBy = createdBy.String
		out = append(out, g)
	}
	return out, rows.Err()
}

const gradingScaleColumns = `id, school_id, version_number, ee_min, me_min, ae_min, be_min, is_active, created_by, created_at`

func (s *SQLStore) ActiveGradingScale(ctx context.Context, schoolID string) (GradingScale, error) {
	var g GradingScale
	var createdBy sql.NullString
	err := s.q().QueryRowContext(ctx, `
		SELECT `+gradingScaleColumns+`
		FROM grading_scale_versions
		WHERE school_id = $1 AND is_active = TRUE
		ORDER BY version_number DESC LIMIT 1`, schoolID).
		Scan(&g.ID, &g.SchoolID, &g.VersionNumber, &g.EEMin, &g.MEMin, &g.AEMin, &g.BEMin, &g.IsActive, &createdBy, &g.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return GradingScale{}, ErrNotFound
	}
	g.CreatedBy = createdBy.String
	return g, err
}

func (s *SQLStore) GradingScaleByID(ctx context.Context, schoolID, id string) (GradingScale, error) {
	var g GradingScale
	var createdBy sql.NullString
	err := s.q().QueryRowContext(ctx, `
		SELECT `+gradingScaleColumns+`
		FROM grading_scale_versions WHERE id = $1 AND school_id = $2`, id, schoolID).
		Scan(&g.ID, &g.SchoolID, &g.VersionNumber, &g.EEMin, &g.MEMin, &g.AEMin, &g.BEMin, &g.IsActive, &createdBy, &g.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return GradingScale{}, ErrNotFound
	}
	g.CreatedBy = createdBy.String
	return g, err
}

func (s *SQLStore) CreateGradingScale(ctx context.Context, schoolID string, g GradingScale) (GradingScale, error) {
	if _, err := s.q().ExecContext(ctx, `
		UPDATE grading_scale_versions SET is_active = FALSE
		WHERE school_id = $1 AND is_active = TRUE`, schoolID); err != nil {
		return GradingScale{}, err
	}
	err := s.q().QueryRowContext(ctx, `
		INSERT INTO grading_scale_versions
			(id, school_id, version_number, ee_min, me_min, ae_min, be_min, is_active, created_by)
		VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, TRUE, $8)
		RETURNING id, created_at`,
		g.ID, schoolID, g.VersionNumber, g.EEMin, g.MEMin, g.AEMin, g.BEMin, nullUUID(g.CreatedBy)).
		Scan(&g.ID, &g.CreatedAt)
	if err != nil {
		return GradingScale{}, err
	}
	g.SchoolID = schoolID
	g.IsActive = true
	return g, nil
}

func (s *SQLStore) ListTimetableSlots(ctx context.Context, schoolID, termID, classID string) ([]TimetableSlot, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT ts.id, ts.school_id, ts.academic_year_id, ts.term_id, ts.class_id, ts.subject_id,
		       COALESCE(sub.name, ''), ts.teacher_membership_id, ts.day_of_week,
		       to_char(ts.start_time, 'HH24:MI'), to_char(ts.end_time, 'HH24:MI'), COALESCE(ts.room, '')
		FROM timetable_slots ts
		LEFT JOIN subjects sub ON sub.id = ts.subject_id AND sub.school_id = ts.school_id
		WHERE ts.school_id = $1 AND ts.term_id = $2
		  AND ($3 = '' OR ts.class_id = NULLIF($3, '')::uuid)
		ORDER BY ts.day_of_week, ts.start_time`, schoolID, termID, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TimetableSlot
	for rows.Next() {
		var slot TimetableSlot
		if err := rows.Scan(&slot.ID, &slot.SchoolID, &slot.AcademicYearID, &slot.TermID, &slot.ClassID,
			&slot.SubjectID, &slot.SubjectName, &slot.TeacherMembershipID, &slot.DayOfWeek,
			&slot.StartTime, &slot.EndTime, &slot.Room); err != nil {
			return nil, err
		}
		out = append(out, slot)
	}
	return out, rows.Err()
}

func (s *SQLStore) ReplaceTimetableSlots(ctx context.Context, schoolID, termID, classID string, slots []TimetableSlot) error {
	if _, err := s.q().ExecContext(ctx, `
		DELETE FROM timetable_slots WHERE school_id = $1 AND term_id = $2 AND class_id = $3`,
		schoolID, termID, classID); err != nil {
		return err
	}
	for _, slot := range slots {
		if _, err := s.q().ExecContext(ctx, `
			INSERT INTO timetable_slots
				(id, school_id, academic_year_id, term_id, class_id, subject_id, teacher_membership_id,
				 day_of_week, start_time, end_time, room)
			VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9::time, $10::time, $11)`,
			slot.ID, schoolID, slot.AcademicYearID, termID, classID, slot.SubjectID,
			slot.TeacherMembershipID, slot.DayOfWeek, slot.StartTime, slot.EndTime, slot.Room); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) ListCalendarEvents(ctx context.Context, schoolID, termID string) ([]CalendarEvent, error) {
	rows, err := s.q().QueryContext(ctx, `
		SELECT id, school_id, academic_year_id, term_id, title, COALESCE(description, ''),
		       COALESCE(event_type, ''), start_at, end_at, created_by, created_at
		FROM calendar_events
		WHERE school_id = $1 AND ($2 = '' OR term_id = NULLIF($2, '')::uuid)
		ORDER BY start_at`, schoolID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		var academicYearID, termID, createdBy sql.NullString
		if err := rows.Scan(&e.ID, &e.SchoolID, &academicYearID, &termID, &e.Title, &e.Description,
			&e.EventType, &e.StartAt, &e.EndAt, &createdBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.AcademicYearID = academicYearID.String
		e.TermID = termID.String
		e.CreatedBy = createdBy.String
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *SQLStore) UpsertCalendarEvents(ctx context.Context, schoolID string, events []CalendarEvent) error {
	for _, e := range events {
		if e.ID == "" {
			if _, err := s.q().ExecContext(ctx, `
				INSERT INTO calendar_events
					(id, school_id, academic_year_id, term_id, title, description, event_type, start_at, end_at, created_by)
				VALUES (gen_random_uuid(), $1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, NULLIF($9, '')::uuid)`,
				schoolID, e.AcademicYearID, e.TermID, e.Title, e.Description, e.EventType, e.StartAt, e.EndAt, e.CreatedBy); err != nil {
				return err
			}
			continue
		}
		if _, err := s.q().ExecContext(ctx, `
			UPDATE calendar_events
			SET academic_year_id = NULLIF($3, '')::uuid, term_id = NULLIF($4, '')::uuid,
			    title = $5, description = $6, event_type = $7, start_at = $8, end_at = $9
			WHERE id = $1 AND school_id = $2`,
			e.ID, schoolID, e.AcademicYearID, e.TermID, e.Title, e.Description, e.EventType, e.StartAt, e.EndAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) UpdateTermDates(ctx context.Context, schoolID, termID string, start, end time.Time) error {
	res, err := s.q().ExecContext(ctx, `
		UPDATE terms SET start_date = $3, end_date = $4
		WHERE id = $1 AND school_id = $2`, termID, schoolID, dateOnly(start), dateOnly(end))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLStore) UpsertAttendance(ctx context.Context, schoolID string, rows []AttendanceRow) (int, error) {
	count := 0
	for _, r := range rows {
		source := r.Source
		if source == "" {
			source = "manual_roll_call"
		}
		if _, err := s.q().ExecContext(ctx, `
			INSERT INTO attendance_daily
				(id, school_id, student_id, class_id, term_id, attendance_date, status, source, recorded_by, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (school_id, student_id, attendance_date)
			DO UPDATE SET class_id = EXCLUDED.class_id, term_id = EXCLUDED.term_id,
			              status = EXCLUDED.status, source = EXCLUDED.source,
			              recorded_by = EXCLUDED.recorded_by, updated_at = NOW()`,
			schoolID, r.StudentID, r.ClassID, r.TermID, dateOnly(r.AttendanceDate), r.Status, source, r.RecordedBy); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *SQLStore) AttendanceForStudent(ctx context.Context, schoolID, studentID string) (AttendanceSummary, error) {
	var summary AttendanceSummary
	err := s.q().QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE status = 'present'),
		       COUNT(*) FILTER (WHERE status = 'absent'),
		       COUNT(*) FILTER (WHERE status = 'late')
		FROM attendance_daily WHERE school_id = $1 AND student_id = $2`, schoolID, studentID).
		Scan(&summary.TotalDays, &summary.Present, &summary.Absent, &summary.Late)
	return summary, err
}

func (s *SQLStore) UpsertMarks(ctx context.Context, schoolID string, rows []Mark, enteredBy string) (int, error) {
	count := 0
	for _, r := range rows {
		if _, err := s.q().ExecContext(ctx, `
			INSERT INTO student_marks
				(id, school_id, student_id, class_id, term_id, subject_id, score, entered_by, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (school_id, student_id, term_id, subject_id)
			DO UPDATE SET class_id = EXCLUDED.class_id, score = EXCLUDED.score,
			              entered_by = EXCLUDED.entered_by, updated_at = NOW()`,
			schoolID, r.StudentID, r.ClassID, r.TermID, r.SubjectID, r.Score, enteredBy); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *SQLStore) MarksForClassTerm(ctx context.Context, schoolID, classID, termID string) ([]Mark, error) {
	return s.queryMarks(ctx, `
		SELECT school_id, student_id, class_id, term_id, subject_id, score, entered_by, updated_at
		FROM student_marks
		WHERE school_id = $1 AND class_id = $2 AND term_id = $3
		ORDER BY student_id, subject_id`, schoolID, classID, termID)
}

func (s *SQLStore) MarksForStudentTerm(ctx context.Context, schoolID, studentID, termID string) ([]Mark, error) {
	return s.queryMarks(ctx, `
		SELECT school_id, student_id, class_id, term_id, subject_id, score, entered_by, updated_at
		FROM student_marks
		WHERE school_id = $1 AND student_id = $2 AND term_id = $3
		ORDER BY subject_id`, schoolID, studentID, termID)
}

func (s *SQLStore) MarksForStudentAllTerms(ctx context.Context, schoolID, studentID string) ([]Mark, error) {
	return s.queryMarks(ctx, `
		SELECT school_id, student_id, class_id, term_id, subject_id, score, entered_by, updated_at
		FROM student_marks WHERE school_id = $1 AND student_id = $2
		ORDER BY term_id, subject_id`, schoolID, studentID)
}

func (s *SQLStore) queryMarks(ctx context.Context, query string, args ...any) ([]Mark, error) {
	rows, err := s.q().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Mark
	for rows.Next() {
		var m Mark
		if err := rows.Scan(&m.SchoolID, &m.StudentID, &m.ClassID, &m.TermID, &m.SubjectID, &m.Score, &m.EnteredBy, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *SQLStore) ResultRelease(ctx context.Context, schoolID, classID, termID string) (ResultRelease, error) {
	var r ResultRelease
	var scaleID, releasedBy, reopenedBy, rereleasedBy sql.NullString
	var releasedAt, reopenedAt, rereleasedAt sql.NullTime
	err := s.q().QueryRowContext(ctx, `
		SELECT id, school_id, class_id, term_id, status, grading_scale_version_id,
		       released_by, released_at, reopened_by, reopened_at, rereleased_by, rereleased_at
		FROM result_releases WHERE school_id = $1 AND class_id = $2 AND term_id = $3`,
		schoolID, classID, termID).
		Scan(&r.ID, &r.SchoolID, &r.ClassID, &r.TermID, &r.Status, &scaleID,
			&releasedBy, &releasedAt, &reopenedBy, &reopenedAt, &rereleasedBy, &rereleasedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ResultRelease{}, ErrNotFound
	}
	if err != nil {
		return ResultRelease{}, err
	}
	r.GradingScaleVersionID = scaleID.String
	r.ReleasedBy = releasedBy.String
	r.ReopenedBy = reopenedBy.String
	r.RereleasedBy = rereleasedBy.String
	if releasedAt.Valid {
		r.ReleasedAt = &releasedAt.Time
	}
	if reopenedAt.Valid {
		r.ReopenedAt = &reopenedAt.Time
	}
	if rereleasedAt.Valid {
		r.RereleasedAt = &rereleasedAt.Time
	}
	return r, nil
}

func (s *SQLStore) SaveResultRelease(ctx context.Context, schoolID string, r ResultRelease) error {
	_, err := s.q().ExecContext(ctx, `
		INSERT INTO result_releases
			(id, school_id, class_id, term_id, status, grading_scale_version_id,
			 released_by, released_at, reopened_by, reopened_at, rereleased_by, rereleased_at)
		VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, NULLIF($6, '')::uuid,
		        NULLIF($7, '')::uuid, $8, NULLIF($9, '')::uuid, $10, NULLIF($11, '')::uuid, $12)
		ON CONFLICT (school_id, class_id, term_id)
		DO UPDATE SET status = EXCLUDED.status,
		              grading_scale_version_id = EXCLUDED.grading_scale_version_id,
		              released_by = EXCLUDED.released_by, released_at = EXCLUDED.released_at,
		              reopened_by = EXCLUDED.reopened_by, reopened_at = EXCLUDED.reopened_at,
		              rereleased_by = EXCLUDED.rereleased_by, rereleased_at = EXCLUDED.rereleased_at`,
		r.ID, schoolID, r.ClassID, r.TermID, r.Status, r.GradingScaleVersionID,
		r.ReleasedBy, nullableTime(r.ReleasedAt), r.ReopenedBy, nullableTime(r.ReopenedAt),
		r.RereleasedBy, nullableTime(r.RereleasedAt))
	return err
}

func (s *SQLStore) ReplaceSnapshots(ctx context.Context, schoolID, releaseID string, snapshots []Snapshot) error {
	if _, err := s.q().ExecContext(ctx, `
		DELETE FROM released_marks_snapshots WHERE school_id = $1 AND result_release_id = $2`,
		schoolID, releaseID); err != nil {
		return err
	}
	for _, snapshot := range snapshots {
		if _, err := s.q().ExecContext(ctx, `
			INSERT INTO released_marks_snapshots
				(id, school_id, result_release_id, student_id, term_id, subject_id, score, cbc_level,
				 is_amended, amendment_note, snapshot_version, report_comment, published_at)
			VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8,
			        $9, NULLIF($10, ''), $11, NULLIF($12, ''), $13)`,
			snapshot.ID, schoolID, releaseID, snapshot.StudentID, snapshot.TermID, snapshot.SubjectID,
			snapshot.Score, snapshot.CBCLevel, snapshot.IsAmended, snapshot.AmendmentNote,
			snapshot.SnapshotVersion, snapshot.ReportComment, snapshot.PublishedAt); err != nil {
			return err
		}
	}
	return nil
}

const snapshotColumns = `id, school_id, result_release_id, student_id, term_id, subject_id, score, cbc_level,
	is_amended, COALESCE(amendment_note, ''), snapshot_version, COALESCE(report_comment, ''), published_at`

func (s *SQLStore) querySnapshots(ctx context.Context, query string, args ...any) ([]Snapshot, error) {
	rows, err := s.q().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Snapshot
	for rows.Next() {
		var snap Snapshot
		if err := rows.Scan(&snap.ID, &snap.SchoolID, &snap.ResultReleaseID, &snap.StudentID, &snap.TermID,
			&snap.SubjectID, &snap.Score, &snap.CBCLevel, &snap.IsAmended, &snap.AmendmentNote,
			&snap.SnapshotVersion, &snap.ReportComment, &snap.PublishedAt); err != nil {
			return nil, err
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

func (s *SQLStore) SnapshotsForRelease(ctx context.Context, schoolID, releaseID string) ([]Snapshot, error) {
	return s.querySnapshots(ctx, `SELECT `+snapshotColumns+` FROM released_marks_snapshots
		WHERE school_id = $1 AND result_release_id = $2 ORDER BY student_id, subject_id`, schoolID, releaseID)
}

func (s *SQLStore) SnapshotsForStudentTerm(ctx context.Context, schoolID, studentID, termID string) ([]Snapshot, error) {
	return s.querySnapshots(ctx, `SELECT `+snapshotColumns+` FROM released_marks_snapshots
		WHERE school_id = $1 AND student_id = $2 AND term_id = $3 ORDER BY subject_id`, schoolID, studentID, termID)
}

func (s *SQLStore) SnapshotsForStudent(ctx context.Context, schoolID, studentID string) ([]Snapshot, error) {
	return s.querySnapshots(ctx, `SELECT `+snapshotColumns+` FROM released_marks_snapshots
		WHERE school_id = $1 AND student_id = $2 ORDER BY term_id, subject_id`, schoolID, studentID)
}

func (s *SQLStore) UpsertReportComments(ctx context.Context, schoolID string, comments []ReportComment) (int, error) {
	count := 0
	for _, c := range comments {
		if _, err := s.q().ExecContext(ctx, `
			INSERT INTO report_comments
				(id, school_id, student_id, class_id, term_id, comment_text, ai_draft_used, approved_by, updated_at)
			VALUES (COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (school_id, student_id, term_id)
			DO UPDATE SET class_id = EXCLUDED.class_id, comment_text = EXCLUDED.comment_text,
			              ai_draft_used = EXCLUDED.ai_draft_used, approved_by = EXCLUDED.approved_by, updated_at = NOW()`,
			c.ID, schoolID, c.StudentID, c.ClassID, c.TermID, c.CommentText, c.AIDraftUsed, c.ApprovedBy); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

const reportCommentColumns = `id, school_id, student_id, class_id, term_id, comment_text, ai_draft_used, approved_by, updated_at`

func scanReportComment(scanner interface{ Scan(dest ...any) error }) (ReportComment, error) {
	var c ReportComment
	err := scanner.Scan(&c.ID, &c.SchoolID, &c.StudentID, &c.ClassID, &c.TermID, &c.CommentText, &c.AIDraftUsed, &c.ApprovedBy, &c.UpdatedAt)
	return c, err
}

func (s *SQLStore) ReportCommentForStudentTerm(ctx context.Context, schoolID, studentID, termID string) (ReportComment, error) {
	c, err := scanReportComment(s.q().QueryRowContext(ctx, `SELECT `+reportCommentColumns+`
		FROM report_comments WHERE school_id = $1 AND student_id = $2 AND term_id = $3`, schoolID, studentID, termID))
	if errors.Is(err, sql.ErrNoRows) {
		return ReportComment{}, ErrNotFound
	}
	return c, err
}

func (s *SQLStore) ReportCommentsForClassTerm(ctx context.Context, schoolID, classID, termID string) ([]ReportComment, error) {
	rows, err := s.q().QueryContext(ctx, `SELECT `+reportCommentColumns+`
		FROM report_comments WHERE school_id = $1 AND class_id = $2 AND term_id = $3`, schoolID, classID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReportComment
	for rows.Next() {
		c, err := scanReportComment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *SQLStore) GuardianIDsForStudent(ctx context.Context, schoolID, studentID string) ([]string, error) {
	return s.queryStrings(ctx, `SELECT user_id::text FROM student_guardians
		WHERE school_id = $1 AND student_id = $2 ORDER BY created_at`, schoolID, studentID)
}

func (s *SQLStore) GuardianIDsForClass(ctx context.Context, schoolID, classID string) ([]string, error) {
	return s.queryStrings(ctx, `
		SELECT DISTINCT sg.user_id::text
		FROM student_guardians sg
		JOIN enrolments e ON e.student_id = sg.student_id AND e.school_id = sg.school_id
		WHERE e.school_id = $1 AND e.class_id = $2 AND e.status = 'active'
		  AND e.academic_year_id = `+activeEnrolmentYear+`
		ORDER BY sg.user_id::text`, schoolID, classID)
}

func (s *SQLStore) GuardianIDsForSchool(ctx context.Context, schoolID string) ([]string, error) {
	return s.queryStrings(ctx, `SELECT DISTINCT user_id::text FROM student_guardians
		WHERE school_id = $1 ORDER BY user_id::text`, schoolID)
}

func (s *SQLStore) StaffUserIDsForSchool(ctx context.Context, schoolID string) ([]string, error) {
	return s.queryStrings(ctx, `SELECT DISTINCT user_id::text FROM memberships
		WHERE school_id = $1 AND status = 'active' AND role <> 'parent' ORDER BY user_id::text`, schoolID)
}

func (s *SQLStore) IsGuardian(ctx context.Context, schoolID, userID, studentID string) (bool, error) {
	var exists bool
	err := s.q().QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM student_guardians
		WHERE school_id = $1 AND user_id = $2 AND student_id = $3)`, schoolID, userID, studentID).Scan(&exists)
	return exists, err
}

func (s *SQLStore) TeacherCanViewStudent(ctx context.Context, schoolID, membershipID, studentID string) (bool, error) {
	var allowed bool
	err := s.q().QueryRowContext(ctx, `
		WITH sc AS (
			SELECT COALESCE(s.current_class_id, (
				SELECT e.class_id FROM enrolments e
				WHERE e.school_id = s.school_id AND e.student_id = s.id AND e.status = 'active'
				ORDER BY e.created_at DESC LIMIT 1
			)) AS class_id
			FROM students s WHERE s.id = $2 AND s.school_id = $1
		)
		SELECT
			EXISTS (SELECT 1 FROM memberships m
			        WHERE m.id = $3 AND m.school_id = $1
			          AND m.role IN ('dean_of_studies','principal','deputy'))
			OR EXISTS (
				SELECT 1 FROM sc
				WHERE sc.class_id IS NOT NULL AND (
					EXISTS (SELECT 1 FROM membership_badges mb
					        JOIN badges b ON b.id = mb.badge_id AND b.school_id = $1
					        WHERE b.school_id = $1 AND b.slug = 'class_teacher'
					          AND mb.scope_id = sc.class_id AND mb.membership_id = $3
					          AND `+activeBadgePredicate+`)
					OR EXISTS (SELECT 1 FROM subject_teacher_assignments a
					           WHERE a.school_id = $1 AND a.class_id = sc.class_id
					             AND a.teacher_membership_id = $3)
				)
			)`, schoolID, studentID, membershipID).Scan(&allowed)
	return allowed, err
}

func (s *SQLStore) AppendAudit(ctx context.Context, schoolID string, event AuditEvent) error {
	details, _ := json.Marshal(event.Details)
	if len(details) == 0 || string(details) == "null" {
		details = []byte("{}")
	}
	_, err := s.q().ExecContext(ctx, `
		INSERT INTO audit_log (school_id, user_id, action, resource_type, resource_id, details)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, ''), $6::jsonb)`,
		schoolID, event.UserID, event.Action, event.ResourceType, event.ResourceID, string(details))
	return err
}

func (s *SQLStore) queryStrings(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := s.q().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

func nullUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

var _ Store = (*SQLStore)(nil)
var _ Store = (*MemoryStore)(nil)

func (s *SQLStore) String() string { return fmt.Sprintf("academics.SQLStore(tx=%v)", s.tx != nil) }
