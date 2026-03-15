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
	log := s.logger.With(
		zap.String("operation", "create_dump"),
		zap.String("user_id", userID.String()),
		zap.Int("raw_text_len", len(rawText)),
	)

	log.Info("Create dump started")

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
		log.Error("Create dump failed", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("create dump: %w", err)
	}

	log.Info("Dump created", zap.String("dump_id", id.String()))

	return id, nil
}

func (s *DumpService) GetUserDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error) {
	log := s.logger.With(
		zap.String("operation", "get_user_dump"),
		zap.String("user_id", userID.String()),
	)

	log.Info("Get user dump started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dump, err := s.repo.GetActiveDump(ctx, userID)
	if err != nil {
		log.Error("Get user dump failed", zap.Error(err))
		return nil, fmt.Errorf("get user dump: %w", err)
	}

	if dump == nil {
		log.Info("Active dump not found")
		return nil, nil
	}

	log.Info("Active dump found", zap.String("dump_id", dump.ID.String()))

	return dump, nil
}

func (s *DumpService) SetDumpStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error {
	log := s.logger.With(
		zap.String("operation", "set_dump_status"),
		zap.String("dump_id", dumpID.String()),
		zap.String("status", string(status)),
	)

	log.Info("Set dump status started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.UpdateStatus(ctx, dumpID, status); err != nil {
		log.Error("Set dump status failed", zap.Error(err))
		return fmt.Errorf("set dump status: %w", err)
	}

	log.Info("Dump status updated")

	return nil
}

func (s *DumpService) AbandonDump(ctx context.Context, dumpID uuid.UUID) error {
	log := s.logger.With(
		zap.String("operation", "abandon_dump"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Abandon dump started")

	if err := s.repo.ClearRawText(ctx, dumpID); err != nil {
		log.Error("Abandon dump failed", zap.Error(err))
		return fmt.Errorf("abandon dump: %w", err)
	}

	log.Info("Dump abandoned")

	return nil
}

func (s *DumpService) CompleteAnalysisStep(ctx context.Context, dumpAnalysis models.DumpAnalysis) error {
	log := s.logger.With(
		zap.String("operation", "complete_analysis_step"),
		zap.String("dump_id", dumpAnalysis.DumpID.String()),
		zap.Int("tasks_count", len(dumpAnalysis.Tasks)),
		zap.Int("questions_count", len(dumpAnalysis.Questions)),
	)

	log.Info("Complete analysis step started")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.repo.CompleteAnalysisStep(ctx, dumpAnalysis)
	if err != nil {
		log.Error("Complete analysis step failed", zap.Error(err))
		return fmt.Errorf("complete analysis step: %w", err)
	}

	log.Info("Analysis step completed")

	return nil
}
