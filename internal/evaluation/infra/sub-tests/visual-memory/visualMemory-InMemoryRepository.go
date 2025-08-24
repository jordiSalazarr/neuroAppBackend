package VIMinfra

import (
	"context"
	"sync"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
)

type InMemoryVisualMemorySubtestRepository struct {
	mu   sync.RWMutex
	data map[string]*VIMdomain.BVMTSubtest
}

func NewInMemoryBVMTRepo() *InMemoryVisualMemorySubtestRepository {
	return &InMemoryVisualMemorySubtestRepository{data: make(map[string]*VIMdomain.BVMTSubtest)}
}

func (r *InMemoryVisualMemorySubtestRepository) Save(ctx context.Context, s *VIMdomain.BVMTSubtest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.PK] = s
	return nil
}

func (r *InMemoryVisualMemorySubtestRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (VIMdomain.BVMTSubtest, error) {
	return VIMdomain.BVMTSubtest{}, nil
}
