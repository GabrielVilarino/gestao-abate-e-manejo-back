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

type proprietarioUseCaseStub struct {
	create     func(*domain.Proprietario) error
	getAll     func() (*[]domain.Proprietario, error)
	update     func(*domain.Proprietario) error
	activate   func(int) error
	deactivate func(int) error
}

func (s proprietarioUseCaseStub) CreateProprietario(proprietario *domain.Proprietario) error {
	return s.create(proprietario)
}
func (s proprietarioUseCaseStub) GetProprietarios() (*[]domain.Proprietario, error) {
	return s.getAll()
}
func (s proprietarioUseCaseStub) UpdateProprietario(proprietario *domain.Proprietario) error {
	return s.update(proprietario)
}
func (s proprietarioUseCaseStub) ActivateProprietario(id int) error   { return s.activate(id) }
func (s proprietarioUseCaseStub) DeactivateProprietario(id int) error { return s.deactivate(id) }

func TestProprietarioControllerCreateStartsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewProprietarioController(proprietarioUseCaseStub{
		create: func(proprietario *domain.Proprietario) error {
			if !proprietario.Ativo {
				t.Fatal("novo proprietário deve iniciar ativo")
			}
			if proprietario.Nome != "João" || proprietario.CPF != "12345678900" {
				t.Fatalf("proprietário recebido = %+v", proprietario)
			}
			return nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"nome":"João","cpf":"12345678900","observacao":""}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.CreateProprietario(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestProprietarioControllerGetIncludesStatus(t *testing.T) {
	proprietarios := []domain.Proprietario{{ID: 1, Nome: "João", CPF: "12345678900", Ativo: false}}
	controller := NewProprietarioController(proprietarioUseCaseStub{
		getAll: func() (*[]domain.Proprietario, error) { return &proprietarios, nil },
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	controller.GetProprietarios(c)

	var body struct {
		Proprietarios []struct {
			ID    int  `json:"id"`
			Ativo bool `json:"ativo"`
		} `json:"proprietarios"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(body.Proprietarios) != 1 || body.Proprietarios[0].ID != 1 || body.Proprietarios[0].Ativo {
		t.Fatalf("resposta inesperada: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestProprietarioControllerRejectsInvalidActivationID(t *testing.T) {
	controller := NewProprietarioController(proprietarioUseCaseStub{
		activate: func(int) error { return errors.New("não deveria ser chamado") },
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalido"}}
	controller.ActivateProprietario(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
	}
}

func TestProprietarioControllerRejectsNegativeUpdateID(t *testing.T) {
	controller := NewProprietarioController(proprietarioUseCaseStub{
		update: func(*domain.Proprietario) error {
			t.Fatal("use case não deveria ser chamado")
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(`{"id":-1,"nome":"João","cpf":"12345678900"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.UpdateProprietario(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
