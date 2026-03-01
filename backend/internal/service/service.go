package service

import (
	"context"

	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

type Dump interface {
	CreateDump(ctx context.Context)
}

type Plan interface {
}

type Service struct {
	Dump
	Plan
}

func NewService(repo repository.Repository, logger *zap.Logger) *Service {
	return &Service{

	}
}