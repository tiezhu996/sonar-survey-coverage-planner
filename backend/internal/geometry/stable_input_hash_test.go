package geometry

import "testing"

func TestStableInputHashDoesNotMutateChecksums(t *testing.T) {
	checksums := []string{"beta", "alpha"}
	original := append([]string(nil), checksums...)
	_ = StableInputHash(7, "EPSG:32650", "grid-v1", checksums, 20)
	if checksums[0] != original[0] || checksums[1] != original[1] {
		t.Fatalf("checksums mutated: %v, want %v", checksums, original)
	}
}
