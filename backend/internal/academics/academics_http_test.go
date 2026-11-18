package academics

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTPAttendanceMarksReleaseReopenRerelease walks the complete CBC
// assessment lifecycle through the registered HTTP routes, not just the
// service methods. It is the closest thing to a frontend-to-backend smoke test
// that can run without a live PostgreSQL instance.
func TestHTTPAttendanceMarksReleaseReopenRerelease(t *testing.T) {
	env := newTestEnv(t)
	mux := http.NewServeMux()
	RegisterRoutes(mux, Deps{
		Store:          env.store,
		Notifier:       env.notifier,
		FeeSyncer:      env.feeSync,
		CommentDrafter: env.drafter,
		Config:         Config{AllowPrincipalDeanOverride: true},
	})

	classTeacher := identityCtx(testUserCT, testMemCT, RoleTeacher)
	mathTeacher := identityCtx(testUserMath, testMemMath, RoleTeacher)
	engTeacher := identityCtx(testUserEng, testMemEng, RoleTeacher)
	otherTeacher := identityCtx(testUserOther, testMemOther, RoleTeacher)
	parentA := identityCtx(testUserParentA, testMemParentA, RoleParent)

	// 1. Daily roll-call through the route.
	status, payload := doJSON(t, mux, http.MethodPost, "/api/v1/classes/"+testClassID+"/attendance", classTeacher, map[string]any{
		"date": "2026-01-10",
		"records": []map[string]string{
			{"student_id": testStudentA, "status": "present"},
			{"student_id": testStudentB, "status": "present"},
		},
	})
	if status != http.StatusOK {
		t.Fatalf("attendance route: %d %v", status, payload)
	}

	// 2. Marks entry through the subject-teacher route.
	for _, entry := range []struct {
		ctx       context.Context
		subjectID string
		scores    map[string]float64
	}{
		{mathTeacher, testMathID, map[string]float64{testStudentA: 70, testStudentB: 70}},
		{engTeacher, testEngID, map[string]float64{testStudentA: 60, testStudentB: 60}},
	} {
		marks := make([]map[string]any, 0, len(entry.scores))
		for studentID, score := range entry.scores {
			marks = append(marks, map[string]any{"student_id": studentID, "score": score})
		}
		status, payload = doJSON(t, mux, http.MethodPut,
			"/api/v1/classes/"+testClassID+"/subjects/"+entry.subjectID+"/marks", entry.ctx,
			map[string]any{"term_id": testTermID, "marks": marks})
		if status != http.StatusOK {
			t.Fatalf("marks route: %d %v", status, payload)
		}
	}

	// 3. Hard-gated release through the class-teacher route.
	status, payload = doJSON(t, mux, http.MethodPost, "/api/v1/classes/"+testClassID+"/results/release", classTeacher, map[string]any{
		"term_id": testTermID,
	})
	if status != http.StatusOK {
		t.Fatalf("release route: %d %v", status, payload)
	}

	// 4. Reopen, correct a mark, re-release without the required note (422).
	status, payload = doJSON(t, mux, http.MethodPost, "/api/v1/classes/"+testClassID+"/results/reopen", classTeacher, map[string]any{
		"term_id": testTermID, "reason": "Correcting Mathematics",
	})
	if status != http.StatusOK {
		t.Fatalf("reopen route: %d %v", status, payload)
	}
	status, payload = doJSON(t, mux, http.MethodPut,
		"/api/v1/classes/"+testClassID+"/subjects/"+testMathID+"/marks", mathTeacher,
		map[string]any{"term_id": testTermID, "marks": []map[string]any{{"student_id": testStudentA, "score": 85}}})
	if status != http.StatusOK {
		t.Fatalf("mark correction route: %d %v", status, payload)
	}
	status, payload = doJSON(t, mux, http.MethodPost, "/api/v1/classes/"+testClassID+"/results/re-release", classTeacher, map[string]any{
		"term_id": testTermID,
	})
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 without teacher note, got %d %v", status, payload)
	}
	if code := errorCode(payload); code != CodeRereleaseNote {
		t.Fatalf("expected %s, got %s (%v)", CodeRereleaseNote, code, payload)
	}

	// 5. Re-release with the teacher note succeeds.
	status, payload = doJSON(t, mux, http.MethodPost, "/api/v1/classes/"+testClassID+"/results/re-release", classTeacher, map[string]any{
		"term_id": testTermID,
		"correction_notes": []map[string]string{{
			"student_id": testStudentA, "subject_id": testMathID, "teacher_note": "Paper 2 addition corrected",
		}},
	})
	if status != http.StatusOK {
		t.Fatalf("re-release route: %d %v", status, payload)
	}

	// 6. The parent route serves the corrected, amended snapshot.
	status, payload = doJSON(t, mux, http.MethodGet,
		"/api/v1/students/"+testStudentA+"/results?term_id="+testTermID, parentA, nil)
	if status != http.StatusOK {
		t.Fatalf("parent results route: %d %v", status, payload)
	}
	subject := findSubjectInJSON(t, payload, testMathID)
	if subject["score"].(float64) != 85 || subject["is_amended"] != true {
		t.Fatalf("expected amended 85 in parent payload, got %v", subject)
	}

	// 7. Performance history route is available to the parent and has no rank.
	status, payload = doJSON(t, mux, http.MethodGet,
		"/api/v1/students/"+testStudentA+"/performance-history", parentA, nil)
	if status != http.StatusOK {
		t.Fatalf("performance route: %d %v", status, payload)
	}
	raw, _ := json.Marshal(payload)
	if bytes.Contains(raw, []byte("rank")) || bytes.Contains(raw, []byte("position")) {
		t.Fatalf("performance payload contains ranking fields: %s", raw)
	}

	// 8. An unrelated teacher can read anonymised stats but not names.
	status, payload = doJSON(t, mux, http.MethodGet,
		"/api/v1/classes/"+testClassID+"/stats?term_id="+testTermID, otherTeacher, nil)
	if status != http.StatusOK {
		t.Fatalf("stats route: %d %v", status, payload)
	}
	raw, _ = json.Marshal(payload)
	for _, forbidden := range []string{"Synthetic Student A", "Synthetic Student B", "ADM-001", testStudentA} {
		if bytes.Contains(raw, []byte(forbidden)) {
			t.Fatalf("stats route leaked %q: %s", forbidden, raw)
		}
	}
}

func doJSON(t *testing.T, mux *http.ServeMux, method, path string, ctx context.Context, body any) (int, map[string]any) {
	t.Helper()
	var buf *bytes.Buffer
	if body == nil {
		buf = bytes.NewBuffer(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		buf = bytes.NewBuffer(raw)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	payload := map[string]any{}
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &payload)
	}
	return rec.Code, payload
}

func errorCode(payload map[string]any) string {
	errObj, _ := payload["error"].(map[string]any)
	code, _ := errObj["code"].(string)
	return code
}

func findSubjectInJSON(t *testing.T, payload map[string]any, subjectID string) map[string]any {
	t.Helper()
	data, _ := payload["data"].(map[string]any)
	subjects, _ := data["subjects"].([]any)
	for _, entry := range subjects {
		subject, _ := entry.(map[string]any)
		if subject["subject_id"] == subjectID {
			return subject
		}
	}
	t.Fatalf("subject %s not found in %v", subjectID, payload)
	return nil
}
