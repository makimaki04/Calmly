package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

type PlanItemRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPlanItemRepo(db *sql.DB, logger *zap.Logger) *PlanItemRepository {
	return &PlanItemRepository{
		db:     db,
		logger: logger.With(zap.String("component", "repository")),
	}
}

const (
	insertPlanItemQuery = `
		INSERT INTO plan_items (
			plan_id, ord, text, created_at
		) VALUES (
			$1,
			$2,
			$3,
			now()
		);
	`
	deleteItemQuery = `
		UPDATE plan_items
		SET deleted_at = now()
		WHERE id = $1;
	`
	toggleItemQuery = `
		UPDATE plan_items
		SET done = $2
		WHERE id = $1;
	`
	updateItemsOrderQuery = `
		UPDATE plan_items
		SET ord = $3
		WHERE id = $1 AND plan_id = $2;
	`
	getItemsByPlanIdsQuery = `
		SELECT id, plan_id, ord, text, done, created_at
		FROM plan_items
		WHERE plan_id = ANY($1) AND deleted_at IS NULL
		ORDER BY plan_id, ord;
	`
)

func (r *PlanItemRepository) CreateItems(ctx context.Context, items []models.PlanItem) error {
	if len(items) == 0 {
		return nil
	}

	log := r.logger.With(
		zap.String("operation", "create_items"),
		zap.String("plan_id", items[0].PlanID.String()),
	)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Error("Begin tx failed", zap.Error(err))
		return fmt.Errorf("begin tx: %w", checkErr(err))
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			_ = tx.Rollback()
			log.Error("Commit tx failed", zap.Error(commitErr))
			err = fmt.Errorf("commit tx: %w", commitErr)
		}
	}()

	for _, item := range items {
		_, err = tx.ExecContext(ctx, insertPlanItemQuery, item.PlanID, item.Ord, item.Text)
		if err != nil {
			log.Error("Create plan item failed", zap.Error(err))
			err = fmt.Errorf("insert plan item: %w", checkErr(err))
			return err
		}
	}

	return nil
}

func (r *PlanItemRepository) AddItem(ctx context.Context, item models.PlanItem) error {
	log := r.logger.With(
		zap.String("operation", "add_item"),
		zap.String("plan_id", item.PlanID.String()),
	)

	_, err := r.db.ExecContext(ctx, insertPlanItemQuery, item.PlanID, item.Ord, item.Text)
	if err != nil {
		log.Error("Add plan item failed", zap.Error(err))
		return fmt.Errorf("insert plan item: %w", checkErr(err))
	}

	return nil
}

func (r *PlanItemRepository) GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error) {
	log := r.logger.With(zap.String("operation", "get_items_by_plan_ids"))

	rows, err := r.db.QueryContext(ctx, getItemsByPlanIdsQuery, planIDs)
	if err != nil {
		log.Error("Get plan items failed", zap.Error(err))
		return nil, fmt.Errorf("query plan items: %w", checkErr(err))
	}
	defer rows.Close()

	var items []models.PlanItem
	for rows.Next() {
		var item models.PlanItem
		if err := rows.Scan(
			&item.ID,
			&item.PlanID,
			&item.Ord,
			&item.Text,
			&item.Done,
			&item.CreatedAt,
		); err != nil {
			log.Error("Scan plan item row failed", zap.Error(err))
			return nil, fmt.Errorf("scan plan item row: %w", checkErr(err))
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Error("Iterate plan item rows failed", zap.Error(err))
		return nil, fmt.Errorf("iterate plan item rows: %w", checkErr(err))
	}

	return items, nil
}

func (r *PlanItemRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "delete_item"),
		zap.String("item_id", itemID.String()),
	)

	_, err := r.db.ExecContext(ctx, deleteItemQuery, itemID)
	if err != nil {
		log.Error("Delete plan item failed", zap.Error(err))
		return fmt.Errorf("soft delete plan item: %w", checkErr(err))
	}

	return nil
}

func (r *PlanItemRepository) ToggleItem(ctx context.Context, itemID uuid.UUID, done bool) error {
	log := r.logger.With(
		zap.String("operation", "toggle_item"),
		zap.String("item_id", itemID.String()),
	)

	_, err := r.db.ExecContext(ctx, toggleItemQuery, itemID, done)
	if err != nil {
		log.Error("Toggle plan item failed", zap.Error(err))
		return fmt.Errorf("toggle plan item: %w", checkErr(err))
	}

	return nil
}

func (r *PlanItemRepository) ReorderItems(ctx context.Context, planID uuid.UUID, itemsIDs []uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "reorder_items"),
		zap.String("plan_id", planID.String()),
	)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Error("Begin tx failed", zap.Error(err))
		return fmt.Errorf("begin tx: %w", checkErr(err))
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			_ = tx.Rollback()
			log.Error("Commit tx failed", zap.Error(commitErr))
			err = fmt.Errorf("commit tx: %w", commitErr)
		}
	}()

	for i, itemID := range itemsIDs {
		ord := i + 1
		_, err = tx.ExecContext(ctx, updateItemsOrderQuery, itemID, planID, ord)
		if err != nil {
			log.Error("Reorder plan item failed", zap.String("item_id", itemID.String()), zap.Error(err))
			err = fmt.Errorf("update plan item order: %w", checkErr(err))
			return err
		}
	}

	return nil
}
