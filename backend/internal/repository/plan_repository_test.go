package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

func TestPlanRepository_SubmitAnswersAndCreatePlan(t *testing.T) {
	dumpID := uuid.New()
	planID := uuid.New()
	itemID := uuid.New()
	questionID := uuid.New()
	createdAt := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	answers := models.DumpAnswers{
		DumpID:  dumpID,
		Answers: []models.Answer{{QuestionID: questionID, Text: "answer"}},
	}
	plan := models.Plan{DumpID: dumpID, Title: "Plan"}
	items := []models.PlanItem{{PlanID: planID, Ord: 1, Text: "item"}}

	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{kind: "exec", match: "INSERT INTO dump_answers", result: driver.RowsAffected(1)},
		{
			kind:  "query",
			match: "INSERT INTO plans",
			rows: &sqlRows{
				columns: []string{"id"},
				values:  [][]driver.Value{{planID[:]}},
			},
		},
		{
			kind:  "query",
			match: "INSERT INTO plan_items",
			rows: &sqlRows{
				columns: []string{"id", "created_at"},
				values:  [][]driver.Value{{itemID[:], createdAt}},
			},
		},
		{kind: "commit"},
	})

	repo := NewPlanRepo(db, zap.NewNop())
	gotPlan, gotItems, err := repo.SubmitAnswersAndCreatePlan(context.Background(), answers, plan, items)
	if err != nil {
		t.Fatalf("SubmitAnswersAndCreatePlan() error = %v", err)
	}
	if gotPlan.ID != planID {
		t.Fatalf("SubmitAnswersAndCreatePlan() plan ID = %v, want %v", gotPlan.ID, planID)
	}
	if len(gotItems) != 1 || gotItems[0].ID != itemID || !gotItems[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("SubmitAnswersAndCreatePlan() items = %+v", gotItems)
	}
}

func TestPlanRepository_SubmitAnswersAndCreatePlan_AnswersAlreadyExist(t *testing.T) {
	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{
			kind: "exec",
			err: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
		},
		{kind: "rollback"},
	})

	repo := NewPlanRepo(db, zap.NewNop())
	_, _, err := repo.SubmitAnswersAndCreatePlan(context.Background(), models.DumpAnswers{}, models.Plan{}, nil)
	if !errors.Is(err, ErrAnswersUniqueViolation) {
		t.Fatalf("SubmitAnswersAndCreatePlan() error = %v, want %v", err, ErrAnswersUniqueViolation)
	}
}

func TestPlanRepository_FinalizeSelectedPlan(t *testing.T) {
	dumpID := uuid.New()
	planID := uuid.New()

	t.Run("success", func(t *testing.T) {
		db := newTestDB(t, []sqlExpectation{
			{kind: "begin"},
			{
				kind:  "query",
				match: "SELECT id",
				rows: &sqlRows{
					columns: []string{"id"},
					values:  [][]driver.Value{{planID[:]}},
				},
			},
			{kind: "exec", match: "saved_at = now()", result: driver.RowsAffected(1)},
			{kind: "exec", match: "DELETE FROM plans", result: driver.RowsAffected(2)},
			{kind: "exec", match: "UPDATE dumps", result: driver.RowsAffected(1)},
			{kind: "commit"},
		})

		repo := NewPlanRepo(db, zap.NewNop())
		if err := repo.FinalizeSelectedPlan(context.Background(), dumpID, planID); err != nil {
			t.Fatalf("FinalizeSelectedPlan() error = %v", err)
		}
	})

	t.Run("plan not found", func(t *testing.T) {
		db := newTestDB(t, []sqlExpectation{
			{kind: "begin"},
			{
				kind: "query",
				rows: &sqlRows{
					columns: []string{"id"},
					values:  nil,
				},
			},
			{kind: "rollback"},
		})

		repo := NewPlanRepo(db, zap.NewNop())
		err := repo.FinalizeSelectedPlan(context.Background(), dumpID, planID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("FinalizeSelectedPlan() error = %v, want %v", err, ErrNotFound)
		}
	})
}
