package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func TestUnauthorizedIs401(t *testing.T) {
	engine, _ := newStatusCodeEngine(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"wrong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", response.Code, response.Body.String())
	}
}

func TestForbiddenIs403(t *testing.T) {
	engine, token := newStatusCodeEngine(t)
	body := `{"area_code":"AUD-CREATE-010","name":"审计员越权创建","boundary_geojson":{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[[[0,0],[100,0],[100,100],[0,100],[0,0]]]}},"target_resolution_m":20,"coordinate_system":"EPSG:32650","default_swath_m":180,"owner_team":"测试组"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/areas", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", response.Code, response.Body.String())
	}
}

func newStatusCodeEngine(t *testing.T) (http.Handler, string) {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:status-code-test?mode=memory&cache=shared",
		JWTSecret:   "status-code-test-secret-with-more-than-24-characters",
		AutoMigrate: true,
		SeedData:    true,
	}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	support := repository.NewSupportRepository(db)
	areas := repository.NewSurveyAreaRepository(db)
	audit := service.NewAuditService(support)
	auth := service.NewAuthService(support, configuration.JWTSecret)
	handlers := router.Handlers{
		Auth: handler.NewAuthHandler(auth, audit),
		Area: handler.NewSurveyAreaHandler(service.NewSurveyAreaService(areas, audit)),
	}
	engine := router.New(log, auth, handlers)
	login, err := auth.Login(dto.LoginRequest{Username: "auditor", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return engine, login.Token
}
