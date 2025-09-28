package fileformatter

import (
	"bytes"
	"fmt"

	"neuro.app.jordi/internal/evaluation/domain"

	pdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/yuin/goldmark"
)

type WKHTMLFileFormatter struct{}

func NewWKHTMLFileFormatter() *WKHTMLFileFormatter {
	return &WKHTMLFileFormatter{}
}

func (f *WKHTMLFileFormatter) GenerateHTML(evaluation domain.Evaluation) (string, error) {
	// Aquí puedes usar el resultado del OpenAIService para meterlo en un template HTML
	htmlAssistantAnalysis, _ := markdownToHTML(evaluation.AssistantAnalysis)
	html := fmt.Sprintf(`
		<html>
		<head><meta charset="utf-8"><title>Informe NeuroApp</title></head>
		<body>
			<h1>Informe Neuropsicológico</h1>
			<p><strong>Paciente:</strong> %s</p>
			<p><strong>Especialista:</strong> %s</p>
			<hr>
			<h2>Resultados</h2>
			<p>%s</p>
		</body>
		</html>
	`, evaluation.PatientName, evaluation.SpecialistMail, htmlAssistantAnalysis)
	return html, nil
}

func (f *WKHTMLFileFormatter) ConvertHTMLtoPDF(html string) ([]byte, error) {
	pdfg, err := pdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to init pdf generator: %w", err)
	}

	page := pdf.NewPageReader(bytes.NewReader([]byte(html)))
	pdfg.AddPage(page)

	err = pdfg.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create pdf: %w", err)
	}

	return pdfg.Bytes(), nil
}

func markdownToHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
