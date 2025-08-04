package createevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
)

func CreateEvaluationCommandHandler(command CreateEvaluationCommand, ctx context.Context, llmService domain.LLMService, fileFormatterService domain.FileFormaterService, evaluationsRepository domain.EvaluationsRepository) error {
	evaluation, err := domain.NewEvaluation(command.TotalScore, command.PatientName, command.SpecialistMail, command.SpecialistID, command.AtentionScore,
		command.MotoreScore, command.SpatialViewScore, command.MemoryScore)
	if err != nil {
		return err
	}
	assistantAnalysis, err := llmService.GenerateAnalysis(evaluation)
	if err != nil {
		return err
	}
	evaluation.AssistantAnalysis = assistantAnalysis
	html, err := fileFormatterService.GenerateHTML(evaluation)
	if err != nil {
		return err
	}
	pdfInBytes, err := fileFormatterService.ConvertHTMLtoPDF(html)
	if err != nil {
		return err
	}
	//TODO: upload to s3 bucket
	err = evaluationsRepository.Save(ctx, evaluation, pdfInBytes)
	return err
	//----------
	//From this ON, is another service, for sending the email with the report (will be another command AND lambda)
	//TODO: send mail with signed url tio s3
	// path := "report-test.pdf"
	// mailService.SendMailWithAttachment("jordisalazarbadia@gmail.com", "test report", "See the new generated report", path)
	// return pdfInBytes, nil
}
