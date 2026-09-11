package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type AbatePort interface {
	Create(abate *domain.Abate) error
	FindByFazendaID(id int) (*domain.Abate, error)
	UpdateDadosGeraisAbate(
		abateID int,
		dados domain.DadosGeraisAbate,
	) error
	UpdateEtapaFazenda(
		abateID int,
		etapa domain.EtapaFazenda,
	) error
	UpdateEtapaFrigorifico(
		abateID int,
		etapa domain.EtapaFrigorifico,
	) error
	Delete(abateID int) error
}
