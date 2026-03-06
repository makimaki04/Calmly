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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.SaveAnalysis(ctx, analysis); err != nil {
		return fmt.Errorf("save dump analysis: %w", err)
	}

	return nil
}

func (s *AnalysisService) GetDumpAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	analysis, err := s.repo.GetAnalysis(ctx, dumpID)
	if err != nil {
		return nil, fmt.Errorf("get dump analysis: %w", err)
	}

	return analysis, nil
}
