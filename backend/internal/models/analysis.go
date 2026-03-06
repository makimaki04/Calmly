package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/pkg/contract"
)

type DumpAnalysis struct {
	DumpID    uuid.UUID
	Tasks     []Task     `json:"tasks"`
	Questions []Question `json:"questions"`
	Mood      *Mood
	Quote     *string
	CreatedAt time.Time
}

type Mood string

const (
	MoodOverwhelmed Mood = "overwhelmed"
	MoodAnxious     Mood = "anxious"
	MoodTired       Mood = "tired"
	MoodNeutral     Mood = "neutral"
	MoodMotivated   Mood = "motivated"
)

type Task struct {
	Text     string `json:"text"`
	Priority string `json:"priority"`
	Category string `json:"category"`
}

type Question struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

func ConvertToAnalysisDTO(analysis DumpAnalysis) contract.AnalysisDTO {
	var tasksDTO []contract.TaskDTO
	for _, t := range analysis.Tasks {
		task := contract.TaskDTO{
			Text:     t.Text,
			Priority: t.Priority,
			Category: t.Category,
		}

		tasksDTO = append(tasksDTO, task)
	}

	var questionsDTO []contract.QuestionDTO
	for _, q := range analysis.Questions {
		question := contract.QuestionDTO{
			ID:   q.ID,
			Text: q.Text,
		}

		questionsDTO = append(questionsDTO, question)
	}

	analysisDTO := contract.AnalysisDTO{
		DumpID:    analysis.DumpID,
		Tasks:     tasksDTO,
		Questions: questionsDTO,
		Mood:      (*string)(analysis.Mood),
		Quote:     analysis.Quote,
	}

	return analysisDTO
}
