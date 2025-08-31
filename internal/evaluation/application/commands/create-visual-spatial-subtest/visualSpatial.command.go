package createvisualspatialsubtest

type EvaluateClockDrawingCommand struct {
	EvaluationID string
	ImageBytes   []byte
	ExpectedHour int // 0–23
	ExpectedMin  int // 0–59
	ReturnDebug  bool
}
