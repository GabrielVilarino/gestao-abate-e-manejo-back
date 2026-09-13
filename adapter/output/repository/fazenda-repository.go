package repository

import (
	"database/sql"
	"fmt"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type FazendaRepository struct {
	db *sql.DB
}

func NewFazendaRepository(db *sql.DB) FazendaRepository {
	return FazendaRepository{db: db}
}

func (f FazendaRepository) CreateFazenda(fazenda *domain.Fazenda) error {
	query := `INSERT INTO public.fazenda (nome, cidade, inscricao_rural, observacao, id_proprietario, ativo) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := f.db.Exec(
		query,
		fazenda.Nome,
		fazenda.Cidade,
		fazenda.InscricaoRural,
		fazenda.Observacao,
		fazenda.IDProprietario,
		fazenda.Ativo,
	)
	return err
}

func (f FazendaRepository) GetFazendas(idProprietario int) (*[]domain.Fazenda, error) {
	query := `SELECT id, nome, cidade, inscricao_rural, observacao, id_proprietario, ativo FROM public.fazenda WHERE id_proprietario = $1`
	rows, err := f.db.Query(query, idProprietario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fazendas := []domain.Fazenda{}
	for rows.Next() {
		fazenda := domain.Fazenda{}
		if err := rows.Scan(
			&fazenda.ID,
			&fazenda.Nome,
			&fazenda.Cidade,
			&fazenda.InscricaoRural,
			&fazenda.Observacao,
			&fazenda.IDProprietario,
			&fazenda.Ativo,
		); err != nil {
			return nil, err
		}
		fazendas = append(fazendas, fazenda)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &fazendas, nil
}

func (f FazendaRepository) UpdateFazenda(fazenda *domain.Fazenda) error {
	query := `UPDATE public.fazenda SET nome = $1, cidade = $2, inscricao_rural = $3, observacao = $4 WHERE id = $5`
	result, err := f.db.Exec(query, fazenda.Nome, fazenda.Cidade, fazenda.InscricaoRural, fazenda.Observacao, fazenda.ID)
	if err != nil {
		return err
	}
	return ensureDataFound(result, fazenda.ID)
}

func (f FazendaRepository) ActivateFazenda(id int) error {
	query := `
		WITH estado AS MATERIALIZED (
			SELECT f.id, p.ativo AS proprietario_ativo
			FROM public.fazenda f
			JOIN public.proprietario p ON p.id = f.id_proprietario
			WHERE f.id = $1
			FOR UPDATE OF p
		), fazenda_ativada AS (
			UPDATE public.fazenda
			SET ativo = true
			WHERE id IN (
				SELECT id FROM estado WHERE proprietario_ativo = true
			)
			RETURNING id
		)
		SELECT
			EXISTS (SELECT 1 FROM estado),
			COALESCE((SELECT proprietario_ativo FROM estado), false),
			EXISTS (SELECT 1 FROM fazenda_ativada)`

	var fazendaEncontrada bool
	var proprietarioAtivo bool
	var fazendaAtivada bool
	if err := f.db.QueryRow(query, id).Scan(&fazendaEncontrada, &proprietarioAtivo, &fazendaAtivada); err != nil {
		return err
	}
	if !fazendaEncontrada {
		return fmt.Errorf("fazenda com id %d não encontrada", id)
	}
	if !proprietarioAtivo {
		return domain.ErrProprietarioInativo
	}
	if !fazendaAtivada {
		return domain.ErrActivateFazenda
	}
	return nil
}

func (f FazendaRepository) DeactivateFazenda(id int) error {
	query := `UPDATE public.fazenda SET ativo = false WHERE id = $1`
	result, err := f.db.Exec(query, id)
	if err != nil {
		return err
	}
	return ensureDataFound(result, id)
}
