package contract

type SubmitAnswersResponse struct {
	Plan      PlanDTO       `json:"plan"`
	PlanItems []PlanItemDTO `json:"plan_items"`
}
