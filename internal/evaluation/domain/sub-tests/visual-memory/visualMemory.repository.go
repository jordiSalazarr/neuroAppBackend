package VIMdomain

import "context"

type VisualMemoryRepository interface {
	Save(ctx context.Context, s *VisualMemorySubtest) error
	GetLastByEvaluationID(ctx context.Context, evaluationID string) (VisualMemorySubtest, error)
	ListByEvaluationID(ctx context.Context, evaluationID string) ([]VisualMemorySubtest, error)
}
