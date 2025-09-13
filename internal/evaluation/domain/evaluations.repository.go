package domain

import (
	"context"
	"errors"
	"time"
)

type EvaluationsRepository interface {
	Save(ctx context.Context, evaluation Evaluation) error
	GetByID(ctx context.Context, id string) (Evaluation, error)

	Update(ctx context.Context, evaluation Evaluation) error

	GetMany(ctx context.Context, fromDate, toDate time.Time, offset, limit int, searchTerm string, specialist_id string, onlyCompleted bool) ([]*Evaluation, error)
}

type MockEvaluationsRepository struct {
	evaluations []Evaluation
}

func NewEvaluationsRepository() *MockEvaluationsRepository {
	return &MockEvaluationsRepository{
		evaluations: []Evaluation{},
	}
}

func (m *MockEvaluationsRepository) Save(ctx context.Context, evaluation Evaluation) error {
	// Mock implementation for testing purposes
	m.evaluations = append(m.evaluations, evaluation)
	return nil
}

func (m *MockEvaluationsRepository) Update(ctx context.Context, evaluation Evaluation) error {
	// Mock implementation for testing purposes
	for idx, eval := range m.evaluations {
		if eval.PK == evaluation.PK {
			m.evaluations[idx] = evaluation
			return nil
		}
	}
	return errors.New("nnot found for update")
}
func (m *MockEvaluationsRepository) GetByID(ctx context.Context, id string) (Evaluation, error) {
	// Mock implementation for testing purposes
	for _, eval := range m.evaluations {
		if eval.PK == id {
			return eval, nil
		}
	}
	return Evaluation{}, nil
}
