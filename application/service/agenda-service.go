package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
)

const (
	agendaNotificationLease = 2 * time.Minute
	agendaNotificationBatch = 50
)

type AgendaService struct {
	AgendaPort       output.AgendaPort
	NotificationPort output.AgendaNotificationPort
	PushSender       output.PushSender
}

func NewAgendaService(agendaPort output.AgendaPort) *AgendaService {
	return &AgendaService{AgendaPort: agendaPort}
}

func NewAgendaNotificationService(
	agendaPort output.AgendaPort,
	notificationPort output.AgendaNotificationPort,
	pushSender output.PushSender,
) *AgendaService {
	return &AgendaService{
		AgendaPort:       agendaPort,
		NotificationPort: notificationPort,
		PushSender:       pushSender,
	}
}

func (a *AgendaService) CreateAgenda(agenda *domain.Agenda) error {
	return a.AgendaPort.CreateAgenda(agenda)
}

func (a *AgendaService) GetAgenda(agendaID, userID int) (*domain.Agenda, error) {
	return a.AgendaPort.GetAgenda(agendaID, userID)
}

func (a *AgendaService) FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
	if filter.DataInicio != nil && filter.DataFim != nil && filter.DataInicio.After(*filter.DataFim) {
		return nil, 0, domain.ErrPeriodoAgendaInvalido
	}
	if filter.Pagina < 1 {
		filter.Pagina = 1
	}
	if filter.Limite < 1 || filter.Limite > 100 {
		filter.Limite = 20
	}
	return a.AgendaPort.FindAgenda(filter)
}

func (a *AgendaService) UpdateAgenda(agenda *domain.Agenda) error {
	return a.AgendaPort.UpdateAgenda(agenda)
}

func (a *AgendaService) DeleteAgenda(agendaID, userID int) error {
	return a.AgendaPort.DeleteAgenda(agendaID, userID)
}

func (a *AgendaService) CreatePushSubscription(subscription *domain.PushSubscription) error {
	if subscription.Endpoint == "" || subscription.P256DH == "" || subscription.Auth == "" {
		return domain.ErrAssinaturaPushInvalida
	}
	return a.AgendaPort.CreatePushSubscription(subscription)
}

func (a *AgendaService) DeletePushSubscription(subscriptionID, userID int) error {
	return a.AgendaPort.DeletePushSubscription(subscriptionID, userID)
}

func (a *AgendaService) StartNotificationWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			if err := a.RunNotificationCycle(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("[AGENDA] - Falha ao processar notificações", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (a *AgendaService) RunNotificationCycle(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	jobs, err := a.NotificationPort.ClaimDueAgendaNotifications(agendaNotificationLease, agendaNotificationBatch)
	if err != nil {
		return err
	}
	failures := make(chan string, len(jobs))
	var workers sync.WaitGroup
	for _, job := range jobs {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := a.processAgendaNotification(ctx, job); err != nil {
				failures <- fmt.Sprintf("notificação %d: %v", job.ID, err)
			}
		}()
	}
	workers.Wait()
	close(failures)
	messages := make([]string, 0, len(failures))
	for failure := range failures {
		messages = append(messages, failure)
	}
	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "; "))
	}
	return nil
}

func (a *AgendaService) processAgendaNotification(ctx context.Context, job domain.AgendaNotification) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := a.NotificationPort.RenewAgendaNotification(job.ID, job.LeaseToken, agendaNotificationLease); err != nil {
		return err
	}
	leaseCtx, cancelLease := context.WithCancel(ctx)
	renewalErrors := make(chan error, 1)
	var renewalWorker sync.WaitGroup
	renewalWorker.Add(1)
	go func() {
		defer renewalWorker.Done()
		ticker := time.NewTicker(agendaNotificationLease / 3)
		defer ticker.Stop()
		for {
			select {
			case <-leaseCtx.Done():
				return
			case <-ticker.C:
				if err := a.NotificationPort.RenewAgendaNotification(job.ID, job.LeaseToken, agendaNotificationLease); err != nil {
					renewalErrors <- err
					cancelLease()
					return
				}
			}
		}
	}()

	sendErr := a.sendAgendaNotification(leaseCtx, job)
	cancelLease()
	renewalWorker.Wait()
	select {
	case err := <-renewalErrors:
		return err
	default:
	}
	message := ""
	if sendErr != nil {
		message = sendErr.Error()
	}
	return a.NotificationPort.FinishAgendaNotification(job.ID, job.LeaseToken, message)
}

func (a *AgendaService) sendAgendaNotification(ctx context.Context, job domain.AgendaNotification) error {
	if time.Until(job.DataHora) <= 0 {
		return errors.New("horário do agendamento atingido antes do envio")
	}
	sendCtx, cancel := context.WithDeadline(ctx, job.DataHora)
	defer cancel()
	subscriptions, err := a.NotificationPort.GetPushSubscriptions(job.UserID)
	if err != nil {
		return err
	}
	if len(subscriptions) == 0 {
		return errors.New("usuário não possui assinatura push ativa")
	}
	payload, err := json.Marshal(map[string]any{
		"title":      "Lembrete de agendamento",
		"body":       fmt.Sprintf("Agendamento da fazenda %d em %s", job.FazendaID, job.DataHora.UTC().Format(time.RFC3339)),
		"agenda_id":  job.AgendaID,
		"fazenda_id": job.FazendaID,
		"data_hora":  job.DataHora.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	var sent int
	var failures []string
	for _, subscription := range subscriptions {
		remaining := time.Until(job.DataHora)
		if remaining <= 0 {
			failures = append(failures, "horário do agendamento atingido antes do envio")
			break
		}
		ttl := int(math.Ceil(remaining.Seconds()))
		if ttl > 60*60 {
			ttl = 60 * 60
		}
		if err := a.PushSender.Send(sendCtx, subscription, payload, ttl); err != nil {
			if errors.Is(err, domain.ErrPushSubscriptionGone) {
				if removeErr := a.NotificationPort.DeletePushSubscriptionByEndpoint(subscription.Endpoint); removeErr != nil {
					failures = append(failures, removeErr.Error())
				}
				continue
			}
			failures = append(failures, err.Error())
			continue
		}
		sent++
	}
	if sent > 0 {
		return nil
	}
	if len(failures) == 0 {
		return errors.New("nenhuma assinatura push pôde receber a notificação")
	}
	return errors.New(strings.Join(failures, "; "))
}
