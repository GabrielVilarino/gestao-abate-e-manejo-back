package service

import (
	"errors"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type fazendaPortStub struct {
	create func(*domain.Fazenda) error
	getAll func(int) (*[]domain.Fazenda, error)
	update func(*domain.Fazenda) error
	delete func(int) error
}

func (s fazendaPortStub) CreateFazenda(fazenda *domain.Fazenda) error {
	return s.create(fazenda)
}
func (s fazendaPortStub) GetFazendas(idProprietario int) (*[]domain.Fazenda, error) {
	return s.getAll(idProprietario)
}
func (s fazendaPortStub) UpdateFazenda(fazenda *domain.Fazenda) error {
	return s.update(fazenda)
}
func (s fazendaPortStub) DeleteFazenda(id int) error { return s.delete(id) }

func TestFazendaServiceCreateFazenda(t *testing.T) {
	fazenda := &domain.Fazenda{ID: 1, Nome: "Fazenda Um"}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewFazendaService(fazendaPortStub{
			create: func(got *domain.Fazenda) error {
				if got != fazenda {
					t.Fatal("CreateFazenda recebeu outro ponteiro")
				}
				return expectedErr
			},
		})
		return service.CreateFazenda(fazenda)
	})
}

func TestFazendaServiceGetFazendas(t *testing.T) {
	fazendas := []domain.Fazenda{{ID: 1}, {ID: 2}}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewFazendaService(fazendaPortStub{
			getAll: func(idProprietario int) (*[]domain.Fazenda, error) {
				if idProprietario != 27 {
					t.Fatalf("id do proprietario = %d", idProprietario)
				}
				return &fazendas, expectedErr
			},
		})

		got, err := service.GetFazendas(27)
		if got != &fazendas {
			t.Fatalf("fazendas recebidas = %v", got)
		}
		return err
	})
}

func TestFazendaServiceUpdateFazenda(t *testing.T) {
	fazenda := &domain.Fazenda{ID: 1, Nome: "Fazenda Atualizada"}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewFazendaService(fazendaPortStub{
			update: func(got *domain.Fazenda) error {
				if got != fazenda {
					t.Fatal("UpdateFazenda recebeu outro ponteiro")
				}
				return expectedErr
			},
		})
		return service.UpdateFazenda(fazenda)
	})
}

func TestFazendaServiceDeleteFazenda(t *testing.T) {
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewFazendaService(fazendaPortStub{
			delete: func(id int) error {
				if id != 42 {
					t.Fatalf("id recebido = %d", id)
				}
				return expectedErr
			},
		})
		return service.DeleteFazenda(42)
	})
}

func testPassthroughErrors(t *testing.T, call func(*testing.T, error) error) {
	t.Helper()
	cases := []struct {
		name string
		err  error
	}{
		{name: "sucesso"},
		{name: "propaga erro", err: errors.New("erro do port")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := call(t, tc.err); !errors.Is(err, tc.err) {
				t.Fatalf("erro = %v, esperado %v", err, tc.err)
			}
		})
	}
}
