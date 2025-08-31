package createlanguagefluencysubtest

type CreateLanguageFluencySubtestCommand struct {
	EvaluationID string   `json:"evaluationId"`
	Category     string   `json:"category"`
	Words        []string `json:"words"`
	Duration     int      `json:"duration"`
	Language     string   `json:"language"`
	Proficiency  string   `json:"proficiency"`
	TotalTime    int      `json:"totalTime"`
}
