package config

import (
	"os"
	"testing"
)

func TestConfigRequestIDPrefixDefault(t *testing.T) {
	_ = os.Setenv("REQUEST_ID_PREFIX", "")
	t.Cleanup(func() { _ = os.Unsetenv("REQUEST_ID_PREFIX") })
	_ = os.Setenv("DB_DRIVER", "sqlite")
	_ = os.Setenv("DB_DSN", "file:request-id-config-test?mode=memory&cache=shared")
	_ = os.Setenv("JWT_SECRET", "request-id-config-test-secret-with-more-than-24-chars")
	configuration, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if configuration.RequestIDPrefix != "sonar" {
		t.Fatalf("request id prefix = %q, want sonar", configuration.RequestIDPrefix)
	}
}
