package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"github.com/makimaki04/Calmly/internal/service"
	"go.uber.org/zap"
)

type dumpStub struct {
	createDumpFn           func(context.Context, uuid.UUID, string) (uuid.UUID, error)
	getUserDumpFn          func(context.Context, uuid.UUID) (*models.Dump, error)
	setDumpStatusFn        func(context.Context, uuid.UUID, models.DumpStatus) error
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
func (s *dumpStub) AbandonDump(context.Context, uuid.UUID) error { return nil }
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

type planStub struct {
	submitAnswersAndCreatePlanFn func(context.Context, models.DumpAnswers, models.Plan, []models.PlanItem) (models.Plan, []models.PlanItem, error)
	createPlanFn                 func(context.Context, uuid.UUID, string) (uuid.UUID, error)
	savePlanFn                   func(context.Context, uuid.UUID, uuid.UUID) error
	getDumpPlansFn               func(context.Context, uuid.UUID) ([]models.Plan, error)
}

func (s *planStub) CreatePlan(ctx context.Context, dumpID uuid.UUID, title string) (uuid.UUID, error) {
	return s.createPlanFn(ctx, dumpID, title)
}
func (s *planStub) SubmitAnswersAndCreatePlan(ctx context.Context, answers models.DumpAnswers, plan models.Plan, items []models.PlanItem) (models.Plan, []models.PlanItem, error) {
	return s.submitAnswersAndCreatePlanFn(ctx, answers, plan, items)
}
func (s *planStub) SavePlan(ctx context.Context, dumpID uuid.UUID, planID uuid.UUID) error {
	return s.savePlanFn(ctx, dumpID, planID)
}
func (s *planStub) GetDumpPlans(ctx context.Context, dumpID uuid.UUID) ([]models.Plan, error) {
	return s.getDumpPlansFn(ctx, dumpID)
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
func (s *planItemStub) GetItemsByPlanIDs(ctx context.Context, ids []uuid.UUID) ([]models.PlanItem, error) {
	return s.getItemsByPlanIDsFn(ctx, ids)
}

type analysisGenStub struct {
	generateAnalysisFn func(context.Context, string) (models.DumpAnalysis, error)
}

func (s *analysisGenStub) GenerateAnalysis(ctx context.Context, rawText string) (models.DumpAnalysis, error) {
	return s.generateAnalysisFn(ctx, rawText)
}

type planGenStub struct {
	generatePlanFn   func(context.Context, string, []models.Task, []models.Question, []models.Answer) (models.Plan, []models.PlanItem, error)
	regeneratePlanFn func(context.Context, string, string) (models.Plan, []models.PlanItem, error)
}

func (s *planGenStub) GeneratePlan(ctx context.Context, rawText string, tasks []models.Task, questions []models.Question, answers []models.Answer) (models.Plan, []models.PlanItem, error) {
	return s.generatePlanFn(ctx, rawText, tasks, questions, answers)
}

func (s *planGenStub) RegeneratePlan(ctx context.Context, rawText string, feedback string) (models.Plan, []models.PlanItem, error) {
	if s.regeneratePlanFn == nil {
		return models.Plan{}, nil, nil
	}
	return s.regeneratePlanFn(ctx, rawText, feedback)
}

func newHandlerForTest(activeDumpID uuid.UUID) *Handler {
	flow := service.NewFlowService(
		&dumpStub{
			createDumpFn: func(context.Context, uuid.UUID, string) (uuid.UUID, error) { return uuid.New(), nil },
			getUserDumpFn: func(context.Context, uuid.UUID) (*models.Dump, error) {
				raw := "raw"
				return &models.Dump{ID: activeDumpID, RawText: &raw}, nil
			},
			setDumpStatusFn:        func(context.Context, uuid.UUID, models.DumpStatus) error { return nil },
			completeAnalysisStepFn: func(context.Context, models.DumpAnalysis) error { return nil },
		},
		&analysisStub{
			getDumpAnalysisFn: func(context.Context, uuid.UUID) (*models.DumpAnalysis, error) {
				return &models.DumpAnalysis{
					DumpID:    activeDumpID,
					Tasks:     []models.Task{{Text: "task from analysis"}},
					Questions: []models.Question{{Text: "question from analysis"}},
				}, nil
			},
		},
		nil,
		&planStub{
			submitAnswersAndCreatePlanFn: func(_ context.Context, answers models.DumpAnswers, plan models.Plan, items []models.PlanItem) (models.Plan, []models.PlanItem, error) {
				if plan.Title != "Generated plan" || len(items) != 1 || items[0].Ord != 1 {
					return models.Plan{}, nil, errors.New("unexpected generated plan payload")
				}
				plan.ID = uuid.New()
				return plan, []models.PlanItem{{ID: uuid.New(), PlanID: plan.ID, Ord: 1, Text: "item"}}, nil
			},
			createPlanFn: func(context.Context, uuid.UUID, string) (uuid.UUID, error) { return uuid.New(), nil },
			savePlanFn:   func(context.Context, uuid.UUID, uuid.UUID) error { return nil },
			getDumpPlansFn: func(context.Context, uuid.UUID) ([]models.Plan, error) {
				return []models.Plan{{ID: uuid.New()}}, nil
			},
		},
		&planItemStub{
			createItemsFn: func(context.Context, []models.PlanItem) ([]models.PlanItem, error) {
				return []models.PlanItem{{ID: uuid.New(), PlanID: uuid.New(), Ord: 1, Text: "item"}}, nil
			},
			getItemsByPlanIDsFn: func(context.Context, []uuid.UUID) ([]models.PlanItem, error) {
				return []models.PlanItem{{ID: uuid.New(), PlanID: uuid.New(), Ord: 1, Text: "item"}}, nil
			},
		},
		&analysisGenStub{
			generateAnalysisFn: func(context.Context, string) (models.DumpAnalysis, error) {
				return models.DumpAnalysis{
					Tasks:     []models.Task{{Text: "task", Priority: "low", Category: "work"}},
					Questions: []models.Question{{Text: "question"}},
					Quote:     ptr("quote"),
					Mood:      moodPtr(models.MoodNeutral),
				}, nil
			},
		},
		&planGenStub{
			generatePlanFn: func(_ context.Context, rawText string, tasks []models.Task, questions []models.Question, answers []models.Answer) (models.Plan, []models.PlanItem, error) {
				if rawText != "raw" {
					return models.Plan{}, nil, errors.New("unexpected raw text")
				}
				if len(tasks) != 1 || len(questions) != 1 || len(answers) != 1 {
					return models.Plan{}, nil, errors.New("unexpected plan generation input")
				}
				return models.Plan{Title: "Generated plan"}, []models.PlanItem{{Text: "first"}}, nil
			},
		},
		zap.NewNop(),
	)
	return NewHandler(nil, flow, zap.NewNop())
}

func TestParseJSONBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		body     string
		wantName string
		wantErr  bool
	}{
		{name: "success", body: `{"name":"ok"}`, wantName: "ok"},
		{name: "empty", body: ``, wantErr: true},
		{name: "unknown field", body: `{"name":"ok","extra":1}`, wantErr: true},
		{name: "multiple objects", body: `{"name":"ok"}{"name":"second"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			got, err := parseJSONBody[payload](rec, req, MaxBodyJSON, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseJSONBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got.Name != tt.wantName {
				t.Fatalf("parseJSONBody() = %+v, want name %q", got, tt.wantName)
			}
		})
	}
}

func TestHandler_StartSession(t *testing.T) {
	h := newHandlerForTest(uuid.New())

	t.Run("missing raw text", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/start", bytes.NewBufferString(`{"dump":{}}`))
		rec := httptest.NewRecorder()

		h.StartSession(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("StartSession() status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/start", bytes.NewBufferString(`{"dump":{"raw_text":"hello"}}`))
		rec := httptest.NewRecorder()

		h.StartSession(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("StartSession() status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_SubmitAnswers(t *testing.T) {
	dumpID := uuid.New()
	questionID := uuid.New()
	h := newHandlerForTest(dumpID)

	t.Run("bad dump id", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`)), "dump_id", "bad")
		rec := httptest.NewRecorder()

		h.SubmitAnswers(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("SubmitAnswers() status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		body := `{"answers":[{"question_id":"` + questionID.String() + `","text":"answer"}]}`
		req := withURLParam(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)), "dump_id", dumpID.String())
		rec := httptest.NewRecorder()

		h.SubmitAnswers(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("SubmitAnswers() status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestHandler_RegeneratePlan(t *testing.T) {
	dumpID := uuid.New()
	h := newHandlerForTest(dumpID)

	req := withURLParam(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"feedback":"less tasks"}`)), "dump_id", dumpID.String())
	rec := httptest.NewRecorder()

	h.RegeneratePlan(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("RegeneratePlan() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_FinalizePlanSelection(t *testing.T) {
	dumpID := uuid.New()
	planID := uuid.New()
	h := newHandlerForTest(dumpID)

	req := withURLParam(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"plan_id":"`+planID.String()+`"}`)), "dump_id", dumpID.String())
	rec := httptest.NewRecorder()

	h.FinalizePlanSelection(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("FinalizePlanSelection() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMapFlowErrorToStatus(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{err: service.ErrActiveDumpNotFound, want: http.StatusNotFound},
		{err: service.ErrDumpNotBelongUser, want: http.StatusForbidden},
		{err: service.ErrAnswersAlreadySubmitted, want: http.StatusConflict},
		{err: service.ErrAnalysisNotFound, want: http.StatusConflict},
		{err: service.ErrNoActiveSessionForRegeneration, want: http.StatusConflict},
		{err: repository.ErrNotFound, want: http.StatusNotFound},
		{err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := mapFlowErrorToStatus(tt.err); got != tt.want {
			t.Fatalf("mapFlowErrorToStatus(%v) = %d, want %d", tt.err, got, tt.want)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{err: service.ErrActiveDumpNotFound, want: "active session not found"},
		{err: service.ErrDumpNotBelongUser, want: "dump does not belong to current user"},
		{err: service.ErrAnalysisNotFound, want: "analysis is missing"},
		{err: service.ErrAnswersAlreadySubmitted, want: "answers already submitted"},
		{err: service.ErrNoActiveSessionForRegeneration, want: "no active session available for plan regeneration"},
		{err: repository.ErrNotFound, want: "resource not found"},
		{err: errors.New("boom"), want: "internal server error"},
	}

	for _, tt := range tests {
		if got := errorMessage(tt.err); got != tt.want {
			t.Fatalf("errorMessage(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestHandler_ErrorResponsesAreJSON(t *testing.T) {
	h := newHandlerForTest(uuid.New())
	req := httptest.NewRequest(http.MethodPost, "/start", bytes.NewBufferString(``))
	rec := httptest.NewRecorder()

	h.StartSession(rec, req)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response json error = %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error message in body, got %v", body)
	}
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func ptr(s string) *string { return &s }

func moodPtr(m models.Mood) *models.Mood { return &m }
