package VIMdomain

import (
	"errors"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/google/uuid"
)

type VisualMemorySubtest struct {
	PK           string            `json:"pk"`
	EvaluationID string            `json:"evaluation_id"`
	Score        VisualMemoryScore `json:"score"`
	Note         VisualMemoryNote  `json:"note"`
	ImageSrc     *string           `json:"image_src"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type VisualMemoryScore struct {
	Val int
}
type VisualMemoryNote struct {
	Val string
}

func newVisualMemoryScore(score int) (*VisualMemoryScore, error) {
	if score < 0 || score > 2 {
		return nil, errors.New("invalid score, must be between 0-2")
	}
	return &VisualMemoryScore{
		Val: score,
	}, nil
}

func newVisualMemoryNote(note string) (*VisualMemoryNote, error) {
	if len(note) >= 2500 {
		return nil, errors.New("note exceeds limit")
	}
	return &VisualMemoryNote{
		Val: note,
	}, nil
}
func NewVisualMemorySubtest(evaluationId string, imageSrc *string, scoreIn int, noteIn string) (VisualMemorySubtest, error) {
	score, err := newVisualMemoryScore(scoreIn)
	if err != nil {
		return VisualMemorySubtest{}, err
	}
	note, err := newVisualMemoryNote(noteIn)
	if err != nil {
		return VisualMemorySubtest{}, err
	}
	pk := uuid.New().String()
	return VisualMemorySubtest{
		PK:           pk,
		EvaluationID: evaluationId,
		Score:        *score,
		Note:         *note,
		ImageSrc:     nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
func NewVisualMemorySubtestFromDB(pk, evaluationId string, imageSrc *string, scoreIn int, noteIn string, createdAt, updatedAt time.Time) (VisualMemorySubtest, error) {
	score, err := newVisualMemoryScore(scoreIn)
	if err != nil {
		return VisualMemorySubtest{}, err
	}
	note, err := newVisualMemoryNote(noteIn)
	if err != nil {
		return VisualMemorySubtest{}, err
	}
	return VisualMemorySubtest{
		PK:           pk,
		EvaluationID: evaluationId,
		Score:        *score,
		Note:         *note,
		ImageSrc:     imageSrc,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
