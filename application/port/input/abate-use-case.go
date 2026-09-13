package input

import (
	"context"
	"io"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type AbateUseCase interface {
	CreateAbate(abate *domain.Abate) error
	FindAbateByID(id int) (*domain.Abate, error)
	FindAbates(filtro domain.FiltroAbate) ([]domain.Abate, error)
	UpdateDadosGeraisAbate(abateID int, dados domain.DadosGeraisAbate) error
	UpdateEtapaFazenda(abateID int, etapa domain.EtapaFazenda) error
	UpdateEtapaFrigorifico(abateID int, etapa domain.EtapaFrigorifico) error
	Delete(ctx context.Context, abateID int) error
	UploadFotoAbate(ctx context.Context, abateID int, etapa, nomeOriginal string, data []byte) (*domain.FotoAbate, error)
	DownloadFotoAbate(ctx context.Context, fotoID int) (*domain.FotoAbate, io.ReadCloser, error)
	DeleteFotoAbate(ctx context.Context, fotoID int) error
}
