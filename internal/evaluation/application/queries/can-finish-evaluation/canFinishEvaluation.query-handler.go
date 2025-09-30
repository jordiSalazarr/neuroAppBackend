package canfinishevaluation

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

func CanFinishEvaluationQueryHandler(ctx context.Context, cmd CanFinishEvaluationQuery, evaluationRepo domain.EvaluationsRepository, verbalMemoryRepository VEMdomain.VerbalMemoryRepository, visualMemoryRepository VIMdomain.VisualMemoryRepository, executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository, letterCancellationRepository LCdomain.LetterCancellationRepository, languageFluencyRepository LFdomain.LanguageFluencyRepository, visualSpatialRepository VPdomain.ResultRepository,
) (bool, error) {
	evaluation, err := evaluationRepo.GetByID(ctx, cmd.EvaluationID)
	if err != nil {
		return false, err
	}

	err = services.PopulateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository, visualSpatialRepository)
	if err != nil {
		return false, err
	}
	return true, nil
}
