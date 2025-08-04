package createevaluation

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	"neuro.app.jordi/internal/shared/mail"
)

func CreateEvaluationCommandHandler(command CreateEvaluationCommand, ctx context.Context, llmService domain.LLMService, fileFormatterService domain.FileFormaterService, evaluationsRepository domain.EvaluationsRepository, mailService mail.MailProvider) error {

	// 🔹 Map DTO → Domain Models
	var sections []domain.Section
	for _, s := range command.Sections {
		var questions []domain.Question
		for _, q := range s.Questions {
			questions = append(questions, domain.Question{
				ID:       q.ID,
				Answer:   q.Answer,
				Response: q.Response,
				Correct:  q.Correct,
				Score:    q.Score,
			})
		}
		sections = append(sections, domain.Section{
			Name:      s.Name,
			Score:     s.Score,
			Questions: questions,
		})
	}

	// 🔹 Use the new constructor
	evaluation, err := domain.NewEvaluation(command.TotalScore, command.PatientName, command.SpecialistMail, command.SpecialistID, sections)
	if err != nil {
		return err
	}

	// 🔹 Generate analysis and PDF
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

	// 🔹 Send the PDF by email
	return mailService.SendEmail(command.SpecialistMail, "New evaluation report for "+command.PatientName, "Please find the evaluation report attached.", command.PatientName+".pdf", pdfInBytes)
}
