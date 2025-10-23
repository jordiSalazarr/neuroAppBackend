package mail

import "context"

type MailProvider interface {
	SendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error
	SendEmailWithAttachment(
		ctx context.Context,
		to, subject, htmlBody, textBody, attachmentName string,
		attachment []byte,
	) error
}

type MockMailService struct{}

func NewMockMailService() *MockMailService {
	return &MockMailService{}
}

func (m *MockMailService) SendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error {
	// Mock implementation: just print the email details
	// fmt.Printf("Sending email to: %s\nSubject: %s\nHTML Body: %s\nText Body: %s\n", to, subject, htmlBody, textBody)
	return nil
}

func (m *MockMailService) SendEmailWithAttachment(ctx context.Context, to, subject, body string, n string, attachmentName string, attachment []byte) error {
	// Mock implementation: just print the email details
	// fmt.Printf("Sending email to: %s\nSubject: %s\nBody: %s\nAttachment Name: %s\nAttachment Size: %d bytes\n", to, subject, body, attachmentName, len(attachment))
	return nil
}
