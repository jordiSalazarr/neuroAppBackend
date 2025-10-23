package fileformatter

import (
	"bytes"
	"fmt"
	"regexp"
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
	// Convierte el análisis (markdown) a HTML simple
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
	// PDF A4 (mm)
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// SIN ASSETS: fuentes core (Helvetica). Para tildes/ñ usamos traductor CP1252.
	pdf.SetFont("Helvetica", "", 12)
	tr := pdf.UnicodeTranslatorFromDescriptor("") // CP1252 (Latin-1 sup.)

	// Sanea el HTML para HTMLBasic (parser básico)
	clean := sanitizeForHTMLBasic(html)

	// Renderer HTML básico
	ht := pdf.HTMLBasicNew()

	// Protege contra pánicos del parser
	var out bytes.Buffer
	var writeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				writeErr = fmt.Errorf("pdf render failed: %v", r)
			}
		}()
		// Importante: traducir el string antes de pasarlo a Write
		ht.Write(6, tr(clean)) // 6 ≈ altura de línea
	}()
	if writeErr != nil {
		return nil, writeErr
	}

	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// --- Helpers ---

func sanitizeForHTMLBasic(s string) string {
	// Trabajamos en minúsculas sólo para tags a limpiar (preserva texto)
	l := strings.ToLower(s)

	// Quita envoltorios que no necesita HTMLBasic
	reMeta := regexp.MustCompile(`(?is)<meta[^>]*>`)
	l = reMeta.ReplaceAllString(l, "")
	removeExact := []string{"<!doctype html>", "<html>", "</html>", "<head>", "</head>", "<body>", "</body>"}
	for _, tag := range removeExact {
		l = strings.ReplaceAll(l, tag, "")
	}

	// <hr> → salto visual simple
	l = strings.ReplaceAll(l, "<hr>", "\n\n")

	// Normaliza strong/em a b/i (HTMLBasic entiende mejor B/I/U)
	l = strings.ReplaceAll(l, "<strong>", "<b>")
	l = strings.ReplaceAll(l, "</strong>", "</b>")
	l = strings.ReplaceAll(l, "<em>", "<i>")
	l = strings.ReplaceAll(l, "</em>", "</i>")

	// Elimina <style> y <link> que no procesa
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	l = reStyle.ReplaceAllString(l, "")
	reLink := regexp.MustCompile(`(?is)<link[^>]*>`)
	l = reLink.ReplaceAllString(l, "")

	// Aplica cambios sobre el original preservando mayúsc./acentos
	s = reMeta.ReplaceAllString(s, "")
	for _, tag := range removeExact {
		s = strings.ReplaceAll(strings.ReplaceAll(s, strings.ToLower(tag), ""), tag, "")
	}
	s = strings.ReplaceAll(s, "<hr>", "\n\n")
	s = strings.ReplaceAll(s, "<strong>", "<b>")
	s = strings.ReplaceAll(s, "</strong>", "</b>")
	s = strings.ReplaceAll(s, "<em>", "<i>")
	s = strings.ReplaceAll(s, "</em>", "</i>")
	s = reStyle.ReplaceAllString(s, "")
	s = reLink.ReplaceAllString(s, "")

	return s
}

func markdownToHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
