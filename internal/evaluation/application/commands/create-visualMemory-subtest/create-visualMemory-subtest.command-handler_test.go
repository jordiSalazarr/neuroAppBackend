package createvisualmemorysubtest

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateVisualMemoryCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()

	valid := CreateVisualMemorySubtestCommand{
		EvaluationID: "eval-123", // usa un ID válido si tu mock lo valida
		Score:        1,
		Note:         "Recuerdo visual dentro de la media.",
	}

	tests := []struct {
		name       string
		cmd        CreateVisualMemorySubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			cmd:        valid,
			shouldPass: true,
		},
		{
			name: "Invalid - missing evaluation id",
			cmd: func() CreateVisualMemorySubtestCommand {
				c := valid
				c.EvaluationID = ""
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative score",
			cmd: func() CreateVisualMemorySubtestCommand {
				c := valid
				c.Score = -1
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - score above max",
			cmd: func() CreateVisualMemorySubtestCommand {
				c := valid
				c.Score = 3
				return c
			}(),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CreateVisualMemoryCommandHandler(
				context.TODO(),
				tt.cmd,
				app.Repositories.VisualMemorySubtestRepository,
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil subtest")
				}
				// Chequeos básicos de consistencia
				if res.EvaluationID != tt.cmd.EvaluationID {
					t.Errorf("expected EvaluationID=%q, got %q", tt.cmd.EvaluationID, res.EvaluationID)
				}
				if res.Score.Val != tt.cmd.Score {
					t.Errorf("expected Score=%d, got %d", tt.cmd.Score, res.Score.Val)
				}
				if res.Note.Val != tt.cmd.Note {
					t.Errorf("expected Note=%q, got %q", tt.cmd.Note, res.Note.Val)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil (cmd=%+v)", tt.cmd)
				}
				if res != nil {
					t.Errorf("expected nil result on error, got %+v", res)
				}
			}
		})
	}
}
