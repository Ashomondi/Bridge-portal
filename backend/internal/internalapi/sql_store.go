package internalapi

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SQLStore is the PostgreSQL implementation of internalapi.Store.
type SQLStore struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

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

func (s *SQLStore) WithTx(ctx context.Context, schoolID string, fn func(tx Store) error) error {
	if s.db == nil {
		return errors.New("internalapi: SQLStore has no database connection")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if schoolID != "" {
		if _, err := tx.ExecContext(ctx, `SELECT set_config('app.school_id', $1, true)`, schoolID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	tenantStore := &SQLStore{db: s.db, tx: tx}
	if err := fn(tenantStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (s *SQLStore) GetSchoolProfile(ctx context.Context, schoolID string) (*SchoolProfile, error) {
	query := `
		SELECT id, code, name, COALESCE(county, ''), COALESCE(logo_url, '')
		FROM schools
		WHERE id = $1
		LIMIT 1;
	`
	var p SchoolProfile
	err := s.q().QueryRowContext(ctx, query, schoolID).Scan(&p.ID, &p.Code, &p.Name, &p.County, &p.LogoURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("school not found")
	}
	return &p, err
}

func (s *SQLStore) ListChildrenForParent(ctx context.Context, schoolID, parentUserID string) ([]LinkedChild, error) {
	query := `
		SELECT s.id, s.admission_number, s.full_name, COALESCE(c.stream_name, ''), sg.relationship
		FROM student_guardians sg
		JOIN students s ON s.id = sg.student_id
		LEFT JOIN classes c ON c.id = s.current_class_id
		WHERE sg.school_id = $1 AND sg.user_id = $2
		ORDER BY s.admission_number ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, parentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []LinkedChild
	for rows.Next() {
		var ch LinkedChild
		if err := rows.Scan(&ch.StudentID, &ch.AdmissionNumber, &ch.FullName, &ch.ClassName, &ch.Relationship); err != nil {
			return nil, err
		}
		children = append(children, ch)
	}
	return children, rows.Err()
}

func (s *SQLStore) IsGuardianOf(ctx context.Context, schoolID, studentID, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM student_guardians
			WHERE school_id = $1 AND student_id = $2 AND user_id = $3
		);
	`
	var exists bool
	err := s.q().QueryRowContext(ctx, query, schoolID, studentID, userID).Scan(&exists)
	return exists, err
}

func (s *SQLStore) GetAttendanceSummary(ctx context.Context, schoolID, studentID, termID string) (*AttendanceSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_days,
			COUNT(*) FILTER (WHERE status = 'present' OR status = 'late') as days_present,
			COUNT(*) FILTER (WHERE status = 'absent') as days_absent
		FROM attendance_daily
		WHERE school_id = $1 AND student_id = $2 AND term_id = $3;
	`
	var total, present, absent int
	err := s.q().QueryRowContext(ctx, query, schoolID, studentID, termID).Scan(&total, &present, &absent)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("attendance summary not found")
	}
	rate := float64(present) / float64(total) * 100.0

	return &AttendanceSummary{
		StudentID:      studentID,
		TermID:         termID,
		TotalDays:      total,
		DaysPresent:    present,
		DaysAbsent:     absent,
		AttendanceRate: rate,
	}, nil
}

func (s *SQLStore) GetMarksSummary(ctx context.Context, schoolID, studentID, termID string) (*StudentMarksSummary, error) {
	query := `
		SELECT s.code, s.name, sm.score, sm.updated_at
		FROM student_marks sm
		JOIN subjects s ON s.id = sm.subject_id
		WHERE sm.school_id = $1 AND sm.student_id = $2 AND sm.term_id = $3
		ORDER BY s.code ASC;
	`
	rows, err := s.q().QueryContext(ctx, query, schoolID, studentID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []SubjectMarkSummary
	var latestPub time.Time
	var totalScore float64

	for rows.Next() {
		var code, name string
		var score float64
		var updatedAt time.Time
		if err := rows.Scan(&code, &name, &score, &updatedAt); err != nil {
			return nil, err
		}
		if updatedAt.After(latestPub) {
			latestPub = updatedAt
		}
		totalScore += score

		level := "BE"
		switch {
		case score >= 80:
			level = "EE"
		case score >= 50:
			level = "ME"
		case score >= 30:
			level = "AE"
		}

		subjects = append(subjects, SubjectMarkSummary{
			SubjectCode: code,
			SubjectName: name,
			Score:       score,
			CBCLevel:    level,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		return nil, errors.New("marks summary not found")
	}

	meanGrade := ""
	avg := totalScore / float64(len(subjects))
	switch {
	case avg >= 80:
		meanGrade = "EE"
	case avg >= 50:
		meanGrade = "ME"
	case avg >= 30:
		meanGrade = "AE"
	default:
		meanGrade = "BE"
	}

	return &StudentMarksSummary{
		StudentID:   studentID,
		TermID:      termID,
		Subjects:    subjects,
		MeanGrade:   meanGrade,
		PublishedAt: latestPub,
	}, nil
}

func (s *SQLStore) SaveAIAuditLog(ctx context.Context, entry *AIAuditLogEntry) error {
	if entry.ID == "" {
		entry.ID = newUUID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	var endpointsJSON, metaJSON []byte
	var err error
	if entry.EndpointsRead != nil {
		endpointsJSON, err = json.Marshal(entry.EndpointsRead)
		if err != nil {
			endpointsJSON = []byte("[]")
		}
	}
	if entry.Metadata != nil {
		metaJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			metaJSON = []byte("{}")
		}
	}

	query := `
		INSERT INTO ai_action_logs (
			id, school_id, user_id, session_id, channel, feature, intent,
			endpoints_read, action_taken, metadata, created_at
		) VALUES (
			$1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7,
			$8, $9, $10, $11
		);
	`
	_, err = s.q().ExecContext(ctx, query,
		entry.ID, entry.SchoolID, entry.UserID, entry.SessionID,
		entry.Channel, entry.Feature, entry.Intent,
		endpointsJSON, entry.ActionTaken, metaJSON, entry.CreatedAt,
	)
	return err
}
