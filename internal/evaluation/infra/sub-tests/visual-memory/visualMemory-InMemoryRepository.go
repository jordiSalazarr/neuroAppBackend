package VIMinfra

import (
	"context"
	"sync"
	"time"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
)

type InMemoryVisualMemorySubtestRepository struct {
	mu   sync.RWMutex
	data map[string]*VIMdomain.BVMTSubtest
}

var MockBVMT = VIMdomain.BVMTSubtest{
	PK:           "bvmt#mock-001",
	EvaluationID: "eval-123456",
	FigureName:   "BVMT-R Figure 1",
	ImageRef:     "s3://neuro-app/uploads/eval-123456/bvmt-1.png",
	ContentType:  "image/png",
	Width:        1024,
	Height:       768,
	// SHA-256 del string vacío (placeholder válido de 64 hex):
	ImageSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	CapturedAt:  time.Date(2025, 8, 24, 12, 0, 0, 0, time.UTC),
	Status:      VIMdomain.BVMTSubtestStatus("processed"), // o tu constante: BVMTStatusProcessed
	Score: &VIMdomain.BVMTScore{
		IoU:        0.71,
		SSIM:       0.83,
		PSNR:       27.4,
		FinalScore: 84,
	},
	CreatedAt: time.Now(),
}

func NewInMemoryBVMTRepo() *InMemoryVisualMemorySubtestRepository {
	return &InMemoryVisualMemorySubtestRepository{data: make(map[string]*VIMdomain.BVMTSubtest)}
}

func (r *InMemoryVisualMemorySubtestRepository) Save(ctx context.Context, s *VIMdomain.BVMTSubtest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.PK] = s
	return nil
}

func (r *InMemoryVisualMemorySubtestRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (VIMdomain.BVMTSubtest, error) {
	return MockBVMT, nil
}
