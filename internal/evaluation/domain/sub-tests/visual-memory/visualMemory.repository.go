package VIMdomain

import "context"

type VisualMemoryRepository interface {
	Save(ctx context.Context, s *BVMTSubtest) error
	GetByEvaluationID(ctx context.Context, evaluationID string) (BVMTSubtest, error)
}
