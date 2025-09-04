package createvisualmemorysubtest

import (
	"context"
	"time"

	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
)

func CreateVisualMemoryCommandHandler(ctx context.Context, cmd CreateBVMTSubtestCommand, repo VIMdomain.VisualMemoryRepository, storage VIMdomain.ImageStorage, maxBytes int, scorer VIMdomain.GeoShapeScorer) (*VIMdomain.GeoShapeScore, error) {
	captured := cmd.CapturedAt
	if captured.IsZero() {
		captured = time.Now().UTC()
	}

	// Construye la entidad desde bytes (valida, detecta MIME, calcula hash, sube a storage…)
	sub, err := VIMdomain.NewGeoFigureSubtestFromBytes(
		cmd.EvaluationID,
		cmd.FigureName,
		cmd.ContentType,
		cmd.ImageBytes,
		captured,
		storage,
		maxBytes,
	)
	if err != nil {
		return nil, err
	}

	// Persiste
	if err := repo.Save(ctx, sub.Score); err != nil {
		return nil, err
	}
	// _, err = templateResolver.Resolve(sub.ImageRef)
	// if err != nil {
	// 	return nil, err
	// }

	// tmp := "/Users/jordisalazarbadia/Desktop/NeuroApp/back/images/triangulo.png"

	// score, err := scorer.Score(tmp, sub.ImageRef)
	// if err != nil {
	// 	return nil, err
	// }

	// 3) Actualiza entidad a "scored" y persiste
	// sub.Score = &score
	sub.Status = VIMdomain.ShapeStatusUploaded
	if err := repo.Save(ctx, sub.Score); err != nil {
		return nil, err
	}
	return sub.Score, nil
}
