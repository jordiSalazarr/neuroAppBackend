package createexecutivefunctionssubtest

import "time"

type CreateExecutiveFunctionsSubtestCommand struct {
	NumberOfItems int           `json:"numberOfItems" bson:"numberOfItems"`
	TotalClicks   int           `json:"totalClicks" bson:"totalClicks"`
	StartAt       time.Time     `json:"startAt" bson:"startAt"`
	TotalErrors   int           `json:"totalErrors" bson:"totalErrors"`
	TotalCorrect  int           `json:"totalCorrect" bson:"totalCorrect"`
	TotalTime     time.Duration `json:"totalTime" bson:"totalTime"`
	Type          string        `json:"type" bson:"type"`
	EvaluationId  string        `json:"evaluationId" bson:"evaluationId"`
	CreatedAt     time.Time     `json:"createdAt" bson:"createdAt"`
}
