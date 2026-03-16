package contract

import (
	"time"

	"github.com/google/uuid"
)

type PlanItemDTO struct {
	ID        uuid.UUID `json:"id"`
	PlanID    uuid.UUID `json:"plan_id"`
	Ord       int       `json:"ord"`
	Text      string    `json:"text"`
	Priority  string    `json:"priority"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
