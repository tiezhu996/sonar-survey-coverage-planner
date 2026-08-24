package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/handler"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
	"sonar-survey-coverage-planner/backend/internal/router"
	"sonar-survey-coverage-planner/backend/internal/service"
)

type concurrentImportEnv struct {
	engine *gin.Engine
	db     *gorm.DB
	token  string
}

func newConcurrentImportEngine(t *testing.T) concurrentImportEnv {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:concurrent-import-test?mode=memory&cache=shared",
		JWTSecret:   "concurrent-import-test-secret-with-more-than-24-characters",
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
	runs := repository.NewSonarRunRepository(db)
	audit := service.NewAuditService(support)
	auth := service.NewAuthService(support, configuration.JWTSecret)
	handlers := router.Handlers{
		Auth: handler.NewAuthHandler(auth, audit),
		Area: handler.NewSurveyAreaHandler(service.NewSurveyAreaService(areas, audit)),
		Plan: handler.NewTransectPlanHandler(service.NewTransectPlanService(plans, areas, audit)),
		Run:  handler.NewSonarRunHandler(service.NewSonarRunService(runs, plans, audit)),
	}
	engine := router.New(log, auth, handlers)
	login, err := auth.Login(dto.LoginRequest{Username: "processor", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return concurrentImportEnv{engine: engine, db: db, token: login.Token}
}

func TestConcurrentImportSameChecksumOneIdempotent(t *testing.T) {
	env := newConcurrentImportEngine(t)
	body := `{"transect_plan_id":1,"run_code":"RUN-CONCURRENT-001","track_geojson":{"type":"Feature","properties":{},"geometry":{"type":"MultiLineString","coordinates":[[[0,0],[10,0]]]}},"actual_swath_m":180,"started_at":"2026-08-01T00:00:00Z","ended_at":"2026-08-01T01:00:00Z","navigation_quality":"good"}`
	const workers = 6
	start := make(chan struct{})
	statuses := make([]int, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for index := 0; index < workers; index++ {
		go func(index int) {
			defer wait.Done()
			<-start
			request := httptest.NewRequest(http.MethodPost, "/api/v1/runs/import", strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer "+env.token)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			env.engine.ServeHTTP(response, request)
			statuses[index] = response.Code
		}(index)
	}
	close(start)
	wait.Wait()
	for _, status := range statuses {
		if status != http.StatusCreated && status != http.StatusOK {
			t.Fatalf("unexpected import status %d (statuses=%v)", status, statuses)
		}
	}
	var count int64
	if err := env.db.Model(&model.SonarRun{}).Where("run_code = ?", "RUN-CONCURRENT-001").Count(&count).Error; err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if count != 1 {
		t.Fatalf("run count = %d, want 1; statuses=%v", count, statuses)
	}
}

func TestConcurrentRateLimitWindowNoRace(t *testing.T) {
	env := newConcurrentImportEngine(t)
	body := `{"transect_plan_id":1,"run_code":"RUN-CONCURRENT-LIMIT-001","track_geojson":{"type":"Feature","properties":{},"geometry":{"type":"MultiLineString","coordinates":[[[0,0],[10,0]]]}},"actual_swath_m":180,"started_at":"2026-08-01T00:00:00Z","ended_at":"2026-08-01T01:00:00Z","navigation_quality":"good"}`
	start := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(4)
	for index := 0; index < 4; index++ {
		go func() {
			defer wait.Done()
			<-start
			request := httptest.NewRequest(http.MethodPost, "/api/v1/runs/import", strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer "+env.token)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			env.engine.ServeHTTP(response, request)
			_ = response.Code
		}()
	}
	close(start)
	wait.Wait()
}
