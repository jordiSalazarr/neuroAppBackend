package VPdomain

import "context"

type ResultRepository interface {
	Save(ctx context.Context, res *ClockDrawResult) error
	GetByEvaluationID(ctx context.Context, id string) (*ClockDrawResult, error)
	GetByID(ctx context.Context, id string) (*ClockDrawResult, error)
}
