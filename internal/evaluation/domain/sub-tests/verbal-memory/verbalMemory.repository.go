package VEMdomain

import (
	"context"
	"time"
)

type VerbalMemoryRepository interface {
	Save(ctx context.Context, subtest VerbalMemorySubtest) error
	GetByID(ctx context.Context, id string) (VerbalMemorySubtest, error)

	GetByEvaluationID(ctx context.Context, evaluationID string) ([]VerbalMemorySubtest, error)
}

type InMemoryVerbalMemoryRepository struct {
	data map[string]VerbalMemorySubtest
}

var mock = VerbalMemorySubtest{
	Pk:               "mock-pk-123",
	SecondsFromStart: 0,
	GivenWords:       []string{"casa", "perro", "sol"},
	RecalledWords:    []string{"casa", "sol"},
	Type:             "verbal-memory",
	EvaluationID:     "eval-456",
	Score: VerbalMemoryScore{
		Score:             2,
		Hits:              2,
		Omissions:         1,
		Intrusions:        0,
		Perseverations:    0,
		Accuracy:          0.66,
		IntrusionRate:     0.0,
		PerseverationRate: 0.0,
	},
	AssistanAnalysis: "Paciente recordó 2 de 3 palabras con un 66% de precisión",
	CreatedAt:        time.Now(),
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
	return mock, nil
}

func (r InMemoryVerbalMemoryRepository) GetByEvaluationID(ctx context.Context, id string) ([]VerbalMemorySubtest, error) {
	for _, test := range r.data {
		if test.EvaluationID == id {
			return []VerbalMemorySubtest{test}, nil
		}
	}
	return []VerbalMemorySubtest{mock}, nil
}
