package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/academics"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type syncRowScanner interface {
	Scan(dest ...any) error
}

type syncRows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

type syncTx interface {
	exec(ctx context.Context, query string, args ...any) error
	queryRow(ctx context.Context, query string, args ...any) syncRowScanner
	query(ctx context.Context, query string, args ...any) (syncRows, error)
	notificationTx() contracts.SQLTx
	logAudit(ctx context.Context, s *Service, event contracts.AuditEvent) error
	isNoRows(err error) bool
}

type pgxSyncTx struct{ tx pgx.Tx }

func (p pgxSyncTx) exec(ctx context.Context, query string, args ...any) error {
	_, err := p.tx.Exec(ctx, query, args...)
	return err
}

func (p pgxSyncTx) queryRow(ctx context.Context, query string, args ...any) syncRowScanner {
	return p.tx.QueryRow(ctx, query, args...)
}

func (p pgxSyncTx) query(ctx context.Context, query string, args ...any) (syncRows, error) {
	rows, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgxSyncRows{rows: rows}, nil
}

func (p pgxSyncTx) notificationTx() contracts.SQLTx {
	return contracts.PgxTx(p.tx)
}

func (p pgxSyncTx) logAudit(ctx context.Context, s *Service, event contracts.AuditEvent) error {
	return s.logAuditPgx(ctx, p.tx, event)
}

func (p pgxSyncTx) isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

type pgxSyncRows struct{ rows pgx.Rows }

func (r pgxSyncRows) Next() bool             { return r.rows.Next() }
func (r pgxSyncRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r pgxSyncRows) Close()                 { r.rows.Close() }
func (r pgxSyncRows) Err() error             { return r.rows.Err() }

type stdSyncTx struct{ tx *sql.Tx }

func (s stdSyncTx) exec(ctx context.Context, query string, args ...any) error {
	_, err := s.tx.ExecContext(ctx, query, args...)
	return err
}

func (s stdSyncTx) queryRow(ctx context.Context, query string, args ...any) syncRowScanner {
	return s.tx.QueryRowContext(ctx, query, args...)
}

func (s stdSyncTx) query(ctx context.Context, query string, args ...any) (syncRows, error) {
	rows, err := s.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return stdSyncRows{rows: rows}, nil
}

func (s stdSyncTx) notificationTx() contracts.SQLTx {
	return contracts.StdSQLTx(s.tx)
}

func (s stdSyncTx) logAudit(ctx context.Context, svc *Service, event contracts.AuditEvent) error {
	return svc.logAuditSQL(ctx, s.tx, event)
}

func (s stdSyncTx) isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

type stdSyncRows struct{ rows *sql.Rows }

func (r stdSyncRows) Next() bool             { return r.rows.Next() }
func (r stdSyncRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r stdSyncRows) Close()                 { _ = r.rows.Close() }
func (r stdSyncRows) Err() error             { return r.rows.Err() }

// OnTermStartDateChanged implements academics.TermDateFeeSyncer. When invoked
// from academics.PutCalendar, it executes inside the caller's active database
// transaction attached via academics.WithSQLTx so term dates, fee due dates,
// change notes, notifications, and audit records commit or roll back together.
func (s *Service) OnTermStartDateChanged(ctx context.Context, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error {
	sqlTx, ok := academics.SQLTxFromContext(ctx)
	if !ok {
		return contracts.Fail(http.StatusInternalServerError, "TERM_SYNC_TX_REQUIRED", "An active database transaction is required to synchronize term fee due dates.")
	}
	return s.syncTermStartDate(ctx, stdSyncTx{tx: sqlTx}, schoolID, termID, newStartDate, actorUserID, note)
}

// OnTermStartDateChangedTx synchronizes fee_periods, fee_charges, fee_ledgers,
// fee_change_notes, guardian notifications, and audit logs within an open pgx.Tx.
func (s *Service) OnTermStartDateChangedTx(ctx context.Context, tx pgx.Tx, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error {
	if tx == nil {
		return contracts.Fail(http.StatusInternalServerError, "TERM_SYNC_TX_REQUIRED", "An active database transaction is required to synchronize term fee due dates.")
	}
	return s.syncTermStartDate(ctx, pgxSyncTx{tx: tx}, schoolID, termID, newStartDate, actorUserID, note)
}

func (s *Service) syncTermStartDate(ctx context.Context, tx syncTx, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error {
	if !contracts.ValidID(schoolID) {
		return validationFail("school_id", "INVALID_SCHOOL", "A valid school_id is required.")
	}
	if !contracts.ValidID(termID) {
		return validationFail("term_id", "INVALID_TERM", "A valid term_id is required.")
	}
	if !contracts.ValidID(actorUserID) {
		return validationFail("actor_user_id", "INVALID_ACTOR", "A valid actor_user_id is required.")
	}
	if newStartDate.IsZero() {
		return validationFail("start_date", "INVALID_START_DATE", "A valid new term start date is required.")
	}
	trimmedNote := strings.TrimSpace(note)
	if trimmedNote == "" {
		return validationFail("fee_change_note", "TERM_DATE_CHANGE_NOTE_REQUIRED", "A fee change note is required when changing a term start date.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return err
	}

	var termExists bool
	if err := tx.queryRow(ctx, `SELECT EXISTS(SELECT 1 FROM terms WHERE school_id = $1 AND id = $2)`, schoolID, termID).Scan(&termExists); err != nil {
		return err
	}
	if !termExists {
		return validationFail("term_id", "INVALID_TERM", "Term does not exist in this school.")
	}

	newDateStr := newStartDate.UTC().Format("2006-01-02")

	var feePeriodID sql.NullString
	var oldDueDate sql.NullTime
	var periodID string
	var dueDate time.Time
	err := tx.queryRow(ctx, `
		SELECT id, due_date
		FROM fee_periods
		WHERE school_id = $1 AND term_id = $2
		FOR UPDATE`,
		schoolID, termID,
	).Scan(&periodID, &dueDate)
	if err != nil && !tx.isNoRows(err) {
		return err
	}

	var affectedStudents []string
	var oldDateStr sql.NullString
	if err == nil {
		feePeriodID = sql.NullString{String: periodID, Valid: true}
		oldDueDate = sql.NullTime{Time: dueDate, Valid: true}
		oldDateStr = sql.NullString{String: dueDate.Format("2006-01-02"), Valid: true}

		if err := tx.exec(ctx, `
			UPDATE fee_periods
			SET due_date = $1::date
			WHERE school_id = $2 AND id = $3`,
			newDateStr, schoolID, periodID,
		); err != nil {
			return err
		}

		rows, err := tx.query(ctx, `
			UPDATE fee_charges
			SET due_date = $1::date
			WHERE school_id = $2 AND fee_period_id = $3
			RETURNING student_id`,
			newDateStr, schoolID, periodID,
		)
		if err != nil {
			return err
		}
		seen := make(map[string]struct{})
		for rows.Next() {
			var sid string
			if err := rows.Scan(&sid); err != nil {
				rows.Close()
				return err
			}
			if _, ok := seen[sid]; !ok {
				seen[sid] = struct{}{}
				affectedStudents = append(affectedStudents, sid)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		sort.Strings(affectedStudents)
	}

	if err := tx.exec(ctx, `
		UPDATE fee_ledgers
		SET due_date = $1::date, updated_at = NOW()
		WHERE school_id = $2 AND term_id = $3`,
		newDateStr, schoolID, termID,
	); err != nil {
		return err
	}

	var noteID string
	if err := tx.queryRow(ctx, `
		INSERT INTO fee_change_notes (
			school_id, term_id, fee_period_id, change_type, amount_delta,
			old_due_date, new_due_date, note, created_by, affected_students_count
		)
		VALUES ($1, $2, NULLIF($3, '')::uuid, 'due_date', 0.00, NULLIF($4, '')::date, $5::date, $6, $7, $8)
		RETURNING id`,
		schoolID, termID, feePeriodID.String, oldDateStr.String, newDateStr, trimmedNote, actorUserID, len(affectedStudents),
	).Scan(&noteID); err != nil {
		return err
	}

	for _, sid := range affectedStudents {
		if err := tx.exec(ctx, `
			INSERT INTO fee_change_note_students (school_id, note_id, student_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING`,
			schoolID, noteID, sid,
		); err != nil {
			return err
		}
	}

	if s.Notifier != nil && len(affectedStudents) > 0 {
		rows, err := tx.query(ctx, `
			SELECT DISTINCT sg.user_id
			FROM student_guardians sg
			JOIN fee_change_note_students fcns
			  ON fcns.school_id = sg.school_id AND fcns.student_id = sg.student_id AND fcns.note_id = $2
			JOIN memberships m ON m.school_id = sg.school_id AND m.user_id = sg.user_id AND m.status = 'active'
			JOIN users u ON u.id = sg.user_id AND u.is_active = TRUE
			WHERE sg.school_id = $1 AND sg.is_fee_payer = TRUE
			ORDER BY sg.user_id ASC`,
			schoolID, noteID,
		)
		if err != nil {
			return err
		}
		var guardianIDs []string
		for rows.Next() {
			var uid string
			if err := rows.Scan(&uid); err != nil {
				rows.Close()
				return err
			}
			guardianIDs = append(guardianIDs, uid)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		body := fmt.Sprintf("The fee due date has been updated to %s: %s", newDateStr, trimmedNote)
		if oldDueDate.Valid {
			body = fmt.Sprintf("The fee due date has been updated from %s to %s: %s", oldDateStr.String, newDateStr, trimmedNote)
		}
		meta := map[string]any{
			"fee_change_note_id": noteID,
			"term_id":            termID,
			"change_type":        "due_date",
			"new_due_date":       newDateStr,
		}
		if feePeriodID.Valid {
			meta["fee_period_id"] = feePeriodID.String
		}
		if oldDateStr.Valid {
			meta["old_due_date"] = oldDateStr.String
		}

		sqlTx := tx.notificationTx()
		for _, uid := range guardianIDs {
			if err := s.Notifier.CreateInAppNotification(ctx, sqlTx, contracts.InAppNotification{
				SchoolID:        schoolID,
				RecipientUserID: uid,
				Category:        "fee_due_date_change",
				Title:           "Fee Due Date Updated",
				Body:            body,
				Metadata:        meta,
			}); err != nil {
				return err
			}
		}
	}

	details := map[string]any{
		"term_id":                 termID,
		"new_due_date":            newDateStr,
		"affected_students_count": len(affectedStudents),
		"note":                    trimmedNote,
	}
	if feePeriodID.Valid {
		details["fee_period_id"] = feePeriodID.String
	}
	if oldDateStr.Valid {
		details["old_due_date"] = oldDateStr.String
	}
	return tx.logAudit(ctx, s, contracts.AuditEvent{
		SchoolID:     schoolID,
		UserID:       actorUserID,
		Action:       "billing.term_due_date.sync",
		ResourceType: "fee_change_note",
		ResourceID:   noteID,
		Details:      details,
	})
}
