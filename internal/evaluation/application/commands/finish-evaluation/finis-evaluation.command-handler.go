package finishevaluation

import (
	"context"
	"errors"

	"neuro.app.jordi/internal/evaluation/application/services"
	"neuro.app.jordi/internal/evaluation/domain"
	reports "neuro.app.jordi/internal/evaluation/domain/services"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

func FinisEvaluationCommanndHandler(
	ctx context.Context, command FinisEvaluationCommannd,
	evaluationRepository domain.EvaluationsRepository, llmService domain.LLMService,
	fileFormatterService domain.FileFormaterService,
	evaluationPublisher reports.Publisher, verbalMemoryRepository VEMdomain.VerbalMemoryRepository,
	visualMemoryRepository VIMdomain.VisualMemoryRepository, executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository,
	letterCancellationRepository LCdomain.LetterCancellationRepository, languageFluencyRepository LFdomain.LanguageFluencyRepository, visualSpatialRepository VPdomain.ResultRepository) (domain.Evaluation, error) {
	if command.EvaluationID == "" {
		return domain.Evaluation{}, errors.New("evaluation ID is required")
	}
	evaluation, err := evaluationRepository.GetByID(ctx, command.EvaluationID)
	if err != nil {
		return domain.Evaluation{}, err
	}

	err = services.PopulateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository, visualSpatialRepository)
	if err != nil {
		return domain.Evaluation{}, err
	}

	res, err := llmService.GenerateAnalysis(evaluation)
	if err != nil {
		return domain.Evaluation{}, err
	}
	evaluation.AssistantAnalysis = res
	evaluation.CurrentStatus = domain.EvaluationCurrentStatusCompleted

	if err = evaluationRepository.Update(ctx, evaluation); err != nil {
		return domain.Evaluation{}, err
	}
	return evaluation, err

}
