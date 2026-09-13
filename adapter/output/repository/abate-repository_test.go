package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

func TestAbateRepositoryCreatePersistsRequiredFieldsAndRollsBackOnChildError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	date := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	abate := &domain.Abate{
		DadosGeraisAbate: domain.DadosGeraisAbate{
			DataAbate: date, FazendaID: 2, NumeroLote: 41,
			NomeFrigorifico: "Frigo", DistanciaFrigorifico: 87.5,
			CategoriaAnimal: "Bovino", PrecoFunrural: 10, PrecoSemFunrural: 11,
		},
		EtapaFazenda: domain.EtapaFazenda{
			PesoTotal: 100, QuantidadeAnimal: []domain.QtdDenticao{{QtdDenticao: 2, QtdAnimais: 5}},
		},
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO public\.abate`).
		WithArgs(2, 41, date, "Frigo", 87.5, "Bovino", 10.0, 11.0, 100.0, 0.0, 0.0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	childErr := errors.New("falha no detalhe")
	mock.ExpectExec(`INSERT INTO public\.abate_denticao`).
		WithArgs(9, 2, 5).
		WillReturnError(childErr)
	mock.ExpectRollback()

	err = NewAbateRepository(db).CreateAbate(abate)
	if !errors.Is(err, childErr) {
		t.Fatalf("erro=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAbateRepositoryFindUsesAllOptionalFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	owner, farm, lot := 1, 2, 3
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	query := `(?s)FROM public\.abate a.*WHERE p\.id = \$1 AND f\.id = \$2 AND a\.numero_lote = \$3 AND a\.data_abate >= \$4 AND a\.data_abate < \$5.*ORDER BY.*LIMIT \$6 OFFSET \$7`
	mock.ExpectQuery(query).
		WithArgs(owner, farm, lot, start, end.AddDate(0, 0, 1), 25, 50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "proprietario_id", "proprietario", "fazenda", "fazenda_id", "numero_lote",
			"data_abate", "frigorifico", "distancia", "categoria", "funrural", "sem_funrural",
			"peso_fazenda", "peso_frigorifico", "balancao",
		}))

	got, err := NewAbateRepository(db).FindAbates(domain.FiltroAbate{
		ProprietarioID: &owner, FazendaID: &farm, NumeroLote: &lot,
		DataInicio: &start, DataFim: &end,
		Limit: 25, Offset: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("abates=%v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAbateRepositoryCreateFotoPersistsPrivateObjectMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	foto := &domain.FotoAbate{
		AbateID: 4, Etapa: "FAZENDA", ObjectKey: "Dono/7/FAZENDA/hash",
		NomeOriginal: "foto.png", ContentType: "image/png", Tamanho: 20, SHA256: "hash",
	}
	mock.ExpectQuery(`(?s)INSERT INTO public\.abate_fotos.*ON CONFLICT.*RETURNING id`).
		WithArgs(4, "FAZENDA", foto.ObjectKey, "foto.png", "image/png", int64(20), "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(13))
	if err := NewAbateRepository(db).CreateFoto(foto); err != nil {
		t.Fatal(err)
	}
	if foto.ID != 13 {
		t.Fatalf("id=%d", foto.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAbateRepositoryDeleteFotoQueuesDurableCleanupInSameTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, abate_id, etapa.*FROM public\.abate_fotos WHERE id=\$1 FOR UPDATE`).
		WithArgs(13).
		WillReturnRows(sqlmock.NewRows([]string{"id", "abate_id", "etapa", "object_key", "nome", "tipo", "tamanho", "sha"}).
			AddRow(13, 4, "FAZENDA", "Dono/7/FAZENDA/hash", "foto.png", "image/png", int64(20), "hash"))
	mock.ExpectExec(`SELECT pg_advisory_xact_lock`).
		WithArgs("Dono/7/FAZENDA/hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM public\.abate_fotos WHERE id=\$1`).
		WithArgs(13).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO public\.abate_storage_cleanup`).
		WithArgs("Dono/7/FAZENDA/hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	foto, err := NewAbateRepository(db).DeleteFoto(13)
	if err != nil {
		t.Fatal(err)
	}
	if foto == nil || foto.ObjectKey != "Dono/7/FAZENDA/hash" {
		t.Fatalf("foto=%+v", foto)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAbateRepositoryPathUpdateUsesIdentityLockAndAtomicPhotoCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dados := domain.DadosGeraisAbate{
		FazendaID: 3, NumeroLote: 4, DataAbate: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		NomeFrigorifico: "Frigo", DistanciaFrigorifico: 5, CategoriaAnimal: "Bovino",
	}
	mock.ExpectQuery(`(?s)WITH lock_identidade AS MATERIALIZED.*pg_advisory_xact_lock.*abate_fotos.*UPDATE public\.abate`).
		WithArgs(3, 4, dados.DataAbate, "Frigo", 5.0, "Bovino", 0.0, 0.0, 9, "identity:abate:9").
		WillReturnRows(sqlmock.NewRows([]string{"encontrado", "bloqueado", "atualizado"}).AddRow(true, true, false))
	err = NewAbateRepository(db).UpdateDadosGeraisAbate(9, dados)
	if !errors.Is(err, domain.ErrDadosGeraisComFotos) {
		t.Fatalf("erro=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAbateRepositoryLoadsListDetailsInFiveBatchQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	baseRows := sqlmock.NewRows([]string{
		"id", "proprietario_id", "proprietario", "fazenda", "fazenda_id", "numero_lote",
		"data_abate", "frigorifico", "distancia", "categoria", "funrural", "sem_funrural",
		"peso_fazenda", "peso_frigorifico", "balancao",
	}).
		AddRow(9, 1, "Dono", "Fazenda", 2, 10, time.Now(), "Frigo", 3.0, "Bovino", 1.0, 2.0, 3.0, 4.0, 5.0).
		AddRow(10, 1, "Dono", "Fazenda", 2, 11, time.Now(), "Frigo", 3.0, "Bovino", 1.0, 2.0, 3.0, 4.0, 5.0)
	mock.ExpectQuery(`(?s)FROM public\.abate a.*LIMIT \$1 OFFSET \$2`).
		WithArgs(50, 0).WillReturnRows(baseRows)
	mock.ExpectQuery(`SELECT abate_id, denticao.*ANY`).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"abate_id", "denticao", "qtd"}))
	mock.ExpectQuery(`SELECT abate_id, acabamento.*ANY`).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"abate_id", "acabamento", "qtd"}))
	mock.ExpectQuery(`SELECT abate_id, classificacao.*abate_classificacao.*ANY`).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"abate_id", "classificacao", "qtd"}))
	mock.ExpectQuery(`SELECT abate_id, classificacao.*peso_total.*ANY`).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"abate_id", "classificacao", "qtd", "peso"}))
	mock.ExpectQuery(`SELECT id, abate_id, etapa.*abate_fotos.*ANY`).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "abate_id", "etapa", "key", "nome", "tipo", "tamanho", "sha"}))
	abates, err := NewAbateRepository(db).FindAbates(domain.FiltroAbate{})
	if err != nil {
		t.Fatal(err)
	}
	if len(abates) != 2 {
		t.Fatalf("quantidade=%d", len(abates))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
