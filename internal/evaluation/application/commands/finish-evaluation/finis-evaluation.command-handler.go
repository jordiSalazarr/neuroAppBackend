package finishevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/application/services"
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

	err = services.PopulateEvaluationWithSubtests(ctx, &evaluation, verbalMemoryRepository, visualMemoryRepository, executiveFunctionsRepository, letterCancellationRepository, languageFluencyRepository)
	if err != nil {
		return domain.Evaluation{}, err
	}

	res, err := llmService.GenerateAnalysis(evaluation)
	if err != nil {
		return domain.Evaluation{}, err
	}
	evaluation.AssistantAnalysis = res
	evaluation.CurrentStatus = domain.EvaluationCurrentStatusCompleted

	//TODO: is sending the pdf really necesary? I just would save the info, if they click send to mail. them we do it (in the historial component)
	// html, err := fileFormatterService.GenerateHTML(evaluation)
	// if err != nil {
	// 	return domain.Evaluation{}, err
	// }

	// pdf, err := fileFormatterService.ConvertHTMLtoPDF(html)
	// if err != nil {
	// 	return domain.Evaluation{}, err
	// }
	// //TODO generate HTML and then PDF and upload it
	// key, url, err := evaluationPublisher.PublishPDF(ctx, evaluation, pdf)
	// if err != nil {
	// 	return domain.Evaluation{}, nil
	// }

	// evaluation.StorageURL = url
	// evaluation.StorageKey = key
	if err = evaluationRepository.Update(ctx, evaluation); err != nil {
		return domain.Evaluation{}, nil
	}
	return evaluation, err

}
