package domain

type LLMService interface {
	GenerateAnalysis(evaluation Evaluation) (string, error)
}

type MockInterface struct{}

func (mi MockInterface) GenerateAnalysis(evaluation Evaluation) (string, error) {
	return "test analysis by chatgpt broder", nil
}
