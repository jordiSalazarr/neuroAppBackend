package createlettercancelationsubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
)

func CreateLetterCancellationSubtestCommandHandler(ctx context.Context, command CreateLetterCancellationSubtestCommand, letterCancellationRepo LCdomain.LetterCancellationRepository, evaluationsRepo domain.EvaluationsRepository, llmService domain.LLMService) (*LCdomain.LettersCancellationSubtest, error) {
	cfg := &LCdomain.CancellationScoreConfig{
		CapErrorFactor: 2.0,
	}
	subtest, err := LCdomain.NewLettersCancellationSubtest(command.TotalTargets, command.Correct, command.Errors, command.TimeInSecs, command.EvaluationID, cfg)
	if err != nil {
		return nil, err
	}
	err = letterCancellationRepo.Save(ctx, subtest)
	if err != nil {
		return nil, err
	}
	return subtest, nil

}
