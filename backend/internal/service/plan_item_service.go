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

// PlanItemService delegates to repository which owns error logging.
// This layer only wraps and propagates errors.
type PlanItemService struct {
	repo   repository.PlanItem
	logger *zap.Logger
}

func NewPlanItemService(repo repository.PlanItem, logger *zap.Logger) *PlanItemService {
	return &PlanItemService{
		repo:   repo,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *PlanItemService) CreateItems(ctx context.Context, items []models.PlanItem) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.CreateItems(ctx, items); err != nil {
		return fmt.Errorf("create items: %w", err)
	}

	return nil
}

func (s *PlanItemService) ToggleItem(ctx context.Context, itemID uuid.UUID, done bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.ToggleItem(ctx, itemID, done); err != nil {
		return fmt.Errorf("toggle item: %w", err)
	}

	return nil
}

func (s *PlanItemService) AddItem(ctx context.Context, item models.PlanItem) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.AddItem(ctx, item); err != nil {
		return fmt.Errorf("add item: %w", err)
	}

	return nil
}

func (s *PlanItemService) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}

func (s *PlanItemService) ReorderItems(ctx context.Context, planID uuid.UUID, itemIDs []uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.ReorderItems(ctx, planID, itemIDs); err != nil {
		return fmt.Errorf("reorder items: %w", err)
	}

	return nil
}

func (s *PlanItemService) GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, err := s.repo.GetItemsByPlanIDs(ctx, planIDs)
	if err != nil {
		return nil, fmt.Errorf("get items by plan ids: %w", err)
	}

	return items, nil
}