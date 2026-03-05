package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
		)
		RETURNING id, created_at;
	`
	deleteItemQuery = `
		UPDATE plan_items
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL;
	`
	toggleItemQuery = `
		UPDATE plan_items
		SET done = $2
		WHERE id = $1 AND deleted_at IS NULL;
	`
	offsetItemsOrdQuery = `
		UPDATE plan_items
		SET ord = ord + 1000000
		WHERE plan_id = $1 AND id = ANY($2::uuid[]) AND deleted_at IS NULL;
	`
	updateItemsOrderQuery = `
		UPDATE plan_items AS pi
		SET ord = v.new_ord
		FROM (
			SELECT *
			FROM unnest($2::uuid[], $3::int[]) AS t(id, new_ord)
		) AS v
		WHERE pi.plan_id = $1 AND pi.id = v.id AND deleted_at IS NULL;
	`
	getItemsByPlanIdsQuery = `
		SELECT id, plan_id, ord, text, done, created_at
		FROM plan_items
		WHERE plan_id = ANY($1) AND deleted_at IS NULL
		ORDER BY plan_id, ord;
	`
)

func (r *PlanItemRepository) CreateItems(ctx context.Context, items []models.PlanItem) (res []models.PlanItem, err error) {
	if len(items) == 0 {
		return make([]models.PlanItem, 0), nil
	}

	log := r.logger.With(
		zap.String("operation", "create_items"),
		zap.String("plan_id", items[0].PlanID.String()),
	)

	log.Info("Create items started")

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Error("Begin tx failed", zap.Error(err))
		return make([]models.PlanItem, 0), fmt.Errorf("begin tx: %w", checkErr(err))
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

	res = make([]models.PlanItem, 0, len(items))
	for _, item := range items {
		var id uuid.UUID
		var createdAt time.Time
		err = tx.QueryRowContext(ctx, insertPlanItemQuery, item.PlanID, item.Ord, item.Text).Scan(&id, &createdAt)
		if err != nil {
			log.Error("Create plan item failed", zap.Error(err))
			err = fmt.Errorf("insert plan item: %w", checkErr(err))
			return make([]models.PlanItem, 0), err
		}

		item.ID = id
		item.CreatedAt = createdAt
		res = append(res, item)
	}

	return res, nil
}

func (r *PlanItemRepository) AddItem(ctx context.Context, item models.PlanItem) (models.PlanItem, error) {
	log := r.logger.With(
		zap.String("operation", "add_item"),
		zap.String("plan_id", item.PlanID.String()),
	)

	log.Info("Add item started")

	var id uuid.UUID
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, insertPlanItemQuery, item.PlanID, item.Ord, item.Text).Scan(&id, &createdAt)
	if err != nil {
		log.Error("Add plan item failed", zap.Error(err))
		return models.PlanItem{}, fmt.Errorf("insert plan item: %w", checkErr(err))
	}

	item.ID = id
	item.CreatedAt = createdAt

	return item, nil
}

func (r *PlanItemRepository) GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error) {
	log := r.logger.With(zap.String("operation", "get_items_by_plan_ids"))

	log.Info("Get items by plan ids started")

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

var ErrItemNotDeleted = errors.New("item was not deleted")

func (r *PlanItemRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "delete_item"),
		zap.String("item_id", itemID.String()),
	)

	log.Info("Delete item started")

	res, err := r.db.ExecContext(ctx, deleteItemQuery, itemID)
	if err != nil {
		log.Error("Delete plan item failed", zap.Error(err))
		return fmt.Errorf("soft delete plan item: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		log.Error("Delete plan item failed", zap.Error(ErrItemNotDeleted))
		return ErrItemNotDeleted
	}

	return nil
}

var ErrItemNotToggled = errors.New("item was not toggled")

func (r *PlanItemRepository) ToggleItem(ctx context.Context, itemID uuid.UUID, done bool) error {
	log := r.logger.With(
		zap.String("operation", "toggle_item"),
		zap.String("item_id", itemID.String()),
	)

	log.Info("Toggle item started")

	res, err := r.db.ExecContext(ctx, toggleItemQuery, itemID, done)
	if err != nil {
		log.Error("Toggle plan item failed", zap.Error(err))
		return fmt.Errorf("toggle plan item: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		log.Error("Toggle plan item failed", zap.Error(ErrItemNotToggled))
		return ErrItemNotToggled
	}

	return nil
}

var ErrItemsNotOffset = errors.New("failed to offset items")
var ErrItemsNotReordered = errors.New("failed to reorder items")

func (r *PlanItemRepository) ReorderItems(ctx context.Context, planID uuid.UUID, itemsIDs []uuid.UUID) (err error) {
	log := r.logger.With(
		zap.String("operation", "reorder_items"),
		zap.String("plan_id", planID.String()),
	)

	log.Info("Reorder items started")

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

	res, err := tx.ExecContext(ctx, offsetItemsOrdQuery, planID, itemsIDs)
	if err != nil {
		log.Error("Offset plan item order failed", zap.Error(err))
		return fmt.Errorf("offset plan item order: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != int64(len(itemsIDs)) {
		log.Error("Reorder plan items failed", zap.Error(ErrItemsNotOffset))
		return ErrItemsNotOffset
	}

	var newOrd []int
	for i := 0; i < len(itemsIDs); i++ {
		newOrd = append(newOrd, i+1)
	}

	res, err = tx.ExecContext(ctx, updateItemsOrderQuery, planID, itemsIDs, newOrd)
	if err != nil {
		log.Error("Update plan item order failed", zap.Error(err))
		return fmt.Errorf("update plan item order: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != int64(len(itemsIDs)) {
		log.Error("Reorder plan items failed", zap.Error(ErrItemsNotReordered))
		return ErrItemsNotReordered
	}

	return nil
}
