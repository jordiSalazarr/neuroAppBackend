package createlanguagefluencysubtest

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateLanguageFluencySubtestCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()

	validCmd := CreateLanguageFluencySubtestCommand{
		EvaluationID: "eval-123",
		Category:     "semantic",
		Words:        []string{"perro", "pez", "pato", "puma"},
		Duration:     60,
		Language:     "es",
		Proficiency:  "native",
		TotalTime:    60,
	}

	tests := []struct {
		name       string
		command    CreateLanguageFluencySubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			command:    validCmd,
			shouldPass: true,
		},
		{
			name: "Invalid - missing evaluation id",
			command: func() CreateLanguageFluencySubtestCommand {
				c := validCmd
				c.EvaluationID = ""
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - empty words",
			command: func() CreateLanguageFluencySubtestCommand {
				c := validCmd
				c.Words = []string{}
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - empty category",
			command: func() CreateLanguageFluencySubtestCommand {
				c := validCmd
				c.Category = ""
				return c
			}(),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CreateLanguageFluencySubtestCommandHandler(
				context.TODO(),
				tt.command,
				app.Repositories.EvaluationsRepository,
				app.Services.LLMService,
				app.Repositories.LanguageFluencyRepository,
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if res.PK == "" {
					t.Errorf("expected persisted entity with PK, got empty")
				}
				if res.Score.Score == 0 {
					t.Errorf("expected non-zero score to be calculated, got 0")
				}
				if res.Category != tt.command.Category {
					t.Errorf("expected category %q, got %q", tt.command.Category, res.Category)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil (command: %+v)", tt.command)
				}
			}
		})
	}
}
