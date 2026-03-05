package contract

import "github.com/google/uuid"

type SubmitAnswersRequest struct {
	DumpID  uuid.UUID   `json:"dump_id"`
	Answers []AnswerDTO `json:"answers"`
}
