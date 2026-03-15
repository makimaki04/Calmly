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

// AnalysisService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type AnalysisService struct {
	repo   repository.DumpAnalysis
	logger *zap.Logger
}

func NewAnalyzeService(repo repository.DumpAnalysis, logger *zap.Logger) *AnalysisService {
	return &AnalysisService{
		repo:   repo,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *AnalysisService) SaveDumpAnalysis(ctx context.Context, analysis models.DumpAnalysis) error {
	log := s.logger.With(
		zap.String("operation", "save_dump_analysis"),
		zap.String("dump_id", analysis.DumpID.String()),
		zap.Int("tasks_count", len(analysis.Tasks)),
		zap.Int("questions_count", len(analysis.Questions)),
	)

	log.Info("Save dump analysis started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.SaveAnalysis(ctx, analysis); err != nil {
		log.Error("Save dump analysis failed", zap.Error(err))
		return fmt.Errorf("save dump analysis: %w", err)
	}

	log.Info("Dump analysis saved")

	return nil
}

func (s *AnalysisService) GetDumpAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error) {
	log := s.logger.With(
		zap.String("operation", "get_dump_analysis"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Get dump analysis started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	analysis, err := s.repo.GetAnalysis(ctx, dumpID)
	if err != nil {
		log.Error("Get dump analysis failed", zap.Error(err))
		return nil, fmt.Errorf("get dump analysis: %w", err)
	}

	if analysis == nil {
		log.Info("Dump analysis not found")
		return nil, nil
	}

	log.Info("Dump analysis found")

	return analysis, nil
}
