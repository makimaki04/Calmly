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

type DumpAnalysisRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDumpAnalysisRepo(db *sql.DB, logger *zap.Logger) *DumpAnalysisRepository {
	return &DumpAnalysisRepository{
		db:     db,
		logger: logger.With(zap.String("component", "repository")),
	}
}

const (
	insertDumpAnalysisQuery = `
		INSERT INTO dump_analysis (
			dump_id, 
			tasks_json, 
			questions_json, 
			mood, 
			quote, 
			created_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			now()
		);
	`
	selectAnalysisByDumpIDQuery = `
		SELECT dump_id, tasks_json, questions_json, mood, quote, created_at
		FROM dump_analysis
		WHERE dump_id = $1;
	`
)

func (r *DumpAnalysisRepository) SaveAnalysis(ctx context.Context, dumpAnalysis models.DumpAnalysis) error {
	log := r.logger.With(
		zap.String("operation", "save_analysis"),
		zap.String("dump_id", dumpAnalysis.DumpID.String()),
	)

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

	_, err = r.db.ExecContext(
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

	log.Info("Analysis saved")

	return nil
}

func (r *DumpAnalysisRepository) GetAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error) {
	log := r.logger.With(
		zap.String("operation", "get_analysis"),
		zap.String("dump_id", dumpID.String()),
	)

	var analysis models.DumpAnalysis
	var tasks []byte
	var questions []byte
	err := r.db.QueryRowContext(ctx, selectAnalysisByDumpIDQuery, dumpID).Scan(
		&analysis.DumpID,
		&tasks,
		&questions,
		&analysis.Mood,
		&analysis.Quote,
		&analysis.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error("Get analysis failed", zap.Error(err))
		return nil, fmt.Errorf("select dump analysis: %w", checkErr(err))
	}

	if err := json.Unmarshal(tasks, &analysis.Tasks); err != nil {
		log.Error("Unmarshal tasks failed", zap.Error(err))
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}

	if err := json.Unmarshal(questions, &analysis.Questions); err != nil {
		log.Error("Unmarshal questions failed", zap.Error(err))
		return nil, fmt.Errorf("unmarshal questions: %w", err)
	}

	return &analysis, nil
}
