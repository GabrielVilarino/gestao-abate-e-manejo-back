package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type AgendaUseCase interface {
	CreateAgenda(agenda *domain.Agenda) error
	GetAgenda(agendaID, userID int) (*domain.Agenda, error)
	FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error)
	UpdateAgenda(agenda *domain.Agenda) error
	DeleteAgenda(agendaID, userID int) error
	CreatePushSubscription(subscription *domain.PushSubscription) error
	DeletePushSubscription(subscriptionID, userID int) error
}
