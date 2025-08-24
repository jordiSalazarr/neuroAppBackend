package LCdomain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"neuro.app.jordi/internal/evaluation/utils"
)

var (
	ErrInvalidTotalTargets = errors.New("totalTargets must be > 0")
	ErrInvalidCorrect      = errors.New("correct must be >= 0 and <= totalTargets")
	ErrInvalidErrors       = errors.New("errors must be >= 0")
	ErrInvalidTimeInSecs   = errors.New("timeInSecs must be > 0")
)

type LettersCancellationSubtest struct {
	PK                string            `json:"pk"`
	TotalTargets      int               `json:"totalTargets"`
	Correct           int               `json:"correct"`
	Errors            int               `json:"errors"`
	TimeInSecs        int               `json:"timeInSecs"`
	EvaluationID      string            `json:"evaluationId"`
	CancellationScore CancellationScore `json:"score"`
	AssistantAnalysis string            `json:"assistantAnalysis"`
	CreatedAt         time.Time         `json:"created_at"`
}

type CancellationScore struct {
	Score          int     `json:"score"`
	CpPerMin       float64 `json:"cpPerMin"`       // (H - C) / min
	Accuracy       float64 `json:"accuracy"`       // H / N_targets
	Omissions      int     `json:"omissions"`      // N_targets - H
	OmissionsRate  float64 `json:"omissionsRate"`  // omissions / N_targets
	CommissionRate float64 `json:"commissionRate"` // C / (H + C) si H+C>0
	HitsPerMin     float64 `json:"hitsPerMin"`
	ErrorsPerMin   float64 `json:"errorsPerMin"`
}

// Configuración de scoring (k: factor de capado de errores; p.ej. 2 => hasta 2× objetivos)
type CancellationScoreConfig struct {
	CapErrorFactor float64 // default 2.0
}

func NewLettersCancellationSubtest(totalTargets, correct, errors, timeInSecs int, evaluationID string, cfg *CancellationScoreConfig) (*LettersCancellationSubtest, error) {
	score, err := calculateCancellationScore(totalTargets, correct, errors, timeInSecs, cfg)
	if err != nil {
		return nil, err
	}
	return &LettersCancellationSubtest{
		PK:                uuid.NewString(),
		TotalTargets:      totalTargets,
		Correct:           correct,
		Errors:            errors,
		TimeInSecs:        timeInSecs,
		EvaluationID:      evaluationID,
		CancellationScore: score,
		AssistantAnalysis: "",
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func calculateCancellationScore(totalTargets, correct, errors, timeInSecs int, cfg *CancellationScoreConfig) (CancellationScore, error) {
	if totalTargets <= 0 {
		return CancellationScore{}, ErrInvalidTotalTargets
	}
	if correct < 0 || correct > totalTargets {
		return CancellationScore{}, ErrInvalidCorrect
	}
	if errors < 0 {
		return CancellationScore{}, ErrInvalidErrors
	}
	if timeInSecs <= 0 {
		return CancellationScore{}, ErrInvalidTimeInSecs
	}

	k := 2.0
	if cfg != nil && cfg.CapErrorFactor > 0 {
		k = cfg.CapErrorFactor
	}

	minutes := float64(timeInSecs) / 60.0
	// Protección contra divisiones absurdas si T es muy bajo
	minutes = math.Max(minutes, 1.0/60.0)

	// Capado de errores para estabilizar la escala
	capErrors := int(math.Round(k * float64(totalTargets)))
	cappedErrors := math.Min(float64(errors), float64(capErrors))

	hitsPerMin := float64(correct) / minutes
	errorsPerMin := float64(cappedErrors) / minutes
	cpPerMin := (float64(correct) - float64(cappedErrors)) / minutes

	omissions := totalTargets - correct
	omissionsRate := float64(omissions) / float64(totalTargets)

	var commissionRate float64
	totalTaps := correct + errors
	if totalTaps > 0 {
		commissionRate = float64(errors) / float64(totalTaps)
	} else {
		commissionRate = 0
	}

	// Normalización 0..100
	cpMax := float64(totalTargets) / minutes // todos aciertos
	cpMin := -float64(capErrors) / minutes   // todos "errores" (capados)
	denom := math.Max(1e-9, cpMax-cpMin)
	norm01 := (cpPerMin - cpMin) / denom
	score := int(math.Round(100 * utils.Clamp01(norm01)))

	return CancellationScore{
		Score:          score,
		CpPerMin:       cpPerMin,
		Accuracy:       float64(correct) / float64(totalTargets),
		Omissions:      omissions,
		OmissionsRate:  omissionsRate,
		CommissionRate: commissionRate,
		HitsPerMin:     hitsPerMin,
		ErrorsPerMin:   errorsPerMin,
	}, nil
}
