package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type agendaControllerUseCaseStub struct {
	create             func(*domain.Agenda) error
	get                func(int, int) (*domain.Agenda, error)
	find               func(domain.AgendaFilter) (*[]domain.Agenda, int, error)
	update             func(*domain.Agenda) error
	delete             func(int, int) error
	createSubscription func(*domain.PushSubscription) error
	deleteSubscription func(int, int) error
}

func (s agendaControllerUseCaseStub) CreateAgenda(agenda *domain.Agenda) error {
	return s.create(agenda)
}

func (s agendaControllerUseCaseStub) GetAgenda(agendaID, userID int) (*domain.Agenda, error) {
	return s.get(agendaID, userID)
}

func (s agendaControllerUseCaseStub) FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
	return s.find(filter)
}

func (s agendaControllerUseCaseStub) UpdateAgenda(agenda *domain.Agenda) error {
	return s.update(agenda)
}

func (s agendaControllerUseCaseStub) DeleteAgenda(agendaID, userID int) error {
	return s.delete(agendaID, userID)
}

func (s agendaControllerUseCaseStub) CreatePushSubscription(subscription *domain.PushSubscription) error {
	return s.createSubscription(subscription)
}

func (s agendaControllerUseCaseStub) DeletePushSubscription(subscriptionID, userID int) error {
	return s.deleteSubscription(subscriptionID, userID)
}

func TestAgendaControllerCreateUsesAuthenticatedUserAndConvertsTimeToUTC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewAgendaController(agendaControllerUseCaseStub{
		create: func(agenda *domain.Agenda) error {
			want := time.Date(2026, 9, 13, 17, 0, 0, 0, time.UTC)
			if agenda.UserID != 7 || agenda.FazendaID != 12 || !agenda.DataHora.Equal(want) {
				t.Fatalf("agenda recebida = %+v", agenda)
			}
			agenda.ID = 31
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &domain.User{ID: 7})
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{
		"fazenda_id":12,
		"data_hora":"2026-09-13T14:00:00-03:00",
		"observacao":"Vacinação"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")

	controller.CreateAgenda(c)

	if w.Code != http.StatusCreated || !bytes.Contains(w.Body.Bytes(), []byte(`"id":31`)) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAgendaControllerFindCombinesFiltersAndOwnership(t *testing.T) {
	controller := NewAgendaController(agendaControllerUseCaseStub{
		find: func(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
			if filter.UserID != 9 || filter.FazendaID == nil || *filter.FazendaID != 4 || filter.Pagina != 2 || filter.Limite != 25 {
				t.Fatalf("filtro recebido = %+v", filter)
			}
			if filter.DataInicio == nil || filter.DataFim == nil || !filter.DataInicio.Before(*filter.DataFim) {
				t.Fatalf("período recebido = %+v", filter)
			}
			items := []domain.Agenda{}
			return &items, 0, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &domain.User{ID: 9})
	c.Request = httptest.NewRequest(http.MethodGet, "/?fazenda_id=4&data_inicio=2026-09-13T08:00:00-03:00&data_fim=2026-09-14T08:00:00-03:00&pagina=2&limite=25", nil)

	controller.FindAgendas(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAgendaControllerCreatesStandardPushSubscription(t *testing.T) {
	expiration := int64(1_800_000_000_000)
	controller := NewAgendaController(agendaControllerUseCaseStub{
		createSubscription: func(subscription *domain.PushSubscription) error {
			if subscription.UserID != 5 || subscription.Endpoint != "https://push.example/device" || subscription.P256DH != "public-key" || subscription.Auth != "auth-key" {
				t.Fatalf("assinatura recebida = %+v", subscription)
			}
			if subscription.ExpiresAt == nil || subscription.ExpiresAt.UnixMilli() != expiration {
				t.Fatalf("expiração recebida = %v", subscription.ExpiresAt)
			}
			subscription.ID = 44
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &domain.User{ID: 5})
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{
		"endpoint":"https://push.example/device",
		"expirationTime":1800000000000,
		"keys":{"p256dh":"public-key","auth":"auth-key"}
	}`))
	c.Request.Header.Set("Content-Type", "application/json")

	controller.CreatePushSubscription(c)

	if w.Code != http.StatusCreated || !bytes.Contains(w.Body.Bytes(), []byte(`"id":44`)) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAgendaControllerRejectsDateWithoutRFC3339Offset(t *testing.T) {
	controller := NewAgendaController(agendaControllerUseCaseStub{
		create: func(*domain.Agenda) error {
			t.Fatal("use case não deveria ser chamado")
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &domain.User{ID: 7})
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"fazenda_id":12,"data_hora":"2026-09-13T14:00:00"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	controller.CreateAgenda(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
