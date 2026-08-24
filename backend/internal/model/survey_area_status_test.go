package model

import "testing"

func TestSurveyAreaHasValidStatus(t *testing.T) {
	valid := SurveyArea{Status: "active"}
	if !valid.HasValidStatus() {
		t.Fatal("active should be valid")
	}
	if (SurveyArea{}).HasValidStatus() {
		t.Fatal("empty status should be invalid")
	}
}
