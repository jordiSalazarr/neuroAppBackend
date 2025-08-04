package createevaluation

type QuestionDTO struct {
	ID       string `json:"id"`
	Answer   string `json:"answer"`
	Response string `json:"response"`
	Correct  string `json:"correct,omitempty"`
	Score    int    `json:"score"`
}

type SectionDTO struct {
	Name      string        `json:"name"`
	Score     int           `json:"score"`
	Questions []QuestionDTO `json:"questions"`
}

type CreateEvaluationCommand struct {
	TotalScore     int          `json:"totalScore"`
	PatientName    string       `json:"patientName"`
	SpecialistMail string       `json:"specialistMail"`
	SpecialistID   string       `json:"specialistId"`
	Sections       []SectionDTO `json:"sections"`
}
