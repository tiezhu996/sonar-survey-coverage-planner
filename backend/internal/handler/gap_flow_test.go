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

type gapTestEnv struct {
	engine         *gin.Engine
	token          string
	reviewerToken string
}

func newGapTestEngine(t *testing.T) gapTestEnv {
	t.Helper()
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_")),
		JWTSecret:   "gap-state-test-secret-with-more-than-24-characters",
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
	gaps := repository.NewCoverageGapRepository(db)
	audit := service.NewAuditService(support)
	auth := service.NewAuthService(support, configuration.JWTSecret)
	handlers := router.Handlers{
		Auth:     handler.NewAuthHandler(auth, audit),
		Area:     handler.NewSurveyAreaHandler(service.NewSurveyAreaService(areas, audit)),
		Plan:     handler.NewTransectPlanHandler(service.NewTransectPlanService(plans, areas, audit)),
		Run:      handler.NewSonarRunHandler(service.NewSonarRunService(runs, plans, audit)),
		Coverage: handler.NewCoverageGapHandler(service.NewCoverageGapService(gaps, areas, runs, audit)),
	}
	engine := router.New(log, auth, handlers)
	processor, err := auth.Login(dto.LoginRequest{Username: "processor", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login processor: %v", err)
	}
	reviewer, err := auth.Login(dto.LoginRequest{Username: "reviewer", Password: "Sonar2026!"})
	if err != nil {
		t.Fatalf("login reviewer: %v", err)
	}
	return gapTestEnv{engine: engine, token: processor.Token, reviewerToken: reviewer.Token}
}

func (e gapTestEnv) request(t *testing.T, method, path, body, idempotency, token string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	if idempotency != "" {
		request.Header.Set("Idempotency-Key", idempotency)
	}
	response := httptest.NewRecorder()
	e.engine.ServeHTTP(response, request)
	return response
}

func detectGapForTest(t *testing.T, env gapTestEnv, key string) uint {
	t.Helper()
	body := `{"survey_area_id":1,"source_run_ids":[1],"algorithm_version":"grid-v1","resolution_m":20}`
	response := env.request(t, http.MethodPost, "/api/v1/coverage-gaps/detect", body, key, env.token)
	if response.Code != http.StatusCreated {
		t.Fatalf("detect status = %d, body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			Gap struct {
				ID uint `json:"id"`
			} `json:"gap"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode detect: %v", err)
	}
	if payload.Data.Gap.ID == 0 {
		t.Fatalf("gap id missing: %s", response.Body.String())
	}
	return payload.Data.Gap.ID
}

func transitionGap(t *testing.T, env gapTestEnv, id uint, target string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"target_state": target, "expected_version": 1, "review_note": "状态流测试需要拦截"})
	return env.request(t, http.MethodPost, fmt.Sprintf("/api/v1/coverage-gaps/%d/transition", id), string(body), "", env.reviewerToken)
}

func TestDetectedGapCannotCloseViaAPI(t *testing.T) {
	env := newGapTestEngine(t)
	id := detectGapForTest(t, env, "detect-006-close-key")
	response := transitionGap(t, env, id, "closed")
	if response.Code != http.StatusConflict {
		t.Fatalf("detected -> closed status=%d, want 409; body=%s", response.Code, response.Body.String())
	}
}

func TestDetectedGapCannotFalsePositiveViaAPI(t *testing.T) {
	env := newGapTestEngine(t)
	id := detectGapForTest(t, env, "detect-006-false-key")
	response := transitionGap(t, env, id, "false_positive")
	if response.Code != http.StatusConflict {
		t.Fatalf("detected -> false_positive status=%d, want 409; body=%s", response.Code, response.Body.String())
	}
}

func TestGapReviewNotePersistsInExplanation(t *testing.T) {
	env := newGapTestEngine(t)
	id := detectGapForTest(t, env, "detect-006-review-note")
	response := transitionGap(t, env, id, "reviewed")
	if response.Code != http.StatusOK {
		t.Fatalf("reviewed status=%d, want 200; body=%s", response.Code, response.Body.String())
	}
	detail := env.request(t, http.MethodGet, fmt.Sprintf("/api/v1/coverage-gaps/%d", id), "", "", env.reviewerToken)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status=%d, want 200; body=%s", detail.Code, detail.Body.String())
	}
	var payload struct {
		Data struct {
			Explanation string `json:"explanation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if !strings.Contains(payload.Data.Explanation, "状态流测试需要拦截") {
		t.Fatalf("explanation = %q, want review note appended", payload.Data.Explanation)
	}
}
