package security

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPortHashAndCompare(t *testing.T) {
	port := NewHashPort()
	hash, err := port.Hash("senha-segura")
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}
	if hash == "senha-segura" || !strings.HasPrefix(hash, "$2") {
		t.Fatalf("hash inesperado: %q", hash)
	}
	if err := port.Compare("senha-segura", hash); err != nil {
		t.Fatalf("senha correta rejeitada: %v", err)
	}
	if err := port.Compare("senha-incorreta", hash); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("erro = %v, esperado %v", err, bcrypt.ErrMismatchedHashAndPassword)
	}
}

func TestHashPortErrors(t *testing.T) {
	port := NewHashPort()

	t.Run("senha longa demais", func(t *testing.T) {
		hash, err := port.Hash(strings.Repeat("a", 73))
		if hash != "" || !errors.Is(err, bcrypt.ErrPasswordTooLong) {
			t.Fatalf("hash = %q, erro = %v", hash, err)
		}
	})

	t.Run("hash invalido", func(t *testing.T) {
		if err := port.Compare("senha", "hash-invalido"); err == nil {
			t.Fatal("hash inválido deveria retornar erro")
		}
	})
}
