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

type DumpAnswersRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDumpAnswersRepo(db *sql.DB, logger *zap.Logger) *DumpAnswersRepo {
	return &DumpAnswersRepo{
		db:     db,
		logger: logger.With(zap.String("component", "repository")),
	}
}

const (
	insertAnswersQuery = `
		INSERT INTO dump_answers (
			dump_id,
			answers_json,
			created_at
		) VALUES (
			$1,
			$2, 
			now()
		);
	`
	selectAnswersByDumpIDQuery = `
		SELECT dump_id, answers_json, created_at
		FROM dump_answers
		WHERE dump_id = $1
	`
)

func (r *DumpAnswersRepo) SaveAnswers(ctx context.Context, answers models.DumpAnswers) error {
	log := r.logger.With(
		zap.String("operation", "save_answers"),
		zap.String("dump_id", answers.DumpID.String()),
	)

	answersJson, err := json.Marshal(answers.Answers)
	if err != nil {
		log.Error("Marshal answers failed", zap.Error(err))
		return fmt.Errorf("marshal answers: %w", err)
	}

	_, err = r.db.ExecContext(ctx, insertAnswersQuery, answers.DumpID, answersJson)
	if err != nil {
		log.Error("Save answers failed", zap.Error(err))
		return fmt.Errorf("insert dump answers: %w", checkErr(err))
	}

	log.Info("Answers saved")

	return nil
}

func (r *DumpAnswersRepo) GetAnswers(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnswers, error) {
	log := r.logger.With(
		zap.String("operation", "get_answers"),
		zap.String("dump_id", dumpID.String()),
	)

	var dumpAnswers models.DumpAnswers
	var answersJson []byte

	err := r.db.QueryRowContext(ctx, selectAnswersByDumpIDQuery, dumpID).Scan(
		&dumpAnswers.DumpID,
		&answersJson,
		&dumpAnswers.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error("Get answers failed", zap.Error(err))
		return nil, fmt.Errorf("select dump answers: %w", checkErr(err))
	}

	if err := json.Unmarshal(answersJson, &dumpAnswers.Answers); err != nil {
		log.Error("Unmarshal answers failed", zap.Error(err))
		return nil, fmt.Errorf("unmarshal answers: %w", err)
	}

	return &dumpAnswers, nil
}
