package contract

type RegeneratePlanResponse struct {
	Plan      PlanDTO       `json:"plan"`
	PlanItems []PlanItemDTO `json:"plan_items"`
}
