package getevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
)

func GetEvaluationQueryHandler(ctx context.Context, query GetEvaluationQuery, evaluationsRepository domain.EvaluationsRepository) (domain.Evaluation, error) {
	//GET EVALUAtION BY ID
	evaluation, err := evaluationsRepository.GetByID(ctx, query.EvaluationID)
	if err != nil {
		return domain.Evaluation{}, err
	}

	return evaluation, nil
}
