package billing_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
)

func TestParseFlexibleIntStrict(t *testing.T) {
	validCases := []struct {
		raw  string
		want int
	}{
		{`0`, 0},
		{`"0"`, 0},
		{`1037`, 1037},
		{`"1037"`, 1037},
		{`" 42 "`, 42},
		{`-1`, -1},
		{`"-1"`, -1},
	}
	for _, tc := range validCases {
		got, err := billing.ParseFlexibleInt(json.RawMessage(tc.raw))
		if err != nil || got != tc.want {
			t.Fatalf("ParseFlexibleInt(%s) = (%d, %v), want (%d, nil)", tc.raw, got, err, tc.want)
		}
	}

	invalidCases := []string{
		``,
		`""`,
		`"   "`,
		`null`,
		`0.5`,
		`"0.5"`,
		`"1.0"`,
		`"0abc"`,
		`"123xyz"`,
		`"0 1"`,
		`"abc"`,
		`true`,
		`{}`,
		`"99999999999999999999999999999"`,
	}
	for _, raw := range invalidCases {
		if got, err := billing.ParseFlexibleInt(json.RawMessage(raw)); err == nil {
			t.Fatalf("ParseFlexibleInt(%s) unexpectedly succeeded with %d", raw, got)
		}
	}
}

func TestDarajaTimestampUsesAfricaNairobi(t *testing.T) {
	// 2026-02-10 22:30:45 UTC is 2026-02-11 01:30:45 EAT (Africa/Nairobi, UTC+3).
	utcInstant := time.Date(2026, 2, 10, 22, 30, 45, 0, time.UTC)
	if got, want := billing.DarajaTimestamp(utcInstant), "20260211013045"; got != want {
		t.Fatalf("DarajaTimestamp(UTC) = %q, want %q", got, want)
	}

	// Verify conversion from another non-UTC zone (e.g. America/New_York UTC-5).
	ny := time.FixedZone("EST", -5*60*60)
	nyInstant := time.Date(2026, 2, 10, 17, 30, 45, 0, ny)
	if got, want := billing.DarajaTimestamp(nyInstant), "20260211013045"; got != want {
		t.Fatalf("DarajaTimestamp(EST) = %q, want %q", got, want)
	}
}

func TestCallbackAuthAndFlatDarajaErrorResponses(t *testing.T) {
	svc := &billing.Service{}
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return contracts.Principal{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "unauth")
	})

	assertFlatDarajaError := func(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
		t.Helper()
		if rec.Code != wantStatus {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, wantStatus, rec.Body.String())
		}
		var raw map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatalf("response is not valid JSON: %v (body=%s)", err, rec.Body.String())
		}
		if _, hasStatus := raw["status"]; hasStatus {
			t.Fatalf("Daraja callback must not use application envelope: %s", rec.Body.String())
		}
		if _, hasData := raw["data"]; hasData {
			t.Fatalf("Daraja callback must not wrap in data envelope: %s", rec.Body.String())
		}
		codeVal, ok := raw["ResultCode"].(float64)
		if !ok || int(codeVal) != 1 {
			t.Fatalf("expected integer ResultCode=1, got %T (%v) in %s", raw["ResultCode"], raw["ResultCode"], rec.Body.String())
		}
		if desc, ok := raw["ResultDesc"].(string); !ok || desc == "" {
			t.Fatalf("expected non-empty string ResultDesc, got %T (%v)", raw["ResultDesc"], raw["ResultDesc"])
		}
	}

	// 1. Fail closed when CallbackAuthToken and VerifyCallback are unset
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusUnauthorized)

	// 2. Fail closed when CallbackAuthToken is set but TrustedCallbackProxies is empty (production misconfig)
	svc.CallbackAuthToken = "top-secret-proxy-token"
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Callback-Token", "top-secret-proxy-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusUnauthorized)

	// 3. Reject spoofed X-Callback-Token from untrusted remote IP
	svc.TrustedCallbackProxies = []string{"10.0.0.0/8", "127.0.0.1"}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "198.51.100.42:44321"
	req.Header.Set("X-Callback-Token", "top-secret-proxy-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusUnauthorized)

	// 4. Reject query-string token even from trusted proxy
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/c2b-confirmation?token=top-secret-proxy-token", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "10.0.1.5:50000"
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusUnauthorized)

	// 5. Reject invalid token from trusted proxy
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/c2b-confirmation", bytes.NewBufferString(`{}`))
	req.RemoteAddr = "10.0.1.5:50000"
	req.Header.Set("X-Callback-Token", "wrong-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusUnauthorized)

	// 6. Malformed JSON from authenticated trusted proxy returns flat Daraja 400
	req = httptest.NewRequest(http.MethodPost, "/api/v1/billing/mpesa/callback", bytes.NewBufferString(`{not-json`))
	req.RemoteAddr = "10.0.1.5:50000"
	req.Header.Set("X-Callback-Token", "top-secret-proxy-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertFlatDarajaError(t, rec, http.StatusBadRequest)
}
