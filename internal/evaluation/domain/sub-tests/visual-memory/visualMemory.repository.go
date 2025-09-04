package VIMdomain

import "context"

type VisualMemoryRepository interface {
	Save(ctx context.Context, s *GeoShapeScore) error
	GetByEvaluationID(ctx context.Context, evaluationID string) (GeoFigureSubtest, error)
}
