package createverbalmemorysubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
)

func CreateVerbalMemorySubtestCommandhandler(ctx context.Context, command CreateVerbalMemorySubtestCommand, evaluationRepository domain.EvaluationsRepository, llmService domain.LLMService, verbalMemorySubtestRepo VEMdomain.VerbalMemoryRepository) (VEMdomain.VerbalMemorySubtest, error) {

	verbalSubtest, err := VEMdomain.NewVerbalMemorySubtest(command.EvaluationID, command.StartAt, command.GivenWords, command.RecalledWords, command.Subtype)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}

	score, err := VEMdomain.ScoreVerbalMemory(verbalSubtest)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}
	verbalSubtest.Score = score

	err = verbalMemorySubtestRepo.Save(ctx, verbalSubtest)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}

	return verbalSubtest, nil
}
