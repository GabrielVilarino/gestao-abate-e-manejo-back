package service

import (
	"context"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type DashboardService struct {
	DashboardPort output.DashboardPort
}

func NewDashboardService(dashboardPort output.DashboardPort) *DashboardService {
	return &DashboardService{DashboardPort: dashboardPort}
}

func (d *DashboardService) GetDashboardAbates(ctx context.Context, filtro domain.FiltroDashboardAbates) (*domain.DashboardAbates, error) {
	if filtro.DataInicio != nil && filtro.DataFim != nil && filtro.DataInicio.After(*filtro.DataFim) {
		return nil, domain.ErrPeriodoInvalido
	}
	return d.DashboardPort.GetDashboardAbates(ctx, filtro)
}
