package LCdomain

import "context"

type LetterCancellationRepository interface {
	Save(ctx context.Context, subtest *LettersCancellationSubtest) error
	GetByEvaluationID(ctx context.Context, evaluationID string) (LettersCancellationSubtest, error)
}

type InMemoryLetterCancellationRepository struct {
	storage map[string]*LettersCancellationSubtest
}

func NewInMemoryLetterCancellationRepository() *InMemoryLetterCancellationRepository {
	return &InMemoryLetterCancellationRepository{
		storage: make(map[string]*LettersCancellationSubtest),
	}
}

func (repo *InMemoryLetterCancellationRepository) Save(ctx context.Context, subtest *LettersCancellationSubtest) error {
	repo.storage[subtest.PK] = subtest
	return nil
}

func (repo *InMemoryLetterCancellationRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (LettersCancellationSubtest, error) {
	return LettersCancellationSubtest{}, nil
}
