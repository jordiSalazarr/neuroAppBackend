package canfinishevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
)

func CanFinishEvaluationQueryHandler(ctx context.Context, cmd CanFinishEvaluationQuery, evaluationRepo domain.EvaluationsRepository) (bool, error) {
	return evaluationRepo.CanFinishEvaluation(ctx, cmd.EvaluationID, cmd.SpecialistID)
}
