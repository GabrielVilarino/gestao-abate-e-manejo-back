package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type fazendaUseCaseStub struct {
	create     func(*domain.Fazenda) error
	getAll     func(int) (*[]domain.Fazenda, error)
	update     func(*domain.Fazenda) error
	activate   func(int) error
	deactivate func(int) error
}

func (s fazendaUseCaseStub) CreateFazenda(fazenda *domain.Fazenda) error { return s.create(fazenda) }
func (s fazendaUseCaseStub) GetFazendas(idProprietario int) (*[]domain.Fazenda, error) {
	return s.getAll(idProprietario)
}
func (s fazendaUseCaseStub) UpdateFazenda(fazenda *domain.Fazenda) error { return s.update(fazenda) }
func (s fazendaUseCaseStub) ActivateFazenda(id int) error                { return s.activate(id) }
func (s fazendaUseCaseStub) DeactivateFazenda(id int) error              { return s.deactivate(id) }

func TestFazendaControllerCreateStartsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewFazendaController(fazendaUseCaseStub{
		create: func(fazenda *domain.Fazenda) error {
			if !fazenda.Ativo {
				t.Fatal("nova fazenda deve iniciar ativa")
			}
			if fazenda.IDProprietario != 9 || fazenda.Nome != "Boa Vista" {
				t.Fatalf("fazenda recebida = %+v", fazenda)
			}
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"nome":"Boa Vista","cidade":"Gurupi","inscricao_rural":"IR-1","id_proprietario":9}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.CreateFazenda(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestFazendaControllerGetIncludesStatus(t *testing.T) {
	fazendas := []domain.Fazenda{{ID: 3, IDProprietario: 9, Ativo: false}}
	controller := NewFazendaController(fazendaUseCaseStub{
		getAll: func(id int) (*[]domain.Fazenda, error) {
			if id != 9 {
				t.Fatalf("id do proprietário = %d", id)
			}
			return &fazendas, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "idProprietario", Value: "9"}}
	controller.GetFazendas(c)

	var body struct {
		Fazendas []struct {
			ID    int  `json:"id"`
			Ativo bool `json:"ativo"`
		} `json:"fazendas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(body.Fazendas) != 1 || body.Fazendas[0].ID != 3 || body.Fazendas[0].Ativo {
		t.Fatalf("resposta inesperada: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFazendaControllerActivationWithInactiveOwnerReturnsConflict(t *testing.T) {
	controller := NewFazendaController(fazendaUseCaseStub{
		activate: func(id int) error {
			if id != 3 {
				t.Fatalf("id da fazenda = %d", id)
			}
			return domain.ErrProprietarioInativo
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	controller.ActivateFazenda(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, esperado %d; body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestFazendaControllerRejectsInvalidOwnerID(t *testing.T) {
	controller := NewFazendaController(fazendaUseCaseStub{
		getAll: func(int) (*[]domain.Fazenda, error) { return nil, errors.New("não deveria ser chamado") },
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "idProprietario", Value: "0"}}
	controller.GetFazendas(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
	}
}

func TestFazendaControllerRejectsNegativeOwnerIDOnCreate(t *testing.T) {
	controller := NewFazendaController(fazendaUseCaseStub{
		create: func(*domain.Fazenda) error {
			t.Fatal("use case não deveria ser chamado")
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"nome":"Boa Vista","cidade":"Gurupi","inscricao_rural":"IR-1","id_proprietario":-1}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.CreateFazenda(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestFazendaControllerUpdateDoesNotAcceptOwnerChange(t *testing.T) {
	controller := NewFazendaController(fazendaUseCaseStub{
		update: func(fazenda *domain.Fazenda) error {
			if fazenda.IDProprietario != 0 {
				t.Fatalf("proprietário não deve fazer parte da atualização: %d", fazenda.IDProprietario)
			}
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(`{"id":11,"nome":"Boa Vista","cidade":"Gurupi","inscricao_rural":"IR-1","id_proprietario":99}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.UpdateFazenda(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
}
