package finishevaluation

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/evaluation/domain"
	reports "neuro.app.jordi/internal/evaluation/domain/services"
	"neuro.app.jordi/internal/pkg"
)

func TestFinisEvaluationCommanndHandler(t *testing.T) {
	app := pkg.NewMockApp()

	valid := FinisEvaluationCommannd{
		EvaluationID: "eval-123", // usa un ID que tu mock resuelva con GetByID
	}

	tests := []struct {
		name         string
		cmd          FinisEvaluationCommannd
		shouldPass   bool
		expectStatus domain.EvaluationCurrentStatus
		expectHasLLM bool
	}{
		{
			name:         "Valid - completes evaluation and sets assistant analysis",
			cmd:          valid,
			shouldPass:   true,
			expectStatus: domain.EvaluationCurrentStatusCompleted,
			expectHasLLM: true,
		},
		{
			name:       "Invalid - missing evaluation id",
			cmd:        FinisEvaluationCommannd{EvaluationID: ""},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FinisEvaluationCommanndHandler(
				context.TODO(),
				tt.cmd,
				app.Repositories.EvaluationsRepository,
				app.Services.LLMService,
				nil,                 // no se usa en el handler actual, pero mantenemos la firma
				reports.Publisher{}, // idem
				app.Repositories.VerbalMemorySubtestRepository,
				app.Repositories.VisualMemorySubtestRepository,
				app.Repositories.ExecutiveFunctionsSubtestRepository,
				app.Repositories.LetterCancellationRepository,
				app.Repositories.LanguageFluencyRepository,
				app.Repositories.VisualSpatialRepository,
			)

			if tt.shouldPass {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				// Estado final
				if got.CurrentStatus != tt.expectStatus {
					t.Errorf("expected status %q, got %q", tt.expectStatus, got.CurrentStatus)
				}
				// LLM
				if tt.expectHasLLM && got.AssistantAnalysis == "" {
					t.Errorf("expected non-empty AssistantAnalysis, got empty")
				}
				// Sanity: es la misma evaluación
				if got.PK == "" {
					t.Errorf("expected evaluation PK to be set, got empty")
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got nil (cmd=%+v)", tt.cmd)
				}
			}
		})
	}
}
