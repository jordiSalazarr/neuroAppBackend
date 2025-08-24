package createevaluation

type CreateEvaluationCommand struct {
	TotalScore     int    `json:"totalScore"`
	PatientName    string `json:"patientName"`
	PatientAge     int    `json:"patientAge"`
	SpecialistMail string `json:"specialistMail"`
	SpecialistID   string `json:"specialistId"`
}
