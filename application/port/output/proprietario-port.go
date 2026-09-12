package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type ProprietarioPort interface {
	CreateProprietario(proprietario *domain.Proprietario) error
	GetProprietarios() (*[]domain.Proprietario, error)
	GetProprietarioByCPF(cpf string) (*domain.Proprietario, error)
	UpdateProprietario(proprietario *domain.Proprietario) error
	ActivateProprietario(id int) error
	DeactivateProprietario(id int) error
}
