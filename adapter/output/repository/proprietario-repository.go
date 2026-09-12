package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type ProprietarioRepository struct {
	db *sql.DB
}

func NewProprietarioRepository(db *sql.DB) ProprietarioRepository {
	return ProprietarioRepository{db: db}
}

func (p ProprietarioRepository) CreateProprietario(proprietario *domain.Proprietario) error {
	query := `INSERT INTO public.proprietario (nome, cpf, observacao, ativo) VALUES ($1, $2, $3, $4)`
	_, err := p.db.Exec(query, proprietario.Nome, proprietario.CPF, proprietario.Observacao, proprietario.Ativo)
	return err
}

func (p ProprietarioRepository) GetProprietarios() (*[]domain.Proprietario, error) {
	query := `SELECT id, nome, cpf, observacao, ativo FROM public.proprietario`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proprietarios := []domain.Proprietario{}
	for rows.Next() {
		proprietario := domain.Proprietario{}
		if err := rows.Scan(
			&proprietario.ID,
			&proprietario.Nome,
			&proprietario.CPF,
			&proprietario.Observacao,
			&proprietario.Ativo,
		); err != nil {
			return nil, err
		}
		proprietarios = append(proprietarios, proprietario)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &proprietarios, nil
}

func (p ProprietarioRepository) GetProprietarioByCPF(cpf string) (*domain.Proprietario, error) {
	query := `SELECT id, nome, cpf, observacao, ativo FROM public.proprietario WHERE cpf = $1`
	proprietario := &domain.Proprietario{}
	err := p.db.QueryRow(query, cpf).Scan(
		&proprietario.ID,
		&proprietario.Nome,
		&proprietario.CPF,
		&proprietario.Observacao,
		&proprietario.Ativo,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return proprietario, nil
}

func (p ProprietarioRepository) UpdateProprietario(proprietario *domain.Proprietario) error {
	query := `UPDATE public.proprietario SET nome = $1, cpf = $2, observacao = $3 WHERE id = $4`
	result, err := p.db.Exec(query, proprietario.Nome, proprietario.CPF, proprietario.Observacao, proprietario.ID)
	if err != nil {
		return err
	}
	return ensureDataFound(result, proprietario.ID)
}

func (p ProprietarioRepository) ActivateProprietario(id int) error {
	query := `UPDATE public.proprietario SET ativo = true WHERE id = $1`
	result, err := p.db.Exec(query, id)
	if err != nil {
		return err
	}
	return ensureDataFound(result, id)
}

func (p ProprietarioRepository) DeactivateProprietario(id int) error {
	query := `
		WITH proprietario_desativado AS (
			UPDATE public.proprietario
			SET ativo = false
			WHERE id = $1
			RETURNING id
		), fazendas_desativadas AS (
			UPDATE public.fazenda
			SET ativo = false
			WHERE id_proprietario IN (SELECT id FROM proprietario_desativado)
			RETURNING id
		)
		SELECT
			EXISTS (SELECT 1 FROM proprietario_desativado),
			(SELECT COUNT(*) FROM fazendas_desativadas)`

	var proprietarioEncontrado bool
	var fazendasDesativadas int
	if err := p.db.QueryRow(query, id).Scan(&proprietarioEncontrado, &fazendasDesativadas); err != nil {
		return err
	}
	if !proprietarioEncontrado {
		return fmt.Errorf("proprietário com id %d não encontrado", id)
	}
	return nil
}
