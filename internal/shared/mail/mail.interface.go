package mail

type MailProvider interface {
	SendEmail(to, subject, body string, pdfName string, pdfData []byte) error
}
