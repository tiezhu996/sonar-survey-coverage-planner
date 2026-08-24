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

func TestHTTPRequestIDUsesConfiguredPrefix(t *testing.T) {
	t.Setenv("REQUEST_ID_PREFIX", "")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "file:request-id-http-test?mode=memory&cache=shared")
	t.Setenv("JWT_SECRET", "request-id-http-test-secret-with-more-than-24-chars")
	configuration, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
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
	engine := router.New(log, auth, handlers, configuration.RequestIDPrefix)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "abcdefgh")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", response.Code, response.Body.String())
	}
	if !strings.HasPrefix(response.Header().Get("X-Request-ID"), "sonar-") {
		t.Fatalf("request id = %q, want sonar- prefix", response.Header().Get("X-Request-ID"))
	}
}
