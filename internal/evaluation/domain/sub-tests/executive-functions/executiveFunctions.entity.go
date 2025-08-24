package EFdomain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"neuro.app.jordi/internal/evaluation/utils"
)

type ExuctiveFunctionSubtestType string

const (
	A  ExuctiveFunctionSubtestType = "a"
	AB ExuctiveFunctionSubtestType = "a+b"
)

type ExecutiveFunctionsSubtest struct {
	PK             string                      `json:"pk" bson:"pk"`
	NumberOfItems  int                         `json:"numberOfItems" bson:"numberOfItems"`
	TotalClicks    int                         `json:"totalClicks" bson:"totalClicks"`
	TotalErrors    int                         `json:"totalErrors" bson:"totalErrors"`
	TotalCorrect   int                         `json:"totalCorrect" bson:"totalCorrect"`
	TotalTime      time.Duration               `json:"totalTime" bson:"totalTime"`
	Type           ExuctiveFunctionSubtestType `json:"type" bson:"type"`
	Score          ExecutiveFunctionsScore     `json:"score" bson:"score"`
	EvauluationId  string                      `json:"evaluationId" bson:"evaluationId"`
	AssistanAnalys string                      `json:"assistantAnalystId" bson:"assistantAnalystId"`
	CreatedAt      time.Time                   `json:"createdAt" bson:"createdAt"`
}

type ExecutiveFunctionsScore struct {
	Score          int     `json:"score" bson:"score"`                   // 0..100
	Accuracy       float64 `json:"accuracy" bson:"accuracy"`             // TotalCorrect / NumberOfItems
	SpeedIndex     float64 `json:"speedIndex" bson:"speedIndex"`         // tIdeal / durationSec (cap 0..1)
	CommissionRate float64 `json:"commissionRate" bson:"commissionRate"` // TotalErrors / TotalClicks (0..1)
	DurationSec    float64 `json:"durationSec" bson:"durationSec"`       // duración total en segundos
}

func NewExecutiveFunctionsSubtest(
	numberOfItems int,
	totalErrors int,
	totalCorrect int,
	totalTime time.Duration,
	subtestType ExuctiveFunctionSubtestType,
	totalClicks int,
	evaluationId string,
	createdAt time.Time,
) (*ExecutiveFunctionsSubtest, error) {
	return &ExecutiveFunctionsSubtest{
		PK:             uuid.NewString(),
		NumberOfItems:  numberOfItems,
		TotalErrors:    totalErrors,
		TotalCorrect:   totalCorrect,
		TotalTime:      totalTime,
		Type:           subtestType,
		EvauluationId:  evaluationId,
		TotalClicks:    totalClicks,
		AssistanAnalys: "",
		CreatedAt:      createdAt,
	}, nil
}

func (s ExecutiveFunctionsSubtest) DurationSeconds() float64 {
	sec := s.TotalTime.Seconds()
	if sec <= 0 {
		return 1 // evita divisiones por 0 o valores negativos
	}
	return sec
}

func ScoreExecutiveFunctions(sub ExecutiveFunctionsSubtest) (ExecutiveFunctionsScore, error) {
	if sub.NumberOfItems <= 0 {
		return ExecutiveFunctionsScore{}, errors.New("numberOfItems must be > 0")
	}
	if sub.TotalCorrect < 0 || sub.TotalErrors < 0 || sub.TotalClicks < 0 {
		return ExecutiveFunctionsScore{}, errors.New("counts must be >= 0")
	}

	clicks := sub.TotalClicks
	if clicks == 0 {
		clicks = sub.TotalCorrect + sub.TotalErrors
		if clicks == 0 {
			clicks = 1
		}
	}

	durationSec := sub.DurationSeconds()
	edges := sub.NumberOfItems - 1
	if edges < 1 {
		edges = 1
	}

	idealPerItem := 1.0
	if sub.Type == AB {
		idealPerItem = 1.5
	}
	tIdeal := float64(edges) * idealPerItem

	accuracy := utils.Clamp01(float64(sub.TotalCorrect) / float64(sub.NumberOfItems))
	speedIdx := utils.Clamp01(tIdeal / durationSec)
	commissionRate := utils.Clamp01(float64(sub.TotalErrors) / float64(clicks))

	score01 := utils.Clamp01(0.7*accuracy + 0.3*speedIdx - 0.2*commissionRate)
	return ExecutiveFunctionsScore{
		Score:          int(math.Round(100 * score01)),
		Accuracy:       accuracy,
		SpeedIndex:     speedIdx,
		CommissionRate: commissionRate,
		DurationSec:    durationSec,
	}, nil
}
