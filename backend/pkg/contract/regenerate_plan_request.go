package contract

import "github.com/google/uuid"

type RegeneratePlanRequest struct {
	DumpID   uuid.UUID `json:"dump_id"`
	Feedback string    `json:"feedback"`
}
