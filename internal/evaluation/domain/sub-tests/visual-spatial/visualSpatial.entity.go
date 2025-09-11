package VPdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type VisualSpatialScore struct {
	Val int
}
type VisualSpatialNote struct {
	Val string
}
type VisualSpatialSubtest struct {
	Id           string
	EvalautionId string
	Score        VisualSpatialScore
	Note         VisualSpatialNote
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func newScore(score int) (VisualSpatialScore, error) {
	if score < 0 || score > 5 {
		return VisualSpatialScore{}, errors.New("test score must be between 0-5")
	}
	return VisualSpatialScore{
		Val: score,
	}, nil
}

func newNote(note string) (VisualSpatialNote, error) {
	if len(note) >= 2500 {
		return VisualSpatialNote{}, errors.New("max of 2500 words")
	}

	return VisualSpatialNote{
		Val: note,
	}, nil
}

func NewVisualSpatialSubtest(evaluationId, note string, score int) (*VisualSpatialSubtest, error) {
	id := uuid.New().String()
	domainScore, err := newScore(score)
	if err != nil {
		return nil, err
	}
	domainNote, err := newNote(note)
	if err != nil {
		return nil, err
	}

	return &VisualSpatialSubtest{
		Id:           id,
		EvalautionId: evaluationId,
		Note:         domainNote,
		Score:        domainScore,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil

}

func NewVisualSpatialSubtestFromExisting(id, evaluationId, note string, score int, createdAt, updatedAt time.Time) (*VisualSpatialSubtest, error) {
	domainScore, err := newScore(score)
	if err != nil {
		return nil, err
	}
	domainNote, err := newNote(note)
	if err != nil {
		return nil, err
	}

	return &VisualSpatialSubtest{
		Id:           id,
		EvalautionId: evaluationId,
		Note:         domainNote,
		Score:        domainScore,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil

}
