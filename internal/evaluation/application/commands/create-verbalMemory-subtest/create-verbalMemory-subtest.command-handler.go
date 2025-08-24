package createverbalmemorysubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
)

func CreateVerbalMemorySubtestCommandhandler(ctx context.Context, command CreateVerbalMemorySubtestCommand, evaluationRepository domain.EvaluationsRepository, llmService domain.LLMService, verbalMemorySubtestRepo VEMdomain.VerbalMemoryRepository) (VEMdomain.VerbalMemorySubtest, error) {
	//Get evaluation by ID
	evaluation, err := evaluationRepository.GetByID(ctx, command.EvaluationID)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}
	//TODO: this should return an error if the type does not match the time since start
	verbalSubtest, err := VEMdomain.NewVerbalMemorySubtest(command.EvaluationID, command.StartAt, command.GivenWords, command.RecalledWords)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}

	score, err := VEMdomain.ScoreVerbalMemory(verbalSubtest)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}
	verbalSubtest.Score = score

	res, err := llmService.VerbalMemoryAnalysis(&verbalSubtest, evaluation.PatientAge)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}
	verbalSubtest.AssistanAnalysis = res

	err = verbalMemorySubtestRepo.Save(ctx, verbalSubtest)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, err
	}

	return verbalSubtest, nil
}
