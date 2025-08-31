package VPdomain

import "time"

type ClockDrawResult struct {
	ID                    string
	EvaluationID          string
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
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
