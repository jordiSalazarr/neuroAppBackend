package encryption

import (
	"crypto/sha256"
	"encoding/base64"

	"neuro.app.jordi/internal/auth/domain"
)

type EncryptionServiceImpl struct{}

func (e *EncryptionServiceImpl) Hash(plain string) string {
	// Genera un hash SHA-256 y lo codifica en base64
	hash := sha256.Sum256([]byte(plain))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (e *EncryptionServiceImpl) Compare(plain, hashed string) bool {
	// Compara el hash generado con el hash proporcionado
	return e.Hash(plain) == hashed
}

func (e *EncryptionServiceImpl) GenerateTokenFromID(userID string) string {
	// Genera un token simple codificando el userID en base64
	return base64.StdEncoding.EncodeToString([]byte(userID))
}

// NewEncryptionService crea una nueva instancia de EncryptionService
func NewEncryptionService() domain.EncryptionService {
	return &EncryptionServiceImpl{}
}
