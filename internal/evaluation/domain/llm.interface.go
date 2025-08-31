package domain

import (
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
)

type LLMService interface {
	GenerateAnalysis(evaluation Evaluation) (string, error)
	LettersCancellationAnalysis(subtest *LCdomain.LettersCancellationSubtest, patientAge int) (string, error)
	VerbalMemoryAnalysis(subtest *VEMdomain.VerbalMemorySubtest, patientAge int) (string, error)

	ExecutiveFunctionsAnalysis(subtest *EFdomain.ExecutiveFunctionsSubtest, patientAge int) (string, error)

	LanguageFluencyAnalysis(subtest *LFdomain.LanguageFluency, patientAge int) (string, error)
}

type MockInterface struct{}

func (mi MockInterface) GenerateAnalysis(evaluation Evaluation) (string, error) {
	return "test analysis by chatgpt broder", nil
}

func (mi MockInterface) LettersCancellationAnalysis(subtest *LCdomain.LettersCancellationSubtest, patientAge int) (string, error) {
	return "test analysis by chatgpt broder", nil
}

func (mi MockInterface) VerbalMemoryAnalysis(subtest *VEMdomain.VerbalMemorySubtest, patientAge int) (string, error) {
	return "test analysis by chatgpt broder", nil
}

func (mi MockInterface) ExecutiveFunctionsAnalysis(subtest *EFdomain.ExecutiveFunctionsSubtest, patientAge int) (string, error) {
	return "test analysis by chatgpt broder", nil
}

func (mi MockInterface) LanguageFluencyAnalysis(subtest *LFdomain.LanguageFluency, patientAge int) (string, error) {
	return "test analysis by chatgpt broder", nil
}
