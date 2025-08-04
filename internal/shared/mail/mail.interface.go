package mail

type MailProvider interface {
	SendMailHTML(to string, subject string, body string) error
	sendWithFile(to string, subject string, body string, filePath string) error
}
