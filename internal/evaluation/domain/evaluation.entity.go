package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

const MaxUserName = 50
const PatientNameHeaderKey = "patient-name"
const SpecialistMailHeaderKey = "specialist-mail"

var ErrInvalidName = errors.New("invalid name")

type EvaluationCurrentStatus string

const (
	EvaluationCurrentStatusCreated    EvaluationCurrentStatus = "CREATED"
	EvaluationCurrentStatusInProgress EvaluationCurrentStatus = "IN_PROGRESS"
	EvaluationCurrentStatusCompleted  EvaluationCurrentStatus = "COMPLETED"
	EvaluationCurrentStatusCancelled  EvaluationCurrentStatus = "CANCELLED"
	EvaluationCurrentStatusFailed     EvaluationCurrentStatus = "FAILED"
	EvaluationCurrentStatusPending    EvaluationCurrentStatus = "PENDING" // subtests finished, waiting for assistant analysis
)

type Evaluation struct {
	PK                        string                  `json:"pk"`
	PatientName               string                  `json:"patientName"`
	PatientAge                int                     `json:"patientAge"`
	SpecialistMail            string                  `json:"specialistMail"`
	SpecialistID              string                  `json:"specialistId"`
	AssistantAnalysis         string                  `json:"assistantAnalysis"`
	StorageURL                string                  `json:"storage_url"`
	StorageKey                string                  `json:"storage_key"`
	CreatedAt                 time.Time               `json:"createdAt"`
	CurrentStatus             EvaluationCurrentStatus `json:"currentStatus"`
	LetterCancellationSubTest LCdomain.LettersCancellationSubtest
	VisualMemorySubTest       VIMdomain.VisualMemorySubtest
	VerbalmemorySubTest       []VEMdomain.VerbalMemorySubtest
	ExecutiveFunctionSubTest  []EFdomain.ExecutiveFunctionsSubtest
	LanguageFluencySubTest    LFdomain.LanguageFluency
	VisualSpatialSubTest      VPdomain.VisualSpatialSubtest
}

func newPatientName(name string) (string, error) {
	if len(name) <= 0 || len(name) >= MaxUserName {
		return "", errors.New("invalid name")
	}
	return name, nil
}

func newPatientAge(age int) (int, error) {
	if age <= 15 || age >= 150 {
		return 0, errors.New("invalid age")
	}
	return age, nil
}

func NewEvaluation(patientNameInput, specialistMailInput, specialistID string, patientAgeInput int) (Evaluation, error) {

	patientName, err := newPatientName(patientNameInput)
	if err != nil {
		return Evaluation{}, err
	}

	patientAge, err := newPatientAge(patientAgeInput)
	if err != nil {
		return Evaluation{}, err
	}

	//TODO: this type should compose all subtests
	evaluation := Evaluation{
		PK:                uuid.NewString(),
		PatientName:       patientName,
		SpecialistMail:    specialistMailInput,
		SpecialistID:      specialistID,
		PatientAge:        patientAge,
		CreatedAt:         time.Now(),
		AssistantAnalysis: "",
		StorageURL:        "",
		CurrentStatus:     EvaluationCurrentStatusCreated,
	}

	return evaluation, nil
}
