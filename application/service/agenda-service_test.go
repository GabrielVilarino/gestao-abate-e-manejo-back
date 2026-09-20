package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type agendaPortStub struct {
	createSubscription func(*domain.PushSubscription) error
	create             func(*domain.Agenda) error
	get                func(int, int) (*domain.Agenda, error)
	find               func(domain.AgendaFilter) (*[]domain.Agenda, int, error)
	update             func(*domain.Agenda) error
	delete             func(int, int) error
	deleteSubscription func(int, int) error
}

func (s agendaPortStub) CreateAgenda(v *domain.Agenda) error              { return s.create(v) }
func (s agendaPortStub) GetAgenda(id, userID int) (*domain.Agenda, error) { return s.get(id, userID) }
func (s agendaPortStub) FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
	return s.find(filter)
}
func (s agendaPortStub) UpdateAgenda(v *domain.Agenda) error { return s.update(v) }
func (s agendaPortStub) DeleteAgenda(id, userID int) error   { return s.delete(id, userID) }
func (s agendaPortStub) CreatePushSubscription(v *domain.PushSubscription) error {
	return s.createSubscription(v)
}
func (s agendaPortStub) DeletePushSubscription(id, userID int) error {
	return s.deleteSubscription(id, userID)
}

type agendaNotificationPortStub struct {
	jobs          []domain.AgendaNotification
	subscriptions []domain.PushSubscription
	finishID      int
	finishToken   string
	finishError   string
	removed       []string
	claimErr      error
	renewErr      error
	renewCount    int
}

func (s *agendaNotificationPortStub) ClaimDueAgendaNotifications(time.Duration, int) ([]domain.AgendaNotification, error) {
	return s.jobs, s.claimErr
}
func (s *agendaNotificationPortStub) RenewAgendaNotification(int, string, time.Duration) error {
	s.renewCount++
	return s.renewErr
}
func (s *agendaNotificationPortStub) GetPushSubscriptions(int) ([]domain.PushSubscription, error) {
	return s.subscriptions, nil
}
func (s *agendaNotificationPortStub) FinishAgendaNotification(id int, token string, sendError string) error {
	s.finishID, s.finishToken, s.finishError = id, token, sendError
	return nil
}
func (s *agendaNotificationPortStub) DeletePushSubscriptionByEndpoint(endpoint string) error {
	s.removed = append(s.removed, endpoint)
	return nil
}

type pushSenderStub struct {
	fail    map[string]error
	sent    []string
	payload []byte
	ttl     int
	send    func(context.Context, domain.PushSubscription, []byte, int) error
}

func (s *pushSenderStub) Send(ctx context.Context, subscription domain.PushSubscription, payload []byte, ttl int) error {
	s.sent = append(s.sent, subscription.Endpoint)
	s.payload = payload
	s.ttl = ttl
	if s.send != nil {
		return s.send(ctx, subscription, payload, ttl)
	}
	return s.fail[subscription.Endpoint]
}

func TestAgendaServiceFindRejectsInvalidPeriod(t *testing.T) {
	start := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	end := start.Add(-time.Minute)
	service := NewAgendaService(agendaPortStub{})
	if _, _, err := service.FindAgenda(domain.AgendaFilter{DataInicio: &start, DataFim: &end}); !errors.Is(err, domain.ErrPeriodoAgendaInvalido) {
		t.Fatalf("erro = %v, esperado período inválido", err)
	}
}

func TestAgendaServiceFindAppliesDefaultPagination(t *testing.T) {
	var got domain.AgendaFilter
	service := NewAgendaService(agendaPortStub{find: func(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
		got = filter
		items := []domain.Agenda{}
		return &items, 0, nil
	}})
	_, _, err := service.FindAgenda(domain.AgendaFilter{UserID: 8})
	if err != nil {
		t.Fatal(err)
	}
	if got.Pagina != 1 || got.Limite != 20 {
		t.Fatalf("paginação = %d/%d", got.Pagina, got.Limite)
	}
}

func TestRunNotificationCycleSendsAndCompletes(t *testing.T) {
	notifications := &agendaNotificationPortStub{
		jobs:          []domain.AgendaNotification{{ID: 11, AgendaID: 23, UserID: 7, FazendaID: 4, DataHora: time.Now().Add(time.Hour), LeaseToken: "lease-11"}},
		subscriptions: []domain.PushSubscription{{Endpoint: "https://push.example/device"}},
	}
	sender := &pushSenderStub{fail: map[string]error{}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)
	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(sender.sent) != 1 || notifications.renewCount != 1 || notifications.finishID != 11 || notifications.finishToken != "lease-11" || notifications.finishError != "" {
		t.Fatalf("sent=%v finish=%d err=%q", sender.sent, notifications.finishID, notifications.finishError)
	}
	if bytes.Contains(sender.payload, []byte("observacao")) {
		t.Fatalf("payload contém observação: %s", sender.payload)
	}
	if sender.ttl <= 0 || sender.ttl > 60*60 {
		t.Fatalf("TTL = %d", sender.ttl)
	}
}

func TestRunNotificationCycleFormatsBodyInBrasiliaTime(t *testing.T) {
	dataHora := time.Now().Add(time.Hour).Truncate(time.Second)
	notifications := &agendaNotificationPortStub{
		jobs:          []domain.AgendaNotification{{ID: 15, AgendaID: 23, UserID: 7, FazendaID: 4, FazendaNome: "Boa Vista", DataHora: dataHora.Add(time.Hour), LeaseToken: "lease-15"}},
		subscriptions: []domain.PushSubscription{{Endpoint: "https://push.example/device"}},
	}
	sender := &pushSenderStub{fail: map[string]error{}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)

	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	wantBody := fmt.Sprintf(`"body":"Você tem um agendamento de abate na fazenda Boa Vista às %s"`, dataHora.Add(time.Hour).In(brasiliaLocation).Format("15:04"))
	if !bytes.Contains(sender.payload, []byte(wantBody)) {
		t.Fatalf("body da notificação = %s", sender.payload)
	}
}

func TestRunNotificationCycleLimitsTTLToRemainingAppointmentTime(t *testing.T) {
	notifications := &agendaNotificationPortStub{
		jobs:          []domain.AgendaNotification{{ID: 13, UserID: 7, DataHora: time.Now().Add(90 * time.Second), LeaseToken: "lease-13"}},
		subscriptions: []domain.PushSubscription{{Endpoint: "https://push.example/device"}},
	}
	sender := &pushSenderStub{fail: map[string]error{}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)

	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.ttl <= 0 || sender.ttl > 90 {
		t.Fatalf("TTL = %d, esperado no máximo 90 segundos", sender.ttl)
	}
}

func TestRunNotificationCycleStopsSendingAtAppointmentTime(t *testing.T) {
	dataHora := time.Now().Add(100 * time.Millisecond)
	notifications := &agendaNotificationPortStub{
		jobs: []domain.AgendaNotification{{
			ID: 14, UserID: 7, DataHora: dataHora, LeaseToken: "lease-14",
		}},
		subscriptions: []domain.PushSubscription{
			{Endpoint: "https://push.example/first"},
			{Endpoint: "https://push.example/second"},
		},
	}
	sender := &pushSenderStub{fail: map[string]error{}}
	var deadlineErr error
	sender.send = func(ctx context.Context, _ domain.PushSubscription, _ []byte, _ int) error {
		deadline, ok := ctx.Deadline()
		if !ok || !deadline.Equal(dataHora) {
			deadlineErr = errors.New("contexto de envio não respeitou o horário do agendamento")
		}
		<-ctx.Done()
		return ctx.Err()
	}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)

	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if deadlineErr != nil {
		t.Fatal(deadlineErr)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("envios = %v, esperado interromper após atingir o horário", sender.sent)
	}
	if notifications.finishError == "" {
		t.Fatal("falha por horário atingido não foi persistida")
	}
}

func TestRunNotificationCyclePersistsDeliveryErrorForRetry(t *testing.T) {
	notifications := &agendaNotificationPortStub{
		jobs:          []domain.AgendaNotification{{ID: 9, AgendaID: 23, UserID: 7, DataHora: time.Now().Add(time.Hour), LeaseToken: "lease-9"}},
		subscriptions: []domain.PushSubscription{{Endpoint: "https://push.example/device"}},
	}
	sender := &pushSenderStub{fail: map[string]error{"https://push.example/device": errors.New("indisponível")}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)
	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if notifications.finishID != 9 || notifications.finishError == "" {
		t.Fatalf("resultado não persistiu falha: id=%d erro=%q", notifications.finishID, notifications.finishError)
	}
}

func TestRunNotificationCycleRemovesExpiredSubscriptionAndStoresFailure(t *testing.T) {
	notifications := &agendaNotificationPortStub{
		jobs:          []domain.AgendaNotification{{ID: 10, UserID: 7, DataHora: time.Now().Add(time.Hour), LeaseToken: "lease-10"}},
		subscriptions: []domain.PushSubscription{{Endpoint: "https://push.example/gone"}},
	}
	sender := &pushSenderStub{fail: map[string]error{"https://push.example/gone": domain.ErrPushSubscriptionGone}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)
	if err := service.RunNotificationCycle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(notifications.removed) != 1 || notifications.finishError == "" {
		t.Fatalf("assinaturas removidas=%v, erro salvo=%q", notifications.removed, notifications.finishError)
	}
}

func TestRunNotificationCycleDoesNotSendAfterLosingLease(t *testing.T) {
	notifications := &agendaNotificationPortStub{
		jobs:     []domain.AgendaNotification{{ID: 12, UserID: 7, LeaseToken: "old-lease"}},
		renewErr: domain.ErrNotificationLeaseLost,
	}
	sender := &pushSenderStub{fail: map[string]error{}}
	service := NewAgendaNotificationService(agendaPortStub{}, notifications, sender)

	err := service.RunNotificationCycle(context.Background())

	if err == nil || len(sender.sent) != 0 || notifications.finishID != 0 {
		t.Fatalf("erro=%v enviados=%v finish=%d", err, sender.sent, notifications.finishID)
	}
}
