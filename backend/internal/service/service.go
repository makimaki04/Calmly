package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

type Dump interface {
	CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error)
	AbandonDump(ctx context.Context, dumpID uuid.UUID) error
}

type Plan interface {
	CreatePlan(ctx context.Context, dumpID uuid.UUID, title string) (uuid.UUID, error)
	SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error
	GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error)
	GetUserSavedPlans(ctx context.Context, userID uuid.UUID) ([]models.Plan, error)
	DeleteSavedPlan(ctx context.Context, planID uuid.UUID) error
	DeleteUnsavedPlans(ctx context.Context, dumpID uuid.UUID) error
}

type Service struct {
	Dump
	Plan
}

func NewService(repo *repository.Repository, dumpExpTime time.Duration, logger *zap.Logger) *Service {
	planSvc := NewPlanService(repo.Plan, logger)
	return &Service{
		Dump: NewDumpService(repo.Dump, planSvc, dumpExpTime, logger),
		Plan: planSvc,
	}
}
