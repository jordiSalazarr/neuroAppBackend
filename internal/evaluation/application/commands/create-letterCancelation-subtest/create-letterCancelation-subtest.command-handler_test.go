package createlettercancelationsubtest

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateLetterCancellationSubtestCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()

	valid := CreateLetterCancellationSubtestCommand{
		TotalTargets: 60,
		Correct:      52,
		Errors:       3,
		TimeInSecs:   75,
		EvaluationID: "eval-123", // usa un ID válido en tu mock si hay validación cruzada
	}

	tests := []struct {
		name       string
		cmd        CreateLetterCancellationSubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			cmd:        valid,
			shouldPass: true,
		},
		{
			name: "Invalid - evaluation id empty",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.EvaluationID = ""
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative total targets",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.TotalTargets = -1
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative correct",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.Correct = -2
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative errors",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.Errors = -3
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - correct greater than total targets",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.TotalTargets = 10
				c.Correct = 11
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative time",
			cmd: func() CreateLetterCancellationSubtestCommand {
				c := valid
				c.TimeInSecs = -5
				return c
			}(),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := CreateLetterCancellationSubtestCommandHandler(
				context.TODO(),
				tt.cmd,
				app.Repositories.LetterCancellationRepository,
				app.Repositories.EvaluationsRepository, // no se usa en el handler, pero seguimos la firma
				app.Services.LLMService,                // idem
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if sub == nil {
					t.Fatalf("expected non-nil subtest")
				}
				// Chequeos básicos de consistencia
				if sub.EvaluationID != tt.cmd.EvaluationID {
					t.Errorf("expected EvaluationID=%q, got %q", tt.cmd.EvaluationID, sub.EvaluationID)
				}
				if sub.TimeInSecs != tt.cmd.TimeInSecs {
					t.Errorf("expected TimeInSecs=%d, got %d", tt.cmd.TimeInSecs, sub.TimeInSecs)
				}
				if sub.TotalTargets != tt.cmd.TotalTargets || sub.Correct != tt.cmd.Correct || sub.Errors != tt.cmd.Errors {
					t.Errorf("entity fields do not match command: got %+v", sub)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil (cmd=%+v)", tt.cmd)
				}
				if sub != nil {
					t.Errorf("expected nil subtest on error, got %+v", sub)
				}
			}
		})
	}
}
