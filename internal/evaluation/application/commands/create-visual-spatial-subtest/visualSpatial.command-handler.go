package createvisualspatialsubtest

import (
	"context"

	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

func CreateViusualSpatialCommandHandler(ctx context.Context, cmd CreateVisualSpatialSubtestCommand, repo VPdomain.ResultRepository) (*VPdomain.VisualSpatialSubtest, error) {
	subtest, err := VPdomain.NewVisualSpatialSubtest(cmd.EvaluationID, cmd.Note, cmd.Score)
	if err != nil {
		return nil, err
	}

	if err = repo.Save(ctx, subtest); err != nil {
		return nil, err
	}
	return subtest, nil
}
