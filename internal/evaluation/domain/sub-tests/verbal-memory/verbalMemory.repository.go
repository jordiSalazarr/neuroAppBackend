package VEMdomain

import (
	"context"
	"errors"
)

type VerbalMemoryRepository interface {
	Save(ctx context.Context, subtest VerbalMemorySubtest) error
	GetByID(ctx context.Context, id string) (VerbalMemorySubtest, error)

	GetByEvaluationID(ctx context.Context, evaluationID string) (VerbalMemorySubtest, error)
}

type InMemoryVerbalMemoryRepository struct {
	data map[string]VerbalMemorySubtest
}

func NewInMemoryVerbalMemoryRepository() *InMemoryVerbalMemoryRepository {
	return &InMemoryVerbalMemoryRepository{
		data: make(map[string]VerbalMemorySubtest),
	}
}

func (r InMemoryVerbalMemoryRepository) Save(ctx context.Context, subtest VerbalMemorySubtest) error {
	r.data[subtest.Pk] = subtest
	return nil
}

func (r InMemoryVerbalMemoryRepository) GetByID(ctx context.Context, id string) (VerbalMemorySubtest, error) {
	subtest, exists := r.data[id]
	if !exists {
		return VerbalMemorySubtest{}, errors.New("verbal memory subtest not found")
	}
	return subtest, nil
}

func (r InMemoryVerbalMemoryRepository) GetByEvaluationID(ctx context.Context, id string) (VerbalMemorySubtest, error) {
	for _, test := range r.data {
		if test.EvaluationID == id {
			return test, nil
		}
	}
	return VerbalMemorySubtest{}, nil
}
