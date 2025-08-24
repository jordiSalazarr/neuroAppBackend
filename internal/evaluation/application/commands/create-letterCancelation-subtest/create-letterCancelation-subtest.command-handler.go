package createlettercancelationsubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
)

func CreateLetterCancellationSubtestCommandHandler(ctx context.Context, command CreateLetterCancellationSubtestCommand, letterCancellationRepo LCdomain.LetterCancellationRepository, evaluationsRepo domain.EvaluationsRepository, llmService domain.LLMService) (*LCdomain.LettersCancellationSubtest, error) {
	cfg := &LCdomain.CancellationScoreConfig{
		CapErrorFactor: 2.0, // factor de capado de errores, por si tocan muy rapido la pantalla, normalizar errores. Se puede jugar con este valor.
	}
	subtest, err := LCdomain.NewLettersCancellationSubtest(command.TotalTargets, command.Correct, command.Errors, command.TimeInSecs, command.EvaluationID, cfg)
	if err != nil {
		return nil, err
	}
	parentEval, err := evaluationsRepo.GetByID(ctx, command.EvaluationID)
	if err != nil {
		return nil, err
	}
	analysis, err := llmService.LettersCancellationAnalysis(subtest, parentEval.PatientAge)
	if err != nil {
		return nil, err
	}
	subtest.AssistantAnalysis = analysis
	err = letterCancellationRepo.Save(ctx, subtest)
	if err != nil {
		return nil, err
	}
	return subtest, nil

}
