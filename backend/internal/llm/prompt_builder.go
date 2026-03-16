package llm

import "strings"

func BuildAnalysisPrompt(rawText string) (string, string) {
	systemPrompt := SystemAnalysisPrompt + "\n\n" + StructuredOutputInstructionAnalysis
	userPrompt := strings.Replace(UserPromptTemplate, "{{raw_text}}", rawText, 1)

	return systemPrompt, userPrompt
}

func BuildAnswersPrompt(rawText string, analysisTasks string, questions string, answers string) (string, string) {
	systemPrompt := SystemSubmitAnswersPrompt + "\n\n" + StructuredOutputInstructionAnswers

	userPrompt := strings.Replace(UserAnswersPromptTamplate, "{{raw_text}}", rawText, 1)
	userPrompt = strings.Replace(userPrompt, "{{analysis_tasks}}", analysisTasks, 1)
	userPrompt = strings.Replace(userPrompt, "{{questions}}", questions, 1)
	userPrompt = strings.Replace(userPrompt, "{{answers}}", answers, 1)

	return systemPrompt, userPrompt
}
