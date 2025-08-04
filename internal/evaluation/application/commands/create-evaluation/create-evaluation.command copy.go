package createevaluation

type CreateEvaluationCommand struct {
	TotalScore       int
	MemoryScore      *int
	AtentionScore    *int
	MotoreScore      *int
	SpatialViewScore *int
	PatientName      string
	SpecialistMail   string
	SpecialistID     string
}
