package contract

import (
	"github.com/google/uuid"
)

type AnalysisDTO struct {
	DumpID    uuid.UUID     `json:"dump_id"`
	Tasks     []TaskDTO     `json:"tasks"`
	Questions []QuestionDTO `json:"questions"`
	Mood      *string       `json:"mood"`
	Quote     *string       `json:"quote"`
}

type TaskDTO struct {
	Text     string `json:"text"`
	Priority string `json:"priority"`
	Category string `json:"category"`
}

type QuestionDTO struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}
