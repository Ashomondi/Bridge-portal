package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/auth"
	"github.com/Ashomondi/ikonex-academy/backend/internal/config"
	"github.com/Ashomondi/ikonex-academy/backend/internal/consent"
	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
	"github.com/Ashomondi/ikonex-academy/backend/internal/onboarding"
)

func main() {
	logger := log.New(os.Stdout, "backend ", log.LstdFlags|log.LUTC)
	if err := run(logger); err != nil {
		logger.Fatalf("startup failed: %v", err)
	}
}

func run(logger *log.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg.DB.DSN(), cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns)
	if err != nil {
		return err
	}
	defer db.Close()

	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiry)
	audit := database.NewAuditLogger(db.DB)
	allowDevelopmentSMS := !strings.EqualFold(cfg.Environment, "production")

	authService := auth.NewService(db.DB, tokens, devSMSDispatcher{logger: logger, enabled: allowDevelopmentSMS}, audit)
	consentService := consent.NewService(db.DB, audit)
	onboardingService := onboarding.NewService(db.DB, audit)
	authMiddleware := middleware.NewAuthMiddleware(db.DB, tokens)

	mux := http.NewServeMux()
	authService.RegisterRoutes(mux)

	consentWrap := func(next http.Handler) http.Handler {
		handler := authMiddleware.RequirePermission(middleware.PermConsentsManage)(next)
		handler = authMiddleware.RequireTenant(handler)
		return authMiddleware.RequireAuth(handler)
	}
	consentService.RegisterRoutes(mux, consentWrap)

	onboardingService.RegisterRoutes(mux, authMiddleware, middleware.RequireInternalService(cfg.InternalServiceToken))

	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s (env=%s)", server.Addr, cfg.Environment)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	select {
	case serveErr := <-errCh:
		return serveErr
	case <-shutdownCtx.Done():
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

// devSMSDispatcher logs transactional SMS in development so OTP login can be
// exercised before the notifications provider is wired. It stays disabled in
// production to avoid leaking one-time codes.
type devSMSDispatcher struct {
	logger  *log.Logger
	enabled bool
}

func (d devSMSDispatcher) SendTransactionalSMS(_ context.Context, toPhoneNumber, message string) error {
	if !d.enabled {
		return errors.New("sms: no dispatcher configured")
	}
	d.logger.Printf("sms to=%s body=%q", toPhoneNumber, message)
	return nil
}
