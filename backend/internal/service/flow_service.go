package service

import (
	"context"
	"errors"
	"fmt"

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
	plangen     llm.PlanGenerator
	logger      *zap.Logger
}

func NewFlowService(
	dumpSvc Dump,
	analysisSvc Analysis,
	answersSvc Answers,
	planSvc Plan,
	planItemSvc PlanItem,
	analysisGen llm.AnalysisGenerator,
	plangen llm.PlanGenerator,
	logger *zap.Logger,
) *FlowService {
	return &FlowService{
		dumpSvc:     dumpSvc,
		analysisSvc: analysisSvc,
		answersSvc:  answersSvc,
		planSvc:     planSvc,
		planItemSvc: planItemSvc,
		analysisGen: analysisGen,
		plangen:     plangen,
		logger:      logger.With(zap.String("component", "service")),
	}
}

func (f *FlowService) StartSession(ctx context.Context, userID uuid.UUID, rawText string) (models.DumpAnalysis, error) {
	log := f.logger.With(
		zap.String("operation", "start_session"),
		zap.String("user_id", userID.String()),
		zap.Int("raw_text_len", len(rawText)),
	)

	log.Info("Start session started")

	dumpID, err := f.dumpSvc.CreateDump(ctx, userID, rawText)
	if err != nil {
		log.Error("Start session failed", zap.Error(err))
		return models.DumpAnalysis{}, fmt.Errorf("create dump: %w", err)
	}

	analysis, err := f.analysisGen.GenerateAnalysis(ctx, rawText)
	if err != nil {
		log.Error("Start session failed", zap.Error(err), zap.String("dump_id", dumpID.String()))
		return models.DumpAnalysis{}, fmt.Errorf("generate analysis: %w", err)
	}

	analysis.DumpID = dumpID

	for i := 0; i < len(analysis.Questions); i++ {
		analysis.Questions[i].ID = uuid.New()
	}

	if err := f.dumpSvc.SetDumpStatus(ctx, dumpID, models.DumpStatusWaitingAnalysis); err != nil {
		log.Error("Start session failed", zap.Error(err), zap.String("dump_id", dumpID.String()))
		return models.DumpAnalysis{}, fmt.Errorf("set waiting analysis status: %w", err)
	}

	if err := f.dumpSvc.CompleteAnalysisStep(ctx, analysis); err != nil {
		log.Error("Start session failed", zap.Error(err), zap.String("dump_id", dumpID.String()))
		return models.DumpAnalysis{}, fmt.Errorf("complete analysis step: %w", err)
	}

	log.Info(
		"Session started",
		zap.String("dump_id", dumpID.String()),
		zap.Int("tasks_count", len(analysis.Tasks)),
		zap.Int("questions_count", len(analysis.Questions)),
	)

	return analysis, nil
}

var (
	ErrActiveDumpNotFound      = errors.New("active session not found")
	ErrDumpNotBelongUser       = errors.New("dump does not belong to current user session")
	ErrAnalysisNotFound        = errors.New("invalid session state: analysis is missing")
	ErrAnswersAlreadySubmitted = errors.New("answers already submitted")
)

func (f *FlowService) SubmitAnswers(ctx context.Context, userID uuid.UUID, answers models.DumpAnswers) (models.Plan, []models.PlanItem, error) {
	log := f.logger.With(
		zap.String("operation", "submit_answers"),
		zap.String("user_id", userID.String()),
		zap.String("dump_id", answers.DumpID.String()),
		zap.Int("answers_count", len(answers.Answers)),
	)

	log.Info("Submit answers started")

	activeDump, err := f.dumpSvc.GetUserDump(ctx, userID)
	if err != nil {
		log.Error("Submit answers failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get user dump: %w", err)
	}

	if activeDump == nil {
		log.Error("Submit answers failed", zap.Error(ErrActiveDumpNotFound))
		return models.Plan{}, []models.PlanItem{}, ErrActiveDumpNotFound
	}

	if activeDump.ID != answers.DumpID {
		log.Error(
			"Submit answers failed",
			zap.Error(ErrDumpNotBelongUser),
			zap.String("active_dump_id", activeDump.ID.String()),
		)
		return models.Plan{}, []models.PlanItem{}, ErrDumpNotBelongUser
	}

	analysis, err := f.analysisSvc.GetDumpAnalysis(ctx, activeDump.ID)
	if err != nil {
		log.Error("Submit answers failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get dump analysis: %w", err)
	}

	if analysis == nil {
		log.Error("Submit answers failed", zap.Error(ErrAnalysisNotFound))
		return models.Plan{}, []models.PlanItem{}, ErrAnalysisNotFound
	}

	_ = analysis
	_ = activeDump.RawText
	// LLM generate plan here using dump.raw_text, analysis and answers

	plan := models.Plan{
		DumpID: activeDump.ID,
		Title:  "Plan",
	}

	mockPlanItems := []models.PlanItem{}

	plan, items, err := f.planSvc.SubmitAnswersAndCreatePlan(ctx, answers, plan, mockPlanItems)
	if err != nil {
		if errors.Is(err, repository.ErrAnswersUniqueViolation) {
			log.Error("Submit answers failed", zap.Error(ErrAnswersAlreadySubmitted))
			return models.Plan{}, []models.PlanItem{}, ErrAnswersAlreadySubmitted
		}
		log.Error("Submit answers failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create plan items: %w", err)
	}

	log.Info(
		"Answers submitted",
		zap.String("plan_id", plan.ID.String()),
		zap.Int("items_count", len(items)),
	)

	return plan, items, nil
}

var ErrNoActiveSessionForRegeneration = errors.New("no active session for regeneration")

func (f *FlowService) GenerateNextPlanCandidate(ctx context.Context, userID uuid.UUID, fb models.UserFeedback) (models.Plan, []models.PlanItem, error) {
	log := f.logger.With(
		zap.String("operation", "generate_next_plan_candidate"),
		zap.String("user_id", userID.String()),
		zap.String("dump_id", fb.DumpID.String()),
	)

	log.Info("Generate next plan candidate started")

	dump, err := f.dumpSvc.GetUserDump(ctx, userID)
	if err != nil {
		log.Error("Generate next plan candidate failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get user dump: %w", err)
	}

	if dump == nil {
		log.Error("Generate next plan candidate failed", zap.Error(ErrNoActiveSessionForRegeneration))
		return models.Plan{}, []models.PlanItem{}, ErrNoActiveSessionForRegeneration
	}

	currPlans, err := f.planSvc.GetDumpPlans(ctx, dump.ID)
	if err != nil {
		log.Error("Generate next plan candidate failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("get current session plans: %w", err)
	}

	planIDs := make([]uuid.UUID, 0, len(currPlans))
	for _, p := range currPlans {
		planIDs = append(planIDs, p.ID)
	}

	planItems, err := f.planItemSvc.GetItemsByPlanIDs(ctx, planIDs)
	if err != nil {
		log.Error("Generate next plan candidate failed", zap.Error(err))
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
		log.Error("Generate next plan candidate failed", zap.Error(err))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create next plan candidate: %w", err)
	}
	newPlan.ID = newPlanID

	mockPlanItems := []models.PlanItem{}

	items, err := f.planItemSvc.CreateItems(ctx, mockPlanItems)
	if err != nil {
		log.Error("Generate next plan candidate failed", zap.Error(err), zap.String("plan_id", newPlan.ID.String()))
		return models.Plan{}, []models.PlanItem{}, fmt.Errorf("create next plan candidate items: %w", err)
	}

	log.Info(
		"Next plan candidate generated",
		zap.String("plan_id", newPlan.ID.String()),
		zap.Int("current_plans_count", len(currPlans)),
		zap.Int("current_items_count", len(planItems)),
		zap.Int("items_count", len(items)),
	)

	return newPlan, items, nil
}

func (f *FlowService) FinalizePlanSelection(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	log := f.logger.With(
		zap.String("operation", "finalize_plan_selection"),
		zap.String("dump_id", dumpID.String()),
		zap.String("plan_id", planID.String()),
	)

	log.Info("Finalize plan selection started")

	if err := f.planSvc.SavePlan(ctx, dumpID, planID); err != nil {
		log.Error("Finalize plan selection failed", zap.Error(err))
		return fmt.Errorf("save selected plan: %w", err)
	}

	log.Info("Plan selection finalized")

	return nil
}
