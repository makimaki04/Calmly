package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

// AnalysisService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type AnalysisService struct {
	db     repository.DumpAnalysis
	logger *zap.Logger
}

func NewAnalyzeService(db repository.DumpAnalysis, logger *zap.Logger) *AnalysisService {
	return &AnalysisService{
		db:     db,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *AnalysisService) SaveDumpAnalysis(ctx context.Context, analysis models.DumpAnalysis) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.db.SaveAnalysis(ctx, analysis); err != nil {
		return err
	}

	return nil
}

func (s *AnalysisService) GetDumpAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	analysis, err := s.db.GetAnalysis(ctx, dumpID)
	if err != nil {
		return nil, err
	}

	if analysis == nil {
		//if dump status analyzed should pull LLM
	}

	return analysis, nil
}
