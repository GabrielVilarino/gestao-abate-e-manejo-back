package domain

import "time"

type Agenda struct {
	ID         int
	FazendaID  int
	DataHora   time.Time
	Observacao string
}
