package createvisualmemorysubtest

import (
	"context"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
)

func CreateVisualMemoryCommandHandler(ctx context.Context, cmd CreateVisualMemorySubtestCommand, repo VIMdomain.VisualMemoryRepository) (*VIMdomain.VisualMemorySubtest, error) {
	sub, err := VIMdomain.NewVisualMemorySubtest(cmd.EvaluationID, nil, cmd.Score, cmd.Note)
	if err != nil {
		return nil, err
	}
	if err := repo.Save(ctx, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}
