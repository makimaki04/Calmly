package contract

type SubmitAnswersRequest struct {
	Answers []AnswerDTO `json:"answers"`
}
