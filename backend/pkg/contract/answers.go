package contract

import "github.com/google/uuid"

type AnswerDTO struct {
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
}
