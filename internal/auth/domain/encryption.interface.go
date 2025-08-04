package domain

type EncryptionService interface {
	Hash(plain string) string
	Compare(plain, hashed string) bool
	GenerateTokenFromID(userID string) string
}
