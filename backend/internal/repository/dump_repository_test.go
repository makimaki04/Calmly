package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

func TestDumpRepository_CreateDump_InsertErrorRollsBack(t *testing.T) {
	userID := uuid.New()
	wantErr := errors.New("insert failed")

	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{
			kind: "query",
			err:  sql.ErrNoRows,
		},
		{
			kind: "query",
			err:  wantErr,
		},
		{kind: "rollback"},
	})

	repo := NewDumpRepo(db, zap.NewNop())
	_, err := repo.CreateDump(context.Background(), userID, models.Dump{})
	if !errors.Is(err, ErrDB) {
		t.Fatalf("CreateDump() error = %v, want wrap %v", err, ErrDB)
	}
}

func TestDumpRepository_CompleteAnalysisStep(t *testing.T) {
	dumpID := uuid.New()
	mood := models.MoodNeutral
	quote := "quote"
	analysis := models.DumpAnalysis{
		DumpID: dumpID,
		Tasks: []models.Task{
			{Text: "task", Priority: "low", Category: "work"},
		},
		Questions: []models.Question{
			{ID: uuid.New(), Text: "question"},
		},
		Mood:  &mood,
		Quote: &quote,
	}

	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{
			kind:  "exec",
			match: "INSERT INTO dump_analysis",
			checkArgs: func(args []driver.NamedValue) error {
				tasksJSON, err := json.Marshal(analysis.Tasks)
				if err != nil {
					return err
				}
				questionsJSON, err := json.Marshal(analysis.Questions)
				if err != nil {
					return err
				}
				if string(namedValueAt(args, 1).([]byte)) != string(tasksJSON) {
					return errors.New("unexpected tasks json")
				}
				if string(namedValueAt(args, 2).([]byte)) != string(questionsJSON) {
					return errors.New("unexpected questions json")
				}
				return nil
			},
			result: driver.RowsAffected(1),
		},
		{
			kind:   "exec",
			match:  "status = $2",
			result: driver.RowsAffected(1),
		},
		{kind: "commit"},
	})

	repo := NewDumpRepo(db, zap.NewNop())
	if err := repo.CompleteAnalysisStep(context.Background(), analysis); err != nil {
		t.Fatalf("CompleteAnalysisStep() error = %v", err)
	}
}

func TestDumpRepository_CompleteAnalysisStep_StatusNotChanged(t *testing.T) {
	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{kind: "exec", result: driver.RowsAffected(1)},
		{kind: "exec", result: driver.RowsAffected(0)},
		{kind: "rollback"},
	})

	repo := NewDumpRepo(db, zap.NewNop())
	err := repo.CompleteAnalysisStep(context.Background(), models.DumpAnalysis{DumpID: uuid.New()})
	if !errors.Is(err, ErrStatusNotChanged) {
		t.Fatalf("CompleteAnalysisStep() error = %v, want %v", err, ErrStatusNotChanged)
	}
}
