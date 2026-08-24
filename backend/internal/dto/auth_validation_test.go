package dto

import "testing"

func TestLoginRequestValidationRejectsWeakCredentials(t *testing.T) {
	request := LoginRequest{Username: "  ", Password: "short"}
	if err := request.Validate(); err == nil {
		t.Fatal("blank username must be rejected")
	}
	request = LoginRequest{Username: "admin", Password: "short"}
	if err := request.Validate(); err == nil {
		t.Fatal("short password must be rejected")
	}
	if err := (LoginRequest{Username: "admin", Password: "long-enough"}).Validate(); err != nil {
		t.Fatalf("valid login rejected: %v", err)
	}
}
