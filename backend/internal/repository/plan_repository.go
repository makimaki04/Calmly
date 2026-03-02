package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

type PlanRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPlanRepo(db *sql.DB, logger *zap.Logger) *PlanRepository {
	return &PlanRepository{
		db:     db,
		logger: logger.With(zap.String("component", "repository")),
	}
}

const (
	insertPlanQuery = `
		INSERT INTO plans (
			dump_id,
			title,
			created_at
		) VALUES (
			$1,
			$2,
			now()
		)
		RETURNING id;
	`
	findPlanByDumpID = `
		SELECT id
		FROM plans
		WHERE id = $1 AND dump_id = $2 AND deleted_at IS NULL;
	`
	selectPlansByDumpIDQuery = `
		SELECT id, dump_id, title, created_at, saved_at, deleted_at
		FROM plans
		WHERE dump_id = $1
	`
	updateSavedPlanQuery = `
		UPDATE plans
		SET 
		saved_at = now()
		WHERE id = $1
			AND deleted_at IS NULL
			AND saved_at IS NULL;
	`
	deleteUnsavedPlansByDumpIDquery = `
		DELETE FROM plans
		WHERE dump_id = $1
			AND id <> $2
			AND saved_at IS NULL;
	`
	selectSavedPlansQuery = `
		SELECT p.id, p.dump_id, p.title, p.created_at
		FROM plans as p
		JOIN dumps as d ON p.dump_id = d.id
		WHERE d.user_id = $1 AND p.saved_at IS NOT NULL AND p.deleted_at IS NULL
		ORDER BY saved_at DESC;
	`
	deleteSavedPlanQuery = `
		UPDATE plans
		SET 
		deleted_at = now()
		WHERE id = $1;
	`
)

func (r *PlanRepository) CreatePlan(ctx context.Context, plan models.Plan) (uuid.UUID, error) {
	log := r.logger.With(
		zap.String("operation", "create_plan"),
		zap.String("dump_id", plan.DumpID.String()),
	)

	log.Info("Create plan started")

	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, insertPlanQuery, plan.DumpID, plan.Title).Scan(&id)
	if err != nil {
		log.Error("Create plan failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("insert plan: %w", checkErr(err))
	}

	log.Info("Plan created", zap.String("plan_id", id.String()))

	return id, nil
}

func (r *PlanRepository) GetPlansByDumpID(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	log := r.logger.With(
		zap.String("operation", "get_plans"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Get plans started")

	rows, err := r.db.QueryContext(ctx, selectPlansByDumpIDQuery, dumpID)
	if err != nil {
		log.Error("Get plans failed", zap.Error(err))
		return nil, fmt.Errorf("query plans: %w", checkErr(err))
	}
	defer rows.Close()

	plans := make([]models.Plan, 0)
	for rows.Next() {
		var plan models.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.DumpID,
			&plan.Title,
			&plan.CreatedAt,
			&plan.SavedAt,
			&plan.DeletedAt,
		); err != nil {
			log.Error("Scan plan row failed", zap.Error(err))
			return nil, fmt.Errorf("scan plan row: %w", checkErr(err))
		}

		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		log.Error("Iterate plan rows failed", zap.Error(err))
		return nil, fmt.Errorf("iterate plan rows: %w", checkErr(err))
	}

	return plans, nil
}

func (r *PlanRepository) FinalizeSelectedPlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) (err error) {
	log := r.logger.With(
		zap.String("operation", "finalize_selected_plan"),
		zap.String("plan_id", planID.String()),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Finalize selected plan started")

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

	err = r.ensurePlanBelongsToDump(ctx, tx, dumpID, planID)
	if err != nil {
		log.Error("Validate plan ownership failed", zap.Error(err))
		return fmt.Errorf("ensure plan belongs to dump: %w", err)
	}

	err = r.markPlanSaved(ctx, tx, planID)
	if err != nil {
		log.Error("Mark plan saved failed", zap.Error(err))
		return fmt.Errorf("mark plan saved: %w", err)
	}

	err = r.deleteUnsavedPlans(ctx, tx, dumpID, planID)
	if err != nil {
		log.Error("Delete unsaved plans failed", zap.Error(err))
		return fmt.Errorf("delete unsaved plans: %w", err)
	}

	return nil
}

func (r *PlanRepository) ensurePlanBelongsToDump(
	ctx context.Context,
	tx *sql.Tx,
	dumpID uuid.UUID,
	planID uuid.UUID,
) error {
	var id uuid.UUID
	if err := tx.QueryRowContext(ctx, findPlanByDumpID, planID, dumpID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		return fmt.Errorf("find plan by dump id: %w", checkErr(err))
	}

	return nil
}

var ErrPlanNotUpdated = errors.New("plan was not updated")

func (r *PlanRepository) markPlanSaved(ctx context.Context, tx *sql.Tx, planID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "save_plan"),
		zap.String("plan_id", planID.String()),
	)

	res, err := tx.ExecContext(ctx, updateSavedPlanQuery, planID)
	if err != nil {
		log.Error("Save plan failed", zap.Error(err))
		return fmt.Errorf("update plan saved_at: %w", checkErr(err))
	}

	if rows, _ := res.RowsAffected(); rows != 1 {
		return ErrPlanNotUpdated
	}

	log.Info("Plan saved")

	return nil
}

func (r *PlanRepository) deleteUnsavedPlans(ctx context.Context, tx *sql.Tx, dumpID uuid.UUID, selectedPlanID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "delete_unsaved_plans"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Delete unsaved plans started")

	_, err := tx.ExecContext(ctx, deleteUnsavedPlansByDumpIDquery, dumpID, selectedPlanID)
	if err != nil {
		log.Error("Delete unsaved plans failed", zap.Error(err))
		return fmt.Errorf("delete unsaved plans: %w", checkErr(err))
	}

	return nil
}

func (r *PlanRepository) GetSavedPlans(ctx context.Context, userID uuid.UUID) ([]models.Plan, error) {
	log := r.logger.With(
		zap.String("operation", "get_saved_plans"),
		zap.String("user_id", userID.String()),
	)

	log.Info("Get saved plans started")

	rows, err := r.db.QueryContext(ctx, selectSavedPlansQuery, userID)
	if err != nil {
		log.Error("Get saved plans failed", zap.Error(err))
		return nil, fmt.Errorf("query saved plans: %w", checkErr(err))
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var plan models.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.DumpID,
			&plan.Title,
			&plan.CreatedAt,
		); err != nil {
			log.Error("Scan saved plan row failed", zap.Error(err))
			return nil, fmt.Errorf("scan saved plan row: %w", checkErr(err))
		}

		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		log.Error("Iterate saved plan rows failed", zap.Error(err))
		return nil, fmt.Errorf("iterate saved plan rows: %w", checkErr(err))
	}

	return plans, nil
}

func (r *PlanRepository) DeleteSavedPlan(ctx context.Context, planID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "delete_saved_plan"),
		zap.String("plan_id", planID.String()),
	)

	log.Info("Delete saved plan started")

	_, err := r.db.ExecContext(ctx, deleteSavedPlanQuery, planID)
	if err != nil {
		log.Error("Delete saved plan failed", zap.Error(err))
		return fmt.Errorf("soft delete plan: %w", checkErr(err))
	}

	return nil
}
