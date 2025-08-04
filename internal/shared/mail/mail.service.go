package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type SMTPMailProvider struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Carga credenciales desde .env.local
func NewSMTPMailProvider() SMTPMailProvider {
	_ = godotenv.Load(".env.local") // no falla si no existe
	return SMTPMailProvider{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

// SendMailHTML envía un correo con cuerpo HTML
func (s SMTPMailProvider) SendMailHTML(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	// Cabeceras + cuerpo HTML
	msg := []byte(strings.Join([]string{
		fmt.Sprintf("From: %s", s.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
	}, "\r\n"))

	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}
