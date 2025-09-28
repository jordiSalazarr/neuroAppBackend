package mail

import "context"

type MailProvider interface {
	SendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error
	SendEmailWithAttachment(ctx context.Context, to, subject, body string, attachmentName string, attachment []byte) error
}
