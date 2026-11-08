package academics

import (
	"context"
	"net/http"
	"strings"
)

// AttendanceRecordInput is one learner's mark for the day.
type AttendanceRecordInput struct {
	StudentID string `json:"student_id"`
	Status    string `json:"status"`
}

// AttendanceRequest is the daily class-teacher roll-call (C35).
type AttendanceRequest struct {
	Date    string                  `json:"date"`
	TermID  string                  `json:"term_id,omitempty"`
	Records []AttendanceRecordInput `json:"records"`
}

// AttendanceResult summarises one saved roll-call.
type AttendanceResult struct {
	ClassID                   string `json:"class_id"`
	Date                      string `json:"date"`
	RecordedCount             int    `json:"recorded_count"`
	AbsentNotificationsQueued int    `json:"absent_notifications_queued"`
}

// RecordAttendance implements POST /api/v1/classes/{id}/attendance. Only the
// class teacher badge scoped to this class may call it (AC-14), and every
// absent learner's guardians receive an in-app notification (C36).
func (s *Service) RecordAttendance(ctx context.Context, classID string, req AttendanceRequest) (AttendanceResult, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return AttendanceResult{}, err
	}
	classID = strings.TrimSpace(classID)
	if classID == "" {
		return AttendanceResult{}, errValidation("class_id is required.")
	}
	date, err := parseDate(req.Date)
	if err != nil {
		return AttendanceResult{}, err
	}
	if len(req.Records) == 0 {
		return AttendanceResult{}, errValidation("At least one attendance record is required.")
	}

	var result AttendanceResult
	var absentStudentIDs []string
	type absenceNotice struct {
		StudentID string
		Guardians []string
	}
	var absenceNotices []absenceNotice
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		allowed, err := tx.HasClassTeacherBadge(ctx, actor.SchoolID, actor.MembershipID, classID)
		if err != nil {
			return err
		}
		if !allowed {
			return errForbidden(CodeClassTeacherScope, "Only the class teacher for this class may record the daily roll-call.")
		}
		if _, err := tx.ClassByID(ctx, actor.SchoolID, classID); err != nil {
			return err
		}

		var term Term
		if strings.TrimSpace(req.TermID) != "" {
			term, err = tx.TermByID(ctx, actor.SchoolID, strings.TrimSpace(req.TermID))
		} else {
			term, err = tx.TermForDate(ctx, actor.SchoolID, date)
		}
		if err != nil {
			return errValidation("The attendance date must belong to a school term.")
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
		rows := make([]AttendanceRow, 0, len(req.Records))
		for _, in := range req.Records {
			in.StudentID = strings.TrimSpace(in.StudentID)
			if in.StudentID == "" || !enrolled[in.StudentID] {
				return errValidation("Attendance includes a student who is not enrolled in this class.")
			}
			if seen[in.StudentID] {
				return errValidation("A student may only appear once in a daily roll-call.")
			}
			seen[in.StudentID] = true
			status := strings.ToLower(strings.TrimSpace(in.Status))
			if status != "present" && status != "absent" && status != "late" {
				return errValidation("Attendance status must be present, absent or late.")
			}
			rows = append(rows, AttendanceRow{
				StudentID:      in.StudentID,
				ClassID:        classID,
				TermID:         term.ID,
				AttendanceDate: date,
				Status:         status,
				Source:         "manual_roll_call",
				RecordedBy:     actor.UserID,
			})
			if status == "absent" {
				absentStudentIDs = append(absentStudentIDs, in.StudentID)
			}
		}
		count, err := tx.UpsertAttendance(ctx, actor.SchoolID, rows)
		if err != nil {
			return err
		}
		result = AttendanceResult{
			ClassID:       classID,
			Date:          date.Format("2006-01-02"),
			RecordedCount: count,
		}
		for _, studentID := range absentStudentIDs {
			guardians, _ := tx.GuardianIDsForStudent(ctx, actor.SchoolID, studentID)
			absenceNotices = append(absenceNotices, absenceNotice{StudentID: studentID, Guardians: guardians})
		}
		return s.audit(ctx, tx, actor, "daily_attendance_recorded", "class", classID, map[string]any{
			"term_id":         term.ID,
			"attendance_date": result.Date,
			"recorded_count":  count,
			"absent_count":    len(absentStudentIDs),
			"source":          "manual_roll_call",
		})
	})
	if err != nil {
		return AttendanceResult{}, err
	}

	// In-app absence notifications are Release 1 only; SMS is Release 2 (C58).
	for _, notice := range absenceNotices {
		queued, _ := s.notifyMany(ctx, actor.SchoolID, notice.Guardians, "academics.attendance",
			"Absence recorded",
			"Your child was marked absent in today's class roll-call.",
			map[string]any{"student_id": notice.StudentID, "class_id": classID, "date": result.Date})
		result.AbsentNotificationsQueued += queued
	}
	return result, nil
}

func (s *Service) handleAttendance(w http.ResponseWriter, r *http.Request) {
	var req AttendanceRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.RecordAttendance(r.Context(), r.PathValue("id"), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodPost, "/api/v1/classes/{id}/attendance", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleAttendance(w, r)
	})
}
