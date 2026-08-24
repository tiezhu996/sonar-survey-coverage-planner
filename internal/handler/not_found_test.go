package handler_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func newNotFoundEngine(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_")),
		JWTSecret:   "not-found-test-secret-with-more-than-24-characters",
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
	plans := repository.NewTransectPlanRepository(db)
	audit := service.NewAuditService(support)
	auth := service.NewAuthService(support, configuration.JWTSecret)
	handlers := router.Handlers{
		Auth: handler.NewAuthHandler(auth, audit),
		Area: handler.NewSurveyAreaHandler(service.NewSurveyAreaService(areas, audit)),
		Plan: handler.NewTransectPlanHandler(service.NewTransectPlanService(plans, areas, audit)),
	}
	engine := router.New(log, auth, handlers)
	response, err := auth.Login(dto.LoginRequest{Username: "admin", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return engine, response.Token
}

func performRequest(t *testing.T, engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func TestMissingSurveyAreaReturnsNotFound(t *testing.T) {
	engine, token := newNotFoundEngine(t)
	response := performRequest(t, engine, http.MethodGet, "/api/v1/areas/999999", token, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != "RESOURCE_NOT_FOUND" {
		t.Fatalf("error code = %q, want RESOURCE_NOT_FOUND", payload.Error.Code)
	}
}

func TestMissingTransectPlanReturnsNotFound(t *testing.T) {
	engine, token := newNotFoundEngine(t)
	response := performRequest(t, engine, http.MethodGet, "/api/v1/plans/999999", token, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", response.Code, response.Body.String())
	}
}

func TestMissingTransectPlanUpdateReturnsNotFound(t *testing.T) {
	engine, token := newNotFoundEngine(t)
	body := `{"name":"不存在的测线","line_geojson":{"type":"Feature","properties":{},"geometry":{"type":"LineString","coordinates":[[0,0],[1,1]]}},"planned_heading":90,"planned_swath_m":180,"line_spacing_m":100,"expected_version":1}`
	response := performRequest(t, engine, http.MethodPut, "/api/v1/plans/999999", token, body)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", response.Code, response.Body.String())
	}
}
