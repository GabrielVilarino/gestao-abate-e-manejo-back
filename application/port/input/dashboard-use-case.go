package input

import (
	"context"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type DashboardUseCase interface {
	GetDashboardAbates(ctx context.Context, filtro domain.FiltroDashboardAbates) (*domain.DashboardAbates, error)
}
