package domain

import "time"

type Agenda struct {
	ID         int
	UserID     int
	FazendaID  int
	DataHora   time.Time
	Observacao string
}

type AgendaFilter struct {
	UserID     int
	FazendaID  *int
	DataInicio *time.Time
	DataFim    *time.Time
	Pagina     int
	Limite     int
}

type AgendaNotification struct {
	ID         int
	AgendaID   int
	UserID     int
	FazendaID  int
	DataHora   time.Time
	Tentativas int
	LeaseToken string
}

type PushSubscription struct {
	ID        int
	UserID    int
	Endpoint  string
	P256DH    string
	Auth      string
	ExpiresAt *time.Time
}
