package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/llm"
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
	analysisGen llm.AnalysisGenerator
	planGen     llm.PlanGenerator
	logger      *zap.Logger
}

func NewFlowService(
	dumpSvc Dump,
	analysisSvc Analysis,
	answersSvc Answers,
	planSvc Plan,
	planItemSvc PlanItem,
	analysisGen llm.AnalysisGenerator,
	planGen llm.PlanGenerator,
	logger *zap.Logger,
) *FlowService {
	return &FlowService{
		dumpSvc:     dumpSvc,
		analysisSvc: analysisSvc,
		answersSvc:  answersSvc,
		planSvc:     planSvc,
		planItemSvc: planItemSvc,
		analysisGen: analysisGen,
		planGen:     planGen,
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
		DumpID:    dumpID,
		Tasks:     []models.Task{},
		Questions: []models.Question{},
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
	ErrEmpryDumpRawText        = errors.New("dump raw text empty")
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

	if activeDump.RawText == nil {
		return models.Plan{}, []models.PlanItem{}, ErrEmpryDumpRawText
	}

	analysis, err := f.analysisSvc.GetDumpAnalysis(ctx, activeDump.ID)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get dump analysis: %w", err)
	}

	if analysis == nil {
		return models.Plan{}, []models.PlanItem{}, ErrAnalysisNotFound
	}

	plan, planItems, err := f.planGen.GeneratePlan(ctx, *activeDump.RawText, analysis.Tasks, analysis.Questions, answers.Answers)
	if err != nil {
		log.Error("Submit answers and generate plan failed", zap.Error(err), zap.String("dump_id", activeDump.ID.String()))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("generate plan: %w", err)
	}

	plan.DumpID = activeDump.ID

	for i := range planItems {
		planItems[i].Ord = i + 1
	}

	plan, items, err := f.planSvc.SubmitAnswersAndCreatePlan(ctx, answers, plan, planItems)
	if err != nil {
		if errors.Is(err, repository.ErrAnswersUniqueViolation) {
			return models.Plan{}, []models.PlanItem{}, ErrAnswersAlreadySubmitted
		}
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

	currPlans, err := f.planSvc.GetDumpPlans(ctx, dump.ID)
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
		DumpID: dump.ID,
		Title:  "Plan",
	}

	newPlanID, err := f.planSvc.CreatePlan(ctx, newPlan.DumpID, newPlan.Title)
	if err != nil {
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create next plan candidate: %w", err)
	}
	newPlan.ID = newPlanID

	mockPlanItems := []models.PlanItem{}

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

	return nil
}
