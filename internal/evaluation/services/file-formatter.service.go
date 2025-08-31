package services

import (
	"neuro.app.jordi/internal/evaluation/domain"
)

type FileFormatterService struct{}

func NewFileFormatter() FileFormatterService {
	return FileFormatterService{}
}

func (ff FileFormatterService) GenerateHTML(evaluation domain.Evaluation) (string, error) {
	// 	var b strings.Builder

	// 	b.WriteString(`<!DOCTYPE html>
	// <html lang="es">
	// <head>
	// <meta charset="UTF-8">
	// <title>Reporte de Evaluación Neurocognitiva</title>
	// <style>
	//     body { font-family: 'Arial', sans-serif; margin: 40px; background: #fafafa; color: #333; }
	//     h1 { color: #003366; text-align: center; font-size: 28px; margin-bottom: 10px; }
	//     h2 { color: #005599; border-bottom: 2px solid #005599; padding-bottom: 4px; margin-top: 30px; }
	//     .info-card { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); margin-bottom: 20px; }
	//     .score-table { width: 100%; border-collapse: collapse; margin-top: 15px; }
	//     .score-table th, .score-table td { border: 1px solid #ddd; padding: 10px; text-align: center; font-size: 14px; }
	//     .score-table th { background: #f0f0f0; color: #333; }
	//     .score-bar { height: 18px; border-radius: 4px; background: #eee; margin: 4px 0; position: relative; }
	//     .score-fill { height: 100%; border-radius: 4px; background: #4CAF50; }
	//     .section-card { background: #fff; padding: 15px; margin-top: 15px; border-radius: 6px; border-left: 4px solid #4CAF50; }
	//     .question { margin: 6px 0; font-size: 14px; }
	//     .analysis { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); line-height: 1.6; }
	// </style>
	// </head>
	// <body>

	// <h1>Informe de Evaluación Neurocognitiva</h1>
	// `)

	// 	// ✅ Tarjeta de información del paciente
	// 	b.WriteString(fmt.Sprintf(`
	// <div class="info-card">
	//   <p><strong>Paciente:</strong> %s</p>
	//   <p><strong>Especialista:</strong> %s</p>
	//   <p><strong>Fecha:</strong> %s</p>
	//   <p><strong>Puntuación Total:</strong> %d / 100</p>
	//   <div class="score-bar"><div class="score-fill" style="width:%d%%;"></div></div>
	// </div>`,
	// 		evaluation.PatientName,
	// 		evaluation.SpecialistMail,
	// 		evaluation.CreatedAt.Format("2006-01-02 15:04"),
	// 		evaluation.TotalScore,
	// 		evaluation.TotalScore,
	// 	))

	// 	// ✅ Secciones
	// 	b.WriteString("<h2>Resultados por Sección</h2>")
	// 	for _, section := range evaluation.Sections {
	// 		if section.Score > 0 {
	// 			b.WriteString(fmt.Sprintf(`
	// <div class="section-card">
	//   <h3>%s — %d/100</h3>
	//   <div class="score-bar"><div class="score-fill" style="width:%d%%;"></div></div>
	// `, section.Name, section.Score, section.Score))

	// 			for _, q := range section.Questions {
	// 				b.WriteString(fmt.Sprintf(`
	//   <div class="question">
	//     <strong>Pregunta:</strong> %s<br>
	//     <em>Respuesta:</em> %s<br>
	//     <em>Correcta:</em> %s | <em>Puntos:</em> %d
	//   </div>`, q.Answer, q.Response, q.Correct, q.Score))
	// 			}
	// 			b.WriteString("</div>")
	// 		}
	// 	}

	// 	// ✅ Análisis Clínico (IA)
	// 	if evaluation.AssistantAnalysis != "" {
	// 		b.WriteString("<h2>Informe Clínico del Especialista Asistente</h2>")
	// 		b.WriteString(fmt.Sprintf("<div class='analysis'>%s</div>", strings.ReplaceAll(evaluation.AssistantAnalysis, "\n", "<br>")))
	// 	}

	// 	b.WriteString("</body></html>")

	// 	return b.String(), nil
	// }

	// func (ff FileFormatterService) ConvertHTMLtoPDF(html string) ([]byte, error) {
	// 	// Create new PDF generator
	// 	pdfg, err := pdfgen.NewPDFGenerator()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to initialize wkhtmltopdf: %w", err)
	// 	}

	// 	// Set page options
	// 	page := pdfgen.NewPageReader(bytes.NewReader([]byte(html)))
	// 	page.EnableLocalFileAccess.Set(true) // Important if you're embedding images or CSS

	// 	// Add to generator
	// 	pdfg.AddPage(page)

	// 	// Global PDF options
	// 	pdfg.Dpi.Set(300)
	// 	pdfg.Orientation.Set(pdfgen.OrientationPortrait)
	// 	pdfg.PageSize.Set(pdfgen.PageSizeA4)
	// 	pdfg.NoCollate.Set(false)
	// 	pdfg.MarginTop.Set(10)
	// 	pdfg.MarginBottom.Set(10)
	// 	pdfg.MarginLeft.Set(10)
	// 	pdfg.MarginRight.Set(10)

	// 	// Generate PDF
	// 	if err := pdfg.Create(); err != nil {
	// 		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	// 	}

	// 	// Get the result as a byte slice
	return "pdfg.Bytes()", nil
}
