package reports

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"time"

	"neuro.app.jordi/internal/evaluation/domain"
)

type Publisher struct {
	Bucket domain.BucketStorage
	KMSKey string // opcional
}

func (p Publisher) PublishPDF(ctx context.Context, eval domain.Evaluation, pdf []byte) (key string, url string, err error) {
	key = path.Join("evaluations", eval.PK, fmt.Sprintf("informe-%s.pdf", eval.PK))

	opts := domain.PutOptions{
		ContentType:        "application/pdf",
		ContentDisposition: fmt.Sprintf(`inline; filename="informe-%s.pdf"`, eval.PK),
		KMSKeyID:           p.KMSKey,
		Tags: map[string]string{
			"evaluation_id": eval.PK,
			"patient_name":  eval.PatientName,
			"env":           "prod",
		},
	}

	res, err := p.Bucket.Put(ctx, key, bytes.NewReader(pdf), int64(len(pdf)), opts)
	if err != nil {
		return "", "", err
	}

	// Genera URL temporal para compartir/adjuntar en email
	u, err := p.Bucket.PresignGet(ctx, key, 24*time.Hour)
	if err != nil {
		return res.Key, "", err
	}
	return res.Key, u, nil
}
