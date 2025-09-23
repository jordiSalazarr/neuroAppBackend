package createverbalmemorysubtest

import (
	"context"
	"testing"
	"time"

	"neuro.app.jordi/internal/pkg"
)

func TestCreateVerbalMemorySubtestCommandhandler(t *testing.T) {
	app := pkg.NewMockApp()

	valid := CreateVerbalMemorySubtestCommand{
		EvaluationID:  "eval-123", // usa un ID que tu mock resuelva si lo necesitases
		StartAt:       time.Now().Add(-1 * time.Minute),
		GivenWords:    []string{"casa", "perro", "mar", "luz", "flor"},
		RecalledWords: []string{"casa", "mar", "flor"},
	}

	tests := []struct {
		name       string
		cmd        CreateVerbalMemorySubtestCommand
		shouldPass bool
	}{
		{
			name:       "Valid command",
			cmd:        valid,
			shouldPass: true,
		},
		{
			name: "Invalid - missing evaluation id",
			cmd: func() CreateVerbalMemorySubtestCommand {
				c := valid
				c.EvaluationID = ""
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - empty given words",
			cmd: func() CreateVerbalMemorySubtestCommand {
				c := valid
				c.GivenWords = []string{}
				return c
			}(),
			shouldPass: false,
		},
		{
			name: "Invalid - zero start time",
			cmd: func() CreateVerbalMemorySubtestCommand {
				c := valid
				c.StartAt = time.Time{} // cero
				return c
			}(),
			shouldPass: false,
		},
		// Descomenta este caso si tu dominio invalida palabras recordadas no presentes en las dadas:
		// {
		// 	name: "Invalid - recalled word not in given words",
		// 	cmd: func() CreateVerbalMemorySubtestCommand {
		// 		c := valid
		// 		c.RecalledWords = []string{"casa", "montaña"} // "montaña" no está en GivenWords
		// 		return c
		// 	}(),
		// 	shouldPass: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CreateVerbalMemorySubtestCommandhandler(
				context.TODO(),
				tt.cmd,
				app.Repositories.EvaluationsRepository,
				app.Services.LLMService,
				app.Repositories.VerbalMemorySubtestRepository,
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				// Sanity checks básicos
				if res.EvaluationID != tt.cmd.EvaluationID {
					t.Errorf("expected EvaluationID=%q, got %q", tt.cmd.EvaluationID, res.EvaluationID)
				}
				if len(res.GivenWords) != len(tt.cmd.GivenWords) {
					t.Errorf("expected %d given words, got %d", len(tt.cmd.GivenWords), len(res.GivenWords))
				}
				// Si en tu dominio el Score puede ser 0 legítimamente, elimina esta aserción
				if res.Score.Score == 0 {
					t.Errorf("expected non-zero score to be calculated, got 0")
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil (cmd=%+v)", tt.cmd)
				}
			}
		})
	}
}
