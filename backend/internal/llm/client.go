package llm

import (
	"context"

	"github.com/makimaki04/Calmly/internal/models"
)

type AnalysisGenerator interface {
	GenerateAnalysis(ctx context.Context, rawText string) (models.DumpAnalysis, error)
}

type PlanGenerator interface {
	GeneratePlan(ctx context.Context, rawText string, tasks []models.Task, questions []models.Question, answers []models.Answer) (models.Plan, []models.PlanItem, error)
	RegeneratePlan(ctx context.Context, rawtext string,  feedback string) (models.Plan, []models.PlanItem, error)
}