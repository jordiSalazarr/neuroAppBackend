package mail

import (
	"bytes"
	"log"

	"github.com/go-mail/mail"
	"neuro.app.jordi/internal/shared/config"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewMailer inicializa la configuración SMTP
func NewMailer() *Mailer {
	return &Mailer{
		Host:     config.GetConfig().SMTP_HOST,
		Port:     config.GetConfig().SMTP_PORT,
		Username: config.GetConfig().SMTP_USERNAME,
		Password: config.GetConfig().SMTP_PASSWORD,
		From:     config.GetConfig().SMTP_FROM,
	}
}

// SendEmail envía un correo con adjunto opcional
func (m *Mailer) SendEmail(to, subject, body string, pdfName string, pdfData []byte) error {
	msg := mail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body) // puedes usar "text/plain" si es texto simple

	// ✅ Si hay PDFData, lo adjuntamos
	if pdfData != nil {
		msg.AttachReader(pdfName, bytes.NewReader(pdfData))
	}

	d := mail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	d.StartTLSPolicy = mail.MandatoryStartTLS // seguridad TLS

	// ✅ Enviar
	if err := d.DialAndSend(msg); err != nil {
		return err
	}
	log.Printf("✅ Email enviado a %s", to)
	return nil
}
