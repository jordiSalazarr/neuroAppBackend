// package: VIMdomain

package VIMdomain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyEvaluationID = errors.New("evaluation_id is required")
	ErrEmptyShapeName    = errors.New("shape_name is required")
	ErrNoImageBytes      = errors.New("image bytes are required")
	ErrUnsupportedMIME   = errors.New("unsupported image content_type")
	ErrTooLarge          = errors.New("image too large")
)

// Reutilizamos el mismo concepto de estado que en BVMT.
type ShapeSubtestStatus string

const (
	ShapeStatusUploaded ShapeSubtestStatus = "uploaded" // imagen recibida, sin score aún
	ShapeStatusScored   ShapeSubtestStatus = "scored"   // ya se calculó score
)

// Figura esperada
type ShapeName string

const (
	ShapeCircle   ShapeName = "circle"
	ShapeSquare   ShapeName = "square"
	ShapeTriangle ShapeName = "triangle"
)

func ParseShapeName(s string) (ShapeName, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "circle", "círculo", "circulo":
		return ShapeCircle, nil
	case "square", "cuadrado":
		return ShapeSquare, nil
	case "triangle", "triángulo", "triangulo":
		return ShapeTriangle, nil
	default:
		return "", ErrEmptyShapeName
	}
}

// Métricas del algoritmo de CV para la figura
type GeoShapeScore struct {
	FinalScore int      `json:"finalScore"`        // 0..100 (agregado)
	Pass       bool     `json:"pass"`              // según umbral
	Reasons    []string `json:"reasons,omitempty"` // explicaciones

	IoU         *float64 `json:"iou,omitempty"`         // intersección/union con ideal
	Circularity *float64 `json:"circularity,omitempty"` // solo círculo
	AngleRMSE   *float64 `json:"angleRMSE,omitempty"`   // polígonos: desvío angular
	SideCV      *float64 `json:"sideCV,omitempty"`      // polígonos: variabilidad de lados

	// Si quieres transportar un overlay para depuración visual
	DebugPNGBase64 string `json:"debugPngBase64,omitempty"`
}

// Puertos (reutilizamos tu ImageStorage)
type ImageStorage interface {
	// Guarda la imagen y devuelve referencia opaca (s3://... o ruta local)
	PutPNGOrJPEG(raw []byte, contentType string) (imageRef string, err error)
}

// Motor de scoring (similar a tu BVMTScorer, pero sin plantilla)
type GeoShapeScorer interface {
	// Calcula métricas/score desde una ruta/ref (o bytes) para la figura esperada.
	// patientImagePath puede ser una ruta local tras almacenar en ImageStorage.
	Score(expected ShapeName, patientImagePath string) (GeoShapeScore, error)
}

// Entidad de dominio del subtest de Figuras Geométricas
type GeoFigureSubtest struct {
	PK           string             `json:"pk"`
	EvaluationID string             `json:"evaluation_id"`
	Shape        ShapeName          `json:"shape"` // circle/square/triangle
	ImageRef     string             `json:"image_ref"`
	ContentType  string             `json:"content_type"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	ImageSHA256  string             `json:"image_sha256"`
	CapturedAt   time.Time          `json:"captured_at"`
	Status       ShapeSubtestStatus `json:"status"`
	Score        *GeoShapeScore     `json:"score,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

// Factory (desde bytes, ideal para multipart)
func NewGeoFigureSubtestFromBytes(
	evaluationID string,
	shapeName string, // "circle" | "square" | "triangle"
	contentType string, // "image/png" | "image/jpeg"
	raw []byte,
	capturedAt time.Time,
	store ImageStorage,
	maxBytes int,
) (*GeoFigureSubtest, error) {
	if strings.TrimSpace(evaluationID) == "" {
		return nil, ErrEmptyEvaluationID
	}
	shape, err := ParseShapeName(shapeName)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, ErrNoImageBytes
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if !(ct == "image/png" || ct == "image/jpeg") {
		// detecta por contenido para ser robusto
		detected := http.DetectContentType(raw)
		if !(strings.HasPrefix(detected, "image/png") || strings.HasPrefix(detected, "image/jpeg")) {
			return nil, ErrUnsupportedMIME
		}
		ct = detected
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return nil, ErrTooLarge
	}

	// Extrae dimensiones
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	// Guarda en storage
	ref, err := store.PutPNGOrJPEG(raw, ct)
	if err != nil {
		return nil, err
	}

	sha := sha256.Sum256(raw)
	now := time.Now().UTC()

	return &GeoFigureSubtest{
		PK:           uuid.NewString(),
		EvaluationID: evaluationID,
		Shape:        shape,
		ImageRef:     ref,
		ContentType:  ct,
		Width:        cfg.Width,
		Height:       cfg.Height,
		ImageSHA256:  hex.EncodeToString(sha[:]),
		CapturedAt:   capturedAt.UTC(),
		Status:       ShapeStatusUploaded,
		CreatedAt:    now,
	}, nil
}

// Helper para marcar como puntuado
func (g *GeoFigureSubtest) AttachScore(s GeoShapeScore) {
	g.Score = &s
	g.Status = ShapeStatusScored
}
