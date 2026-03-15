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
	log := s.logger.With(
		zap.String("operation", "create_plan"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Create plan started")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	planID, err := s.repo.CreatePlan(ctx, models.Plan{
		DumpID: dumpID,
		Title:  title,
	})
	if err != nil {
		log.Error("Create plan failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("create plan: %w", err)
	}

	log.Info("Plan created", zap.String("plan_id", planID.String()))

	return planID, nil
}

func (s *PlanService) SubmitAnswersAndCreatePlan(ctx context.Context, answers models.DumpAnswers, plan models.Plan, planItems []models.PlanItem) (models.Plan, []models.PlanItem, error) {
	log := s.logger.With(
		zap.String("operation", "submit_answers_and_create_plan"),
		zap.String("dump_id", answers.DumpID.String()),
		zap.Int("answers_count", len(answers.Answers)),
		zap.Int("plan_items_count", len(planItems)),
	)

	log.Info("Submit answers and create plan started")

	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	plan, items, err := s.repo.SubmitAnswersAndCreatePlan(ctx, answers, plan, planItems)
	if err != nil {
		log.Error("Submit answers and create plan failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("submit answers and create plan: %w", err)
	}

	log.Info(
		"Plan created from answers",
		zap.String("plan_id", plan.ID.String()),
		zap.Int("items_count", len(items)),
	)

	return plan, items, nil
}

func (s *PlanService) SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	log := s.logger.With(
		zap.String("operation", "save_plan"),
		zap.String("dump_id", dumpID.String()),
		zap.String("plan_id", planID.String()),
	)

	log.Info("Save plan started")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.repo.FinalizeSelectedPlan(ctx, dumpID, planID); err != nil {
		log.Error("Save plan failed", zap.Error(err))
		return fmt.Errorf("save plan: %w", err)
	}

	log.Info("Plan saved")

	return nil
}

func (s *PlanService) GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	log := s.logger.With(
		zap.String("operation", "get_dump_plans"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Get dump plans started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	plans, err := s.repo.GetCurrentSessionsPlans(ctx, dumpID)
	if err != nil {
		log.Error("Get dump plans failed", zap.Error(err))
		return nil, fmt.Errorf("get dump plans: %w", err)
	}

	log.Info("Dump plans loaded", zap.Int("plans_count", len(plans)))

	return plans, nil
}

func (s *PlanService) GetUserSavedPlans(ctx context.Context, userID uuid.UUID) ([]models.Plan, error) {
	log := s.logger.With(
		zap.String("operation", "get_user_saved_plans"),
		zap.String("user_id", userID.String()),
	)

	log.Info("Get user saved plans started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	plans, err := s.repo.GetSavedPlans(ctx, userID)
	if err != nil {
		log.Error("Get user saved plans failed", zap.Error(err))
		return nil, fmt.Errorf("get saved plans: %w", err)
	}

	log.Info("Saved plans loaded", zap.Int("plans_count", len(plans)))

	return plans, nil
}

func (s *PlanService) DeleteSavedPlan(ctx context.Context, planID uuid.UUID) error {
	log := s.logger.With(
		zap.String("operation", "delete_saved_plan"),
		zap.String("plan_id", planID.String()),
	)

	log.Info("Delete saved plan started")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.repo.DeleteSavedPlan(ctx, planID); err != nil {
		log.Error("Delete saved plan failed", zap.Error(err))
		return fmt.Errorf("delete saved plan: %w", err)
	}

	log.Info("Saved plan deleted")

	return nil
}
