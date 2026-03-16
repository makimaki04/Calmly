package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

func NewDumpRepo(db *sql.DB, logger *zap.Logger) *DumpRepository {
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
	updateCompleteAnalysisStatusQuery = `
		UPDATE dumps
		SET 
		status = $2
		WHERE id = $1 AND status = 'waiting_analysis';
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
	clearExpiredRawTextsQuery = `
		UPDATE dumps
		SET
		raw_text = NULL,
		raw_text_deleted_at = now(),
		status = CASE
				WHEN status NOT IN ('abandoned', 'planned') THEN 'abandoned'
				ELSE status
			END
		WHERE raw_text IS NOT NULL AND raw_text_expires_at <= now();
	`
)

func (r *DumpRepository) CreateDump(ctx context.Context, userID uuid.UUID, dump models.Dump) (id uuid.UUID, err error) {
	log := r.logger.With(zap.String("operation", "create_dump"))

	log.Info("Create dump started")

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Error("Begin tx failed", zap.Error(err))
		return uuid.Nil, fmt.Errorf("begin tx: %w", checkErr(err))
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

	var activeDump models.Dump
	err = tx.QueryRowContext(ctx, selectDumpByUserIDQuery, userID).Scan(
		&activeDump.ID,
	)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("Get active dump failed", zap.Error(err))
		return uuid.Nil, fmt.Errorf("select active dump: %w", checkErr(err))
	}

	if activeDump.ID != uuid.Nil {
		res, err := tx.ExecContext(ctx, updateStatusQuery, activeDump.ID, string(models.DumpStatusAbandoned))
		if err != nil {
			log.Error("Update status failed", zap.Error(err))
			return uuid.Nil, fmt.Errorf("update dump status: %w", checkErr(err))
		}

		if row, _ := res.RowsAffected(); row != 1 {
			log.Error("update dump status", zap.Error(ErrStatusNotChanged))
			return uuid.Nil, ErrStatusNotChanged
		}
	}

	if err := tx.QueryRowContext(
		ctx,
		insertDumpQuery,
		dump.UserID,
		dump.GuestID,
		dump.Status,
		dump.RawText,
		dump.RawExpiresAt,
	).Scan(&id); err != nil {
		log.Error("Create dump failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("insert dump: %w", checkErr(err))
	}

	log.Info("Dump created", zap.String("dump_id", id.String()))

	return id, nil
}

var (
	ErrStatusNotChanged  = errors.New("failed to change dump status")
	ErrRawTextNotCleared = errors.New("failed to clear raw text")
)

func (r *DumpRepository) UpdateStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error {
	log := r.logger.With(
		zap.String("operation", "update_status"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Update status started")

	res, err := r.db.ExecContext(ctx, updateStatusQuery, dumpID, status)
	if err != nil {
		log.Error("Update status failed", zap.Error(err))
		return fmt.Errorf("update dump status: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		log.Error("update dump status", zap.Error(ErrStatusNotChanged))
		return ErrStatusNotChanged
	}

	return nil
}

func (r *DumpRepository) ClearRawText(ctx context.Context, dumpID uuid.UUID) error {
	log := r.logger.With(
		zap.String("operation", "clear_raw_text"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Clear raw text started")

	res, err := r.db.ExecContext(ctx, clearRawTextQuery, dumpID)
	if err != nil {
		log.Error("Clear raw text failed", zap.Error(err))
		return fmt.Errorf("clear raw text: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		log.Error("Clear raw text failed", zap.Error(ErrRawTextNotCleared))
		return ErrRawTextNotCleared
	}

	log.Info("Raw text cleared")

	return nil
}

func (r *DumpRepository) ClearExpiredRawTexts(ctx context.Context) error {
	log := r.logger.With(zap.String("operation", "clear_expired_raw_texts"))

	log.Info("Clear expired raw texts started")

	_, err := r.db.ExecContext(ctx, clearExpiredRawTextsQuery)
	if err != nil {
		log.Error("Clear expired raw texts failed", zap.Error(err))
		err = fmt.Errorf("clear expired raw texts: %w", checkErr(err))
		return err
	}

	log.Info("Expired raw texts cleared")

	return nil
}

func (r *DumpRepository) GetActiveDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error) {
	log := r.logger.With(
		zap.String("operation", "get_active_dump"),
		zap.String("user_id", userID.String()),
	)

	log.Info("Get active dump started")

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

func (r *DumpRepository) CompleteAnalysisStep(ctx context.Context, dumpAnalysis models.DumpAnalysis) (err error) {
	log := r.logger.With(
		zap.String("operation", "complete_analysis_step"),
		zap.String("dump_id", dumpAnalysis.DumpID.String()),
	)

	log.Info("Complete analysis step started")

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

	tJson, err := json.Marshal(dumpAnalysis.Tasks)
	if err != nil {
		log.Error("Marshal tasks failed", zap.Error(err))
		return fmt.Errorf("marshal tasks: %w", err)
	}

	qJson, err := json.Marshal(dumpAnalysis.Questions)
	if err != nil {
		log.Error("Marshal questions failed", zap.Error(err))
		return fmt.Errorf("marshal questions: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		insertDumpAnalysisQuery,
		dumpAnalysis.DumpID,
		tJson,
		qJson,
		dumpAnalysis.Mood,
		dumpAnalysis.Quote,
	)
	if err != nil {
		log.Error("Save analysis failed", zap.Error(err))
		return fmt.Errorf("insert dump analysis: %w", checkErr(err))
	}

	res, err := tx.ExecContext(
		ctx,
		updateCompleteAnalysisStatusQuery,
		dumpAnalysis.DumpID,
		models.DumpStatusWaitingAnswers,
	)
	if err != nil {
		log.Error("Update status failed", zap.Error(err))
		return fmt.Errorf("update dump status: %w", checkErr(err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		log.Error("update dump status", zap.Error(ErrStatusNotChanged))
		return ErrStatusNotChanged
	}

	return nil
}
