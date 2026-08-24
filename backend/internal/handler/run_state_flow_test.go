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

type runStateTestEnv struct {
	engine *gin.Engine
	token  string
}

func newRunStateEngine(t *testing.T) runStateTestEnv {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_")),
		JWTSecret:   "run-state-test-secret-with-more-than-24-characters",
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
	response, err := auth.Login(dto.LoginRequest{Username: "processor", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return runStateTestEnv{engine: engine, token: response.Token}
}

func (e runStateTestEnv) request(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+e.token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	e.engine.ServeHTTP(response, request)
	return response
}

func importRunForStateTest(t *testing.T, env runStateTestEnv, runCode string) uint {
	t.Helper()
	body := `{"transect_plan_id":1,"run_code":"` + runCode + `","track_geojson":{"type":"Feature","properties":{},"geometry":{"type":"MultiLineString","coordinates":[[[0,0],[10,0]]]}},"actual_swath_m":180,"started_at":"2026-08-01T00:00:00Z","ended_at":"2026-08-01T01:00:00Z","navigation_quality":"good"}`
	response := env.request(t, http.MethodPost, "/api/v1/runs/import", body)
	if response.Code != http.StatusCreated {
		t.Fatalf("import status = %d, body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			Run struct {
				ID uint `json:"id"`
			} `json:"run"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode import: %v", err)
	}
	if payload.Data.Run.ID == 0 {
		t.Fatalf("imported run id missing: %s", response.Body.String())
	}
	return payload.Data.Run.ID
}

func transitionRun(t *testing.T, env runStateTestEnv, id uint, version uint, target string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"target_state": target, "expected_version": version, "reason": "状态流测试"})
	return env.request(t, http.MethodPost, fmt.Sprintf("/api/v1/runs/%d/transition", id), string(body))
}

func TestProcessedRunSupersedeViaAPI(t *testing.T) {
	env := newRunStateEngine(t)
	id := importRunForStateTest(t, env, "RUN-STATE-GREEN-005")
	if response := transitionRun(t, env, id, 1, "quality_checked"); response.Code != http.StatusOK {
		t.Fatalf("quality_checked status=%d body=%s", response.Code, response.Body.String())
	}
	if response := transitionRun(t, env, id, 2, "processing"); response.Code != http.StatusOK {
		t.Fatalf("processing status=%d body=%s", response.Code, response.Body.String())
	}
	if response := transitionRun(t, env, id, 3, "processed"); response.Code != http.StatusOK {
		t.Fatalf("processed status=%d body=%s", response.Code, response.Body.String())
	}
	response := transitionRun(t, env, id, 4, "superseded")
	if response.Code != http.StatusOK {
		t.Fatalf("superseded status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestSupersedeOnlyFromProcessed(t *testing.T) {
	env := newRunStateEngine(t)
	id := importRunForStateTest(t, env, "RUN-STATE-SUPERSEDE-005")
	if response := transitionRun(t, env, id, 1, "quality_checked"); response.Code != http.StatusOK {
		t.Fatalf("quality_checked status=%d body=%s", response.Code, response.Body.String())
	}
	response := transitionRun(t, env, id, 2, "superseded")
	if response.Code != http.StatusConflict {
		t.Fatalf("superseded from quality_checked status=%d, want 409; body=%s", response.Code, response.Body.String())
	}
}
