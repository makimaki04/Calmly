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
	Tasks     []Task
	Questions []Question
	Mood      *Mood
	Quote     *string
	CreatedAt time.Time
}

type Task struct {
	Text     string
	Priority string
	Category string
}

type Question struct {
	Text string
}

type DumpAnswers struct {
	DumpID    uuid.UUID
	Answers   []Answer
	CreatedAt time.Time
}

type Answer struct {
	QuestionIdx int
	Text        string
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
