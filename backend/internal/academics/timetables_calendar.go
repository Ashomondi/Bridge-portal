package academics

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"
)

// TimetableSlotInput is one lesson slot supplied by the Dean.
type TimetableSlotInput struct {
	SubjectID           string `json:"subject_id"`
	TeacherMembershipID string `json:"teacher_membership_id,omitempty"`
	DayOfWeek           int    `json:"day_of_week"`
	StartTime           string `json:"start_time"`
	EndTime             string `json:"end_time"`
	Room                string `json:"room,omitempty"`
}

// TimetableRequest replaces a class's timetable for one term.
type TimetableRequest struct {
	TermID  string               `json:"term_id"`
	ClassID string               `json:"class_id"`
	Slots   []TimetableSlotInput `json:"slots"`
}

// TimetableResponse is the GET/PUT timetable payload.
type TimetableResponse struct {
	TermID  string          `json:"term_id"`
	ClassID string          `json:"class_id,omitempty"`
	Slots   []TimetableSlot `json:"slots"`
}

// GetTimetable implements GET /api/v1/timetables.
func (s *Service) GetTimetable(ctx context.Context, termID, classID string) (TimetableResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return TimetableResponse{}, err
	}
	termID = strings.TrimSpace(termID)
	if termID == "" {
		return TimetableResponse{}, errValidation("term_id is required.")
	}
	var out TimetableResponse
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		if _, err := tx.TermByID(ctx, actor.SchoolID, termID); err != nil {
			return err
		}
		slots, err := tx.ListTimetableSlots(ctx, actor.SchoolID, termID, strings.TrimSpace(classID))
		if err != nil {
			return err
		}
		out = TimetableResponse{TermID: termID, ClassID: classID, Slots: slots}
		return nil
	})
	return out, err
}

// PutTimetable replaces a class timetable, validates clashes and notifies the
// affected teachers and parents in-app (C59).
func (s *Service) PutTimetable(ctx context.Context, req TimetableRequest) (TimetableResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return TimetableResponse{}, err
	}
	if !s.canManage(actor, PermTimetablesManage) {
		return TimetableResponse{}, errForbidden(CodeForbidden, "Only the Dean of Studies may edit timetables.")
	}
	req.TermID = strings.TrimSpace(req.TermID)
	req.ClassID = strings.TrimSpace(req.ClassID)
	if req.TermID == "" || req.ClassID == "" {
		return TimetableResponse{}, errValidation("term_id and class_id are required.")
	}

	var saved []TimetableSlot
	var teacherRecipients []string
	var guardianRecipients []string
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		term, err := tx.TermByID(ctx, actor.SchoolID, req.TermID)
		if err != nil {
			return err
		}
		class, err := tx.ClassByID(ctx, actor.SchoolID, req.ClassID)
		if err != nil {
			return err
		}
		subjects, err := tx.SubjectsForClass(ctx, actor.SchoolID, req.ClassID)
		if err != nil {
			return err
		}
		subjectByID := map[string]Subject{}
		for _, sub := range subjects {
			subjectByID[sub.ID] = sub
		}

		slots := make([]TimetableSlot, 0, len(req.Slots))
		for _, in := range req.Slots {
			sub, ok := subjectByID[strings.TrimSpace(in.SubjectID)]
			if !ok {
				return errValidation("A timetable subject does not belong to this class's grade level.")
			}
			if in.DayOfWeek < 1 || in.DayOfWeek > 7 {
				return errValidation("day_of_week must be between 1 (Monday) and 7 (Sunday).")
			}
			start, err := normalizeClock(in.StartTime)
			if err != nil {
				return err
			}
			end, err := normalizeClock(in.EndTime)
			if err != nil {
				return err
			}
			if start >= end {
				return errValidation("A lesson's end_time must be after its start_time.")
			}
			teacherID := strings.TrimSpace(in.TeacherMembershipID)
			if teacherID == "" {
				if a, aerr := tx.SubjectAssignment(ctx, actor.SchoolID, req.ClassID, sub.ID, term.AcademicYearID); aerr == nil {
					teacherID = a.TeacherMembershipID
				}
			}
			if teacherID == "" {
				return errValidation("Every timetable slot needs an assigned teacher.")
			}
			if _, err := tx.MembershipByID(ctx, actor.SchoolID, teacherID); err != nil {
				return errValidation("A timetable slot references a teacher who is not in this school.")
			}
			slots = append(slots, TimetableSlot{
				AcademicYearID:      term.AcademicYearID,
				TermID:              req.TermID,
				ClassID:             req.ClassID,
				SubjectID:           sub.ID,
				SubjectName:         sub.Name,
				TeacherMembershipID: teacherID,
				DayOfWeek:           in.DayOfWeek,
				StartTime:           start,
				EndTime:             end,
				Room:                strings.TrimSpace(in.Room),
			})
		}
		if err := validateNoOverlaps(slots); err != nil {
			return err
		}
		if err := tx.ReplaceTimetableSlots(ctx, actor.SchoolID, req.TermID, req.ClassID, slots); err != nil {
			return err
		}
		saved = slots

		for _, slot := range slots {
			if mem, err := tx.MembershipByID(ctx, actor.SchoolID, slot.TeacherMembershipID); err == nil {
				teacherRecipients = append(teacherRecipients, mem.UserID)
			}
		}
		guardianRecipients, _ = tx.GuardianIDsForClass(ctx, actor.SchoolID, req.ClassID)

		return s.audit(ctx, tx, actor, "timetable_updated", "class", class.ID, map[string]any{
			"term_id":    req.TermID,
			"slot_count": len(slots),
		})
	})
	if err != nil {
		return TimetableResponse{}, err
	}

	recipients := append(append([]string{}, teacherRecipients...), guardianRecipients...)
	_, _ = s.notifyMany(ctx, actor.SchoolID, recipients, "academics.timetable",
		"Timetable updated",
		"The class timetable has been updated. Open the app to see the new lesson times.",
		map[string]any{"class_id": req.ClassID, "term_id": req.TermID})

	return TimetableResponse{TermID: req.TermID, ClassID: req.ClassID, Slots: saved}, nil
}

func normalizeClock(value string) (string, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04", "15:04:05", "3:04 PM", "03:04 PM"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format("15:04"), nil
		}
	}
	return "", errValidation("Times must use 24-hour HH:MM format.")
}

func validateNoOverlaps(slots []TimetableSlot) error {
	byDay := map[int][]TimetableSlot{}
	for _, slot := range slots {
		byDay[slot.DayOfWeek] = append(byDay[slot.DayOfWeek], slot)
	}
	for _, daySlots := range byDay {
		sort.Slice(daySlots, func(i, j int) bool { return daySlots[i].StartTime < daySlots[j].StartTime })
		for i := 1; i < len(daySlots); i++ {
			if daySlots[i].StartTime < daySlots[i-1].EndTime {
				return errValidation("Timetable slots overlap on day " + itoa(daySlots[i].DayOfWeek) + ".")
			}
		}
	}
	return nil
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	return string(digits)
}

// TermUpdateInput is one term-date edit made after the setup wizard (C60/C61).
type TermUpdateInput struct {
	TermID        string `json:"term_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	FeeChangeNote string `json:"fee_change_note"`
}

// CalendarEventInput is one school event.
type CalendarEventInput struct {
	ID          string `json:"id,omitempty"`
	TermID      string `json:"term_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventType   string `json:"event_type"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at"`
}

// CalendarRequest updates term dates and school events.
type CalendarRequest struct {
	TermUpdates []TermUpdateInput    `json:"term_updates,omitempty"`
	Events      []CalendarEventInput `json:"events,omitempty"`
}

// CalendarResponse is the GET/PUT calendar payload.
type CalendarResponse struct {
	Terms              []Term          `json:"terms"`
	Events             []CalendarEvent `json:"events"`
	Updated            bool            `json:"updated,omitempty"`
	FeeDueDatesShifted bool            `json:"fee_due_dates_shifted"`
	EventsSaved        int             `json:"events_saved,omitempty"`
}

// GetCalendar returns term dates and events.
func (s *Service) GetCalendar(ctx context.Context, termID string) (CalendarResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return CalendarResponse{}, err
	}
	var out CalendarResponse
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		terms, err := tx.ListTerms(ctx, actor.SchoolID)
		if err != nil {
			return err
		}
		events, err := tx.ListCalendarEvents(ctx, actor.SchoolID, strings.TrimSpace(termID))
		if err != nil {
			return err
		}
		out.Terms = terms
		out.Events = events
		return nil
	})
	return out, err
}

// PutCalendar updates term dates and events. A start-date change requires a
// non-empty fee_change_note, calls billing's fee-sync hook and is audited
// (C61, AC-22).
func (s *Service) PutCalendar(ctx context.Context, req CalendarRequest) (CalendarResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return CalendarResponse{}, err
	}
	if !s.canManage(actor, PermCalendarManage) {
		return CalendarResponse{}, errForbidden(CodeForbidden, "Only the Dean of Studies may edit the school calendar.")
	}

	type validatedTermUpdate struct {
		input TermUpdateInput
		start time.Time
		end   time.Time
		old   Term
	}
	validated := make([]validatedTermUpdate, 0, len(req.TermUpdates))

	// Validate everything before writing, so one bad update cannot partially
	// move term dates.
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		for _, in := range req.TermUpdates {
			in.TermID = strings.TrimSpace(in.TermID)
			if in.TermID == "" {
				return errValidation("Every term update needs term_id.")
			}
			old, err := tx.TermByID(ctx, actor.SchoolID, in.TermID)
			if err != nil {
				return err
			}
			start, err := parseDate(in.StartDate)
			if err != nil {
				return err
			}
			end := dateOnly(old.EndDate)
			if strings.TrimSpace(in.EndDate) != "" {
				end, err = parseDate(in.EndDate)
				if err != nil {
					return err
				}
			}
			if end.Before(start) {
				return errValidation("A term's end_date cannot be before its start_date.")
			}
			startChanged := !dateOnly(start).Equal(dateOnly(old.StartDate))
			if startChanged && strings.TrimSpace(in.FeeChangeNote) == "" {
				return &APIError{
					Status:  http.StatusUnprocessableEntity,
					Code:    CodeTermDateNote,
					Message: "Changing a term start date moves fee due dates and requires a fee_change_note.",
					Details: map[string]any{"term_id": in.TermID},
				}
			}
			validated = append(validated, validatedTermUpdate{input: in, start: start, end: end, old: old})
		}
		return nil
	})
	if err != nil {
		return CalendarResponse{}, err
	}

	out := CalendarResponse{Updated: true}
	var recipients []string
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		for _, update := range validated {
			startChanged := !dateOnly(update.start).Equal(dateOnly(update.old.StartDate))
			if err := tx.UpdateTermDates(ctx, actor.SchoolID, update.input.TermID, update.start, update.end); err != nil {
				return err
			}
			details := map[string]any{
				"term_id":   update.input.TermID,
				"old_start": dateOnly(update.old.StartDate).Format("2006-01-02"),
				"new_start": dateOnly(update.start).Format("2006-01-02"),
				"old_end":   dateOnly(update.old.EndDate).Format("2006-01-02"),
				"new_end":   dateOnly(update.end).Format("2006-01-02"),
			}
			if err := s.audit(ctx, tx, actor, "term_dates_updated", "term", update.input.TermID, details); err != nil {
				return err
			}
			if startChanged {
				out.FeeDueDatesShifted = true
				details["fee_change_note"] = update.input.FeeChangeNote
				if err := s.feeSyncer.OnTermStartDateChanged(ctx, actor.SchoolID, update.input.TermID, update.start, actor.UserID, update.input.FeeChangeNote); err != nil {
					return err
				}
			}
		}

		events := make([]CalendarEvent, 0, len(req.Events))
		for _, in := range req.Events {
			in.Title = strings.TrimSpace(in.Title)
			if in.Title == "" {
				return errValidation("Every calendar event needs a title.")
			}
			start, err := parseDateTime(in.StartAt)
			if err != nil {
				return err
			}
			end, err := parseDateTime(in.EndAt)
			if err != nil {
				return err
			}
			if end.Before(start) {
				return errValidation("A calendar event's end_at cannot be before start_at.")
			}
			termID := strings.TrimSpace(in.TermID)
			academicYearID := ""
			if termID != "" {
				term, err := tx.TermByID(ctx, actor.SchoolID, termID)
				if err != nil {
					return err
				}
				academicYearID = term.AcademicYearID
			}
			events = append(events, CalendarEvent{
				ID:             strings.TrimSpace(in.ID),
				TermID:         termID,
				AcademicYearID: academicYearID,
				Title:          in.Title,
				Description:    in.Description,
				EventType:      in.EventType,
				StartAt:        start,
				EndAt:          end,
				CreatedBy:      actor.UserID,
			})
		}
		if len(events) > 0 {
			if err := tx.UpsertCalendarEvents(ctx, actor.SchoolID, events); err != nil {
				return err
			}
			out.EventsSaved = len(events)
			if err := s.audit(ctx, tx, actor, "calendar_events_updated", "calendar", "", map[string]any{"event_count": len(events)}); err != nil {
				return err
			}
		}
		if len(events) > 0 {
			recipients, _ = tx.GuardianIDsForSchool(ctx, actor.SchoolID)
			staff, _ := tx.StaffUserIDsForSchool(ctx, actor.SchoolID)
			recipients = append(recipients, staff...)
		}
		return nil
	})
	if err != nil {
		return CalendarResponse{}, err
	}
	if len(recipients) > 0 {
		_, _ = s.notifyMany(ctx, actor.SchoolID, recipients, "academics.calendar",
			"School calendar updated",
			"The school calendar or term dates have been updated. Open the app for details.",
			map[string]any{})
	}

	// Return the fresh calendar after updates.
	fresh, err := s.GetCalendar(ctx, "")
	if err != nil {
		return CalendarResponse{}, err
	}
	fresh.Updated = true
	fresh.FeeDueDatesShifted = out.FeeDueDatesShifted
	fresh.EventsSaved = out.EventsSaved
	return fresh, nil
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errValidation("A date is required in YYYY-MM-DD format.")
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errValidation("Dates must use YYYY-MM-DD format.")
	}
	return dateOnly(t), nil
}

func parseDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errValidation("A timestamp is required.")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errValidation("Timestamps must use ISO-8601 format.")
}

func (s *Service) handleGetTimetable(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetTimetable(r.Context(), r.URL.Query().Get("term_id"), r.URL.Query().Get("class_id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handlePutTimetable(w http.ResponseWriter, r *http.Request) {
	var req TimetableRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.PutTimetable(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	result, err := s.GetCalendar(r.Context(), r.URL.Query().Get("term_id"))
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handlePutCalendar(w http.ResponseWriter, r *http.Request) {
	var req CalendarRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.PutCalendar(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func init() {
	registerRoute(http.MethodGet, "/api/v1/timetables", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleGetTimetable(w, r)
	})
	registerRoute(http.MethodPut, "/api/v1/timetables", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handlePutTimetable(w, r)
	})
	registerRoute(http.MethodGet, "/api/v1/calendar", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleGetCalendar(w, r)
	})
	registerRoute(http.MethodPut, "/api/v1/calendar", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handlePutCalendar(w, r)
	})
}
