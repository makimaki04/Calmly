package service

import (
	"context"
	"time"

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
	mood := models.MoodTired
	quote := ""
	mockAnalysis := models.DumpAnalysis{
		DumpID: dumpID,
		Tasks: []models.Task{
			{
				Text:     "Task 1",
				Priority: "low",
				Category: "work",
			},
			{
				Text:     "Task 2",
				Priority: "high",
				Category: "mind",
			},
			{
				Text:     "Task 3",
				Priority: "high",
				Category: "life balance",
			},
		},
		Questions: []models.Question{
			{
				Text: "What's going on",
			},
		},
		Mood:      &mood,
		Quote:     &quote,
		CreatedAt: time.Now(),
	}

	if err := f.analysisSvc.SaveDumpAnalysis(ctx, mockAnalysis); err != nil {
		return models.DumpAnalysis{}, err
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusAnalyzed); err != nil {
		return models.DumpAnalysis{}, err
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusWaitingAnswers); err != nil {
		return models.DumpAnalysis{}, err
	}

	return mockAnalysis, nil
}

func (f *FlowService) SubmitAnswers(ctx context.Context, answers models.DumpAnswers) (models.Plan, []models.PlanItem, error) {
	if err := f.answersSvc.SaveAnswers(ctx, answers); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	dumpID := answers.DumpID
	analysis, err :=f.analysisSvc.GetDumpAnalysis(ctx, dumpID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	_ = analysis
	// LLM generate plan here using analysis and answers

	plan := models.Plan{
		DumpID: dumpID,
		Title:  "Plan",
	}

	planID, err := f.planSvc.CreatePlan(ctx, plan.DumpID, plan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}
	plan.ID = planID

	mockPlanItems := []models.PlanItem{
		{
			ID:     uuid.New(),
			PlanID: planID,
			Ord:    1,
			Text:   "Make some food",
			Done:   false,
		},
		{
			ID:     uuid.New(),
			PlanID: planID,
			Ord:    2,
			Text:   "Sleep",
			Done:   false,
		},
		{
			ID:     uuid.New(),
			PlanID: planID,
			Ord:    3,
			Text:   "Gym",
			Done:   false,
		},
	}
	if err := f.planItemSvc.CreateItems(ctx, mockPlanItems); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	return plan, mockPlanItems, nil
}

func (f *FlowService) GenerateNextPlanCandidate(ctx context.Context, fb models.UserFeedback) (models.Plan, []models.PlanItem, error) {
	currPlans, err := f.planSvc.GetDumpPlans(ctx, fb.DumpID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	planIDs := make([]uuid.UUID, 0, len(currPlans))
	for _, p := range currPlans {
		planIDs = append(planIDs, p.ID)
	}

	planItems, err := f.planItemSvc.GetItemsByPlanIDs(ctx, planIDs)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	_ = planItems
	// LLM generate new plan and plan_items here following user feedback and current plans and planItems

	newPlan := models.Plan{
		DumpID: fb.DumpID,
		Title:  "Plan",
	}

	newPlanID, err := f.planSvc.CreatePlan(ctx, newPlan.DumpID, newPlan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}
	newPlan.ID = newPlanID

	newPlanItems := []models.PlanItem{}
	if err := f.planItemSvc.CreateItems(ctx, newPlanItems); err != nil {
		return models.Plan{}, []models.PlanItem{}, err
	}

	return newPlan, newPlanItems, nil
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
