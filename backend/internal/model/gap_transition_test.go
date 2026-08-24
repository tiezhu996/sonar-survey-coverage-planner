package model

import "testing"

func TestDetectedGapModelRejectsDirectTerminal(t *testing.T) {
	gap := CoverageGap{GapState: "detected"}
	if gap.CanTransitionTo("closed") {
		t.Fatal("detected model must reject closed")
	}
	if gap.CanTransitionTo("false_positive") {
		t.Fatal("detected model must reject false_positive")
	}
	gap.GapState = "reviewed"
	if !gap.CanTransitionTo("false_positive") {
		t.Fatal("reviewed model should allow false_positive")
	}
}
