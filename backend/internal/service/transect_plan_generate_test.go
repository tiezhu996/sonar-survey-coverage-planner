package service

import (
	"testing"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/repository"
)

func TestGenerateDoesNotMutateInputHeadings(t *testing.T) {
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:transect-generate-test?mode=memory&cache=shared",
		JWTSecret:   "transect-generate-test-secret-with-more-than-24-chars",
		AutoMigrate: true,
		SeedData:    true,
	}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	areas := repository.NewSurveyAreaRepository(db)
	plans := repository.NewTransectPlanRepository(db)
	audit := NewAuditService(repository.NewSupportRepository(db))
	service := NewTransectPlanService(plans, areas, audit)
	headings := make([]float64, 2, 4)
	headings[0] = 10
	headings[1] = 20
	request := dto.GenerateLinesRequest{
		SurveyAreaID:  1,
		Name:          "批量方向测试",
		Heading:       90,
		LineSpacingM:  100,
		PlannedSwathM: 180,
		Headings:      headings,
	}
	if _, err := service.Generate(request, Actor{UserID: 2, Username: "planner", Role: "survey_planner"}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(headings) != 2 || headings[0] != 10 || headings[1] != 20 {
		t.Fatalf("input headings mutated to %v", headings)
	}
}
