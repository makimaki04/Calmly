package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"go.uber.org/zap"
)

type dumpStub struct {
	createDumpFn           func(context.Context, uuid.UUID, string) (uuid.UUID, error)
	getUserDumpFn          func(context.Context, uuid.UUID) (*models.Dump, error)
	setDumpStatusFn        func(context.Context, uuid.UUID, models.DumpStatus) error
	abandonDumpFn          func(context.Context, uuid.UUID) error
	completeAnalysisStepFn func(context.Context, models.DumpAnalysis) error
}

func (s *dumpStub) CreateDump(ctx context.Context, userID uuid.UUID, rawText string) (uuid.UUID, error) {
	return s.createDumpFn(ctx, userID, rawText)
}
func (s *dumpStub) GetUserDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error) {
	return s.getUserDumpFn(ctx, userID)
}
func (s *dumpStub) SetDumpStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error {
	return s.setDumpStatusFn(ctx, dumpID, status)
}
func (s *dumpStub) AbandonDump(ctx context.Context, dumpID uuid.UUID) error {
	if s.abandonDumpFn == nil {
		return nil
	}
	return s.abandonDumpFn(ctx, dumpID)
}
func (s *dumpStub) CompleteAnalysisStep(ctx context.Context, analysis models.DumpAnalysis) error {
	return s.completeAnalysisStepFn(ctx, analysis)
}

type analysisStub struct {
	getDumpAnalysisFn func(context.Context, uuid.UUID) (*models.DumpAnalysis, error)
}

func (s *analysisStub) SaveDumpAnalysis(context.Context, models.DumpAnalysis) error { return nil }
func (s *analysisStub) GetDumpAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error) {
	return s.getDumpAnalysisFn(ctx, dumpID)
}

type answersStub struct {
	getAnswersFn func(context.Context, uuid.UUID) (*models.DumpAnswers, error)
}

func (s *answersStub) SaveAnswers(context.Context, models.DumpAnswers) error { return nil }
func (s *answersStub) GetAnswers(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnswers, error) {
	return s.getAnswersFn(ctx, dumpID)
}

type planStub struct {
	submitAnswersAndCreatePlanFn func(context.Context, models.DumpAnswers, models.Plan, []models.PlanItem) (models.Plan, []models.PlanItem, error)
	createNewPlanCandidateFn     func(context.Context, models.Plan, []models.PlanItem) (models.Plan, []models.PlanItem, error)
	createPlanFn                 func(context.Context, uuid.UUID, string) (uuid.UUID, error)
	savePlanFn                   func(context.Context, uuid.UUID, uuid.UUID) error
	getDumpPlansFn               func(context.Context, uuid.UUID) ([]models.Plan, error)
	getLastGeneratedPlanFn       func(context.Context, uuid.UUID) (models.Plan, error)
}

func (s *planStub) CreatePlan(ctx context.Context, dumpID uuid.UUID, title string) (uuid.UUID, error) {
	return s.createPlanFn(ctx, dumpID, title)
}
func (s *planStub) SubmitAnswersAndCreatePlan(ctx context.Context, answers models.DumpAnswers, plan models.Plan, items []models.PlanItem) (models.Plan, []models.PlanItem, error) {
	return s.submitAnswersAndCreatePlanFn(ctx, answers, plan, items)
}
func (s *planStub) CreateNewPlanCandidate(ctx context.Context, plan models.Plan, items []models.PlanItem) (models.Plan, []models.PlanItem, error) {
	if s.createNewPlanCandidateFn == nil {
		return models.Plan{}, nil, nil
	}
	return s.createNewPlanCandidateFn(ctx, plan, items)
}
func (s *planStub) SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	return s.savePlanFn(ctx, dumpID, planID)
}
func (s *planStub) GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	return s.getDumpPlansFn(ctx, dumpID)
}
func (s *planStub) GetLastGeneratedPlan(ctx context.Context, dumpID uuid.UUID) (models.Plan, error) {
	return s.getLastGeneratedPlanFn(ctx, dumpID)
}
func (s *planStub) GetUserSavedPlans(context.Context, uuid.UUID) ([]models.Plan, error) { return nil, nil }
func (s *planStub) DeleteSavedPlan(context.Context, uuid.UUID) error                     { return nil }

type planItemStub struct {
	createItemsFn       func(context.Context, []models.PlanItem) ([]models.PlanItem, error)
	getItemsByPlanIDsFn func(context.Context, []uuid.UUID) ([]models.PlanItem, error)
}

func (s *planItemStub) CreateItems(ctx context.Context, items []models.PlanItem) ([]models.PlanItem, error) {
	return s.createItemsFn(ctx, items)
}
func (s *planItemStub) ToggleItem(context.Context, uuid.UUID, bool) error { return nil }
func (s *planItemStub) AddItem(context.Context, models.PlanItem) (models.PlanItem, error) {
	return models.PlanItem{}, nil
}
func (s *planItemStub) DeleteItem(context.Context, uuid.UUID) error                { return nil }
func (s *planItemStub) ReorderItems(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (s *planItemStub) GetItemsByPlanIDs(ctx context.Context, planIDs []uuid.UUID) ([]models.PlanItem, error) {
	return s.getItemsByPlanIDsFn(ctx, planIDs)
}

type analysisGenStub struct {
	generateAnalysisFn func(context.Context, string) (models.DumpAnalysis, error)
}

func (s *analysisGenStub) GenerateAnalysis(ctx context.Context, rawText string) (models.DumpAnalysis, error) {
	return s.generateAnalysisFn(ctx, rawText)
}

type planGenStub struct {
	generatePlanFn   func(context.Context, string, []models.Task, []models.Question, []models.Answer) (models.Plan, []models.PlanItem, error)
	regeneratePlanFn func(context.Context, string, models.DumpAnalysis, models.DumpAnswers, models.Plan, []models.PlanItem, string) (models.Plan, []models.PlanItem, error)
}

func (s *planGenStub) GeneratePlan(ctx context.Context, rawText string, tasks []models.Task, questions []models.Question, answers []models.Answer) (models.Plan, []models.PlanItem, error) {
	return s.generatePlanFn(ctx, rawText, tasks, questions, answers)
}

func (s *planGenStub) RegeneratePlan(ctx context.Context, rawText string, analysis models.DumpAnalysis, answers models.DumpAnswers, plan models.Plan, planItems []models.PlanItem, feedback string) (models.Plan, []models.PlanItem, error) {
	if s.regeneratePlanFn == nil {
		return models.Plan{}, nil, nil
	}
	return s.regeneratePlanFn(ctx, rawText, analysis, answers, plan, planItems, feedback)
}

func TestFlowService_StartSession(t *testing.T) {
	userID := uuid.New()
	dumpID := uuid.New()
	var completed models.DumpAnalysis

	svc := NewFlowService(
		&dumpStub{
			createDumpFn: func(_ context.Context, gotUserID uuid.UUID, rawText string) (uuid.UUID, error) {
				if gotUserID != userID || rawText != "raw text" {
					t.Fatalf("CreateDump() got (%v, %q)", gotUserID, rawText)
				}
				return dumpID, nil
			},
			setDumpStatusFn: func(_ context.Context, gotDumpID uuid.UUID, status models.DumpStatus) error {
				if gotDumpID != dumpID || status != models.DumpStatusWaitingAnalysis {
					t.Fatalf("SetDumpStatus() got (%v, %v)", gotDumpID, status)
				}
				return nil
			},
			completeAnalysisStepFn: func(_ context.Context, analysis models.DumpAnalysis) error {
				completed = analysis
				return nil
			},
		},
		&analysisStub{},
		&answersStub{},
		&planStub{},
		&planItemStub{},
		&analysisGenStub{
			generateAnalysisFn: func(_ context.Context, rawText string) (models.DumpAnalysis, error) {
				if rawText != "raw text" {
					t.Fatalf("GenerateAnalysis() rawText = %q", rawText)
				}
				return models.DumpAnalysis{
					Tasks:     []models.Task{{Text: "task"}},
					Questions: []models.Question{{Text: "q1"}, {Text: "q2"}},
				}, nil
			},
		},
		nil,
		zap.NewNop(),
	)

	got, err := svc.StartSession(context.Background(), userID, "raw text")
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}

	if got.DumpID != dumpID {
		t.Fatalf("StartSession() dumpID = %v, want %v", got.DumpID, dumpID)
	}
	for i, q := range got.Questions {
		if q.ID == uuid.Nil {
			t.Fatalf("StartSession() question %d id is nil", i)
		}
	}
	if completed.DumpID != dumpID {
		t.Fatalf("CompleteAnalysisStep() dumpID = %v, want %v", completed.DumpID, dumpID)
	}
}

func TestFlowService_SubmitAnswers(t *testing.T) {
	userID := uuid.New()
	dumpID := uuid.New()
	otherDumpID := uuid.New()
	questionID := uuid.New()

	tests := []struct {
		name     string
		dumpSvc  *dumpStub
		analysis *analysisStub
		planSvc  *planStub
		wantErr  error
	}{
		{
			name: "active dump not found",
			dumpSvc: &dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) { return nil, nil },
			},
			analysis: &analysisStub{},
			planSvc:  &planStub{},
			wantErr:  ErrActiveDumpNotFound,
		},
		{
			name: "dump belongs to another session",
			dumpSvc: &dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
					return &models.Dump{ID: otherDumpID}, nil
				},
			},
			analysis: &analysisStub{},
			planSvc:  &planStub{},
			wantErr:  ErrDumpNotBelongUser,
		},
		{
			name: "analysis missing",
			dumpSvc: &dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
					raw := "raw"
					return &models.Dump{ID: dumpID, RawText: &raw}, nil
				},
			},
			analysis: &analysisStub{
				getDumpAnalysisFn: func(context.Context, uuid.UUID) (*models.DumpAnalysis, error) { return nil, nil },
			},
			planSvc: &planStub{},
			wantErr: ErrAnalysisNotFound,
		},
		{
			name: "raw text missing",
			dumpSvc: &dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
					return &models.Dump{ID: dumpID, RawText: nil}, nil
				},
			},
			analysis: &analysisStub{},
			planSvc:  &planStub{},
			wantErr:  ErrEmpryDumpRawText,
		},
		{
			name: "answers already submitted",
			dumpSvc: &dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
					raw := "raw"
					return &models.Dump{ID: dumpID, RawText: &raw}, nil
				},
			},
			analysis: &analysisStub{
				getDumpAnalysisFn: func(context.Context, uuid.UUID) (*models.DumpAnalysis, error) {
					return &models.DumpAnalysis{DumpID: dumpID}, nil
				},
			},
			planSvc: &planStub{
				submitAnswersAndCreatePlanFn: func(context.Context, models.DumpAnswers, models.Plan, []models.PlanItem) (models.Plan, []models.PlanItem, error) {
					return models.Plan{}, nil, repository.ErrAnswersUniqueViolation
				},
			},
			wantErr: ErrAnswersAlreadySubmitted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewFlowService(tt.dumpSvc, tt.analysis, &answersStub{}, tt.planSvc, &planItemStub{}, &analysisGenStub{}, &planGenStub{
				generatePlanFn: func(context.Context, string, []models.Task, []models.Question, []models.Answer) (models.Plan, []models.PlanItem, error) {
					return models.Plan{}, nil, nil
				},
			}, zap.NewNop())
			_, _, err := svc.SubmitAnswers(context.Background(), userID, models.DumpAnswers{
				DumpID:  dumpID,
				Answers: []models.Answer{{QuestionID: questionID, Text: "answer"}},
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SubmitAnswers() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlowService_SubmitAnswers_Success(t *testing.T) {
	userID := uuid.New()
	dumpID := uuid.New()
	planID := uuid.New()
	questionID := uuid.New()

	svc := NewFlowService(
		&dumpStub{
			getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
				raw := "raw"
				return &models.Dump{ID: dumpID, RawText: &raw}, nil
			},
		},
		&analysisStub{
			getDumpAnalysisFn: func(context.Context, uuid.UUID) (*models.DumpAnalysis, error) {
				return &models.DumpAnalysis{
					DumpID:    dumpID,
					Tasks:     []models.Task{{Text: "task from analysis"}},
					Questions: []models.Question{{Text: "question from analysis"}},
				}, nil
			},
		},
		&answersStub{},
		&planStub{
			submitAnswersAndCreatePlanFn: func(_ context.Context, answers models.DumpAnswers, plan models.Plan, items []models.PlanItem) (models.Plan, []models.PlanItem, error) {
				if answers.DumpID != dumpID || plan.DumpID != dumpID || plan.Title != "Generated plan" {
					t.Fatalf("unexpected input: answers=%+v plan=%+v items=%+v", answers, plan, items)
				}
				if len(items) != 2 || items[0].Ord != 1 || items[1].Ord != 2 {
					t.Fatalf("unexpected generated items: %+v", items)
				}
				plan.ID = planID
				return plan, []models.PlanItem{{ID: uuid.New(), PlanID: planID, Ord: 1, Text: "item-1"}}, nil
			},
		},
		&planItemStub{},
		&analysisGenStub{},
		&planGenStub{
			generatePlanFn: func(_ context.Context, rawText string, tasks []models.Task, questions []models.Question, answers []models.Answer) (models.Plan, []models.PlanItem, error) {
				if rawText != "raw" {
					t.Fatalf("GeneratePlan() rawText = %q", rawText)
				}
				if len(tasks) != 1 || tasks[0].Text != "task from analysis" {
					t.Fatalf("GeneratePlan() tasks = %+v", tasks)
				}
				if len(questions) != 1 || questions[0].Text != "question from analysis" {
					t.Fatalf("GeneratePlan() questions = %+v", questions)
				}
				if len(answers) != 1 || answers[0].QuestionID != questionID {
					t.Fatalf("GeneratePlan() answers = %+v", answers)
				}
				return models.Plan{
						Title: "Generated plan",
					}, []models.PlanItem{
						{Text: "first"},
						{Text: "second"},
					}, nil
			},
		},
		zap.NewNop(),
	)

	plan, items, err := svc.SubmitAnswers(context.Background(), userID, models.DumpAnswers{
		DumpID:  dumpID,
		Answers: []models.Answer{{QuestionID: questionID, Text: "answer"}},
	})
	if err != nil {
		t.Fatalf("SubmitAnswers() error = %v", err)
	}
	if plan.ID != planID || len(items) != 1 {
		t.Fatalf("SubmitAnswers() got (%+v, %+v)", plan, items)
	}
}

func TestFlowService_GenerateNextPlanCandidate(t *testing.T) {
	userID := uuid.New()
	dumpID := uuid.New()
	planID := uuid.New()

	t.Run("no active session", func(t *testing.T) {
		svc := NewFlowService(
			&dumpStub{getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) { return nil, nil }},
			&analysisStub{},
			&answersStub{},
			&planStub{},
			&planItemStub{},
			&analysisGenStub{},
			nil,
			zap.NewNop(),
		)

		_, _, err := svc.GenerateNextPlanCandidate(context.Background(), userID, models.UserFeedback{DumpID: dumpID, Text: "feedback"})
		if !errors.Is(err, ErrNoActiveSessionForRegeneration) {
			t.Fatalf("GenerateNextPlanCandidate() error = %v, want %v", err, ErrNoActiveSessionForRegeneration)
		}
	})

	t.Run("success", func(t *testing.T) {
		lastPlanID := uuid.New()
		questionID := uuid.New()

		svc := NewFlowService(
			&dumpStub{
				getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
					raw := "raw"
					return &models.Dump{ID: dumpID, RawText: &raw}, nil
				},
			},
			&analysisStub{
				getDumpAnalysisFn: func(context.Context, uuid.UUID) (*models.DumpAnalysis, error) {
					return &models.DumpAnalysis{
						DumpID:    dumpID,
						Tasks:     []models.Task{{Text: "task from analysis"}},
						Questions: []models.Question{{ID: questionID, Text: "question from analysis"}},
					}, nil
				},
			},
			&answersStub{
				getAnswersFn: func(context.Context, uuid.UUID) (*models.DumpAnswers, error) {
					return &models.DumpAnswers{
						DumpID:  dumpID,
						Answers: []models.Answer{{QuestionID: questionID, Text: "answer"}},
					}, nil
				},
			},
			&planStub{
				getLastGeneratedPlanFn: func(context.Context, uuid.UUID) (models.Plan, error) {
					return models.Plan{ID: lastPlanID, DumpID: dumpID, Title: "Last plan"}, nil
				},
				createPlanFn: func(_ context.Context, gotDumpID uuid.UUID, title string) (uuid.UUID, error) {
					if gotDumpID != dumpID || title != "Regenerated plan" {
						t.Fatalf("CreatePlan() got (%v, %q)", gotDumpID, title)
					}
					return planID, nil
				},
			},
			&planItemStub{
				createItemsFn: func(_ context.Context, items []models.PlanItem) ([]models.PlanItem, error) {
					if len(items) != 1 || items[0].Text != "regenerated item" {
						t.Fatalf("CreateItems() items = %+v", items)
					}
					return []models.PlanItem{{ID: uuid.New(), PlanID: planID, Ord: 1, Text: "regenerated item"}}, nil
				},
				getItemsByPlanIDsFn: func(_ context.Context, ids []uuid.UUID) ([]models.PlanItem, error) {
					if len(ids) != 1 || ids[0] != lastPlanID {
						t.Fatalf("GetItemsByPlanIDs() ids = %v", ids)
					}
					return []models.PlanItem{{ID: uuid.New(), PlanID: ids[0], Ord: 1, Text: "old item"}}, nil
				},
			},
			&analysisGenStub{},
			&planGenStub{
				regeneratePlanFn: func(_ context.Context, rawText string, analysis models.DumpAnalysis, answers models.DumpAnswers, plan models.Plan, planItems []models.PlanItem, feedback string) (models.Plan, []models.PlanItem, error) {
					if rawText != "raw" || feedback != "feedback" {
						t.Fatalf("RegeneratePlan() got rawText=%q feedback=%q", rawText, feedback)
					}
					if analysis.DumpID != dumpID || answers.DumpID != dumpID || plan.ID != lastPlanID || len(planItems) != 1 {
						t.Fatalf("RegeneratePlan() got analysis=%+v answers=%+v plan=%+v planItems=%+v", analysis, answers, plan, planItems)
					}
					return models.Plan{Title: "Regenerated plan"}, []models.PlanItem{{Text: "regenerated item"}}, nil
				},
			},
			zap.NewNop(),
		)

		plan, items, err := svc.GenerateNextPlanCandidate(context.Background(), userID, models.UserFeedback{DumpID: dumpID, Text: "feedback"})
		if err != nil {
			t.Fatalf("GenerateNextPlanCandidate() error = %v", err)
		}
		if plan.ID != planID || len(items) != 1 {
			t.Fatalf("GenerateNextPlanCandidate() got (%+v, %+v)", plan, items)
		}
	})
}

func TestFlowService_FinalizePlanSelection(t *testing.T) {
	dumpID := uuid.New()
	planID := uuid.New()
	wantErr := errors.New("boom")

	t.Run("success", func(t *testing.T) {
		svc := NewFlowService(
			&dumpStub{},
			&analysisStub{},
			nil,
			&planStub{
				savePlanFn: func(_ context.Context, gotDumpID, gotPlanID uuid.UUID) error {
					if gotDumpID != dumpID || gotPlanID != planID {
						t.Fatalf("SavePlan() got (%v, %v)", gotDumpID, gotPlanID)
					}
					return nil
				},
			},
			&planItemStub{},
			&analysisGenStub{},
			nil,
			zap.NewNop(),
		)

		if err := svc.FinalizePlanSelection(context.Background(), dumpID, planID); err != nil {
			t.Fatalf("FinalizePlanSelection() error = %v", err)
		}
	})

	t.Run("wraps error", func(t *testing.T) {
		svc := NewFlowService(
			&dumpStub{},
			&analysisStub{},
			nil,
			&planStub{
				savePlanFn: func(context.Context, uuid.UUID, uuid.UUID) error { return wantErr },
			},
			&planItemStub{},
			&analysisGenStub{},
			nil,
			zap.NewNop(),
		)

		err := svc.FinalizePlanSelection(context.Background(), dumpID, planID)
		if !errors.Is(err, wantErr) {
			t.Fatalf("FinalizePlanSelection() error = %v, want wrap %v", err, wantErr)
		}
	})
}
