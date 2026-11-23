package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONAcceptsSingleValue(t *testing.T) {
	var body struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
	rec := httptest.NewRecorder()

	if !DecodeJSON(rec, req, &body) {
		t.Fatalf("DecodeJSON rejected a valid body: %s", rec.Body.String())
	}
	if body.Name != "test" {
		t.Fatalf("Name = %q, want test", body.Name)
	}
}

func TestDecodeJSONRejectsTrailingValue(t *testing.T) {
	var body struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}{"name":"second"}`))
	rec := httptest.NewRecorder()

	if DecodeJSON(rec, req, &body) {
		t.Fatal("DecodeJSON accepted trailing JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
