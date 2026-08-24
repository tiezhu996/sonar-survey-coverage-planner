package service

import (
	"testing"

	"sonar-survey-coverage-planner/backend/internal/config"
	"sonar-survey-coverage-planner/backend/internal/model"
	"sonar-survey-coverage-planner/backend/internal/repository"
)

func TestAuditRecordMissingRequestIDNoPanic(t *testing.T) {
	configuration := config.Config{
		DBDriver:    "sqlite",
		DBDSN:       "file:audit-request-id-test?mode=memory&cache=shared",
		JWTSecret:   "audit-request-id-test-secret-with-more-than-24-characters",
		AutoMigrate: true,
		SeedData:    true,
	}
	db, err := config.OpenDatabase(configuration)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	service := NewAuditService(repository.NewSupportRepository(db))
	if err := service.Record(Actor{}, "audit.missing-id", "test_entity", 1, nil, model.AuditEvent{}, map[string]any{"note": "fallback"}); err != nil {
		t.Fatalf("record with missing request id: %v", err)
	}
}
