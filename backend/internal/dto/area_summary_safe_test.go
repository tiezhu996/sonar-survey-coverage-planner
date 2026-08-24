package dto

import "testing"

func TestAreaSummarySafeLatestCoverage(t *testing.T) {
	summary := AreaSummary{LatestCoverage: 0.32}
	if summary.SafeLatestCoverage() != 0.32 {
		t.Fatalf("safe coverage = %v, want 0.32", summary.SafeLatestCoverage())
	}
	if (AreaSummary{}).SafeLatestCoverage() != 0 {
		t.Fatal("empty summary should return zero safely")
	}
}
