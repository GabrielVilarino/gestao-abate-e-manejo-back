package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type FazendaUseCase interface {
	CreateFazenda(fazenda *domain.Fazenda) error
	GetFazendas(idProprietario int) (*[]domain.Fazenda, error)
	UpdateFazenda(fazenda *domain.Fazenda) error
	ActivateFazenda(id int) error
	DeactivateFazenda(id int) error
}
