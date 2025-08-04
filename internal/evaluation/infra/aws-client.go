package infra

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client *s3.Client
	once     sync.Once
)

// getS3Client devuelve siempre la misma instancia del cliente S3
func GetS3Client() *s3.Client {
	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("Error cargando configuración AWS: %v", err)
		}
		s3Client = s3.NewFromConfig(cfg)
	})
	return s3Client
}
