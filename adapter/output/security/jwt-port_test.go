package security

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTPortGenerateAndValidate(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")
	port := NewTokenPort()
	user := domain.User{ID: 7, Nome: "Maria", Email: "maria@example.com", Role: domain.RoleUser, Ativo: true}
	before := time.Now()

	tokenString, err := port.Generate(user)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte("segredo-de-teste"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("token inválido: %v", err)
	}
	if claims.Issuer != "gestao_abate" || claims.Subject != user.Email || claims.UserID != user.ID || claims.Nome != user.Nome || claims.Email != user.Email || claims.Role != user.Role || claims.Ativo != user.Ativo {
		t.Fatalf("claims inesperadas: %+v", claims)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil || claims.IssuedAt.Time.Before(before.Add(-time.Second)) || claims.IssuedAt.Time.After(time.Now().Add(time.Second)) {
		t.Fatalf("issued_at inesperado: %v", claims.IssuedAt)
	}
	expectedExpiration := claims.IssuedAt.Time.Add(72 * time.Hour)
	if difference := claims.ExpiresAt.Time.Sub(expectedExpiration); difference < -time.Second || difference > time.Second {
		t.Fatalf("expiração = %v, esperada próxima de %v", claims.ExpiresAt.Time, expectedExpiration)
	}

	validatedUser, err := port.Validate(tokenString)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}
	if *validatedUser != user {
		t.Fatalf("usuário validado = %+v, esperado %+v", validatedUser, user)
	}
}

func TestJWTPortRejectsInvalidTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-original")
	port := NewTokenPort()
	user := domain.User{ID: 7, Email: "user@example.com", Ativo: true}

	t.Run("token adulterado", func(t *testing.T) {
		tokenString, err := port.Generate(user)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(tokenString, ".")
		if len(parts) != 3 {
			t.Fatalf("token inesperado: %q", tokenString)
		}
		if parts[2][0] == 'a' {
			parts[2] = "b" + parts[2][1:]
		} else {
			parts[2] = "a" + parts[2][1:]
		}
		validatedUser, err := port.Validate(strings.Join(parts, "."))
		if validatedUser != nil || err == nil || !strings.Contains(err.Error(), "token inválido") {
			t.Fatalf("usuário = %+v, erro = %v", validatedUser, err)
		}
	})

	t.Run("token expirado", func(t *testing.T) {
		expiredPort := &JWTPort{secretKey: []byte("segredo-original"), expiration: -time.Hour}
		tokenString, err := expiredPort.Generate(user)
		if err != nil {
			t.Fatal(err)
		}
		validatedUser, err := port.Validate(tokenString)
		if validatedUser != nil || !errors.Is(err, jwt.ErrTokenExpired) {
			t.Fatalf("usuário = %+v, erro = %v", validatedUser, err)
		}
	})

	t.Run("segredo diferente", func(t *testing.T) {
		tokenString, err := port.Generate(user)
		if err != nil {
			t.Fatal(err)
		}
		otherPort := &JWTPort{secretKey: []byte("outro-segredo"), expiration: 72 * time.Hour}
		validatedUser, err := otherPort.Validate(tokenString)
		if validatedUser != nil || err == nil || !strings.Contains(err.Error(), "token inválido") {
			t.Fatalf("usuário = %+v, erro = %v", validatedUser, err)
		}
	})
}

func TestNewTokenPortWithoutSecretPanics(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	defer func() {
		panicValue := recover()
		if panicValue == nil || !strings.Contains(panicValue.(string), "JWT_SECRET") {
			t.Fatalf("panic inesperado: %v", panicValue)
		}
	}()
	NewTokenPort()
}
