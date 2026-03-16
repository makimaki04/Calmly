package llm

type AnalysisResponse struct {
	Tasks     []AnalysisTaskResponse     `json:"tasks"`
	Questions []AnalysisQuestionResponse `json:"questions"`
	Mood      string                     `json:"mood"`
	Quote     string                     `json:"quote"`
}

type AnalysisTaskResponse struct {
	Text         string `json:"text"`
	TaskPriority string `json:"priority"`
	Category     string `json:"category"`
}

type AnalysisQuestionResponse struct {
	Text string `json:"text"`
}

type SubmitAnswersResponse struct {
	PlanTitle string     `json:"plan_title"`
	Items     []PlanItem `json:"items"`
}

type PlanItem struct {
	Text     string `json:"text"`
	Priority string `json:"priority"`
}
