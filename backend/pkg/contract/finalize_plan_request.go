package contract

import "github.com/google/uuid"

type FinalizePlanRequest struct {
	DumpID uuid.UUID `json:"dump_id"`
	PlanID uuid.UUID `json:"plan_id"`
}
