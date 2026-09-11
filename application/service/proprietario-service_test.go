package service

import (
	"errors"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type proprietarioPortStub struct {
	create   func(*domain.Proprietario) error
	getAll   func() (*[]domain.Proprietario, error)
	getByCPF func(string) (*domain.Proprietario, error)
	update   func(*domain.Proprietario) error
	delete   func(int) error
}

func (s proprietarioPortStub) CreateProprietario(proprietario *domain.Proprietario) error {
	return s.create(proprietario)
}
func (s proprietarioPortStub) GetProprietarios() (*[]domain.Proprietario, error) {
	return s.getAll()
}
func (s proprietarioPortStub) GetProprietarioByCPF(cpf string) (*domain.Proprietario, error) {
	return s.getByCPF(cpf)
}
func (s proprietarioPortStub) UpdateProprietario(proprietario *domain.Proprietario) error {
	return s.update(proprietario)
}
func (s proprietarioPortStub) DeleteProprietario(id int) error { return s.delete(id) }

func TestProprietarioServiceCreateProprietario(t *testing.T) {
	t.Run("retorna erro da consulta", func(t *testing.T) {
		expectedErr := errors.New("erro de consulta")
		service := NewProprietarioService(proprietarioPortStub{
			getByCPF: func(cpf string) (*domain.Proprietario, error) {
				if cpf != "12345678900" {
					t.Fatalf("cpf recebido = %q", cpf)
				}
				return nil, expectedErr
			},
		})

		err := service.CreateProprietario(&domain.Proprietario{CPF: "12345678900"})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})

	t.Run("rejeita cpf existente", func(t *testing.T) {
		existing := domain.Proprietario{ID: 1}
		service := NewProprietarioService(proprietarioPortStub{
			getByCPF: func(string) (*domain.Proprietario, error) { return &existing, nil },
		})

		err := service.CreateProprietario(&domain.Proprietario{CPF: "12345678900"})
		if !errors.Is(err, domain.ErrCpfAlreadyExists) {
			t.Fatalf("erro = %v, esperado %v", err, domain.ErrCpfAlreadyExists)
		}
	})

	t.Run("cria proprietario", func(t *testing.T) {
		proprietario := &domain.Proprietario{Nome: "Fulano", CPF: "12345678900"}
		service := NewProprietarioService(proprietarioPortStub{
			getByCPF: func(cpf string) (*domain.Proprietario, error) {
				if cpf != proprietario.CPF {
					t.Fatalf("cpf recebido = %q", cpf)
				}
				return nil, nil
			},
			create: func(got *domain.Proprietario) error {
				if got != proprietario {
					t.Fatal("CreateProprietario recebeu outro ponteiro")
				}
				return nil
			},
		})

		if err := service.CreateProprietario(proprietario); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
	})

	t.Run("propaga erro ao criar", func(t *testing.T) {
		expectedErr := errors.New("erro ao criar")
		service := NewProprietarioService(proprietarioPortStub{
			getByCPF: func(string) (*domain.Proprietario, error) { return nil, nil },
			create:   func(*domain.Proprietario) error { return expectedErr },
		})

		err := service.CreateProprietario(&domain.Proprietario{CPF: "12345678900"})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})
}

func TestProprietarioServiceCRUD(t *testing.T) {
	proprietarios := []domain.Proprietario{{ID: 1, Nome: "Fulano"}}
	proprietario := &domain.Proprietario{ID: 2, Nome: "Atualizado"}

	t.Run("lista proprietarios", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewProprietarioService(proprietarioPortStub{
				getAll: func() (*[]domain.Proprietario, error) { return &proprietarios, expectedErr },
			})

			got, err := service.GetProprietarios()
			if got != &proprietarios {
				t.Fatalf("resultado = %v", got)
			}
			return err
		})
	})

	t.Run("atualiza proprietario", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewProprietarioService(proprietarioPortStub{
				update: func(got *domain.Proprietario) error {
					if got != proprietario {
						t.Fatal("UpdateProprietario recebeu outro ponteiro")
					}
					return expectedErr
				},
			})
			return service.UpdateProprietario(proprietario)
		})
	})

	t.Run("exclui proprietario", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewProprietarioService(proprietarioPortStub{
				delete: func(id int) error {
					if id != 42 {
						t.Fatalf("id recebido = %d", id)
					}
					return expectedErr
				},
			})
			return service.DeleteProprietario(42)
		})
	})
}
