package createevaluation

type CreateEvaluationCommand struct {
	PatientName    string `json:"patientName"`
	PatientAge     int    `json:"patientAge"`
	SpecialistMail string `json:"specialistMail"`
	SpecialistID   string `json:"specialistId"`
}
