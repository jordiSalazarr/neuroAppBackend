package createverbalmemorysubtest

import "time"

type CreateVerbalMemorySubtestCommand struct {
	EvaluationID  string    `json:"evaluation_id"`
	StartAt       time.Time `json:"start_at"`
	GivenWords    []string  `json:"given_words"`
	RecalledWords []string  `json:"recalled_words"`
}
