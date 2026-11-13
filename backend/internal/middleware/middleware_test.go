package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondAuthError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"school required", errSchoolRequired, http.StatusBadRequest, "SCHOOL_REQUIRED"},
		{"membership required", errMembershipRequired, http.StatusBadRequest, "MEMBERSHIP_REQUIRED"},
		{"no membership", errNoMembership, http.StatusForbidden, "FORBIDDEN"},
		{"database error", errors.New("database unavailable"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondAuthError(rec, tc.err)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var envelope errorEnvelope
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if envelope.Error.Code != tc.code {
				t.Fatalf("code = %q, want %q", envelope.Error.Code, tc.code)
			}
		})
	}
}

func TestFilterMemberships(t *testing.T) {
	memberships := []membershipRow{
		{id: "1", schoolID: "school-a", role: "teacher"},
		{id: "2", schoolID: "school-b", role: "parent"},
		{id: "3", schoolID: "school-a", role: "parent"},
	}

	got := filterMemberships(memberships, "school-a")
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, membership := range got {
		if membership.schoolID != "school-a" {
			t.Fatalf("unexpected school id %q", membership.schoolID)
		}
	}
}

func TestPathScope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/classes/class-1", nil)
	req.SetPathValue("class_id", "class-1")

	if got := PathScope("class_id")(req); got != "class-1" {
		t.Fatalf("PathScope = %q, want class-1", got)
	}
}
