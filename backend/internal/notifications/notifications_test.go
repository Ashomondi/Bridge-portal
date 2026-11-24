package notifications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/academics"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func TestCreateNotificationCommitsOrRollsBackWithCaller(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	s := &notifications.Service{DB: contracts.TenantTxFunc(db.WithTenantTx)}
	ctx := context.Background()

	principalA := contracts.Principal{UserID: a.UserID, SchoolID: a.ID, MembershipID: a.MembershipID, Role: "parent"}
	principalB := contracts.Principal{UserID: b.UserID, SchoolID: b.ID, MembershipID: b.MembershipID, Role: "parent"}

	// Seed a second active user in school A to verify multiple notifications in one transaction.
	var secondUserA string
	if err := db.Admin.QueryRow(ctx, "INSERT INTO users(phone_number,full_name) VALUES('+254700000101','Synthetic Second Guardian') RETURNING id").Scan(&secondUserA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, "INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'parent','active')", a.ID, secondUserA); err != nil {
		t.Fatal(err)
	}

	input := contracts.InAppNotification{
		SchoolID:        a.ID,
		RecipientUserID: a.UserID,
		Category:        "fee_change",
		Title:           "Fee change",
		Body:            "Synthetic fee notice",
		Metadata:        map[string]any{"student_id": a.StudentID},
	}

	// Single notification commits with caller transaction.
	if err := s.DB.WithTenantTx(ctx, principalA, func(tx pgx.Tx) error {
		return s.CreateInAppNotification(ctx, contracts.PgxTx(tx), input)
	}); err != nil {
		t.Fatal(err)
	}

	// Multiple notifications in a single transaction commit together.
	inputSecond := input
	inputSecond.RecipientUserID = secondUserA
	if err := s.DB.WithTenantTx(ctx, principalA, func(tx pgx.Tx) error {
		sqlTx := contracts.PgxTx(tx)
		if err := s.CreateInAppNotification(ctx, sqlTx, input); err != nil {
			return err
		}
		return s.CreateInAppNotification(ctx, sqlTx, inputSecond)
	}); err != nil {
		t.Fatal(err)
	}

	// Caller rollback reverts the notification created inside that transaction.
	rollback := errors.New("caller transaction failed")
	err := s.DB.WithTenantTx(ctx, principalA, func(tx pgx.Tx) error {
		if err := s.CreateInAppNotification(ctx, contracts.PgxTx(tx), input); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("caller rollback not propagated: %v", err)
	}

	// Failure on the second notification rolls back the entire transaction including the first.
	badSecond := input
	badSecond.RecipientUserID = b.UserID // wrong school -> INVALID_NOTIFICATION_RECIPIENT
	err = s.DB.WithTenantTx(ctx, principalA, func(tx pgx.Tx) error {
		sqlTx := contracts.PgxTx(tx)
		if err := s.CreateInAppNotification(ctx, sqlTx, input); err != nil {
			return err
		}
		return s.CreateInAppNotification(ctx, sqlTx, badSecond)
	})
	if err == nil {
		t.Fatal("partial multi-notification transaction succeeded")
	}

	var countA, countSecond int
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", a.ID, a.UserID).Scan(&countA); err != nil || countA != 2 {
		t.Fatalf("unexpected notification count for primary user after commits/rollbacks: %d %v", countA, err)
	}
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", a.ID, secondUserA).Scan(&countSecond); err != nil || countSecond != 1 {
		t.Fatalf("unexpected notification count for second user: %d %v", countSecond, err)
	}

	// Cross-school recipient rejected.
	input.RecipientUserID = b.UserID
	if err := s.DB.WithTenantTx(ctx, principalA, func(tx pgx.Tx) error {
		return s.CreateInAppNotification(ctx, contracts.PgxTx(tx), input)
	}); err == nil {
		t.Fatal("cross-school recipient accepted")
	}

	// Cross-school transaction rejected by RLS and membership validation.
	input.RecipientUserID = a.UserID
	if err := s.DB.WithTenantTx(ctx, principalB, func(tx pgx.Tx) error {
		return s.CreateInAppNotification(ctx, contracts.PgxTx(tx), input)
	}); err == nil {
		t.Fatal("cross-school transaction accepted")
	}

	// Nil transaction rejected explicitly.
	if err := s.CreateInAppNotification(ctx, nil, input); err == nil {
		t.Fatal("nil transaction accepted")
	}

	// Blank title rejected before touching DB.
	input.Title = " "
	if err := s.CreateInAppNotification(ctx, nil, input); err == nil {
		t.Fatal("blank notification accepted")
	}
}

func TestAcademicsServiceWithNotificationsTransactionAtomicity(t *testing.T) {
	db := testenv.New(t, true)
	a := db.SeedSchool(t, "01")
	ctx := context.Background()

	// Seed a class teacher and a second student with an inactive guardian so we
	// can verify both atomic commit and atomic rollback across academics + notifications.
	var teacherUserID, teacherMemID, secondStudentID, inactiveGuardianID string
	if err := db.Admin.QueryRow(ctx, "INSERT INTO users(phone_number,full_name) VALUES('+254700000201','Synthetic Class Teacher') RETURNING id").Scan(&teacherUserID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, "INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'teacher','active') RETURNING id", a.ID, teacherUserID).Scan(&teacherMemID); err != nil {
		t.Fatal(err)
	}
	var badgeID string
	if err := db.Admin.QueryRow(ctx, "INSERT INTO badges(school_id,name,slug,scope_type) VALUES($1,'Class Teacher','class_teacher','class') RETURNING id", a.ID).Scan(&badgeID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, "INSERT INTO membership_badges(membership_id,badge_id,scope_id) VALUES($1,$2,$3)", teacherMemID, badgeID, a.ClassID); err != nil {
		t.Fatal(err)
	}

	if err := db.Admin.QueryRow(ctx, "INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'DEMO-002','Synthetic Student Two','2013-02-01','male',$2) RETURNING id", a.ID, a.ClassID).Scan(&secondStudentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, "INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id) VALUES($1,$2,$3,$4,$5)", a.ID, secondStudentID, a.YearID, a.GradeID, a.ClassID); err != nil {
		t.Fatal(err)
	}
	if err := db.Admin.QueryRow(ctx, "INSERT INTO users(phone_number,full_name,is_active) VALUES('+254700000202','Inactive Guardian',false) RETURNING id").Scan(&inactiveGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, "INSERT INTO memberships(school_id,user_id,role,status) VALUES($1,$2,'parent','active')", a.ID, inactiveGuardianID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Admin.Exec(ctx, "INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'guardian')", a.ID, secondStudentID, inactiveGuardianID); err != nil {
		t.Fatal(err)
	}

	sqlDB := stdlib.OpenDBFromPool(db.Admin)
	t.Cleanup(func() { _ = sqlDB.Close() })

	notifSvc := &notifications.Service{DB: contracts.TenantTxFunc(db.WithTenantTx)}
	acadSvc := academics.NewService(academics.Deps{
		Store:    academics.NewSQLStore(sqlDB),
		Notifier: notifSvc,
	})
	teacherCtx := academics.WithIdentity(ctx, academics.Identity{
		UserID:       teacherUserID,
		SchoolID:     a.ID,
		MembershipID: teacherMemID,
		Role:         academics.RoleTeacher,
	})

	// 1. Successful roll-call commits both attendance and in-app absence notice atomically.
	res, err := acadSvc.RecordAttendance(teacherCtx, a.ClassID, academics.AttendanceRequest{
		Date:   "2026-01-12",
		TermID: a.TermID,
		Records: []academics.AttendanceRecordInput{
			{StudentID: a.StudentID, Status: "absent"},
		},
	})
	if err != nil {
		t.Fatalf("RecordAttendance failed: %v", err)
	}
	if res.RecordedCount != 1 || res.AbsentNotificationsQueued != 1 {
		t.Fatalf("unexpected attendance result: %+v", res)
	}

	// 2. When the second absent student's notification fails (inactive guardian),
	// the entire roll-call transaction must roll back: no attendance row for
	// 2026-01-13, no first-student notification, and no new audit row.
	_, err = acadSvc.RecordAttendance(teacherCtx, a.ClassID, academics.AttendanceRequest{
		Date:   "2026-01-13",
		TermID: a.TermID,
		Records: []academics.AttendanceRecordInput{
			{StudentID: a.StudentID, Status: "absent"},
			{StudentID: secondStudentID, Status: "absent"},
		},
	})
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "INVALID_NOTIFICATION_RECIPIENT" {
		t.Fatalf("expected INVALID_NOTIFICATION_RECIPIENT error propagation, got %v", err)
	}

	var rolledBackAttendance, totalNotices, totalAudits int
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM attendance_daily WHERE school_id=$1 AND attendance_date='2026-01-13'", a.ID).Scan(&rolledBackAttendance); err != nil || rolledBackAttendance != 0 {
		t.Fatalf("attendance was not rolled back on notification failure: %d %v", rolledBackAttendance, err)
	}
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1", a.ID).Scan(&totalNotices); err != nil || totalNotices != 1 {
		t.Fatalf("partial notification committed despite rollback: %d %v", totalNotices, err)
	}
	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE school_id=$1 AND action='daily_attendance_recorded'", a.ID).Scan(&totalAudits); err != nil || totalAudits != 1 {
		t.Fatalf("audit row was not rolled back on notification failure: %d %v", totalAudits, err)
	}
}

func TestNotificationRoutesScopePaginationAndReadOwnership(t *testing.T) {
	db := testenv.New(t, true)
	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
	ctx := context.Background()
	var ownID, otherUserID, otherSchoolID string
	for _, row := range []struct {
		school, user string
		id           *string
	}{{a.ID, a.UserID, &ownID}, {a.ID, b.UserID, &otherUserID}, {b.ID, b.UserID, &otherSchoolID}} {
		if err := db.Admin.QueryRow(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Synthetic','Synthetic') RETURNING id", row.school, row.user).Scan(row.id); err != nil {
			t.Fatal(err)
		}
	}
	s := &notifications.Service{DB: contracts.TenantTxFunc(db.WithTenantTx)}
	p := contracts.Principal{UserID: a.UserID, SchoolID: a.ID, MembershipID: a.MembershipID, Role: "parent"}
	mux := http.NewServeMux()
	notifications.RegisterRoutes(mux, s, func(*http.Request) (contracts.Principal, error) { return p, nil })

	// Request-supplied school headers/query params must not override Principal.SchoolID.
	req := httptest.NewRequest("GET", "/api/v1/notifications?per_page=1&school_id="+b.ID, nil)
	req.Header.Set("X-School-ID", b.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var result struct {
		Data []notifications.Notification `json:"data"`
		Meta contracts.Page               `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != 200 || len(result.Data) != 1 || result.Meta.Total != 1 || result.Data[0].ID != ownID {
		t.Fatalf("list scoping/pagination: %d %s %v", w.Code, w.Body.String(), err)
	}

	for _, id := range []string{otherUserID, otherSchoolID} {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("PATCH", "/api/v1/notifications/"+id+"/read", nil))
		if w.Code != 404 {
			t.Fatalf("another user's notice changed: %d", w.Code)
		}
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("PATCH", "/api/v1/notifications/not-a-uuid/read", nil))
	if w.Code != 400 {
		t.Fatalf("invalid notification ID status: %d", w.Code)
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications?per_page=101", nil))
	if w.Code != 422 {
		t.Fatalf("invalid pagination status: %d", w.Code)
	}

	var firstRead time.Time
	for i := 0; i < 2; i++ {
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("PATCH", "/api/v1/notifications/"+ownID+"/read", nil))
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		var readAt time.Time
		if err := db.Admin.QueryRow(ctx, "SELECT read_at FROM in_app_notifications WHERE school_id=$1 AND id=$2 AND is_read", a.ID, ownID).Scan(&readAt); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			firstRead = readAt
		} else if !readAt.Equal(firstRead) {
			t.Fatal("repeat read replaced original timestamp")
		}
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications?page=2&per_page=1", nil))
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Data) != 0 || result.Meta.Total != 1 {
		t.Fatal(w.Body.String())
	}

	unauthenticated := http.NewServeMux()
	notifications.RegisterRoutes(unauthenticated, s, nil)
	w = httptest.NewRecorder()
	unauthenticated.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/notifications", nil))
	if w.Code != 401 {
		t.Fatal("unauthenticated list was allowed")
	}
}
