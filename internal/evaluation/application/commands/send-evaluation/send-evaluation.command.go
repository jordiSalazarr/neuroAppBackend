package sendevaluation

type SendEvaluationCommand struct {
	PatientName       string
	SpecialistMail    string
	EvaluationContent []byte
	StoredPDFPath     string
}
