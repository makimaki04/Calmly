package models

import (
	"time"

	"github.com/google/uuid"
)

type DumpAnswers struct {
	DumpID    uuid.UUID
	Answers   []Answer `json:"answers"`
	CreatedAt time.Time
}

type Answer struct {
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
}
