package fileformatter

import (
	"bytes"
	"fmt"
	"strings"

	"neuro.app.jordi/internal/evaluation/domain"

	fpdf "github.com/go-pdf/fpdf"
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
	// Config A4 en mm
	pdf := fpdf.New("P", "mm", "A4", "")

	// Márgenes
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Fuente UTF-8 (necesitas un TTF en tu repo; recomiendo DejaVu Sans)
	// Coloca DejaVuSans.ttf en ./assets/fonts/DejaVuSans.ttf
	pdf.AddUTF8Font("DejaVu", "", "assets/fonts/DejaVuSans.ttf")

	pdf.AddUTF8Font("DejaVu", "B", "assets/fonts/DejaVuSans-Bold.ttf")
	// si no tienes bold, puedes omitir y usar solo regular

	pdf.SetFont("DejaVu", "", 12)

	// fpdf.HTMLBasic es MUY básico; quita <html>, <head>, etc. si existen
	clean := stripOuterHTML(html)

	ht := pdf.HTMLBasicNew()
	// interlineado aproximado (altura de línea)
	ht.Write(6, clean)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func stripOuterHTML(s string) string {
	// fpdf.HTMLBasic no necesita <html><body> y a veces le molestan estilos
	// Limpieza muy simple:
	s = strings.ReplaceAll(s, "<!doctype html>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "<html>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "</html>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "<head>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "</head>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "<body>", "")
	s = strings.ReplaceAll(strings.ToLower(s), "</body>", "")
	return s
}

func markdownToHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
