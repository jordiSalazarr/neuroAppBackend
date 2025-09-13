package listevaluations

import "time"

type ListEvaluationsQuery struct {
	SpecialistID  string    `json:"specialist_id"`
	FromDate      time.Time `json:"from_date"`
	ToDate        time.Time `json:"to_date"`
	SearchTerm    string    `json:"search_term"`
	Offset        int       `json:"offset"`
	Limit         int       `json:"limit"`
	OnlyCompleted bool
}
