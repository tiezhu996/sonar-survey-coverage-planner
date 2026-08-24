package constants

import "testing"

func TestCanMutateAuditorFalse(t *testing.T) {
	if CanMutate(RoleAuditor) {
		t.Fatal("auditor must be read-only")
	}
	if !CanMutate(RoleSurveyPlanner) {
		t.Fatal("survey planner should be able to mutate")
	}
}
