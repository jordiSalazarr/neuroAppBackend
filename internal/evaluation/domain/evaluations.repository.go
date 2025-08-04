package domain

import "context"

type EvaluationsRepository interface {
	Save(ctx context.Context, evaluation Evaluation, PDFcontent []byte) error
}

type MockEvaluationsRepository struct{}

func (m MockEvaluationsRepository) Save(ctx context.Context, evaluation Evaluation, PDFcontent []byte) error {
	// Mock implementation for testing purposes
	return nil
}
