package createlanguagefluencysubtest

type CreateLanguageFluencySubtestCommand struct {
	EvaluationID string
	Category     string
	Words        []string
	Duration     int // Duration in seconds
	Language     string
	Proficiency  string
	TotalTime    int // Total time taken to complete the subtest
}
