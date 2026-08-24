package constants

import "testing"

func TestProcessedRunCanSupersede(t *testing.T) {
	if !RunProcessed.CanTransition(RunSuperseded) {
		t.Fatal("processed -> superseded should be allowed")
	}
}

func TestRejectedRunCannotProcess(t *testing.T) {
	if RunRejected.CanTransition(RunProcessing) {
		t.Fatal("rejected -> processing must be rejected")
	}
}
