package repository

import (
	"context"
	"database/sql"
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
	selectPlansByDumpIDQuery = `
		SELECT id, dump_id, title, created_at, saved_at, deleted_at
		FROM plans
		WHERE dump_id = $1
	`
	updateSavedPlanQuery = `
		UPDATE plans
		SET 
		saved_at = now()
		WHERE id = $1;
	`
	deleteUnsavePlansByDumpIDquery = `
		DELETE FROM plans
		WHERE dump_id = $1 AND saved_at IS NULL;
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

	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, insertPlanQuery, plan.DumpID, plan.Title).Scan(&id)
	if err != nil {
		log.Error("Create plan failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("insert plan: %w", checkErr(err))
	}

	log.Info("Plan created", zap.String("plan_id", id.String()))

	return id, nil
}

func (r *PlanRepository) GetPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	log := r.logger.With(
		zap.String("operation", "get_plans"),
		zap.String("dump_id", dumpID.String()),
	)

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

func (r *PlanRepository) SavePlan(ctx context.Context, planID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "save_plan"),
		zap.String("plan_id", planID.String()),
	)

	_, err := r.db.ExecContext(ctx, updateSavedPlanQuery, planID)
	if err != nil {
		log.Error("Save plan failed", zap.Error(err))
		return fmt.Errorf("update plan saved_at: %w", checkErr(err))
	}

	log.Info("Plan saved")

	return nil
}

func (r *PlanRepository) DeleteUnsavedPlans(ctx context.Context, dumpID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "delete_unsaved_plans"),
		zap.String("dump_id", dumpID.String()),
	)

	_, err := r.db.ExecContext(ctx, deleteUnsavePlansByDumpIDquery, dumpID)
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

	_, err := r.db.ExecContext(ctx, deleteSavedPlanQuery, planID)
	if err != nil {
		log.Error("Delete saved plan failed", zap.Error(err))
		return fmt.Errorf("soft delete plan: %w", checkErr(err))
	}

	return nil
}
