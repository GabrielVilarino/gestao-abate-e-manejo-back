package service

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type FazendaService struct {
	FazendaPort output.FazendaPort
}

func NewFazendaService(
	fazendaPort output.FazendaPort,
) *FazendaService {
	return &FazendaService{
		FazendaPort: fazendaPort,
	}
}

func (f *FazendaService) CreateFazenda(fazenda *domain.Fazenda) error {
	return f.FazendaPort.CreateFazenda(fazenda)
}

func (f *FazendaService) GetFazendas(idProprietario int) (*[]domain.Fazenda, error) {
	return f.FazendaPort.GetFazendas(idProprietario)
}

func (f *FazendaService) UpdateFazenda(fazenda *domain.Fazenda) error {
	return f.FazendaPort.UpdateFazenda(fazenda)
}

func (f *FazendaService) DeleteFazenda(id string) error {
	return f.FazendaPort.DeleteFazenda(id)
}
