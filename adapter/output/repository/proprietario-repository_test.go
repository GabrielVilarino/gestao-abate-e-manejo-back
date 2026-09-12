package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProprietarioRepositoryDeactivateUsesAtomicCascade(t *testing.T) {
	tests := []struct {
		name      string
		exists    bool
		farmCount int
		wantErr   bool
	}{
		{name: "desativa com fazendas", exists: true, farmCount: 2},
		{name: "desativa sem fazendas", exists: true, farmCount: 0},
		{name: "proprietario inexistente", exists: false, farmCount: 0, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			atomicQuery := `(?s)WITH proprietario_desativado AS \(.*UPDATE public\.proprietario.*fazendas_desativadas AS \(.*UPDATE public\.fazenda.*SELECT.*EXISTS`
			mock.ExpectQuery(atomicQuery).
				WithArgs(7).
				WillReturnRows(sqlmock.NewRows([]string{"proprietario_encontrado", "fazendas_desativadas"}).AddRow(tc.exists, tc.farmCount))

			err = NewProprietarioRepository(db).DeactivateProprietario(7)
			if (err != nil) != tc.wantErr {
				t.Fatalf("erro = %v, wantErr = %v", err, tc.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProprietarioRepositoryActivateDoesNotActivateFarms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := regexp.QuoteMeta(`UPDATE public.proprietario SET ativo = true WHERE id = $1`)
	mock.ExpectExec(query).WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := NewProprietarioRepository(db).ActivateProprietario(7); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProprietarioRepositoryDeactivatePropagatesStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectedErr := errors.New("falha atômica")
	mock.ExpectQuery(`(?s)WITH proprietario_desativado AS`).WithArgs(7).WillReturnError(expectedErr)
	err = NewProprietarioRepository(db).DeactivateProprietario(7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("erro = %v, esperado %v", err, expectedErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
