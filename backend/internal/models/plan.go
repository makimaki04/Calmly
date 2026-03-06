package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/pkg/contract"
)

type Plan struct {
	ID        uuid.UUID
	DumpID    uuid.UUID
	Title     string
	CreatedAt time.Time
	SavedAt   *time.Time
	DeletedAt *time.Time
}

func ConvertToPlanDTO(plan Plan) contract.PlanDTO {
	planDTO := contract.PlanDTO{
		ID:        plan.ID,
		DumpID:    plan.DumpID,
		Title:     plan.Title,
		CreatedAt: plan.CreatedAt,
		SavedAt:   plan.SavedAt,
	}

	return planDTO
}
