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

// ======== PÚBLICO ========

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
	// ==== Branding / estilos ====
	const (
		brandName        = "NeuroApp"
		pageW, pageH     = 210.0, 297.0 // A4 mm
		marginL, marginT = 15.0, 18.0
		marginR, marginB = 15.0, 18.0
		titleSize        = 18.0
		sectionTitleSize = 13.0
		bodySize         = 11.0
		infoLabelSize    = 10.0
		infoValueSize    = 11.0
		brandBarHeight   = 12.0
		sectionBarHeight = 8.0
		cardRadius       = 2.5
		lineH            = 5.5
	)

	// Colores (RGB)
	brandColor := color{32, 140, 140} // teal
	darkText := color{30, 30, 30}
	mutedText := color{90, 90, 90}
	lightGray := color{245, 247, 248}
	sectionBg := color{230, 244, 244}

	// ==== Parsear HTML de entrada ====
	patient, specialist, plainResults := extractFromHTML(html)

	// ==== PDF base ====
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginL, marginT, marginR)
	pdf.SetAutoPageBreak(true, marginB)
	pdf.AliasNbPages("")

	// Traductor CP1252 para tildes/ñ sin TTF
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Header
	pdf.SetHeaderFunc(func() {
		// Banda superior de marca
		setFill(pdf, brandColor)
		pdf.Rect(0, 0, pageW, brandBarHeight, "F")

		// Nombre de marca
		setText(pdf, color{255, 255, 255})
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetXY(marginL, 4.0)
		pdf.CellFormat(0, 6, tr(brandName), "", 0, "L", false, 0, "")
		// Línea de separación bajo la banda
		setDraw(pdf, color{230, 230, 230})
		pdf.SetLineWidth(0.2)
		pdf.Line(marginL, brandBarHeight, pageW-marginR, brandBarHeight)
	})

	// Footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		setText(pdf, mutedText)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 10, tr(fmt.Sprintf("Página %d/{nb}", pdf.PageNo())), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()

	// ==== Título ====
	setText(pdf, darkText)
	pdf.SetFont("Helvetica", "B", titleSize)
	pdf.CellFormat(0, 10, tr("Informe Neuropsicológico"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// ==== Tarjeta de información Paciente/Especialista ====
	drawInfoCard(pdf, tr, patient, specialist, darkText, mutedText, lightGray, cardRadius, infoLabelSize, infoValueSize, lineH)

	pdf.Ln(3)

	// ==== Sección: Resultados ====
	drawSectionTitle(pdf, tr, "Resultados", brandColor, sectionBg, sectionBarHeight, sectionTitleSize)
	pdf.Ln(2)

	// Cuerpo de resultados (multicell bonito)
	setText(pdf, darkText)
	pdf.SetFont("Helvetica", "", bodySize)
	writeBody(pdf, tr, plainResults, lineH)

	// ==== Salida ====
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// ======== PRIVADO: helpers de estilo ========

type color struct{ R, G, B int }

func setFill(pdf *fpdf.Fpdf, c color) { pdf.SetFillColor(c.R, c.G, c.B) }
func setDraw(pdf *fpdf.Fpdf, c color) { pdf.SetDrawColor(c.R, c.G, c.B) }
func setText(pdf *fpdf.Fpdf, c color) { pdf.SetTextColor(c.R, c.G, c.B) }

func drawInfoCard(
	pdf *fpdf.Fpdf,
	tr func(string) string,
	patient, specialist string,
	labelCol, mutedCol, bg color,
	radius, labelSize, valueSize, lineH float64,
) {
	x := pdf.GetX()
	y := pdf.GetY()
	w := 0.0 // ancho automático (hasta margen derecho)
	h := 22.0
	// Fondo tarjeta (rounded)
	setFill(pdf, bg)
	setDraw(pdf, color{220, 225, 228})
	pdf.SetLineWidth(0.2)
	roundedRect(pdf, x, y, "") // reset? no
	// Calcular el ancho disponible usando GetPageSize y GetMargins
	pageW, _ := pdf.GetPageSize()
	_, _, right, _ := pdf.GetMargins()
	pdf.RoundedRect(x, y, pageW-right-x, h, radius, "1234", "DF")
	pdf.SetXY(x+6, y+4)

	// Labels/values
	setText(pdf, mutedCol)
	pdf.SetFont("Helvetica", "", labelSize)
	pdf.CellFormat(w, lineH, tr("Paciente"), "", 0, "L", false, 0, "")
	pdf.SetX(x + 35)
	setText(pdf, labelCol)
	pdf.SetFont("Helvetica", "B", valueSize)
	pdf.CellFormat(w, lineH, tr(nonEmpty(patient, "—")), "", 1, "L", false, 0, "")

	pdf.SetX(x + 6)
	setText(pdf, mutedCol)
	pdf.SetFont("Helvetica", "", labelSize)
	pdf.CellFormat(w, lineH, tr("Especialista"), "", 0, "L", false, 0, "")
	pdf.SetX(x + 35)
	setText(pdf, labelCol)
	pdf.SetFont("Helvetica", "B", valueSize)
	pdf.CellFormat(w, lineH, tr(nonEmpty(specialist, "—")), "", 1, "L", false, 0, "")

	// Mueve el cursor bajo la tarjeta
	pdf.SetY(y + h + 2)
}

func drawSectionTitle(pdf *fpdf.Fpdf, tr func(string) string, title string, brand, bg color, barH float64, size float64) {
	// Barra sutil de fondo
	setFill(pdf, bg)
	left, top, right, bottom := pdf.GetMargins() // left, top, right, bottom
	_ = top
	_ = bottom // si no los usas

	x := left
	y := pdf.GetY()

	pageW, _ := pdf.GetPageSize() // ancho y alto de la página
	w := pageW - left - right     // ancho disponible entre márgenes

	pdf.Rect(x, y, w, barH, "F") // barra de fondo entre márgenes

	// Texto
	pdf.SetY(pdf.GetY() + 1.5)
	setText(pdf, brand)
	pdf.SetFont("Helvetica", "B", size)
	pdf.CellFormat(0, barH, tr(title), "", 1, "L", false, 0, "")
}

func writeBody(pdf *fpdf.Fpdf, tr func(string) string, text string, lineH float64) {
	// Soporte simple de listas y párrafos
	lines := strings.Split(normalizeWhitespace(text), "\n")
	for _, ln := range lines {
		l := strings.TrimSpace(ln)
		if l == "" {
			pdf.Ln(lineH / 2)
			continue
		}
		if strings.HasPrefix(l, "- ") || strings.HasPrefix(l, "• ") || strings.HasPrefix(l, "* ") {
			// viñeta
			pdf.SetX(pdf.GetX() + 2)
			pdf.CellFormat(3, lineH, tr("•"), "", 0, "", false, 0, "")
			pdf.MultiCell(0, lineH, tr(strings.TrimSpace(l[2:])), "", "L", false)
		} else {
			pdf.MultiCell(0, lineH, tr(l), "", "L", false)
		}
	}
}

func nonEmpty(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// colapsar múltiples \n
	re := regexp.MustCompile(`\n{3,}`)
	return re.ReplaceAllString(s, "\n\n")
}

// ======== PRIVADO: parser HTML MUY sencillo ========

func extractFromHTML(html string) (patient string, specialist string, plainResults string) {
	// Paciente
	rePac := regexp.MustCompile(`(?is)<strong>\s*Paciente:\s*</strong>\s*([^<]+)`)
	if m := rePac.FindStringSubmatch(html); len(m) > 1 {
		patient = strings.TrimSpace(htmlToText(m[1]))
	}
	// Especialista
	reSpec := regexp.MustCompile(`(?is)<strong>\s*Especialista:\s*</strong>\s*([^<]+)`)
	if m := reSpec.FindStringSubmatch(html); len(m) > 1 {
		specialist = strings.TrimSpace(htmlToText(m[1]))
	}
	// Resultados (del <h2>Resultados</h2> hasta el final)
	reRes := regexp.MustCompile(`(?is)<h2[^>]*>\s*Resultados\s*</h2>(.*)$`)
	if m := reRes.FindStringSubmatch(html); len(m) > 1 {
		plainResults = strings.TrimSpace(htmlToText(m[1]))
	} else {
		plainResults = strings.TrimSpace(htmlToText(html))
	}
	return
}

func htmlToText(s string) string {
	// Cambios rápidos para listas
	s = strings.ReplaceAll(s, "</li>", "\n")
	reLi := regexp.MustCompile(`(?is)<li[^>]*>`)
	s = reLi.ReplaceAllString(s, "- ")

	// Sustituir <br> y cerrar párrafos por saltos de línea
	reBr := regexp.MustCompile(`(?is)<br\s*/?>`)
	s = reBr.ReplaceAllString(s, "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</h1>", "\n\n")
	s = strings.ReplaceAll(s, "</h2>", "\n\n")
	s = strings.ReplaceAll(s, "</h2 >", "\n\n")

	// Quitar el resto de tags
	reTags := regexp.MustCompile(`(?is)<[^>]+>`)
	s = reTags.ReplaceAllString(s, "")

	// Normalizar espacios
	s = strings.TrimSpace(s)
	reSpace := regexp.MustCompile(`[ \t]+`)
	s = reSpace.ReplaceAllString(s, " ")

	// Colapsa saltos múltiples
	s = normalizeWhitespace(s)
	return s
}

// ======== PRIVADO: util ========

func roundedRect(pdf *fpdf.Fpdf, x, y float64, _ string) {
	_ = x
	_ = y
	// No hace nada: dejamos la API por si quieres evolucionar
}

// ======== PRIVADO: markdown ========

func markdownToHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
