package sendevaluation

import (
	"neuro.app.jordi/internal/evaluation/domain"
	"neuro.app.jordi/internal/shared/mail"
)

func SendEvaluationCommandHandler(command SendEvaluationCommand, fileSigner domain.FileSigner, mailService mail.MailProvider) error {
	//this will be a lambda, will be triggered by the S3 bucket, we need to get the signed url from the S3 bucket
	// and send the email with the report
	_, err := fileSigner.SignFile(command.StoredPDFPath)
	if err != nil {
		return err
	}

	// //now, we will use amazon SES to send the email
	// subject := "New evaluation report for " + command.PatientName
	// body := "Please find the evaluation report attached. You can download it from the following link:" + url
	// err = mailService.SendMailHTML(command.SpecialistMail, subject, body)
	// if err != nil {
	// 	return err
	// }
	return nil
}
