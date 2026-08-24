package dto

import "testing"

func TestGapTransitionValidationRejectsDirectClose(t *testing.T) {
	request := GapTransitionRequest{TargetState: "closed"}
	if err := request.ValidateDirectTerminal("detected"); err == nil {
		t.Fatal("detected -> closed must be rejected")
	}
	if err := request.ValidateDirectTerminal("reviewed"); err != nil {
		t.Fatalf("reviewed -> closed rejected: %v", err)
	}
}

func TestGapTransitionValidationRejectsDirectFalsePositive(t *testing.T) {
	request := GapTransitionRequest{TargetState: "false_positive"}
	if err := request.ValidateDirectTerminal("detected"); err == nil {
		t.Fatal("detected -> false_positive must be rejected")
	}
}
