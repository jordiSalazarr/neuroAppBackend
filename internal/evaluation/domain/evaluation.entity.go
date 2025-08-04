package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const MaxUserName = 50
const PatientNameHeaderKey = "patient-name"
const SpecialistMailHeaderKey = "specialist-mail"

var ErrInvalidName = errors.New("invalid name")

type Evaluation struct {
	PK                string    `json:"pk"`
	PatientName       string    `json:"patientName"`
	SpecialistMail    string    `json:"specialistMail"`
	SpecialistID      string    `json:"specialistId"`
	AssistantAnalysis string    `json:"assistantAnalysis"`
	CreatedAt         time.Time `json:"createdAt"`

	TotalScore int       `json:"totalScore"`
	Sections   []Section `json:"sections"` // ✅ New nested structure
}

type Section struct {
	Name      string     `json:"name"`      // e.g. "Memory"
	Score     int        `json:"score"`     // sum of questions
	Questions []Question `json:"questions"` // ✅ Each question
}

type Question struct {
	ID       string `json:"id"`       // unique per question
	Answer   string `json:"text"`     // question asked
	Response string `json:"response"` // patient's answer
	Correct  string `json:"correct"`  // expected answer (optional)
	Score    int    `json:"score"`    // score for this question
}

func newEvaluationTotalScore(score int) (int, error) {
	if score <= 0 || score >= 100000 {
		return 0, errors.New("invalid evaluation score")
	}
	return score, nil
}

func newPatientName(name string) (string, error) {
	if len(name) <= 0 || len(name) >= MaxUserName {
		return "", errors.New("invalid name")
	}
	return name, nil
}

func NewEvaluation(totalScore int, patientNameInput string, specialistMailInput, specialitID string, atentionScore, motoreScore, spatialViewScore, memoryScore *int) (Evaluation, error) {
	finalScore, err := newEvaluationTotalScore(totalScore)
	if err != nil {
		return Evaluation{}, err
	}

	patientName, err := newPatientName(patientNameInput)
	if err != nil {
		return Evaluation{}, err
	}
	id := uuid.NewString()
	evaluation := Evaluation{
		PK:                id,
		TotalScore:        finalScore,
		PatientName:       patientName,
		SpecialistMail:    specialistMailInput,
		SpecialistID:      specialitID,
		CreatedAt:         time.Now(),
		AssistantAnalysis: "",
		Sections:          []Section{},
	}

	return evaluation, nil
}
