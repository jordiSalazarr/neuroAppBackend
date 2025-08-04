package createevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	"neuro.app.jordi/internal/shared/mail"
)

func CreateEvaluationCommandHandler(command CreateEvaluationCommand, ctx context.Context, llmService domain.LLMService, fileFormatterService domain.FileFormaterService, evaluationsRepository domain.EvaluationsRepository, mailService mail.MailProvider) error {
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
	return mailService.SendEmail(command.SpecialistMail, "New evaluation report for "+command.PatientName, "Please find the evaluation report attached.", command.PatientName+".pdf", pdfInBytes)
}
