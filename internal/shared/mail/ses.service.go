package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESEmailSender struct {
	client     *ses.Client
	senderAddr string
}

func NewSESEmailSender(ctx context.Context) (*SESEmailSender, error) {
	awsRegion := os.Getenv("AWS_REGION")
	sender := os.Getenv("SES_SENDER")

	if awsRegion == "" || sender == "" {
		return nil, fmt.Errorf("missing AWS_REGION or SES_SENDER env vars")
	}

	// Carga la configuración con credenciales de entorno
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ses.NewFromConfig(cfg)

	return &SESEmailSender{
		client:     client,
		senderAddr: sender,
	}, nil
}

func (s *SESEmailSender) SendEmail(ctx context.Context, to string, subject string, htmlBody string, textBody string) error {
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(subject),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBody),
				},
				Text: &types.Content{
					Data: aws.String(textBody),
				},
			},
		},
		Source: aws.String(s.senderAddr),
	}

	_, err := s.client.SendEmail(ctx, input)
	return err
}

func (s *SESEmailSender) SendEmailWithAttachment(ctx context.Context, to, subject, body string, attachmentName string, attachment []byte) error {
	// MIME básico
	var emailRaw bytes.Buffer
	boundary := "NextPartBoundary"

	emailRaw.WriteString(fmt.Sprintf("From: %s\r\n", s.senderAddr))
	emailRaw.WriteString(fmt.Sprintf("To: %s\r\n", to))
	emailRaw.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	emailRaw.WriteString("MIME-Version: 1.0\r\n")
	emailRaw.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	emailRaw.WriteString("\r\n--" + boundary + "\r\n")
	emailRaw.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	emailRaw.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	emailRaw.WriteString(body + "\r\n")

	// Attachment
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(attachment)))
	base64.StdEncoding.Encode(encoded, attachment)

	emailRaw.WriteString("\r\n--" + boundary + "\r\n")
	emailRaw.WriteString("Content-Type: application/pdf\r\n")
	emailRaw.WriteString("Content-Disposition: attachment; filename=\"" + attachmentName + "\"\r\n")
	emailRaw.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	emailRaw.Write(encoded)
	emailRaw.WriteString("\r\n--" + boundary + "--")

	input := &ses.SendRawEmailInput{
		RawMessage: &types.RawMessage{Data: emailRaw.Bytes()},
	}

	_, err := s.client.SendRawEmail(ctx, input)
	return err
}
