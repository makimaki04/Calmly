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

// DumpService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type DumpService struct {
	repo        repository.Dump
	dumpExpTime time.Duration
	logger      *zap.Logger
}

func NewDumpService(repo repository.Dump, dumpExpTime time.Duration, logger *zap.Logger) *DumpService {
	return &DumpService{
		repo:        repo,
		dumpExpTime: dumpExpTime,
		logger:      logger.With(zap.String("component", "service")),
	}
}

func (s *DumpService) CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	expTime := time.Now().Add(s.dumpExpTime)
	id, err := s.repo.CreateDump(ctx, userID, models.Dump{
		UserID:       &userID,
		GuestID:      nil,
		Status:       models.DumpStatusNew,
		RawText:      &rawText,
		RawExpiresAt: &expTime,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create dump: %w", err)
	}

	return id, nil
}

func (s *DumpService) GetUserDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dump, err := s.repo.GetActiveDump(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user dump: %w", err)
	}

	return dump, nil
}

func (s *DumpService) SetDumpStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.UpdateStatus(ctx, dumpID, status); err != nil {
		return fmt.Errorf("set dump status: %w", err)
	}

	return nil
}

func (s *DumpService) AbandonDump(ctx context.Context, dumpID uuid.UUID) error {
	if err := s.repo.ClearRawText(ctx, dumpID); err != nil {
		return fmt.Errorf("abandon dump: %w", err)
	}

	return nil
}

func (s *DumpService) CompleteAnalysisStep(ctx context.Context, dumpAnalysis models.DumpAnalysis) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.repo.CompleteAnalysisStep(ctx, dumpAnalysis)
	if err != nil {
		return fmt.Errorf("complete analysis step: %w", err)
	}

	return nil
}
