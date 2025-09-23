package createexecutivefunctionssubtest

import (
	"context"
	"testing"
	"time"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateExecutiveFunctionsSubtestCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()
	validCommand := CreateExecutiveFunctionsSubtestCommand{
		NumberOfItems: 10,
		TotalClicks:   20,
		StartAt:       time.Now(),
		TotalErrors:   2,
		TotalCorrect:  8,
		TotalTime:     30 * time.Second,
		Type:          "a",
		EvaluationId:  "eval-123",
		CreatedAt:     time.Now(),
	}

	tests := []struct {
		name       string
		command    CreateExecutiveFunctionsSubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			command:    validCommand,
			shouldPass: true,
		},
		{
			name: "Invalid command - negative items",
			command: func() CreateExecutiveFunctionsSubtestCommand {
				c := validCommand
				c.NumberOfItems = -5
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid command - invalid type",
			command: func() CreateExecutiveFunctionsSubtestCommand {
				c := validCommand
				c.Type = "invalid"
				return c
			}(),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CreateExecutiveFunctionsSubtestCommandHandler(
				context.TODO(),
				tt.command,
				app.Repositories.EvaluationsRepository,
				app.Services.LLMService,
				app.Repositories.ExecutiveFunctionsSubtestRepository,
			)

			if tt.shouldPass {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}
				if result.Score.Score == 0 {
					t.Errorf("expected score to be calculated, got 0")
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
		})
	}
}
