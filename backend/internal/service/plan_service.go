package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

// PlanService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type PlanService struct {
	repo   repository.Plan
	logger *zap.Logger
}

func NewPlanService(repo repository.Plan, logger *zap.Logger) *PlanService {
	return &PlanService{
		repo:   repo,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *PlanService) CreatePlan(ctx context.Context, dumpID uuid.UUID, title string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	planID, err := s.repo.CreatePlan(ctx, models.Plan{
		DumpID: dumpID,
		Title:  title,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create plan: %w", err)
	}

	return planID, nil
}

func (s *PlanService) SubmitAnswersAndCreatePlan(ctx context.Context, answers models.DumpAnswers, plan models.Plan, planItems []models.PlanItem) (models.Plan, []models.PlanItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	plan, items, err := s.repo.SubmitAnswersAndCreatePlan(ctx, answers, plan, planItems)
	if err != nil {
		return  models.Plan{}, []models.PlanItem{}, err
	}

	return plan, items, nil 
}

func (s *PlanService) SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.repo.FinalizeSelectedPlan(ctx, dumpID, planID); err != nil {
		return fmt.Errorf("save plan: %w", err)
	}

	return nil
}

func (s *PlanService) GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	plans, err := s.repo.GetCurrentSessionsPlans(ctx, dumpID)
	if err != nil {
		return nil, fmt.Errorf("get dump plans: %w", err)
	}

	return plans, nil
}

func (s *PlanService) GetUserSavedPlans(ctx context.Context, userID uuid.UUID) ([]models.Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	plans, err := s.repo.GetSavedPlans(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get saved plans: %w", err)
	}

	return plans, nil
}

func (s *PlanService) DeleteSavedPlan(ctx context.Context, planID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.repo.DeleteSavedPlan(ctx, planID); err != nil {
		return fmt.Errorf("delete saved plan: %w", err)
	}

	return nil
}
