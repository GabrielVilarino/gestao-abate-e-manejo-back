package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type ProprietarioUseCase interface {
	CreateProprietario(proprietario *domain.Proprietario) error
	GetProprietarios() (*[]domain.Proprietario, error)
	UpdateProprietario(proprietario *domain.Proprietario) error
	DeleteProprietario(id string) error
}
