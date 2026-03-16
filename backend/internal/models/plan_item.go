package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/pkg/contract"
)

type PlanItem struct {
	ID        uuid.UUID
	PlanID    uuid.UUID
	Ord       int
	Text      string
	Priority  string
	Done      bool
	CreatedAt time.Time
	DeletedAt *time.Time
}

func ConvertToPlanItemsDTO(items []PlanItem) []contract.PlanItemDTO {
	var itemsDTO []contract.PlanItemDTO
	for _, i := range items {
		itemDTO := contract.PlanItemDTO{
			ID:        i.ID,
			PlanID:    i.PlanID,
			Ord:       i.Ord,
			Text:      i.Text,
			Priority:  i.Priority,
			Done:      i.Done,
			CreatedAt: i.CreatedAt,
		}

		itemsDTO = append(itemsDTO, itemDTO)
	}

	return itemsDTO
}
