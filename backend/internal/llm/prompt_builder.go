package llm

import "strings"

func BuildAnalysisPrompt(rawText string) (string, string) {
	systemPrompt := SystemAnalysisPrompt + "\n\n" + StructuredOutputInstruction
	userPrompt := strings.Replace(UserPromptTemplate, "{{raw_text}}", rawText, 1)

	return systemPrompt, userPrompt
}
