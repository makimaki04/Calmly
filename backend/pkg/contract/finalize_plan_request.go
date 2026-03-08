package contract

import "github.com/google/uuid"

type FinalizePlanRequest struct {
	PlanID uuid.UUID `json:"plan_id"`
}
