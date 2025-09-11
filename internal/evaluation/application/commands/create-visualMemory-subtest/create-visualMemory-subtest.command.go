package createvisualmemorysubtest

type CreateVisualMemorySubtestCommand struct {
	EvaluationID string `json:"evaluation_id"`
	Score        int    `json:"score_id"`
	Note         string `json:"note"`
}
