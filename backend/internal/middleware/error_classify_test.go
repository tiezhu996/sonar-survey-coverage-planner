package middleware

import (
	"errors"
	"net/http"
	"testing"

	"sonar-survey-coverage-planner/backend/pkg/api"
)

func TestErrorHandlerClassifiesClientError(t *testing.T) {
	if !IsClientError(api.NewError(http.StatusBadRequest, "TEST", "bad", nil)) {
		t.Fatal("4xx app error should be classified as client error")
	}
	if IsClientError(errors.New("boom")) {
		t.Fatal("plain error must not be classified as client error")
	}
}
