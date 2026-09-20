package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

func newAgendaMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestAgendaRepositoryCreateAgendaPersistsAgendaAndNotificationAtomically(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	agenda := &domain.Agenda{UserID: 7, FazendaID: 12, DataHora: time.Date(2026, 9, 13, 15, 0, 0, 0, time.FixedZone("BRT", -3*60*60)), Observacao: "Visita"}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO public.agenda (user_id, fazenda_id, data_hora, observacao)")).
		WithArgs(7, 12, agenda.DataHora.UTC(), "Visita").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(34))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO public.agenda_notificacao (agenda_id, user_id, status, next_attempt_at)")).
		WithArgs(34, 7, agenda.DataHora.UTC()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := repository.CreateAgenda(agenda); err != nil {
		t.Fatal(err)
	}
	if agenda.ID != 34 {
		t.Fatalf("id agenda = %d, esperado 34", agenda.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryGetAgendaHidesAnotherUsersAgenda(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, fazenda_id, data_hora, observacao")).
		WithArgs(5, 91).WillReturnError(sql.ErrNoRows)
	agenda, err := repository.GetAgenda(5, 91)
	if agenda != nil || !errors.Is(err, domain.ErrAgendaNaoEncontrada) {
		t.Fatalf("agenda=%v erro=%v", agenda, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryClaimsDueJobsWithLeaseAndSkipLocked(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE public.agenda_notificacao n")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("WITH due AS (")).
		WithArgs(float64(120), 25, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "agenda_id", "user_id", "tentativas", "lease_token", "fazenda_id", "nome", "data_hora"}).
			AddRow(3, 11, 8, 2, "lease-1", 4, "Boa Vista", now.Add(time.Hour)))
	mock.ExpectCommit()
	jobs, err := repository.ClaimDueAgendaNotifications(2*time.Minute, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != 3 || jobs[0].Tentativas != 2 || jobs[0].UserID != 8 || jobs[0].LeaseToken != "lease-1" {
		t.Fatalf("jobs = %+v", jobs)
	}
	if jobs[0].FazendaNome != "Boa Vista" {
		t.Fatalf("nome da fazenda = %q", jobs[0].FazendaNome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryFinishNotificationSchedulesRetryBeforeAppointment(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE public.agenda_notificacao n")).
		WithArgs(44, "push indisponível", "lease-44").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repository.FinishAgendaNotification(44, "lease-44", "push indisponível"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryUpdateRejectsUnknownFarm(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	agenda := &domain.Agenda{ID: 4, UserID: 8, FazendaID: 999, DataHora: time.Now()}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT n.status, COALESCE(n.locked_until > CURRENT_TIMESTAMP, TRUE)")).
		WithArgs(4, 8).WillReturnRows(sqlmock.NewRows([]string{"status", "lease_ativa"}).AddRow("pendente", true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM public.fazenda WHERE id = $1)")).
		WithArgs(999).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	err := repository.UpdateAgenda(agenda)

	if !errors.Is(err, domain.ErrFazendaAgendaInvalida) {
		t.Fatalf("erro = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryDeletesOnlyOwnersSubscription(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.push_subscription WHERE id = $1 AND user_id = $2")).
		WithArgs(31, 7).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.DeletePushSubscription(31, 7)

	if !errors.Is(err, domain.ErrAssinaturaNaoEncontrada) {
		t.Fatalf("erro = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryRenewRequiresCurrentLeaseToken(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE public.agenda_notificacao")).
		WithArgs(7, "expired-lease", float64(120)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.RenewAgendaNotification(7, "expired-lease", 2*time.Minute)

	if !errors.Is(err, domain.ErrNotificationLeaseLost) {
		t.Fatalf("erro = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryBlocksDeleteWhileNotificationIsOwned(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT n.status, COALESCE(n.locked_until > CURRENT_TIMESTAMP, TRUE)")).
		WithArgs(21, 5).
		WillReturnRows(sqlmock.NewRows([]string{"status", "lease_ativa"}).AddRow("processando", true))
	mock.ExpectRollback()

	err := repository.DeleteAgenda(21, 5)

	if !errors.Is(err, domain.ErrAgendaEmProcessamento) {
		t.Fatalf("erro = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAgendaRepositoryLoadsSubscriptionsOnlyForActiveUser(t *testing.T) {
	db, mock := newAgendaMock(t)
	repository := NewAgendaRepository(db)
	mock.ExpectQuery(`(?s)JOIN public\.usuario u.*u\.ativo = TRUE`).
		WithArgs(14).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "endpoint", "p256dh", "auth", "expires_at"}))

	subscriptions, err := repository.GetPushSubscriptions(14)

	if err != nil || len(subscriptions) != 0 {
		t.Fatalf("assinaturas=%v erro=%v", subscriptions, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
