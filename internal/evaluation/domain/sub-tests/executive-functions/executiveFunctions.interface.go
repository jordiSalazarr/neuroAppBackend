package EFdomain

import "context"

type ExecutiveFunctionsSubtestRepository interface {
	Save(ctx context.Context, subtest ExecutiveFunctionsSubtest) error
	GetByID(ctx context.Context, id string) (ExecutiveFunctionsSubtest, error)

	GetByEvaluationID(ctx context.Context, evaluationID string) (ExecutiveFunctionsSubtest, error)
}

type MockExecutiveFunctionsSubtestRepository struct {
	subtests []ExecutiveFunctionsSubtest
}

func NewExecutiveFunctionsSubtestRepository() *MockExecutiveFunctionsSubtestRepository {
	return &MockExecutiveFunctionsSubtestRepository{
		subtests: []ExecutiveFunctionsSubtest{},
	}
}
func (m *MockExecutiveFunctionsSubtestRepository) Save(ctx context.Context, subtest ExecutiveFunctionsSubtest) error {
	// Mock implementation for testing purposes
	m.subtests = append(m.subtests, subtest)
	return nil
}
func (m *MockExecutiveFunctionsSubtestRepository) GetByID(ctx context.Context, id string) (ExecutiveFunctionsSubtest, error) {
	// Mock implementation for testing purposes
	for _, st := range m.subtests {
		if st.PK == id {
			return st, nil
		}
	}
	return ExecutiveFunctionsSubtest{}, nil
}

func (m *MockExecutiveFunctionsSubtestRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (ExecutiveFunctionsSubtest, error) {
	// Mock implementation for testing purposes
	for _, st := range m.subtests {
		if st.EvauluationId == evaluationID {
			return st, nil
		}
	}
	return ExecutiveFunctionsSubtest{}, nil
}
