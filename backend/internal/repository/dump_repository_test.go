package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

func TestDumpRepository_CreateDump(t *testing.T) {
	userID := uuid.New()
	activeDumpID := uuid.New()
	newDumpID := uuid.New()
	rawText := "raw"

	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{
			kind:  "query",
			match: "FROM dumps",
			checkArgs: func(args []driver.NamedValue) error {
				if got := namedValueAt(args, 0); got != userID {
					return errors.New("unexpected user id")
				}
				return nil
			},
			rows: &sqlRows{
				columns: []string{"id"},
				values:  [][]driver.Value{{activeDumpID[:]}},
			},
		},
		{
			kind:   "exec",
			match:  "UPDATE dumps",
			result: driver.RowsAffected(1),
		},
		{
			kind:  "query",
			match: "INSERT INTO dumps",
			checkArgs: func(args []driver.NamedValue) error {
				if got := namedValueAt(args, 0); got != userID {
					return errors.New("unexpected dump user id")
				}
				if got := namedValueAt(args, 3); got != rawText {
					return errors.New("unexpected raw text")
				}
				return nil
			},
			rows: &sqlRows{
				columns: []string{"id"},
				values:  [][]driver.Value{{newDumpID[:]}},
			},
		},
		{kind: "commit"},
	})

	repo := NewDumpRepo(db, zap.NewNop())
	got, err := repo.CreateDump(context.Background(), userID, models.Dump{
		UserID:  &userID,
		Status:  models.DumpStatusNew,
		RawText: &rawText,
	})
	if err != nil {
		t.Fatalf("CreateDump() error = %v", err)
	}
	if got != newDumpID {
		t.Fatalf("CreateDump() = %v, want %v", got, newDumpID)
	}
}

func TestDumpRepository_CreateDump_StatusNotChanged(t *testing.T) {
	userID := uuid.New()
	activeDumpID := uuid.New()

	db := newTestDB(t, []sqlExpectation{
		{kind: "begin"},
		{
			kind: "query",
			rows: &sqlRows{
				columns: []string{"id"},
				values:  [][]driver.Value{{activeDumpID[:]}},
			},
		},
		{
			kind:   "exec",
			result: driver.RowsAffected(0),
		},
		{kind: "rollback"},
	})

	repo := NewDumpRepo(db, zap.NewNop())
	_, err := repo.CreateDump(context.Background(), userID, models.Dump{})
	if !errors.Is(err, ErrStatusNotChanged) {
		t.Fatalf("CreateDump() error = %v, want %v", err, ErrStatusNotChanged)
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
