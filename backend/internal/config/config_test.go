package config

import (
	"strings"
	"testing"
)

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("DB_NAME", "ikonex_academy")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
}

func TestLoadRequiresEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("ENVIRONMENT", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error when ENVIRONMENT is missing")
	}
}

func TestLoadRejectsUnknownEnvironment(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("ENVIRONMENT", "prod-ish")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for an unknown environment")
	}
}

func TestLoadRequiresLongJWTSecret(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("JWT_SECRET", "too-short")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a short JWT secret")
	}
}

func TestLoadDefaultsAndLists(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://app.ikonex.test")
	t.Setenv("TRUST_PROXY_HEADERS", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Environment != "development" {
		t.Fatalf("Environment = %q, want development", cfg.Environment)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[1] != "https://app.ikonex.test" {
		t.Fatalf("unexpected CORS origins: %#v", cfg.CORS.AllowedOrigins)
	}
	if !cfg.TrustProxyHeaders {
		t.Fatal("TrustProxyHeaders = false, want true")
	}
}

func TestLoadDefaultsCORSOrigins(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://localhost:3000" {
		t.Fatalf("unexpected default CORS origins: %#v", cfg.CORS.AllowedOrigins)
	}
}
