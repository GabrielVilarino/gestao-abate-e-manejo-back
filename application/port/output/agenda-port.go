package output

import (
	"context"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type AgendaPort interface {
	CreateAgenda(agenda *domain.Agenda) error
	GetAgenda(agendaID, userID int) (*domain.Agenda, error)
	FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error)
	UpdateAgenda(agenda *domain.Agenda) error
	DeleteAgenda(agendaID, userID int) error
	CreatePushSubscription(subscription *domain.PushSubscription) error
	DeletePushSubscription(subscriptionID, userID int) error
}

type AgendaNotificationPort interface {
	ClaimDueAgendaNotifications(lease time.Duration, limit int) ([]domain.AgendaNotification, error)
	RenewAgendaNotification(notificationID int, leaseToken string, lease time.Duration) error
	GetPushSubscriptions(userID int) ([]domain.PushSubscription, error)
	FinishAgendaNotification(notificationID int, leaseToken string, sendError string) error
	DeletePushSubscriptionByEndpoint(endpoint string) error
}

type PushSender interface {
	Send(ctx context.Context, subscription domain.PushSubscription, payload []byte, ttl int) error
}
