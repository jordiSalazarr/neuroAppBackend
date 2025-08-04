package infra

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"neuro.app.jordi/internal/evaluation/domain"
	"neuro.app.jordi/internal/shared/config"
)

type EvaluationsS3Bucket struct{}

func (m EvaluationsS3Bucket) Save(ctx context.Context, evaluation domain.Evaluation, PDFcontent []byte) error {
	s3Client := GetS3Client()
	bucketName, err := config.GetS3BucketName()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("evaluations/%s/%s.pdf", evaluation.PatientName, evaluation.PK)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(PDFcontent),
		Metadata: map[string]string{
			domain.PatientNameHeaderKey:    evaluation.PatientName,
			domain.SpecialistMailHeaderKey: evaluation.SpecialistMail,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
