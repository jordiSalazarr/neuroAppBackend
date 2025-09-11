package createexecutivefunctionssubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
)

func CreateExecutiveFunctionsSubtestCommandHandler(ctx context.Context, cmd CreateExecutiveFunctionsSubtestCommand, evaluationRepo domain.EvaluationsRepository, llmService domain.LLMService, executiveFunctionsSubtestRepo EFdomain.ExecutiveFunctionsSubtestRepository) (EFdomain.ExecutiveFunctionsSubtest, error) {
	executiveFunctionsSubtest, err := EFdomain.NewExecutiveFunctionsSubtest(cmd.NumberOfItems, cmd.TotalErrors, cmd.TotalCorrect, cmd.TotalTime, EFdomain.ExuctiveFunctionSubtestType(cmd.Type), cmd.TotalClicks, cmd.EvaluationId, cmd.CreatedAt)
	if err != nil {
		return EFdomain.ExecutiveFunctionsSubtest{}, err
	}

	score, err := EFdomain.ScoreExecutiveFunctions(*executiveFunctionsSubtest)
	if err != nil {
		return EFdomain.ExecutiveFunctionsSubtest{}, err
	}
	executiveFunctionsSubtest.Score = score

	err = executiveFunctionsSubtestRepo.Save(ctx, *executiveFunctionsSubtest)
	if err != nil {
		return EFdomain.ExecutiveFunctionsSubtest{}, err
	}

	return *executiveFunctionsSubtest, nil
}
