package getevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/application/services"
	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

func GetEvaluationQueryHandler(ctx context.Context, query GetEvaluationQuery,
	evaluationsRepository domain.EvaluationsRepository,
	verbalMemoryRepository VEMdomain.VerbalMemoryRepository,
	visualMemoryRepository VIMdomain.VisualMemoryRepository,
	executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository,
	letterCancellationRepository LCdomain.LetterCancellationRepository,
	languageFluencyRepository LFdomain.LanguageFluencyRepository,
	visualSpatialRepository VPdomain.ResultRepository,
) (domain.Evaluation, error) {
	//GET EVALUAtION BY ID
	evaluation, err := evaluationsRepository.GetByID(ctx, query.EvaluationID)
	if err != nil {
		return domain.Evaluation{}, err
	}

	err = services.PopulateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository, visualSpatialRepository)
	if err != nil {
		return domain.Evaluation{}, err
	}
	return evaluation, nil
}
