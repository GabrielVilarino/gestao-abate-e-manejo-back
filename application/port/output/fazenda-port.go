package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type FazendaPort interface {
	CreateFazenda(fazenda *domain.Fazenda) error
	GetFazendas(idProprietario int) (*[]domain.Fazenda, error)
	UpdateFazenda(fazenda *domain.Fazenda) error
	DeleteFazenda(id string) error
}
