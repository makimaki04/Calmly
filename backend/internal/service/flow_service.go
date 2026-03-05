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

func (f *FlowService) SubmitAnswers(ctx context.Context, answers models.DumpAnswers) (models.Plan, []models.PlanItem, error) {
	if err := f.answersSvc.SaveAnswers(ctx, answers); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	dumpID := answers.DumpID

	// LLM generate plan here

	plan := models.Plan{
		DumpID: dumpID,
		Title:  "Plan",
	}

	planID, err := f.planSvc.CreatePlan(ctx, plan.DumpID, plan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}
	plan.ID = planID

	planItems := []models.PlanItem{}
	if err := f.planItemSvc.CreateItems(ctx, planItems); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	return plan, planItems, nil
}

func (f *FlowService) GenerateNextPlanCandidate(ctx context.Context, fb models.UserFeedback) (models.Plan, []models.PlanItem, error) {
	currPlans, err := f.planSvc.GetDumpPlans(ctx, fb.DumpID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	_ = currPlans
	// LLM generate new plan and plan_items here folowing user feedback and current plans

	plan := models.Plan{
		DumpID: fb.DumpID,
		Title:  "Plan",
	}

	planID, err := f.planSvc.CreatePlan(ctx, plan.DumpID, plan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}
	plan.ID = planID

	planItems := []models.PlanItem{}
	if err := f.planItemSvc.CreateItems(ctx, planItems); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	return plan, planItems, nil
}

func (f *FlowService) FinalizePlanSelection(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	if err := f.planSvc.SavePlan(ctx, dumpID, planID); err != nil {
		return err
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusPlanned); err != nil {
		return err
	}

	return nil
}
