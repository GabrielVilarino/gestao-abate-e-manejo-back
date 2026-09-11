package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type AgendaUseCase interface {
	CreateAgenda(agenda *domain.Agenda) error
	FindAgenda(userID *int) (*[]domain.Agenda, error)
	UpdateAgenda(agenda *domain.Agenda) error
	DeleteAgenda(agendaID int) error
}
