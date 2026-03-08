package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
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
		return models.DumpAnalysis{}, fmt.Errorf("create dump: %w", err)
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
				ID:   uuid.New(),
				Text: "What's going on",
			},
			{
				ID:   uuid.New(),
				Text: "Another qusetion",
			},
			{
				ID:   uuid.New(),
				Text: "Question number 3",
			},
		},
		Mood:      &mood,
		Quote:     &quote,
		CreatedAt: time.Now(),
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusWaitingAnalysis); err != nil {
		return models.DumpAnalysis{}, fmt.Errorf("set waiting analysis status: %w", err)
	}

	if err := f.dumpSvc.CompleteAnalysisStep(ctx, mockAnalysis); err != nil {
		return models.DumpAnalysis{}, fmt.Errorf("complete analysis step: %w", err)
	}

	return mockAnalysis, nil
}

var (
	ErrActiveDumpNotFound      = errors.New("active session not found")
	ErrDumpNotBelongUser       = errors.New("dump does not belong to current user session")
	ErrAnalysisNotFound        = errors.New("invalid session state: analysis is missing")
	ErrAnswersAlreadySubmitted = errors.New("answers already submitted")
)

func (f *FlowService) SubmitAnswers(ctx context.Context, userID uuid.UUID, answers models.DumpAnswers) (models.Plan, []models.PlanItem, error) {
	activeDump, err := f.dumpSvc.GetUserDump(ctx, userID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get user dump: %w", err)
	}

	if activeDump == nil {
		return models.Plan{}, []models.PlanItem{}, ErrActiveDumpNotFound
	}

	if activeDump.ID != answers.DumpID {
		return models.Plan{}, []models.PlanItem{}, ErrDumpNotBelongUser
	}

	analysis, err := f.analysisSvc.GetDumpAnalysis(ctx, activeDump.ID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get dump analysis: %w", err)
	}

	if analysis == nil {
		return models.Plan{}, []models.PlanItem{}, ErrAnalysisNotFound
	}

	if err := f.answersSvc.SaveAnswers(ctx, answers); err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			return models.Plan{}, []models.PlanItem{}, ErrAnswersAlreadySubmitted
		}
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("save answers: %w", err)
	}

	_ = analysis
	_ = activeDump.RawText
	// LLM generate plan here using dump.raw_text, analysis and answers

	plan := models.Plan{
		DumpID: activeDump.ID,
		Title:  "Plan",
	}

	planID, err := f.planSvc.CreatePlan(ctx, plan.DumpID, plan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create plan: %w", err)
	}
	plan.ID = planID

	mockPlanItems := []models.PlanItem{
		{
			PlanID: planID,
			Ord:    1,
			Text:   "Make some food",
			Done:   false,
		},
		{
			PlanID: planID,
			Ord:    2,
			Text:   "Sleep",
			Done:   false,
		},
		{
			PlanID: planID,
			Ord:    3,
			Text:   "Gym",
			Done:   false,
		},
	}
	items, err := f.planItemSvc.CreateItems(ctx, mockPlanItems)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create plan items: %w", err)
	}

	return plan, items, nil
}

var ErrNoActiveSessionForRegeneration = errors.New("no active session for regeneration")

func (f *FlowService) GenerateNextPlanCandidate(ctx context.Context, userID uuid.UUID, fb models.UserFeedback) (models.Plan, []models.PlanItem, error) {
	dump, err := f.dumpSvc.GetUserDump(ctx, userID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get user dump: %w", err)
	}

	if dump == nil {
		return models.Plan{}, []models.PlanItem{}, ErrNoActiveSessionForRegeneration
	}

	currPlans, err := f.planSvc.GetDumpPlans(ctx, fb.DumpID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get current session plans: %w", err)
	}

	planIDs := make([]uuid.UUID, 0, len(currPlans))
	for _, p := range currPlans {
		planIDs = append(planIDs, p.ID)
	}

	planItems, err := f.planItemSvc.GetItemsByPlanIDs(ctx, planIDs)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get plan items by plan ids: %w", err)
	}

	_ = planItems
	// LLM generate new plan and plan_items here following user feedback and current plans and planItems

	newPlan := models.Plan{
		DumpID: fb.DumpID,
		Title:  "Plan",
	}

	newPlanID, err := f.planSvc.CreatePlan(ctx, newPlan.DumpID, newPlan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create next plan candidate: %w", err)
	}
	newPlan.ID = newPlanID

	mockPlanItems := []models.PlanItem{
		{
			PlanID: newPlan.ID,
			Ord:    1,
			Text:   "Wake up",
			Done:   false,
		},
		{
			PlanID: newPlan.ID,
			Ord:    2,
			Text:   "Football",
			Done:   false,
		},
		{
			PlanID: newPlan.ID,
			Ord:    3,
			Text:   "Home work",
			Done:   false,
		},
	}

	items, err := f.planItemSvc.CreateItems(ctx, mockPlanItems)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create next plan candidate items: %w", err)
	}

	return newPlan, items, nil
}

func (f *FlowService) FinalizePlanSelection(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	if err := f.planSvc.SavePlan(ctx, dumpID, planID); err != nil {
		return fmt.Errorf("save selected plan: %w", err)
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusPlanned); err != nil {
		return fmt.Errorf("set planned status: %w", err)
	}

	return nil
}
