package createvisualspatialsubtest

type CreateVisualSpatialSubtestCommand struct {
	EvaluationID string `json:"evaluation_id"`
	Note         string `json:"note"`
	Score        int    `json:"score"`
}
