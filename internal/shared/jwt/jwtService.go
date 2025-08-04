package jwtService

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	TokenDuration time.Duration
	Secret        string
}

func New() *Service {
	duration := time.Hour * 2
	secretKey := os.Getenv("JWT_SECRET")
	return &Service{
		TokenDuration: duration,
		Secret:        secretKey,
	}
}

type Claims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

func (s *Service) GenerateToken(id string) (string, error) {
	claims := &Claims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Secret))
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Secret), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
