package contract

import (
	"time"

	"github.com/google/uuid"
)

type PlanDTO struct {
	ID        uuid.UUID  `json:"id"`
	DumpID    uuid.UUID  `json:"dump_id"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	SavedAt   *time.Time `json:"saved_at"`
}
