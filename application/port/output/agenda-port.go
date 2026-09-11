package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type AgendaPort interface {
	CreateAgenda(agenda *domain.Agenda) error
	FindAgenda(userID *int) (*[]domain.Agenda, error)
	UpdateAgenda(agenda *domain.Agenda) error
	DeleteAgenda(agendaID int) error
}
