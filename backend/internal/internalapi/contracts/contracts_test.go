package contracts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

const syntheticID = "10000000-0000-0000-0000-000000000001"

func TestIDRequiresCanonicalNonzeroUUID(t *testing.T) {
	for _, id := range []string{"", "00000000-0000-0000-0000-000000000000", "10000000x0000x0000x0000x000000000001", "10000000000000000000000000000001"} {
		if ValidID(id) {
			t.Fatalf("invalid ID accepted: %q", id)
		}
	}
	if !ValidID(syntheticID) {
		t.Fatal("canonical ID rejected")
	}
}

func TestIdentityRejectsUntrustedOrIncompleteContext(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-ID", syntheticID)
	for _, identity := range []Identity{nil, func(*http.Request) (Principal, error) {
		return Principal{UserID: syntheticID}, nil
	}, func(*http.Request) (Principal, error) { return Principal{}, errors.New("private token") }} {
		if _, err := RequireIdentity(r, identity); err == nil || strings.Contains(err.Error(), "private") {
			t.Fatalf("invalid identity: %v", err)
		}
	}
	p := Principal{syntheticID, syntheticID, syntheticID, "parent"}
	if got, err := RequireIdentity(r, func(*http.Request) (Principal, error) { return p, nil }); err != nil || got != p {
		t.Fatalf("trusted identity: %+v %v", got, err)
	}
}

func TestTenantAdapterPropagatesTransactionFailure(t *testing.T) {
	failure := errors.New("rollback")
	called := 0
	db := TenantTxFunc(func(ctx context.Context, school string, fn func(pgx.Tx) error) error {
		called++
		if school != syntheticID {
			t.Fatal("tenant changed")
		}
		return fn(nil)
	})
	for _, principal := range []Principal{{}, {SchoolID: syntheticID}, {UserID: syntheticID, SchoolID: "invalid", MembershipID: syntheticID, Role: "parent"}} {
		var apiError *APIError
		if err := db.WithTenantTx(context.Background(), principal, nil); !errors.As(err, &apiError) || apiError.Code != "UNAUTHENTICATED" || called != 0 {
			t.Fatalf("invalid principal reached adapter: %v", err)
		}
	}
	p := Principal{syntheticID, syntheticID, syntheticID, "parent"}
	if err := db.WithTenantTx(context.Background(), p, func(pgx.Tx) error { return failure }); !errors.Is(err, failure) || called != 1 {
		t.Fatalf("transaction error lost: %v", err)
	}
}

func TestTenantAdapterUsesAuthenticatedSchool(t *testing.T) {
	const otherSchool = "20000000-0000-0000-0000-000000000002"
	r := httptest.NewRequest("POST", "/?school_id="+otherSchool, strings.NewReader(`{"school_id":"`+otherSchool+`"}`))
	r.Header.Set("X-School-ID", otherSchool)
	p, err := RequireIdentity(r, func(*http.Request) (Principal, error) {
		return Principal{syntheticID, syntheticID, syntheticID, "parent"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var usedSchool string
	db := TenantTxFunc(func(ctx context.Context, school string, fn func(pgx.Tx) error) error {
		usedSchool = school
		return fn(nil)
	})
	if err := db.WithTenantTx(r.Context(), p, func(pgx.Tx) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if usedSchool != syntheticID {
		t.Fatalf("request-selected tenant reached transaction: %s", usedSchool)
	}
}

func TestResponseEnvelopesAndSafeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, errors.New("database password private"))
	if w.Code != 500 || strings.Contains(w.Body.String(), "password") || !strings.Contains(w.Body.String(), `"details":{}`) {
		t.Fatalf("unsafe error: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	WriteError(w, Fail(403, "FORBIDDEN", "Access denied."))
	if w.Code != 403 || !strings.Contains(w.Body.String(), `"code":"FORBIDDEN"`) {
		t.Fatal(w.Body.String())
	}
	w = httptest.NewRecorder()
	Success(w, 200, []string{}, &Page{1, 25, 0})
	if w.Body.String() != `{"status":"success","data":[],"meta":{"page":1,"per_page":25,"total":0}}` || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Body.String())
	}
}

func TestDecodeAndPagination(t *testing.T) {
	for _, body := range []string{`{"unexpected":true}`, `{}`, `{} {}`, strings.Repeat("x", (1<<20)+1)} {
		var input struct{}
		err := Decode(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), &input)
		if (err == nil) != (body == "{}") {
			t.Fatalf("decode result %v", err)
		}
	}
	for _, query := range []string{"page=0", "per_page=101", "page=1000001", "page=bad"} {
		if _, err := Pagination(httptest.NewRequest("GET", "/?"+query, nil)); err == nil {
			t.Fatalf("invalid pagination accepted: %s", query)
		}
	}
	if p, err := Pagination(httptest.NewRequest("GET", "/", nil)); err != nil || p.Page != 1 || p.PerPage != 25 {
		t.Fatalf("defaults: %+v %v", p, err)
	}
}
