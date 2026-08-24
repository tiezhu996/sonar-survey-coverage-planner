package dto

import "testing"

func TestCloneHeadingsCopiesInput(t *testing.T) {
	request := GenerateLinesRequest{Headings: []float64{10, 20}}
	clone := request.CloneHeadings()
	clone = append(clone, 30)
	if len(request.Headings) != 2 {
		t.Fatalf("input headings mutated to %v", request.Headings)
	}
}
