package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type FazendaUseCase interface {
	CreateFazenda(fazenda *domain.Fazenda) error
	GetFazendas(idProprietario int) (*[]domain.Fazenda, error)
	UpdateFazenda(fazenda *domain.Fazenda) error
	DeleteFazenda(id int) error
}
