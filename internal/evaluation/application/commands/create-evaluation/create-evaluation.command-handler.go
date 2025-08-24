package createevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
)

func CreateEvaluationCommandHandler(command CreateEvaluationCommand, ctx context.Context, evaluationsRepository domain.EvaluationsRepository) (domain.Evaluation, error) {

	evaluation, err := domain.NewEvaluation(command.PatientName, command.SpecialistMail, command.SpecialistID, command.PatientAge)
	if err != nil {
		return domain.Evaluation{}, err
	}

	err = evaluationsRepository.Save(ctx, evaluation)
	return domain.Evaluation{}, err
}
