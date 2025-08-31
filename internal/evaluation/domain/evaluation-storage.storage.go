package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

type PutOptions struct {
	ContentType        string            // "application/pdf"
	ContentDisposition string            // e.g. inline; filename="informe.pdf"
	CacheControl       string            // opcional
	Tags               map[string]string // key=value
	KMSKeyID           string            // si usas SSE-KMS
	StorageClass       string            // STANDARD, INTELLIGENT_TIERING...
}

type PutResult struct {
	Key       string // s3://bucket/key
	ETag      string
	VersionID string
	URL       string // opcional: URL pública si procede (normalmente vacío)
}

type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
	VersionID   string
}

type BucketStorage interface {
	// Sube contenido por streaming (mejor que []byte)
	Put(ctx context.Context, key string, r io.Reader, size int64, opts PutOptions) (PutResult, error)
	// Lectura/metadata
	Head(ctx context.Context, key string) (ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	// URLs presignadas (útil para descarga/envío por email sin exponer el bucket)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	PresignPut(ctx context.Context, key string, ttl time.Duration, opts PutOptions) (string, error)
}

type MockBucket struct {
	Files map[string][]byte
	Err   error // si quieres simular fallo
}

func NewMockBucket() *MockBucket {
	return &MockBucket{Files: make(map[string][]byte)}
}

func (m *MockBucket) Put(ctx context.Context, key string, r io.Reader, size int64, opts PutOptions) (PutResult, error) {
	if m.Err != nil {
		return PutResult{}, m.Err
	}
	data, _ := io.ReadAll(r)
	m.Files[key] = data
	return PutResult{
		Key: key,
		// simula ETag
		ETag:      "mock-etag",
		VersionID: "v1",
	}, nil
}

func (m *MockBucket) Head(ctx context.Context, key string) (ObjectInfo, error) {
	if m.Err != nil {
		return ObjectInfo{}, m.Err
	}
	data, ok := m.Files[key]
	if !ok {
		return ObjectInfo{}, errors.New("not found")
	}
	return ObjectInfo{
		Key:         key,
		Size:        int64(len(data)),
		ContentType: "application/pdf",
		ETag:        "mock-etag",
		VersionID:   "v1",
	}, nil
}

func (m *MockBucket) Delete(ctx context.Context, key string) error {
	if m.Err != nil {
		return m.Err
	}
	delete(m.Files, key)
	return nil
}

func (m *MockBucket) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	// devuelve URL simulada
	return "https://mock-bucket.local/" + key, nil
}

func (m *MockBucket) PresignPut(ctx context.Context, key string, ttl time.Duration, opts PutOptions) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return "https://mock-bucket.local/upload/" + key, nil
}
