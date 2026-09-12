package security

import (
	"fmt"
	"os"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTPort struct {
	secretKey  []byte
	expiration time.Duration
}

func NewTokenPort() *JWTPort {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required but not set")
	}

	return &JWTPort{
		secretKey:  []byte(secret),
		expiration: 72 * time.Hour,
	}
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int    `json:"user_id"`
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Role   int    `json:"role"`
	Ativo  bool   `json:"ativo"`
}

func (j *JWTPort) Generate(user domain.User) (string, error) {
	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "gestao_abate",
			Subject:   fmt.Sprintf("%s", user.Email),
		},
		UserID: user.ID,
		Nome:   user.Nome,
		Email:  user.Email,
		Role:   int(user.Role),
		Ativo:  user.Ativo,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTPort) Validate(tokenString string) (*domain.User, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	user := &domain.User{
		ID:    claims.UserID,
		Nome:  claims.Nome,
		Email: claims.Email,
		Role:  claims.Role,
		Ativo: claims.Ativo,
	}

	return user, nil
}
