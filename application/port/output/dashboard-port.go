package output

import (
	"context"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type DashboardPort interface {
	GetDashboardAbates(ctx context.Context, filtro domain.FiltroDashboardAbates) (*domain.DashboardAbates, error)
}
