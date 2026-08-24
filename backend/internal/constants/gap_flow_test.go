package constants

import "testing"

func TestDetectedGapMustReviewFirst(t *testing.T) {
	if GapDetected.CanTransition(GapClosed) {
		t.Fatal("detected -> closed must be rejected")
	}
	if GapDetected.CanTransition(GapFalsePositive) {
		t.Fatal("detected -> false_positive must be rejected")
	}
}
