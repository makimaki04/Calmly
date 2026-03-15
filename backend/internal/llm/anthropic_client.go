package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

type AnthropicClient struct {
	client anthropic.Client
	logger *zap.Logger
}

func NewAnthropicClient(apiKey string, logger *zap.Logger) *AnthropicClient {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &AnthropicClient{
		client: client,
		logger: logger.With(zap.String("component", "llm_client")),
	}
}

func (c *AnthropicClient) GenerateAnalysis(ctx context.Context, rawText string) (models.DumpAnalysis, error) {
	log := c.logger.With(
		zap.String("operation", "generate_analysis"),
		zap.Int("raw_text_len", len(rawText)),
	)

	log.Info("Generate analysis started")

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	systemPrompt, userPrompt := BuildAnalysisPrompt(rawText)

	message, err := c.client.Messages.New(
		ctx,
		anthropic.MessageNewParams{
			MaxTokens: 600,
			Model:     anthropic.ModelClaudeSonnet4_6,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
			},
		},
	)
	if err != nil {
		log.Error("Generate analysis failed", zap.Error(err))
		return models.DumpAnalysis{}, fmt.Errorf("request analysis from anthropic: %w", err)
	}

	content := message.Content
	parsedLLMText, err := ParseLLMResponse(content)
	if err != nil {
		log.Error("parse llm response to string", zap.Error(err))
		return models.DumpAnalysis{}, err
	}

	extractedLLMText, err := ExtractJSONCandidate(parsedLLMText)
	if err != nil {
		log.Debug("invalid llm text after extraction", zap.String("llm raw text", string(parsedLLMText)))
		return models.DumpAnalysis{}, err
	}

	llmAnalysis, err := decodeJSON[AnalysisResponse](extractedLLMText)
	if err != nil {
		log.Error("decode extracted llm text", zap.Error(err))
		return models.DumpAnalysis{}, fmt.Errorf("decode JSON: %w", err)
	}

	if err := validateAndNormalizeAnalysisResponse(&llmAnalysis); err != nil {
		log.Error("validate analysis response", zap.Error(err))
		return models.DumpAnalysis{}, err
	}

	analysis := mapAnalysisResponseToDomain(llmAnalysis)

	log.Info(
		"Analysis generated",
		zap.Int("tasks_count", len(analysis.Tasks)),
		zap.Int("questions_count", len(analysis.Questions)),
	)

	return analysis, nil
}

var (
	ErrInvalidTaskPriority = errors.New("invalid task priority value")
	ErrEmptyTaskFields     = errors.New("task contains empty required fields")
	ErrEmptyQuestionText   = errors.New("question text is empty")
	ErrInvalidMoodValue    = errors.New("invalid mood value")
	ErrEmptyQuote          = errors.New("empty quote value")
)

func validateAndNormalizeAnalysisResponse(llmAnalysis *AnalysisResponse) error {
	for i, t := range llmAnalysis.Tasks {
		t.Text = strings.TrimSpace(t.Text)
		t.TaskPriority = strings.TrimSpace(t.TaskPriority)
		t.Category = strings.TrimSpace(t.Category)

		if t.Text == "" || t.TaskPriority == "" || t.Category == "" {
			return ErrEmptyTaskFields
		}

		switch t.TaskPriority {
		case "low":
		case "medium":
		case "high":
		default:
			return ErrInvalidTaskPriority
		}

		llmAnalysis.Tasks[i] = t
	}

	for i, q := range llmAnalysis.Questions {
		q.Text = strings.TrimSpace(q.Text)
		if q.Text == "" {
			return ErrEmptyQuestionText
		}

		llmAnalysis.Questions[i] = q
	}

	llmAnalysis.Quote = strings.TrimSpace(llmAnalysis.Quote)
	if llmAnalysis.Quote == "" {
		return ErrEmptyQuote
	}

	llmAnalysis.Mood = strings.TrimSpace(llmAnalysis.Mood)
	switch models.Mood(llmAnalysis.Mood) {
	case models.MoodAnxious:
	case models.MoodMotivated:
	case models.MoodNeutral:
	case models.MoodOverwhelmed:
	case models.MoodTired:
	default:
		return ErrInvalidMoodValue
	}

	return nil
}

func mapAnalysisResponseToDomain(llmAnalysis AnalysisResponse) models.DumpAnalysis {
	var tasks []models.Task
	for _, t := range llmAnalysis.Tasks {
		task := models.Task{
			Text:     t.Text,
			Priority: t.TaskPriority,
			Category: t.Category,
		}

		tasks = append(tasks, task)
	}

	var questions []models.Question
	for _, q := range llmAnalysis.Questions {
		question := models.Question{
			Text: q.Text,
		}

		questions = append(questions, question)
	}

	var analysis models.DumpAnalysis
	analysis.Tasks = tasks
	analysis.Questions = questions

	quote := llmAnalysis.Quote
	analysis.Quote = &quote

	switch models.Mood(llmAnalysis.Mood) {
	case models.MoodAnxious:
		m := models.MoodAnxious
		analysis.Mood = &m
	case models.MoodMotivated:
		m := models.MoodMotivated
		analysis.Mood = &m
	case models.MoodNeutral:
		m := models.MoodNeutral
		analysis.Mood = &m
	case models.MoodOverwhelmed:
		m := models.MoodOverwhelmed
		analysis.Mood = &m
	case models.MoodTired:
		m := models.MoodTired
		analysis.Mood = &m
	}

	return analysis
}

func (c *AnthropicClient) GeneratePlan(
	ctx context.Context,
	rawText string,
	questions []models.Question,
	answers []models.Answer,
) (models.Plan, []models.PlanItem, error) {

	return models.Plan{}, []models.PlanItem{}, nil
}

func (c *AnthropicClient) RegeneratePlan(
	ctx context.Context,
	rawtext string,
	feedback string,
) (models.Plan, []models.PlanItem, error) {
	return models.Plan{}, []models.PlanItem{}, nil
}

var ErrEmptyLLMResponse = errors.New("empty llm response")

func ParseLLMResponse(content []anthropic.ContentBlockUnion) (string, error) {
	var builder strings.Builder
	for _, block := range content {
		text := strings.TrimSpace(block.Text)

		if text == "" {
			continue
		}

		builder.WriteString(text)
		builder.WriteByte('\n')
	}

	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", ErrEmptyLLMResponse
	}

	return result, nil
}

var ErrInvalidLLMResponse = errors.New("invalid llm response format")

func ExtractJSONCandidate(text string) ([]byte, error) {
	trimText := strings.TrimSpace(text)

	if strings.HasPrefix(trimText, "{") && json.Valid([]byte(trimText)) {
		return []byte(trimText), nil
	}

	for i := 0; i < len(trimText); i++ {
		if trimText[i] != '{' {
			continue
		}

		depth := 0
		inString := false
		escaped := false

		for j := i; j < len(trimText); j++ {
			ch := trimText[j]

			if !inString {
				switch ch {
				case '"':
					inString = true
				case '{':
					depth++
				case '}':
					depth--
					if depth == 0 {
						candidate := trimText[i : j+1]
						candidate = strings.TrimSpace(candidate)

						if json.Valid([]byte(candidate)) {
							return []byte(candidate), nil
						} else {
							break
						}
					} else if depth < 0 {
						break
					}
				}

				continue
			}

			if escaped {
				escaped = false
				continue
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
		}
	}

	return nil, ErrInvalidLLMResponse
}

func decodeJSON[T any](text []byte) (T, error) {
	var result T
	if err := json.Unmarshal(text, &result); err != nil {
		return result, err
	}

	return result, nil
}
