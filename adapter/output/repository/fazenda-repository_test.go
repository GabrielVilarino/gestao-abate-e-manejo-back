package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

func TestFazendaRepositoryActivateIsConditionalOnOwnerStatus(t *testing.T) {
	tests := []struct {
		name          string
		farmExists    bool
		ownerActive   bool
		farmActivated bool
		wantError     error
		wantAnyError  bool
	}{
		{name: "ativa quando proprietario ativo", farmExists: true, ownerActive: true, farmActivated: true},
		{name: "bloqueia quando proprietario inativo", farmExists: true, ownerActive: false, wantError: domain.ErrProprietarioInativo},
		{name: "fazenda inexistente", wantAnyError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			conditionalQuery := `(?s)WITH estado AS MATERIALIZED \(.*FOR UPDATE OF p.*fazenda_ativada AS \(.*UPDATE public\.fazenda.*proprietario_ativo = true.*SELECT.*EXISTS`
			mock.ExpectQuery(conditionalQuery).
				WithArgs(11).
				WillReturnRows(sqlmock.NewRows([]string{"fazenda_encontrada", "proprietario_ativo", "fazenda_ativada"}).
					AddRow(tc.farmExists, tc.ownerActive, tc.farmActivated))

			err = NewFazendaRepository(db).ActivateFazenda(11)
			if tc.wantError != nil && !errors.Is(err, tc.wantError) {
				t.Fatalf("erro = %v, esperado %v", err, tc.wantError)
			}
			if tc.wantError == nil && tc.wantAnyError != (err != nil) {
				t.Fatalf("erro = %v, wantAnyError = %v", err, tc.wantAnyError)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFazendaRepositoryUpdateDoesNotChangeStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	fazenda := &domain.Fazenda{
		ID: 11, Nome: "Boa Vista", Cidade: "Gurupi", InscricaoRural: "IR-1", Observacao: "", Ativo: true,
	}
	query := `UPDATE public\.fazenda SET nome = \$1, cidade = \$2, inscricao_rural = \$3, observacao = \$4 WHERE id = \$5`
	mock.ExpectExec(query).
		WithArgs(fazenda.Nome, fazenda.Cidade, fazenda.InscricaoRural, fazenda.Observacao, fazenda.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewFazendaRepository(db).UpdateFazenda(fazenda); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
