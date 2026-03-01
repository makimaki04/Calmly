package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

type Dump interface {
	CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error)
	AbandonDump(ctx context.Context, dumpID uuid.UUID) error
}

type Plan interface {
}

type Service struct {
	Dump
	Plan
}

func NewService(repo *repository.Repository, dumpExpTime time.Duration,logger *zap.Logger) *Service {
	return &Service{
		Dump: NewDumpService(repo.Dump, repo.Plan, dumpExpTime, logger),
	}
}