package llm

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/makimaki04/Calmly/internal/models"
)

func TestParseLLMResponse(t *testing.T) {
	content := []anthropic.ContentBlockUnion{
		{Text: "  first line  "},
		{Text: ""},
		{Text: " second line "},
	}

	got, err := ParseLLMResponse(content)
	if err != nil {
		t.Fatalf("ParseLLMResponse() error = %v", err)
	}

	want := "first line\nsecond line"
	if got != want {
		t.Fatalf("ParseLLMResponse() = %q, want %q", got, want)
	}
}

func TestParseLLMResponse_Empty(t *testing.T) {
	_, err := ParseLLMResponse([]anthropic.ContentBlockUnion{{Text: "   "}})
	if !errors.Is(err, ErrEmptyLLMResponse) {
		t.Fatalf("ParseLLMResponse() error = %v, want %v", err, ErrEmptyLLMResponse)
	}
}

func TestExtractJSONCandidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "plain json",
			input: `{"tasks":[],"questions":[],"mood":"neutral","quote":"ok"}`,
			want:  `{"tasks":[],"questions":[],"mood":"neutral","quote":"ok"}`,
		},
		{
			name: "json inside markdown",
			input: "```json\nbefore\n{\"tasks\":[{\"text\":\"a { brace }\",\"priority\":\"low\",\"category\":\"work\"}]," +
				"\"questions\":[{\"text\":\"q\"}],\"mood\":\"neutral\",\"quote\":\"ok\"}\nafter\n```",
			want: "{\"tasks\":[{\"text\":\"a { brace }\",\"priority\":\"low\",\"category\":\"work\"}]," +
				"\"questions\":[{\"text\":\"q\"}],\"mood\":\"neutral\",\"quote\":\"ok\"}",
		},
		{
			name:    "missing valid json",
			input:   "no json here",
			wantErr: ErrInvalidLLMResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractJSONCandidate(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ExtractJSONCandidate() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if string(got) != tt.want {
				t.Fatalf("ExtractJSONCandidate() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestValidateAndNormalizeAnalysisResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   AnalysisResponse
		wantErr error
		check   func(t *testing.T, got AnalysisResponse)
	}{
		{
			name: "success trims fields",
			input: AnalysisResponse{
				Tasks: []AnalysisTaskResponse{
					{Text: " task ", TaskPriority: " medium ", Category: " home "},
				},
				Questions: []AnalysisQuestionResponse{{Text: " question "}},
				Mood:      " neutral ",
				Quote:     " quote ",
			},
			check: func(t *testing.T, got AnalysisResponse) {
				t.Helper()
				if got.Tasks[0].Text != "task" || got.Tasks[0].TaskPriority != "medium" || got.Tasks[0].Category != "home" {
					t.Fatalf("task was not normalized: %+v", got.Tasks[0])
				}
				if got.Questions[0].Text != "question" {
					t.Fatalf("question was not normalized: %+v", got.Questions[0])
				}
				if got.Mood != "neutral" || got.Quote != "quote" {
					t.Fatalf("response was not normalized: %+v", got)
				}
			},
		},
		{
			name: "empty task field",
			input: AnalysisResponse{
				Tasks:     []AnalysisTaskResponse{{Text: " ", TaskPriority: "low", Category: "work"}},
				Questions: []AnalysisQuestionResponse{{Text: "q"}},
				Mood:      "neutral",
				Quote:     "ok",
			},
			wantErr: ErrEmptyTaskFields,
		},
		{
			name: "invalid priority",
			input: AnalysisResponse{
				Tasks:     []AnalysisTaskResponse{{Text: "task", TaskPriority: "urgent", Category: "work"}},
				Questions: []AnalysisQuestionResponse{{Text: "q"}},
				Mood:      "neutral",
				Quote:     "ok",
			},
			wantErr: ErrInvalidTaskPriority,
		},
		{
			name: "empty question",
			input: AnalysisResponse{
				Tasks:     []AnalysisTaskResponse{{Text: "task", TaskPriority: "low", Category: "work"}},
				Questions: []AnalysisQuestionResponse{{Text: " "}},
				Mood:      "neutral",
				Quote:     "ok",
			},
			wantErr: ErrEmptyQuestionText,
		},
		{
			name: "empty quote",
			input: AnalysisResponse{
				Tasks:     []AnalysisTaskResponse{{Text: "task", TaskPriority: "low", Category: "work"}},
				Questions: []AnalysisQuestionResponse{{Text: "q"}},
				Mood:      "neutral",
				Quote:     " ",
			},
			wantErr: ErrEmptyQuote,
		},
		{
			name: "invalid mood",
			input: AnalysisResponse{
				Tasks:     []AnalysisTaskResponse{{Text: "task", TaskPriority: "low", Category: "work"}},
				Questions: []AnalysisQuestionResponse{{Text: "q"}},
				Mood:      "happy",
				Quote:     "ok",
			},
			wantErr: ErrInvalidMoodValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input
			err := valAndNormAnalysisResp(&got)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateAndNormalizeAnalysisResponse() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestMapAnalysisResponseToDomain(t *testing.T) {
	input := AnalysisResponse{
		Tasks: []AnalysisTaskResponse{
			{Text: "task", TaskPriority: "high", Category: "work"},
		},
		Questions: []AnalysisQuestionResponse{
			{Text: "question"},
		},
		Mood:  string(models.MoodOverwhelmed),
		Quote: "quote",
	}

	got := mapAnalysisResponseToDomain(input)

	if len(got.Tasks) != 1 || got.Tasks[0].Priority != "high" {
		t.Fatalf("unexpected tasks: %+v", got.Tasks)
	}
	if len(got.Questions) != 1 || got.Questions[0].Text != "question" {
		t.Fatalf("unexpected questions: %+v", got.Questions)
	}
	if got.Mood == nil || *got.Mood != models.MoodOverwhelmed {
		t.Fatalf("unexpected mood: %+v", got.Mood)
	}
	if got.Quote == nil || *got.Quote != "quote" {
		t.Fatalf("unexpected quote: %+v", got.Quote)
	}
}

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Value string `json:"value"`
	}

	got, err := decodeJSON[payload]([]byte(`{"value":"ok"}`))
	if err != nil {
		t.Fatalf("decodeJSON() error = %v", err)
	}
	if got.Value != "ok" {
		t.Fatalf("decodeJSON() = %+v", got)
	}

	_, err = decodeJSON[payload]([]byte(`{"value":`))
	if err == nil {
		t.Fatal("decodeJSON() error = nil, want error")
	}

	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("decodeJSON() error = %T, want *json.SyntaxError", err)
	}
}
