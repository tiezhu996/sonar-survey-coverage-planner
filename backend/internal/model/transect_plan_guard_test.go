package model

import "testing"

func TestTransectPlanCanCreatePlanForArea(t *testing.T) {
	plan := TransectPlan{}
	if plan.CanCreatePlanForArea("archived") {
		t.Fatal("archived area must not accept new plans")
	}
	if !plan.CanCreatePlanForArea("active") {
		t.Fatal("active area should accept new plans")
	}
}
