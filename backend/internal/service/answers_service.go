package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

// AnswerService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type AnswerService struct {
	repo     repository.DumpAnswers
	logger *zap.Logger
}

func NewAnswerService(repo repository.DumpAnswers, logger *zap.Logger) *AnswerService {
	return &AnswerService{
		repo:     repo,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *AnswerService) SaveAnswers(ctx context.Context, answers models.DumpAnswers) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.SaveAnswers(ctx, answers); err != nil {
		return err
	}

	return nil
}

func (s *AnswerService) GetAnswers(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnswers, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	answers, err := s.repo.GetAnswers(ctx, dumpID) 
	if err != nil {
		return nil, err
	}

	if answers == nil {
		//if dump status waiting_answers should pull LLM
	}

	return answers, nil
}
