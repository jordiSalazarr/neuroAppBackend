package services

import (
	"bytes"
	"fmt"
	"strings"

	pdfgen "github.com/SebastiaanKlippert/go-wkhtmltopdf"

	"neuro.app.jordi/internal/evaluation/domain"
)

type FileFormatterService struct{}

func NewFileFormatter() FileFormatterService {
	return FileFormatterService{}
}

func (ff FileFormatterService) GenerateHTML(evaluation domain.Evaluation) (string, error) {
	var b strings.Builder

	// Header
	b.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		h1 { color: #222; }
		h2 { color: #444; margin-top: 30px; }
		p, li { font-size: 16px; color: #333; }
		.score-table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		.score-table th, .score-table td { border: 1px solid #ccc; padding: 10px; text-align: left; }
		.score-table th { background-color: #f4f4f4; }
		</style><title>Patient Evaluation Report</title></head><body>`)

	// Title & patient info
	b.WriteString(fmt.Sprintf(`<h1>Neurocognitive Evaluation Report</h1>
		<p><strong>Patient Name:</strong> %s<br>
		<strong>Specilist mail:</strong> %s<br>
		<strong>Evaluation Date:</strong> %s</p>`,
		evaluation.PatientName,
		evaluation.SpecialistMail,
		evaluation.CreatedAt.Format("2006-01-02 15:04"),
	))

	// Scores Table
	b.WriteString("<h2>Test Scores</h2><table class='score-table'><tr><th>Subtest</th><th>Score</th></tr>")
	b.WriteString(fmt.Sprintf("<tr><td><strong>Total Score</strong></td><td><strong>%d</strong></td></tr>", evaluation.TotalScore))
	b.WriteString("</table>")

	// Assistant's analysis (if present)
	if evaluation.AssistantAnalysis != "" {
		b.WriteString("<h2>AI Clinical Analysis</h2>")
		b.WriteString(fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(evaluation.AssistantAnalysis, "\n", "<br>")))
	}
	b.WriteString("</body></html>")

	return b.String(), nil
}

func (ff FileFormatterService) ConvertHTMLtoPDF(html string) ([]byte, error) {
	// Create new PDF generator
	pdfg, err := pdfgen.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wkhtmltopdf: %w", err)
	}

	// Set page options
	page := pdfgen.NewPageReader(bytes.NewReader([]byte(html)))
	page.EnableLocalFileAccess.Set(true) // Important if you're embedding images or CSS

	// Add to generator
	pdfg.AddPage(page)

	// Global PDF options
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(pdfgen.OrientationPortrait)
	pdfg.PageSize.Set(pdfgen.PageSizeA4)
	pdfg.NoCollate.Set(false)
	pdfg.MarginTop.Set(10)
	pdfg.MarginBottom.Set(10)
	pdfg.MarginLeft.Set(10)
	pdfg.MarginRight.Set(10)

	// Generate PDF
	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Get the result as a byte slice
	return pdfg.Bytes(), nil
}
