package service

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type ProprietarioService struct {
	ProprietarioPort output.ProprietarioPort
}

func NewProprietarioService(
	proprietarioPort output.ProprietarioPort,
) *ProprietarioService {
	return &ProprietarioService{
		ProprietarioPort: proprietarioPort,
	}
}

func (p *ProprietarioService) CreateProprietario(proprietario *domain.Proprietario) error {
	proprietarioData, err := p.ProprietarioPort.GetProprietarioByCPF(proprietario.CPF)
	if err != nil {
		return err
	}

	if proprietarioData != nil {
		return domain.ErrCpfAlreadyExists
	}

	return p.ProprietarioPort.CreateProprietario(proprietario)
}

func (p *ProprietarioService) GetProprietarios() (*[]domain.Proprietario, error) {
	return p.ProprietarioPort.GetProprietarios()
}

func (p *ProprietarioService) UpdateProprietario(proprietario *domain.Proprietario) error {
	return p.ProprietarioPort.UpdateProprietario(proprietario)
}

func (p *ProprietarioService) DeleteProprietario(id int) error {
	return p.ProprietarioPort.DeleteProprietario(id)
}
