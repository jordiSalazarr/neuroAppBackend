package listevaluations

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
)

func GetEvaluationsCommandHandler(ctx context.Context, query ListEvaluationsQuery, evaluationRepo domain.EvaluationsRepository) ([]*domain.Evaluation, error) {
	return evaluationRepo.GetMany(ctx, query.FromDate, query.ToDate, query.Offset, query.Limit, query.SearchTerm, query.SpecialistID)
}
