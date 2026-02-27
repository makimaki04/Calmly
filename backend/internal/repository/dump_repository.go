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

type DumpRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDumpRepository(db *sql.DB, logger *zap.Logger) *DumpRepository {
	return &DumpRepository{
		db:     db,
		logger: logger.With(zap.String("component", "repository")),
	}
}

const (
	insertDumpQuery = `
		INSERT INTO dumps (
			user_id, 
			guest_id, 
			status, 
			raw_text, 
			raw_text_expires_at, 
			created_at, 
			updated_at
		) VALUES (
			$1, 
			$2, 
			$3, 
			$4, 
			$5, 
			now(), 
			now()
		)
		RETURNING id;
	`
	updateStatusQuery = `
		UPDATE dumps
		SET 
		status = $2
		WHERE id = $1;
	`
	clearRawTextQuery = `
		UPDATE dumps
		SET
		raw_text = NULL,
		raw_text_deleted_at = now(),
		status = CASE
				WHEN status NOT IN ('abandoned', 'planned') THEN 'abandoned'
				ELSE status
			END
		WHERE id = $1;
	`
	selectDumpByUserIDQuery = `
		SELECT id, user_id, guest_id, status, raw_text, raw_text_deleted_at, raw_text_expires_at, created_at, updated_at
		FROM dumps
		WHERE user_id = $1 AND status <> 'abandoned' AND status <> 'planned'
		ORDER BY updated_at DESC 
		LIMIT 1;
	`
)

func (r *DumpRepository) CreateDump(ctx context.Context, dump models.Dump) (uuid.UUID, error) {
	log := r.logger.With(zap.String("operation", "create_dump"))

	var id uuid.UUID

	err := r.db.QueryRowContext(
		ctx,
		insertDumpQuery,
		dump.UserID,
		dump.GuestID,
		dump.Status,
		dump.RawText,
		dump.RawExpiresAt,
	).Scan(&id)
	if err != nil {
		log.Error("Create dump failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("insert dump: %w", checkErr(err))
	}

	log.Info("Dump created", zap.String("dump_id", id.String()))

	return id, nil
}

func (r *DumpRepository) UpdateStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error {
	log := r.logger.With(
		zap.String("operation", "update_status"),
		zap.String("dump_id", dumpID.String()),
	)

	_, err := r.db.ExecContext(ctx, updateStatusQuery, dumpID, status)
	if err != nil {
		log.Error("Update status failed", zap.Error(err))
		return fmt.Errorf("update dump status: %w", checkErr(err))
	}

	return nil
}

func (r *DumpRepository) ClearRawText(ctx context.Context, dumpID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "clear_raw_text"),
		zap.String("dump_id", dumpID.String()),
	)

	_, err := r.db.ExecContext(ctx, clearRawTextQuery, dumpID)
	if err != nil {
		log.Error("Clear raw text failed", zap.Error(err))
		return fmt.Errorf("clear raw text: %w", checkErr(err))
	}

	log.Info("Raw text cleared")

	return nil
}

func (r *DumpRepository) GetActiveDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error) {
	log := r.logger.With(
		zap.String("operation", "get_active_dump"),
		zap.String("user_id", userID.String()),
	)

	var dump models.Dump

	err := r.db.QueryRowContext(ctx, selectDumpByUserIDQuery, userID).Scan(
		&dump.ID,
		&dump.UserID,
		&dump.GuestID,
		&dump.Status,
		&dump.RawText,
		&dump.RawDeletedAt,
		&dump.RawExpiresAt,
		&dump.CreatedAt,
		&dump.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error("Get active dump failed", zap.Error(err))
		return nil, fmt.Errorf("select active dump: %w", checkErr(err))
	}

	return &dump, nil
}
