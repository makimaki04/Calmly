package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

type Dump interface {
	CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error)
	AbandonDump(ctx context.Context, dumpID uuid.UUID) error
}

type Analyze interface {
	SaveDumpAnalysis(ctx context.Context, analysis models.DumpAnalysis) error
	GetDumpAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error)
}

type Plan interface {
	CreatePlan(ctx context.Context, dumpID uuid.UUID, title string) (uuid.UUID, error)
	SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error
	GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error)
	GetUserSavedPlans(ctx context.Context, userID uuid.UUID) ([]models.Plan, error)
	DeleteSavedPlan(ctx context.Context, planID uuid.UUID) error
}

type PlanItem interface {
	CreateItems(ctx context.Context, items []models.PlanItem) error
	ToggleItem(ctx context.Context, itemID uuid.UUID, done bool) error
	AddItem(ctx context.Context, item models.PlanItem) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error
	ReorderItems(ctx context.Context, planID uuid.UUID, itemIDs []uuid.UUID) error
	GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error)
}

type Service struct {
	Dump
	Analyze
	Plan
	PlanItem
}

func NewService(repo *repository.Repository, dumpExpTime time.Duration, logger *zap.Logger) *Service {
	return &Service{
		Dump:     NewDumpService(repo.Dump, dumpExpTime, logger),
		Analyze:  NewAnalyzeService(repo.DumpAnalysis, logger),
		Plan:     NewPlanService(repo.Plan, logger),
		PlanItem: NewPlanItemService(repo.PlanItem, logger),
	}
}
