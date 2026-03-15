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

func (s *PlanItemService) CreateItems(ctx context.Context, items []models.PlanItem) ([]models.PlanItem, error) {
	log := s.logger.With(
		zap.String("operation", "create_items"),
		zap.Int("items_count", len(items)),
	)

	log.Info("Create items started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, err := s.repo.CreateItems(ctx, items)
	if err != nil {
		log.Error("Create items failed", zap.Error(err))
		return nil, fmt.Errorf("create items: %w", err)
	}

	log.Info("Items created", zap.Int("items_count", len(items)))

	return items, nil
}

func (s *PlanItemService) ToggleItem(ctx context.Context, itemID uuid.UUID, done bool) error {
	log := s.logger.With(
		zap.String("operation", "toggle_item"),
		zap.String("item_id", itemID.String()),
		zap.Bool("done", done),
	)

	log.Info("Toggle item started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.ToggleItem(ctx, itemID, done); err != nil {
		log.Error("Toggle item failed", zap.Error(err))
		return fmt.Errorf("toggle item: %w", err)
	}

	log.Info("Item toggled")

	return nil
}

func (s *PlanItemService) AddItem(ctx context.Context, item models.PlanItem) (models.PlanItem, error) {
	log := s.logger.With(
		zap.String("operation", "add_item"),
		zap.String("plan_id", item.PlanID.String()),
	)

	log.Info("Add item started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	item, err := s.repo.AddItem(ctx, item)
	if err != nil {
		log.Error("Add item failed", zap.Error(err))
		return models.PlanItem{}, fmt.Errorf("add item: %w", err)
	}

	log.Info("Item added", zap.String("item_id", item.ID.String()))

	return item, nil
}

func (s *PlanItemService) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	log := s.logger.With(
		zap.String("operation", "delete_item"),
		zap.String("item_id", itemID.String()),
	)

	log.Info("Delete item started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.DeleteItem(ctx, itemID); err != nil {
		log.Error("Delete item failed", zap.Error(err))
		return fmt.Errorf("delete item: %w", err)
	}

	log.Info("Item deleted")

	return nil
}

func (s *PlanItemService) ReorderItems(ctx context.Context, planID uuid.UUID, itemIDs []uuid.UUID) error {
	log := s.logger.With(
		zap.String("operation", "reorder_items"),
		zap.String("plan_id", planID.String()),
		zap.Int("items_count", len(itemIDs)),
	)

	log.Info("Reorder items started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.ReorderItems(ctx, planID, itemIDs); err != nil {
		log.Error("Reorder items failed", zap.Error(err))
		return fmt.Errorf("reorder items: %w", err)
	}

	log.Info("Items reordered")

	return nil
}

func (s *PlanItemService) GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error) {
	log := s.logger.With(
		zap.String("operation", "get_items_by_plan_ids"),
		zap.Int("plans_count", len(planIDs)),
	)

	log.Info("Get items by plan ids started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, err := s.repo.GetItemsByPlanIDs(ctx, planIDs)
	if err != nil {
		log.Error("Get items by plan ids failed", zap.Error(err))
		return nil, fmt.Errorf("get items by plan ids: %w", err)
	}

	log.Info("Plan items loaded", zap.Int("items_count", len(items)))

	return items, nil
}
