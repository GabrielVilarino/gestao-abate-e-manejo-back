package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) DashboardRepository {
	return DashboardRepository{db: db}
}

func (d DashboardRepository) GetDashboardAbates(ctx context.Context, filtro domain.FiltroDashboardAbates) (*domain.DashboardAbates, error) {
	where, args := dashboardFilters(filtro)
	dashboard := &domain.DashboardAbates{}
	if err := d.db.QueryRowContext(ctx, `
		SELECT COUNT(*)::int,
			COALESCE(SUM(COALESCE(a.quantidade_total_cabecas, 0)), 0)::int,
			COALESCE(AVG(COALESCE(a.quantidade_total_cabecas, 0)), 0),
			COALESCE(AVG(COALESCE(a.peso_total_fazenda, 0)), 0),
			COALESCE(AVG(COALESCE(a.peso_total_frigorifico, 0)), 0),
			COALESCE(AVG(CASE WHEN a.peso_total_fazenda > 0 THEN a.peso_total_frigorifico / a.peso_total_fazenda * 100 END), 0),
			COALESCE(AVG(COALESCE(a.distancia_frigorifico, 0)), 0)
		FROM public.abate a
		JOIN public.fazenda f ON f.id = a.fazenda_id
		JOIN public.proprietario p ON p.id = f.id_proprietario`+where,
		args...,
	).Scan(
		&dashboard.Resumo.TotalAbates, &dashboard.Resumo.TotalCabecas,
		&dashboard.Resumo.MediaCabecasPorAbate, &dashboard.Resumo.PesoMedioFazenda,
		&dashboard.Resumo.PesoMedioFrigorifico, &dashboard.Resumo.RendimentoMedioCarcaca,
		&dashboard.Resumo.DistanciaMediaFrigorifico,
	); err != nil {
		return nil, err
	}

	var err error
	dashboard.PorFazenda, err = d.dashboardPorFazenda(ctx, where, args)
	if err != nil {
		return nil, err
	}
	dashboard.Evolucao, err = d.dashboardEvolucao(ctx, where, args)
	if err != nil {
		return nil, err
	}
	dashboard.PorCategoriaAnimal, err = d.dashboardPorCategoria(ctx, where, args)
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}

func (d DashboardRepository) dashboardPorFazenda(ctx context.Context, where string, args []any) ([]domain.DashboardPorFazenda, error) {
	query := `
		SELECT f.id, f.nome, COUNT(*)::int,
			COALESCE(SUM(COALESCE(a.quantidade_total_cabecas, 0)), 0)::int,
			COALESCE(AVG(COALESCE(a.quantidade_total_cabecas, 0)), 0)
		FROM public.abate a
		JOIN public.fazenda f ON f.id = a.fazenda_id
		JOIN public.proprietario p ON p.id = f.id_proprietario` + where + `
		GROUP BY f.id, f.nome
		ORDER BY f.nome, f.id`
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.DashboardPorFazenda, 0)
	for rows.Next() {
		item := domain.DashboardPorFazenda{}
		if err := rows.Scan(&item.FazendaID, &item.NomeFazenda, &item.TotalAbates, &item.TotalCabecas, &item.MediaCabecasPorAbate); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (d DashboardRepository) dashboardEvolucao(ctx context.Context, where string, args []any) ([]domain.DashboardEvolucao, error) {
	query := `
		SELECT a.data_abate, COUNT(*)::int,
			COALESCE(SUM(COALESCE(a.quantidade_total_cabecas, 0)), 0)::int
		FROM public.abate a
		JOIN public.fazenda f ON f.id = a.fazenda_id
		JOIN public.proprietario p ON p.id = f.id_proprietario` + where + `
		GROUP BY a.data_abate
		ORDER BY a.data_abate`
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.DashboardEvolucao, 0)
	for rows.Next() {
		item := domain.DashboardEvolucao{}
		if err := rows.Scan(&item.Data, &item.TotalAbates, &item.TotalCabecas); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (d DashboardRepository) dashboardPorCategoria(ctx context.Context, where string, args []any) ([]domain.DashboardPorCategoria, error) {
	query := `
		SELECT a.categoria_animal, COUNT(*)::int,
			COALESCE(SUM(COALESCE(a.quantidade_total_cabecas, 0)), 0)::int
		FROM public.abate a
		JOIN public.fazenda f ON f.id = a.fazenda_id
		JOIN public.proprietario p ON p.id = f.id_proprietario` + where + `
		GROUP BY a.categoria_animal
		ORDER BY a.categoria_animal`
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.DashboardPorCategoria, 0)
	for rows.Next() {
		item := domain.DashboardPorCategoria{}
		if err := rows.Scan(&item.CategoriaAnimal, &item.TotalAbates, &item.TotalCabecas); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func dashboardFilters(filtro domain.FiltroDashboardAbates) (string, []any) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filtro.ProprietarioID != nil {
		add("p.id = $%d", *filtro.ProprietarioID)
	}
	if filtro.FazendaID != nil {
		add("f.id = $%d", *filtro.FazendaID)
	}
	if filtro.DataInicio != nil {
		add("a.data_abate >= $%d", *filtro.DataInicio)
	}
	if filtro.DataFim != nil {
		add("a.data_abate < $%d", filtro.DataFim.AddDate(0, 0, 1))
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
