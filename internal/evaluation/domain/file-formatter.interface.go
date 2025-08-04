package domain

type FileFormaterService interface {
	GenerateHTML(evaluation Evaluation) (string, error)
	ConvertHTMLtoPDF(html string) ([]byte, error)
}

type MockFileFormatterService struct{}

func (m MockFileFormatterService) GenerateHTML(evaluation Evaluation) (string, error) {
	return "<html><body>Test HTML</body></html>", nil
}
func (m MockFileFormatterService) ConvertHTMLtoPDF(html string) ([]byte, error) {
	return []byte("PDF content"), nil
}
