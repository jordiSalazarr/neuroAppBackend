package createvisualspatialsubtest

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateViusualSpatialCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()

	valid := CreateVisualSpatialSubtestCommand{
		EvaluationID: "eval-123", // usa un ID válido en tu mock si aplica
		Note:         "Paciente con dificultades leves en copia de figuras.",
		Score:        4,
	}

	tests := []struct {
		name       string
		cmd        CreateVisualSpatialSubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			cmd:        valid,
			shouldPass: true,
		},
		{
			name: "Invalid - missing evaluation id",
			cmd: func() CreateVisualSpatialSubtestCommand {
				c := valid
				c.EvaluationID = ""
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - negative score",
			cmd: func() CreateVisualSpatialSubtestCommand {
				c := valid
				c.Score = -1
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - score above max range",
			cmd: func() CreateVisualSpatialSubtestCommand {
				c := valid
				c.Score = 99
				return c
			}(),
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CreateViusualSpatialCommandHandler(
				context.TODO(),
				tt.cmd,
				app.Repositories.VisualSpatialRepository, // ajusta el nombre si tu MockApp expone otro
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil subtest")
				}
				// Chequeos básicos
				if res.EvalautionId != tt.cmd.EvaluationID {
					t.Errorf("expected EvaluationID=%q, got %q", tt.cmd.EvaluationID, res.EvalautionId)
				}
				if res.Note.Val != tt.cmd.Note {
					t.Errorf("expected Note=%q, got %q", tt.cmd.Note, res.Note)
				}
				if res.Score.Val != tt.cmd.Score {
					t.Errorf("expected Score=%d, got %d", tt.cmd.Score, res.Score)
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
