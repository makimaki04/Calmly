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

// AnswerService delegates to repository which owns error logging.
// This layer only wraps and propagates errors — no duplicate logs.
type AnswerService struct {
	repo   repository.DumpAnswers
	logger *zap.Logger
}

func NewAnswerService(repo repository.DumpAnswers, logger *zap.Logger) *AnswerService {
	return &AnswerService{
		repo:   repo,
		logger: logger.With(zap.String("component", "service")),
	}
}

func (s *AnswerService) SaveAnswers(ctx context.Context, answers models.DumpAnswers) error {
	log := s.logger.With(
		zap.String("operation", "save_answers"),
		zap.String("dump_id", answers.DumpID.String()),
		zap.Int("answers_count", len(answers.Answers)),
	)

	log.Info("Save answers started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.repo.SaveAnswers(ctx, answers); err != nil {
		log.Error("Save answers failed", zap.Error(err))
		return fmt.Errorf("save answers: %w", err)
	}

	log.Info("Answers saved")

	return nil
}

func (s *AnswerService) GetAnswers(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnswers, error) {
	log := s.logger.With(
		zap.String("operation", "get_answers"),
		zap.String("dump_id", dumpID.String()),
	)

	log.Info("Get answers started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	answers, err := s.repo.GetAnswers(ctx, dumpID)
	if err != nil {
		log.Error("Get answers failed", zap.Error(err))
		return nil, fmt.Errorf("get answers: %w", err)
	}

	if answers == nil {
		log.Info("Answers not found")
		//if dump status waiting_answers should pull LLM
		return nil, nil
	}

	log.Info("Answers found", zap.Int("answers_count", len(answers.Answers)))

	return answers, nil
}
