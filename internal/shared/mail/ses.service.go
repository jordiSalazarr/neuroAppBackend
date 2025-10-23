package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/google/uuid"
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

func (s *SESEmailSender) SendEmailWithAttachment(
	ctx context.Context,
	to, subject, htmlBody, textBody, attachmentName string,
	attachment []byte,
) error {

	// Encode Subject (RFC 2047)
	encSubject := mime.QEncoding.Encode("utf-8", subject)

	// Boundaries
	mixedB := "MIX-" + uuid.NewString()
	altB := "ALT-" + uuid.NewString()

	var buf bytes.Buffer

	// Headers
	from := s.senderAddr
	now := time.Now().UTC().Format(time.RFC1123Z)
	msgID := fmt.Sprintf("<%s@%s>", uuid.NewString(), strings.SplitN(from, "@", 2)[1])

	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", encSubject)
	fmt.Fprintf(&buf, "Date: %s\r\n", now)
	fmt.Fprintf(&buf, "Message-ID: %s\r\n", msgID)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedB)
	fmt.Fprintf(&buf, "\r\n") // end headers

	// ---- multipart/alternative (text + html)
	fmt.Fprintf(&buf, "--%s\r\n", mixedB)
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altB)

	// text/plain (quoted-printable)
	fmt.Fprintf(&buf, "--%s\r\n", altB)
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qpTxt := quotedprintable.NewWriter(&buf)
	_, _ = qpTxt.Write([]byte(textBody))
	_ = qpTxt.Close()
	fmt.Fprintf(&buf, "\r\n")

	// text/html (quoted-printable)
	fmt.Fprintf(&buf, "--%s\r\n", altB)
	fmt.Fprintf(&buf, "Content-Type: text/html; charset=utf-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qpHTML := quotedprintable.NewWriter(&buf)
	_, _ = qpHTML.Write([]byte(htmlBody))
	_ = qpHTML.Close()
	fmt.Fprintf(&buf, "\r\n")

	// cierre del alternative
	fmt.Fprintf(&buf, "--%s--\r\n", altB)

	// ---- Adjunto PDF (base64 con wrap 76)
	fmt.Fprintf(&buf, "--%s\r\n", mixedB)
	safeName := attachmentName
	if safeName == "" {
		safeName = "informe.pdf"
	}
	fmt.Fprintf(&buf, "Content-Type: application/pdf; name=\"%s\"\r\n", safeName)
	fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", safeName)
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n\r\n")

	b64 := base64.StdEncoding.EncodeToString(attachment)
	// wrap a 76 chars por línea con CRLF
	for i := 0; i < len(b64); i += 76 {
		end := i + 76
		if end > len(b64) {
			end = len(b64)
		}
		buf.WriteString(b64[i:end])
		buf.WriteString("\r\n")
	}
	fmt.Fprintf(&buf, "\r\n")

	// cierre del mixed
	fmt.Fprintf(&buf, "--%s--\r\n", mixedB)

	// Envío SES
	input := &ses.SendRawEmailInput{
		RawMessage: &types.RawMessage{Data: buf.Bytes()},
	}
	_, err := s.client.SendRawEmail(ctx, input)
	return err
}
