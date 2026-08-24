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

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func TestGenerateRejectsTooManyHeadings(t *testing.T) {
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:transect-heading-limit-test?mode=memory&cache=shared",
		JWTSecret:   "transect-heading-limit-test-secret-with-more-than-24-characters",
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
		Plan: handler.NewTransectPlanHandler(service.NewTransectPlanService(plans, areas, audit)),
	}
	engine := router.New(log, auth, handlers)
	login, err := auth.Login(dto.LoginRequest{Username: "planner", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	headings := make([]float64, 21)
	for index := range headings {
		headings[index] = float64(index * 10)
	}
	headingsJSON, _ := json.Marshal(headings)
	payload := fmt.Sprintf(`{"survey_area_id":1,"name":"方向过多测试","heading":90,"line_spacing_m":100,"planned_swath_m":180,"headings":%s}`, headingsJSON)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/plans/generate", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+login.Token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", response.Code, response.Body.String())
	}
}
