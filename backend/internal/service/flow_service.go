package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

type FlowService struct {
	dumpSvc     Dump
	analysisSvc Analysis
	answersSvc  Answers
	planSvc     Plan
	planItemSvc PlanItem
	logger      *zap.Logger
}

func NewFlowService(
	dumpSvc Dump,
	analysisSvc Analysis,
	answersSvc Answers,
	planSvc Plan,
	planItemSvc PlanItem,
	logger *zap.Logger,
) *FlowService {
	return &FlowService{
		dumpSvc:     dumpSvc,
		analysisSvc: analysisSvc,
		answersSvc:  answersSvc,
		planSvc:     planSvc,
		planItemSvc: planItemSvc,
		logger:      logger.With(zap.String("component", "service")),
	}
}

func (f *FlowService) StartSession(ctx context.Context, userID uuid.UUID, rawText string) (models.DumpAnalysis, error) {
	dumpID, err := f.dumpSvc.CreateDump(ctx, userID, rawText)
	if err != nil {
		return models.DumpAnalysis{}, err
	}

	// LLM generate analysis here

	analysis := models.DumpAnalysis{
		DumpID: dumpID,
	}

	if err := f.analysisSvc.SaveDumpAnalysis(ctx, analysis); err != nil {
		return models.DumpAnalysis{}, err
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusAnalyzed); err != nil {
		return models.DumpAnalysis{}, err
	}

	return analysis, nil
}

func (f *FlowService) SubmitAnswers() {

}

func (f *FlowService) GenerateNextPlanCandidate() {

}

func (f *FlowService) FinalizePlanSelection() {

}
