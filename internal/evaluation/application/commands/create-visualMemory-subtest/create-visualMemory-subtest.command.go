package createvisualmemorysubtest

import "time"

// CreateBVMTSubtestCommand es el DTO interno que usa la capa de aplicación.
// La imagen ya viene en bytes (parseada por el HTTP handler).
type CreateBVMTSubtestCommand struct {
	EvaluationID string
	FigureName   string
	CapturedAt   time.Time // opcional; si es zero, la app puede usar time.Now().UTC()
	ImageBytes   []byte    // bytes del archivo subido
	ContentType  string    // Content-Type del archivo (puede venir vacío)
}
