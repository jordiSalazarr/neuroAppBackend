package createvisualspatialsubtest

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

// Puerto secundario (servicio de análisis)
type Analyzer interface {
	Analyze(imageBytes []byte, expectedHour, expectedMin int) (Analysis, []byte, error)
}

type Analysis struct {
	Pass                  bool
	Reasons               []string
	CenterX               int
	CenterY               int
	Radius                float64
	DialCircularity       float64
	MinuteAngleDeg        float64
	HourAngleDeg          float64
	ExpectedMinuteAngle   float64
	ExpectedHourAngle     float64
	MinuteAngularErrorDeg float64
	HourAngularErrorDeg   float64
}

type EvaluateClockDrawingHandler struct {
	Repo     VPdomain.ResultRepository
	Analyzer Analyzer
}

type EvaluateClockDrawingResult struct {
	ID                    string   `json:"id"`
	EvaluationID          string   `json:"evaluationId"`
	Pass                  bool     `json:"pass"`
	Reasons               []string `json:"reasons"`
	CenterX               int      `json:"centerX"`
	CenterY               int      `json:"centerY"`
	Radius                float64  `json:"radius"`
	DialCircularity       float64  `json:"dialCircularity"`
	MinuteAngleDeg        float64  `json:"minuteAngleDeg"`
	HourAngleDeg          float64  `json:"hourAngleDeg"`
	ExpectedMinuteAngle   float64  `json:"expectedMinuteAngle"`
	ExpectedHourAngle     float64  `json:"expectedHourAngle"`
	MinuteAngularErrorDeg float64  `json:"minuteAngularErrorDeg"`
	HourAngularErrorDeg   float64  `json:"hourAngularErrorDeg"`
	DebugPNGBase64        string   `json:"debugPngBase64,omitempty"`
}

func CreateViusualSpatialCommandHandler(ctx context.Context, cmd EvaluateClockDrawingCommand, analyzer Analyzer, repo VPdomain.ResultRepository) (EvaluateClockDrawingResult, error) {
	if cmd.ExpectedHour < 0 || cmd.ExpectedHour > 23 {
		cmd.ExpectedHour = 11
	}
	if cmd.ExpectedMin < 0 || cmd.ExpectedMin > 59 {
		cmd.ExpectedMin = 10
	}

	analysis, debugPNG, err := analyzer.Analyze(cmd.ImageBytes, cmd.ExpectedHour, cmd.ExpectedMin)
	if err != nil {
		return EvaluateClockDrawingResult{}, err
	}

	entity := &VPdomain.ClockDrawResult{
		ID:                    uuid.New().String(),
		EvaluationID:          cmd.EvaluationID,
		Pass:                  analysis.Pass,
		Reasons:               analysis.Reasons,
		CenterX:               analysis.CenterX,
		CenterY:               analysis.CenterY,
		Radius:                analysis.Radius,
		DialCircularity:       analysis.DialCircularity,
		MinuteAngleDeg:        analysis.MinuteAngleDeg,
		HourAngleDeg:          analysis.HourAngleDeg,
		ExpectedMinuteAngle:   analysis.ExpectedMinuteAngle,
		ExpectedHourAngle:     analysis.ExpectedHourAngle,
		MinuteAngularErrorDeg: analysis.MinuteAngularErrorDeg,
		HourAngularErrorDeg:   analysis.HourAngularErrorDeg,
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}

	if err := repo.Save(ctx, entity); err != nil {
		return EvaluateClockDrawingResult{}, err
	}

	out := EvaluateClockDrawingResult{
		ID:                    entity.ID,
		EvaluationID:          entity.EvaluationID,
		Pass:                  entity.Pass,
		Reasons:               entity.Reasons,
		CenterX:               entity.CenterX,
		CenterY:               entity.CenterY,
		Radius:                entity.Radius,
		DialCircularity:       entity.DialCircularity,
		MinuteAngleDeg:        entity.MinuteAngleDeg,
		HourAngleDeg:          entity.HourAngleDeg,
		ExpectedMinuteAngle:   entity.ExpectedMinuteAngle,
		ExpectedHourAngle:     entity.ExpectedHourAngle,
		MinuteAngularErrorDeg: entity.MinuteAngularErrorDeg,
		HourAngularErrorDeg:   entity.HourAngularErrorDeg,
	}

	if cmd.ReturnDebug && len(debugPNG) > 0 {
		out.DebugPNGBase64 = base64.StdEncoding.EncodeToString(debugPNG)
	}

	return out, nil
}
