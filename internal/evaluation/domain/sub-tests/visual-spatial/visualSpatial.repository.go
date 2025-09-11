package VPdomain

import "context"

type ResultRepository interface {
	Save(ctx context.Context, res *VisualSpatialSubtest) error
	GetByEvaluationID(ctx context.Context, id string) (*VisualSpatialSubtest, error)
	GetByID(ctx context.Context, id string) (*VisualSpatialSubtest, error)
}
