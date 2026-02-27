package models

import (
	"time"

	"github.com/google/uuid"
)

type DumpStatus string

const (
	DumpStatusNew            DumpStatus = "new"
	DumpStatusAnalyzed       DumpStatus = "analyzed"
	DumpStatusWaitingAnswers DumpStatus = "waiting_answers"
	DumpStatusPlanned        DumpStatus = "planned"
	DumpStatusAbandoned      DumpStatus = "abandoned"
)

type Dump struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	GuestID      *uuid.UUID
	Status       DumpStatus
	RawText      *string
	RawDeletedAt *time.Time
	RawExpiresAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Mood string

const (
	MoodOverwhelmed Mood = "overwhelmed"
	MoodAnxious     Mood = "anxious"
	MoodTired       Mood = "tired"
	MoodNeutral     Mood = "neutral"
	MoodMotivated   Mood = "motivated"
)

type DumpAnalysis struct {
	DumpID    uuid.UUID
	Tasks     []Task     `json:"tasks"`
	Questions []Question `json:"questions"`
	Mood      *Mood
	Quote     *string
	CreatedAt time.Time
}

type Task struct {
	Text     string `json:"text"`
	Priority string `json:"priority"`
	Category string `json:"category"`
}

type Question struct {
	Text string `json:"text"`
}

type DumpAnswers struct {
	DumpID    uuid.UUID
	Answers   []Answer `json:"answers"`
	CreatedAt time.Time
}

type Answer struct {
	QuestionIdx int    `json:"question_idx"`
	Text        string `json:"text"`
}

type Plan struct {
	ID        uuid.UUID
	DumpID    uuid.UUID
	Title     string
	CreatedAt time.Time
	SavedAt   *time.Time
	DeletedAt *time.Time
}

type PlanItem struct {
	ID        uuid.UUID
	PlanID    uuid.UUID
	Ord       int
	Text      string
	Done      bool
	CreatedAt time.Time
	DeletedAt *time.Time
}
