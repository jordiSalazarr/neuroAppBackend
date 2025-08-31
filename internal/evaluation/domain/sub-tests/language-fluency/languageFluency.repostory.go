package LFdomain

import (
	"context"
	"errors"
)

type LanguageFluencyRepository interface {
	Save(ctx context.Context, lf LanguageFluency) error
	GetByID(ctx context.Context, id string) (LanguageFluency, error)

	GetByEvaluationID(ctx context.Context, id string) (LanguageFluency, error)
}

type LanguageFluencyMock struct {
	Data map[string]LanguageFluency
}

func NewLanguageFluencyMock() *LanguageFluencyMock {
	return &LanguageFluencyMock{
		Data: make(map[string]LanguageFluency),
	}
}
func (repo *LanguageFluencyMock) Save(ctx context.Context, lf LanguageFluency) error {
	repo.Data[lf.PK] = lf
	return nil
}
func (repo *LanguageFluencyMock) GetByID(ctx context.Context, id string) (LanguageFluency, error) {
	lf, ok := repo.Data[id]
	if !ok {
		return LanguageFluency{}, errors.New("not found")
	}
	return lf, nil
}

func (repo *LanguageFluencyMock) GetByEvaluationID(ctx context.Context, evaluationID string) (LanguageFluency, error) {
	for _, test := range repo.Data {
		if test.EvaluationID == evaluationID {
			return test, nil
		}
	}
	return LanguageFluency{}, nil
}
