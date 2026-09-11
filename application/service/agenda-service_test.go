package service

import (
	"testing"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type agendaPortStub struct {
	create func(*domain.Agenda) error
	find   func(*int) (*[]domain.Agenda, error)
	update func(*domain.Agenda) error
	delete func(int) error
}

func (s agendaPortStub) CreateAgenda(agenda *domain.Agenda) error { return s.create(agenda) }
func (s agendaPortStub) FindAgenda(userID *int) (*[]domain.Agenda, error) {
	return s.find(userID)
}
func (s agendaPortStub) UpdateAgenda(agenda *domain.Agenda) error { return s.update(agenda) }
func (s agendaPortStub) DeleteAgenda(agendaID int) error          { return s.delete(agendaID) }

func TestAgendaServiceCreateAgenda(t *testing.T) {
	agenda := &domain.Agenda{ID: 1, DataHora: time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAgendaService(agendaPortStub{
			create: func(got *domain.Agenda) error {
				if got != agenda {
					t.Fatal("CreateAgenda recebeu outro ponteiro")
				}
				return expectedErr
			},
		})
		return service.CreateAgenda(agenda)
	})
}

func TestAgendaServiceFindAgenda(t *testing.T) {
	userID := 27
	agendas := []domain.Agenda{{ID: 1}, {ID: 2}}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAgendaService(agendaPortStub{
			find: func(gotUserID *int) (*[]domain.Agenda, error) {
				if gotUserID != &userID {
					t.Fatal("FindAgenda recebeu outro ponteiro de userID")
				}
				return &agendas, expectedErr
			},
		})

		got, err := service.FindAgenda(&userID)
		if got != &agendas {
			t.Fatalf("agendas recebidas = %v", got)
		}
		return err
	})
}

func TestAgendaServiceUpdateAgenda(t *testing.T) {
	agenda := &domain.Agenda{ID: 1, Observacao: "Atualizada"}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAgendaService(agendaPortStub{
			update: func(got *domain.Agenda) error {
				if got != agenda {
					t.Fatal("UpdateAgenda recebeu outro ponteiro")
				}
				return expectedErr
			},
		})
		return service.UpdateAgenda(agenda)
	})
}

func TestAgendaServiceDeleteAgenda(t *testing.T) {
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAgendaService(agendaPortStub{
			delete: func(id int) error {
				if id != 42 {
					t.Fatalf("id recebido = %d", id)
				}
				return expectedErr
			},
		})
		return service.DeleteAgenda(42)
	})
}
