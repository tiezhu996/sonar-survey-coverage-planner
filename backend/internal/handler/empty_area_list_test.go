package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

func TestEmptyAreaListNoPanic(t *testing.T) {
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:empty-area-list-test?mode=memory&cache=shared",
		JWTSecret:   "empty-area-list-test-secret-with-more-than-24-characters",
		AutoMigrate: true,
		SeedData:    true,
	}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.Exec("DELETE FROM coverage_gaps").Error; err != nil {
		t.Fatalf("delete gaps: %v", err)
	}
	if err := db.Exec("DELETE FROM sonar_runs").Error; err != nil {
		t.Fatalf("delete runs: %v", err)
	}
	if err := db.Exec("DELETE FROM transect_plans").Error; err != nil {
		t.Fatalf("delete plans: %v", err)
	}
	if err := db.Exec("DELETE FROM survey_areas").Error; err != nil {
		t.Fatalf("delete areas: %v", err)
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
	login, err := auth.Login(dto.LoginRequest{Username: "admin", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/areas", nil)
	request.Header.Set("Authorization", "Bearer "+login.Token)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", response.Code, response.Body.String())
	}
}
