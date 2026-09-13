package output

import (
	"context"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type AbatePort interface {
	CreateAbate(abate *domain.Abate) error
	FindAbateByID(id int) (*domain.Abate, error)
	FindAbates(filtro domain.FiltroAbate) ([]domain.Abate, error)
	UpdateDadosGeraisAbate(abateID int, dados domain.DadosGeraisAbate) error
	UpdateEtapaFazenda(abateID int, etapa domain.EtapaFazenda) error
	UpdateEtapaFrigorifico(abateID int, etapa domain.EtapaFrigorifico) error
	Delete(abateID int) ([]string, error)
	GetContextoFoto(abateID int) (*domain.ContextoFotoAbate, error)
	FindFotoByHash(abateID int, etapa, hash string) (*domain.FotoAbate, error)
	CreateFoto(foto *domain.FotoAbate) error
	FindFotoByID(fotoID int) (*domain.FotoAbate, error)
	DeleteFoto(fotoID int) (*domain.FotoAbate, error)
	CountFotosByObjectKey(objectKey string) (int, error)
	EnqueueObjectCleanup(objectKey string) error
	FindPendingObjectCleanup(limit int) ([]string, error)
	DeletePendingObjectCleanup(objectKey string) error
	MarkObjectCleanupFailed(objectKey, message string) error
	AcquireObjectLock(ctx context.Context, objectKey string) (func() error, error)
}
