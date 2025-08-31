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
	ErrEmptyFigureName   = errors.New("figure_name is required")
	ErrNoImageBytes      = errors.New("image bytes are required")
	ErrUnsupportedMIME   = errors.New("unsupported image content_type")
	ErrTooLarge          = errors.New("image too large")
)

type BVMTSubtestStatus string

const (
	BVMTStatusUploaded BVMTSubtestStatus = "uploaded" // imagen recibida, sin score aún
	BVMTStatusScored   BVMTSubtestStatus = "scored"   // ya se calculó similitud
)

// domain/bvmt_ports.go
type TemplateResolver interface {
	// Devuelve la ruta local (o ref accesible) de la plantilla correcta de esa figura
	Resolve(figureName string) (string, error)
}

type BVMTScorer interface {
	// Calcula IoU/SSIM/PSNR y score final (0..100) a partir de dos rutas/refs
	Score(templatePath string, patientImagePath string) (BVMTScore, error)
}

type BVMTScore struct {
	IoU        float64 `json:"iou"`
	SSIM       float64 `json:"ssim"`
	PSNR       float64 `json:"psnr"`
	FinalScore int     `json:"finalScore"` // 0..100
}

type BVMTSubtest struct {
	PK           string            `json:"pk"`
	EvaluationID string            `json:"evaluation_id"`
	FigureName   string            `json:"figure_name"`
	ImageRef     string            `json:"image_ref"`
	ContentType  string            `json:"content_type"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	ImageSHA256  string            `json:"image_sha256"`
	CapturedAt   time.Time         `json:"captured_at"`
	Status       BVMTSubtestStatus `json:"status"`
	Score        *BVMTScore        `json:"score,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// Puertos
type ImageStorage interface {
	// Guarda la imagen y devuelve referencia opaca (p.ej. s3://bucket/key o ruta local).
	PutPNGOrJPEG(raw []byte, contentType string) (imageRef string, err error)
}

// Factory (desde bytes, ideal para multipart)
func NewBVMTSubtestFromBytes(
	evaluationID, figureName, contentType string,
	raw []byte,
	capturedAt time.Time,
	store ImageStorage,
	maxBytes int,
) (*BVMTSubtest, error) {
	if strings.TrimSpace(evaluationID) == "" {
		return nil, ErrEmptyEvaluationID
	}
	if strings.TrimSpace(figureName) == "" {
		return nil, ErrEmptyFigureName
	}
	if len(raw) == 0 {
		return nil, ErrNoImageBytes
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if !(ct == "image/png" || ct == "image/jpeg") {
		// detecta por contenido para ser más robusto
		detected := http.DetectContentType(raw)
		if !(strings.HasPrefix(detected, "image/png") || strings.HasPrefix(detected, "image/jpeg")) {
			return nil, ErrUnsupportedMIME
		}
		ct = detected
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return nil, ErrTooLarge
	}

	// Extrae dimensiones (valida que es imagen)
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
	return &BVMTSubtest{
		PK:           uuid.NewString(),
		EvaluationID: evaluationID,
		FigureName:   figureName,
		ImageRef:     ref,
		ContentType:  ct,
		Width:        cfg.Width,
		Height:       cfg.Height,
		ImageSHA256:  hex.EncodeToString(sha[:]),
		CapturedAt:   capturedAt.UTC(),
		Status:       BVMTStatusUploaded,
		CreatedAt:    now,
	}, nil
}
