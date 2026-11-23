package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var allowedEnvironments = map[string]bool{
	"development": true,
	"test":        true,
	"staging":     true,
	"production":  true,
}

var allowedSSLModes = map[string]bool{
	"disable":     true,
	"allow":       true,
	"prefer":      true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

// Config is the fully resolved runtime configuration for the backend.
type Config struct {
	Environment string
	LogLevel    string

	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig

	Server            ServerConfig
	CORS              CORSConfig
	TrustProxyHeaders bool
	TrustedProxyCIDRs []string

	InternalServiceToken string
	RetentionDays        int
	OTPHMACSecret        string
}

// CORSConfig describes which browser origins may call the API.
type CORSConfig struct {
	AllowedOrigins []string
}

// DBConfig describes the PostgreSQL connection and pool sizing.
type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

// DSN builds a libpq-compatible connection string with escaped credentials.
func (d DBConfig) DSN() string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     net.JoinHostPort(d.Host, d.Port),
		Path:     d.Name,
		RawQuery: url.Values{"sslmode": {d.SSLMode}}.Encode(),
	}
	return dsn.String()
}

// RedisConfig describes the job queue / cache connection.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// JWTConfig describes the bearer token signing configuration.
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// ServerConfig describes the HTTP listener.
type ServerConfig struct {
	Host string
	Port string
}

// Load reads configuration from the environment and validates the values that
// the service cannot start without.
func Load() (*Config, error) {
	expiryHours, err := envInt("JWT_EXPIRY_HOURS", 24)
	if err != nil {
		return nil, err
	}
	maxOpen, err := envInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	maxIdle, err := envInt("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return nil, err
	}
	retentionDays, err := envInt("UNVERIFIED_SUBMISSION_RETENTION_DAYS", 7)
	if err != nil {
		return nil, err
	}

	environment := strings.ToLower(strings.TrimSpace(env("ENVIRONMENT", "")))
	if environment == "" {
		return nil, fmt.Errorf("config: ENVIRONMENT is required")
	}
	if !allowedEnvironments[environment] {
		return nil, fmt.Errorf("config: ENVIRONMENT must be one of development, test, staging or production")
	}

	sslMode := strings.ToLower(strings.TrimSpace(env("DB_SSLMODE", "")))
	if sslMode == "" {
		if environment == "production" || environment == "staging" {
			sslMode = "require"
		} else {
			sslMode = "disable"
		}
	}
	if !allowedSSLModes[sslMode] {
		return nil, fmt.Errorf("config: DB_SSLMODE must be one of disable, allow, prefer, require, verify-ca or verify-full")
	}
	if (environment == "production" || environment == "staging") &&
		sslMode != "require" && sslMode != "verify-ca" && sslMode != "verify-full" {
		return nil, fmt.Errorf("config: DB_SSLMODE must require TLS in %s", environment)
	}

	trustProxyHeaders, err := envBool("TRUST_PROXY_HEADERS")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Environment: environment,
		LogLevel:    env("LOG_LEVEL", "debug"),
		DB: DBConfig{
			Host:         env("DB_HOST", "localhost"),
			Port:         env("DB_PORT", "5432"),
			User:         env("DB_USER", ""),
			Password:     env("DB_PASSWORD", ""),
			Name:         env("DB_NAME", ""),
			SSLMode:      sslMode,
			MaxOpenConns: maxOpen,
			MaxIdleConns: maxIdle,
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "localhost"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret: env("JWT_SECRET", ""),
			Expiry: time.Duration(expiryHours) * time.Hour,
		},
		Server: ServerConfig{
			Host: env("BACKEND_HOST", "0.0.0.0"),
			Port: env("BACKEND_PORT", "8080"),
		},
		CORS: CORSConfig{
			AllowedOrigins: envList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		TrustProxyHeaders:    trustProxyHeaders,
		TrustedProxyCIDRs:    envList("TRUSTED_PROXY_CIDRS", nil),
		InternalServiceToken: env("INTERNAL_SERVICE_TOKEN", ""),
		RetentionDays:        retentionDays,
		OTPHMACSecret:        env("OTP_HMAC_SECRET", ""),
	}

	if cfg.DB.Name == "" {
		return nil, fmt.Errorf("config: DB_NAME is required")
	}
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	if len(strings.TrimSpace(cfg.JWT.Secret)) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be at least 32 characters")
	}
	if cfg.JWT.Expiry <= 0 {
		return nil, fmt.Errorf("config: JWT_EXPIRY_HOURS must be positive")
	}
	if cfg.DB.MaxOpenConns <= 0 {
		return nil, fmt.Errorf("config: DB_MAX_OPEN_CONNS must be positive")
	}
	if cfg.DB.MaxIdleConns <= 0 {
		return nil, fmt.Errorf("config: DB_MAX_IDLE_CONNS must be positive")
	}
	if cfg.DB.MaxIdleConns > cfg.DB.MaxOpenConns {
		return nil, fmt.Errorf("config: DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS")
	}
	if cfg.RetentionDays <= 0 {
		return nil, fmt.Errorf("config: UNVERIFIED_SUBMISSION_RETENTION_DAYS must be positive")
	}
	if strings.TrimSpace(cfg.OTPHMACSecret) == "" {
		return nil, fmt.Errorf("config: OTP_HMAC_SECRET is required")
	}
	if len(strings.TrimSpace(cfg.OTPHMACSecret)) < 32 {
		return nil, fmt.Errorf("config: OTP_HMAC_SECRET must be at least 32 characters")
	}
	if cfg.TrustProxyHeaders && len(cfg.TrustedProxyCIDRs) == 0 {
		return nil, fmt.Errorf("config: TRUSTED_PROXY_CIDRS is required when TRUST_PROXY_HEADERS is enabled")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer: %w", key, err)
	}
	return n, nil
}

func envList(key string, fallback []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return fallback
	}
	return values
}

func envBool(key string) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a boolean: %w", key, err)
	}
	return value, nil
}
