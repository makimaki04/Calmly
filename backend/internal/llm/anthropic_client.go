package llm

import (
	"context"
	"encoding/json"
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
		return models.DumpAnalysis{}, err
	}

	content := message.Content
	var rawLLMText string
	for _, block := range content {
		rawLLMText += block.Text
	}

	var llmAnalysis AnalysisResponse
	if err := json.Unmarshal([]byte(rawLLMText), &llmAnalysis); err != nil {
		return models.DumpAnalysis{}, err
	}

	return models.DumpAnalysis{}, nil
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
