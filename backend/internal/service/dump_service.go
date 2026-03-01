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

type PlanCleaner interface {
	DeleteUnsavedPlans(ctx context.Context, dumpID uuid.UUID) error
}

type DumpService struct {
	repo        repository.Dump
	planCleaner PlanCleaner
	dumpExpTime time.Duration
	logger      *zap.Logger
}

func NewDumpService(repo repository.Dump, planCleaner PlanCleaner, dumpExpTime time.Duration, logger *zap.Logger) *DumpService {
	return &DumpService{
		repo:        repo,
		planCleaner: planCleaner,
		dumpExpTime: dumpExpTime,
		logger:      logger.With(zap.String("component", "service")),
	}
}

func (s *DumpService) CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	activeDump, err := s.repo.GetActiveDump(ctx, userID)
	if err != nil {
		// Error is logged inside repository. Avoid duplicates here.
		return uuid.UUID{}, fmt.Errorf("get active dump: %w", err)
	}

	if activeDump != nil {
		if err := s.AbandonDump(ctx, activeDump.ID); err != nil {
			return uuid.UUID{}, fmt.Errorf("abandon previous dump: %w", err)
		}
	}

	expTime := time.Now().Add(s.dumpExpTime)
	id, err := s.repo.CreateDump(ctx, models.Dump{
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

func (s *DumpService) AbandonDump(ctx context.Context, dumpID uuid.UUID) error {
	if err := s.repo.ClearRawText(ctx, dumpID); err != nil {
		return fmt.Errorf("abandon dump: %w", err)
	}

	if err := s.planCleaner.DeleteUnsavedPlans(ctx, dumpID); err != nil {
		return fmt.Errorf("delete unsaved plans: %w", err)
	}

	return nil
}
