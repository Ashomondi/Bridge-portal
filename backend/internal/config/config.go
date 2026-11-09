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

// Config is the fully resolved runtime configuration for the backend.
type Config struct {
	Environment string
	LogLevel    string

	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig

	Server               ServerConfig
	InternalServiceToken string
	RetentionDays        int
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

	cfg := &Config{
		Environment: env("ENVIRONMENT", "development"),
		LogLevel:    env("LOG_LEVEL", "debug"),
		DB: DBConfig{
			Host:         env("DB_HOST", "localhost"),
			Port:         env("DB_PORT", "5432"),
			User:         env("DB_USER", ""),
			Password:     env("DB_PASSWORD", ""),
			Name:         env("DB_NAME", ""),
			SSLMode:      env("DB_SSLMODE", "disable"),
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
		InternalServiceToken: env("INTERNAL_SERVICE_TOKEN", ""),
		RetentionDays:        retentionDays,
	}

	if cfg.DB.Name == "" {
		return nil, fmt.Errorf("config: DB_NAME is required")
	}
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	if cfg.JWT.Expiry <= 0 {
		return nil, fmt.Errorf("config: JWT_EXPIRY_HOURS must be positive")
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
