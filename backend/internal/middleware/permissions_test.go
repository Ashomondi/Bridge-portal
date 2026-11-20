package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHasPermissionRoleDefaults(t *testing.T) {
	principal := Identity{Role: "principal"}
	if !principal.HasPermission(PermParentsQRManage, "") {
		t.Fatalf("principal should hold %s", PermParentsQRManage)
	}
	if !principal.HasPermission(PermStaffInvitesCreate, "") {
		t.Fatalf("principal should hold %s", PermStaffInvitesCreate)
	}

	teacher := Identity{Role: "teacher"}
	if teacher.HasPermission(PermParentsQRManage, "") {
		t.Fatalf("teacher must not hold %s", PermParentsQRManage)
	}
	if !teacher.HasPermission(PermClassStatsView, "") {
		t.Fatalf("teacher should hold %s", PermClassStatsView)
	}

	unknown := Identity{Role: "visitor"}
	if unknown.HasPermission(PermClassStatsView, "") {
		t.Fatalf("unknown role must not hold permissions")
	}
}

func TestHasPermissionScopedBadge(t *testing.T) {
	identity := Identity{
		Role: "teacher",
		Badges: []BadgeGrant{{
			ID:          "badge-1",
			Slug:        "class_teacher",
			ScopeType:   "class",
			ScopeID:     "class-a",
			Permissions: []string{PermClassStudentsView},
		}},
	}

	if !identity.HasPermission(PermClassStudentsView, "class-a") {
		t.Fatalf("scoped badge should allow its own class")
	}
	if identity.HasPermission(PermClassStudentsView, "class-b") {
		t.Fatalf("scoped badge must not allow another class")
	}
	if identity.HasPermission(PermClassStudentsView, "") {
		t.Fatalf("scoped badge must not allow a missing scope")
	}
}

func TestHasPermissionRejectsExpiredBadge(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	identity := Identity{
		Role: "teacher",
		Badges: []BadgeGrant{{
			ID:          "badge-1",
			Slug:        "delegated_verifier",
			ScopeType:   "none",
			Permissions: []string{PermRegistrationsVerify},
			ValidUntil:  &past,
		}},
	}

	if identity.HasPermission(PermRegistrationsVerify, "") {
		t.Fatalf("expired badge was accepted")
	}
}

func TestHasPermissionRejectsFutureBadge(t *testing.T) {
	future := time.Now().Add(time.Minute)
	identity := Identity{
		Role: "teacher",
		Badges: []BadgeGrant{{
			ID:          "badge-1",
			Slug:        "delegated_verifier",
			ScopeType:   "none",
			Permissions: []string{PermRegistrationsVerify},
			ValidFrom:   &future,
		}},
	}

	if identity.HasPermission(PermRegistrationsVerify, "") {
		t.Fatalf("badge that is not yet valid was accepted")
	}
}

func TestUnscopedBadgeIgnoresScope(t *testing.T) {
	identity := Identity{
		Badges: []BadgeGrant{{
			ID:          "badge-1",
			ScopeType:   "none",
			Permissions: []string{PermRegistrationsVerify},
		}},
	}
	if !identity.HasPermission(PermRegistrationsVerify, "anything") {
		t.Fatalf("unscoped badge should allow the permission")
	}
}

func TestRequirePermissionHTTP(t *testing.T) {
	middleware := NewAuthMiddleware(nil, nil)

	allowed := middleware.RequirePermission(PermParentsQRManage)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/parent-qr", nil)
	request = request.WithContext(WithIdentity(request.Context(), Identity{Role: "principal"}))
	recorder := httptest.NewRecorder()
	allowed.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("principal status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/parent-qr", nil)
	request = request.WithContext(WithIdentity(request.Context(), Identity{Role: "teacher"}))
	recorder = httptest.NewRecorder()
	allowed.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("teacher status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestRequirePermissionHTTPWithoutIdentity(t *testing.T) {
	middleware := NewAuthMiddleware(nil, nil)
	handler := middleware.RequirePermission(PermParentsQRManage)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/parent-qr", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
