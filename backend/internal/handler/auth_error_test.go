package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func TestWrongPasswordReturns401(t *testing.T) {
	engine := newAuthErrorEngine(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"wrong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", response.Code, response.Body.String())
	}
}

func TestMalformedBearerReturns401(t *testing.T) {
	engine := newAuthErrorEngine(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer not-a-token")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", response.Code, response.Body.String())
	}
}

func newAuthErrorEngine(t *testing.T) http.Handler {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:auth-error-test?mode=memory&cache=shared",
		JWTSecret:   "auth-error-test-secret-with-more-than-24-characters",
		AutoMigrate: true,
		SeedData:    true,
	}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	support := repository.NewSupportRepository(db)
	audit := service.NewAuditService(support)
	auth := service.NewAuthService(support, configuration.JWTSecret)
	handlers := router.Handlers{Auth: handler.NewAuthHandler(auth, audit)}
	return router.New(log, auth, handlers)
}
