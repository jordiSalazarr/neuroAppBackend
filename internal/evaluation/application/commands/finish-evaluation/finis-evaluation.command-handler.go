package finishevaluation

import (
	"context"
	"errors"

	"neuro.app.jordi/internal/evaluation/domain"
	reports "neuro.app.jordi/internal/evaluation/domain/services"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
)

func FinisEvaluationCommanndHandler(
	ctx context.Context, command FinisEvaluationCommannd,
	evaluationRepository domain.EvaluationsRepository, llmService domain.LLMService,
	fileFormatterService domain.FileFormaterService,
	evaluationPublisher reports.Publisher, verbalMemoryRepository VEMdomain.VerbalMemoryRepository,
	visualMemoryRepository VIMdomain.VisualMemoryRepository, executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository,
	letterCancellationRepository LCdomain.LetterCancellationRepository, languageFluencyRepository LFdomain.LanguageFluencyRepository) (domain.Evaluation, error) {
	evaluation, err := evaluationRepository.GetByID(ctx, command.EvaluationID)
	if err != nil {
		return domain.Evaluation{}, err
	}

	err = populateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository)
	if err != nil {
		return domain.Evaluation{}, err
	}

	res, err := llmService.GenerateAnalysis(evaluation)
	if err != nil {
		return domain.Evaluation{}, err
	}
	evaluation.AssistantAnalysis = res

	html, err := fileFormatterService.GenerateHTML(evaluation)
	if err != nil {
		return domain.Evaluation{}, err
	}

	pdf, err := fileFormatterService.ConvertHTMLtoPDF(html)
	if err != nil {
		return domain.Evaluation{}, err
	}
	//TODO generate HTML and then PDF and upload it
	key, url, err := evaluationPublisher.PublishPDF(ctx, evaluation, pdf)
	if err != nil {
		return domain.Evaluation{}, nil
	}

	evaluation.StorageURL = url
	evaluation.StorageKey = key
	if err = evaluationRepository.Update(ctx, evaluation); err != nil {
		return domain.Evaluation{}, nil
	}
	err = populateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository)
	return evaluation, err

}

func populateEvaluationWithSubtests(ctx context.Context,
	evaluation *domain.Evaluation,
	verbalMemoryRepository VEMdomain.VerbalMemoryRepository,
	visualMemoryRepository VIMdomain.VisualMemoryRepository,
	executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository,
	letterCancellationRepository LCdomain.LetterCancellationRepository,
	languageFluencyRepository LFdomain.LanguageFluencyRepository,
) error {
	if evaluation == nil {
		return errors.New("populateEvaluationWithSubtests: evaluation is nil")
	}

	var merr error

	// 1) Verbal Memory
	vm, err := verbalMemoryRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.VerbalmemorySubTest = vm

	// 2) Visual Memory
	vim, err := visualMemoryRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.VisualMemorySubTest = vim

	// 3) Executive Functions (TMT, etc.)
	ef, err := executiveFunctionsRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.ExecutiveFunctionSubTest = ef

	// 4) Letter Cancellation (Atención sostenida)
	lc, err := letterCancellationRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.LetterCancellationSubTest = lc

	// 5) Language Fluency
	lf, err := languageFluencyRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.LanguageFluencySubTest = lf

	return merr
}
