package service

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type AgendaService struct {
	AgendaPort output.AgendaPort
}

func NewAgendaService(
	agendaPort output.AgendaPort,
) *AgendaService {
	return &AgendaService{
		AgendaPort: agendaPort,
	}
}

func (a *AgendaService) CreateAgenda(agenda *domain.Agenda) error {
	return a.AgendaPort.CreateAgenda(agenda)
}

func (a *AgendaService) FindAgenda(userID *int) (*[]domain.Agenda, error) {
	return a.AgendaPort.FindAgenda(userID)
}

func (a *AgendaService) UpdateAgenda(agenda *domain.Agenda) error {
	return a.AgendaPort.UpdateAgenda(agenda)
}

func (a *AgendaService) DeleteAgenda(agendaID int) error {
	return a.AgendaPort.DeleteAgenda(agendaID)
}
