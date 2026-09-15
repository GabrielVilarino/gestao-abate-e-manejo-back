package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type abateControllerUseCaseStub struct {
	create            func(*domain.Abate) error
	findByID          func(int) (*domain.Abate, error)
	find              func(domain.FiltroAbate) ([]domain.Abate, error)
	updateDados       func(int, domain.DadosGeraisAbate) error
	updateFazenda     func(int, domain.EtapaFazenda) error
	updateFrigorifico func(int, domain.EtapaFrigorifico) error
	deleteAbate       func(context.Context, int) error
	upload            func(context.Context, int, string, string, []byte) (*domain.FotoAbate, error)
	download          func(context.Context, int) (*domain.FotoAbate, io.ReadCloser, error)
	deleteFoto        func(context.Context, int) error
	generateReport    func(context.Context, []int) ([]byte, error)
}

func (s abateControllerUseCaseStub) CreateAbate(v *domain.Abate) error { return s.create(v) }
func (s abateControllerUseCaseStub) FindAbateByID(id int) (*domain.Abate, error) {
	return s.findByID(id)
}
func (s abateControllerUseCaseStub) FindAbates(f domain.FiltroAbate) ([]domain.Abate, error) {
	return s.find(f)
}
func (s abateControllerUseCaseStub) UpdateDadosGeraisAbate(id int, v domain.DadosGeraisAbate) error {
	return s.updateDados(id, v)
}
func (s abateControllerUseCaseStub) UpdateEtapaFazenda(id int, v domain.EtapaFazenda) error {
	return s.updateFazenda(id, v)
}
func (s abateControllerUseCaseStub) UpdateEtapaFrigorifico(id int, v domain.EtapaFrigorifico) error {
	return s.updateFrigorifico(id, v)
}
func (s abateControllerUseCaseStub) Delete(ctx context.Context, id int) error {
	return s.deleteAbate(ctx, id)
}
func (s abateControllerUseCaseStub) UploadFotoAbate(ctx context.Context, id int, etapa, nome string, data []byte) (*domain.FotoAbate, error) {
	return s.upload(ctx, id, etapa, nome, data)
}
func (s abateControllerUseCaseStub) DownloadFotoAbate(ctx context.Context, id int) (*domain.FotoAbate, io.ReadCloser, error) {
	return s.download(ctx, id)
}
func (s abateControllerUseCaseStub) DeleteFotoAbate(ctx context.Context, id int) error {
	return s.deleteFoto(ctx, id)
}
func (s abateControllerUseCaseStub) GenerateAbateReport(ctx context.Context, ids []int) ([]byte, error) {
	return s.generateReport(ctx, ids)
}

func TestAbateControllerCreateRequiresLotAndDistance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewAbateController(abateControllerUseCaseStub{
		create: func(*domain.Abate) error {
			t.Fatal("use case não deveria ser chamado")
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{
		"dados_gerais":{"data_abate":"2026-09-12","fazenda_id":2,"nome_frigorifico":"F","categoria_animal":"Bovino"}
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.CreateAbate(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAbateControllerCreateValidatesRequiredDenticaoIncludingZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		denticao   string
		wantStatus int
	}{
		{name: "ausente", denticao: ``, wantStatus: http.StatusBadRequest},
		{name: "nula", denticao: `"denticao":null,`, wantStatus: http.StatusBadRequest},
		{name: "fora do catálogo", denticao: `"denticao":3,`, wantStatus: http.StatusBadRequest},
		{name: "zero", denticao: `"denticao":0,`, wantStatus: http.StatusCreated},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			controller := NewAbateController(abateControllerUseCaseStub{
				create: func(abate *domain.Abate) error {
					called = true
					if len(abate.EtapaFazenda.QuantidadeAnimal) != 1 || abate.EtapaFazenda.QuantidadeAnimal[0].QtdDenticao != 0 {
						t.Fatalf("dentição mapeada=%+v", abate.EtapaFazenda.QuantidadeAnimal)
					}
					return nil
				},
			})
			body := fmt.Sprintf(`{
				"dados_gerais":{"data_abate":"2026-09-12","fazenda_id":2,"numero_lote":1,"nome_frigorifico":"F","distancia_frigorifico":1,"categoria_animal":"Bovino"},
				"etapa_fazenda":{"quantidade_animal":[{%s"qtd_animais":1}]}
			}`, test.denticao)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.CreateAbate(c)

			if w.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if called != (test.wantStatus == http.StatusCreated) {
				t.Fatalf("use case chamado=%v", called)
			}
		})
	}
}

func TestAbateControllerUpdateEtapaFazendaValidatesRequiredDenticaoIncludingZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		denticao   string
		wantStatus int
	}{
		{name: "ausente", denticao: ``, wantStatus: http.StatusBadRequest},
		{name: "nula", denticao: `"denticao":null,`, wantStatus: http.StatusBadRequest},
		{name: "fora do catálogo", denticao: `"denticao":3,`, wantStatus: http.StatusBadRequest},
		{name: "zero", denticao: `"denticao":0,`, wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			controller := NewAbateController(abateControllerUseCaseStub{
				updateFazenda: func(id int, etapa domain.EtapaFazenda) error {
					called = true
					if id != 7 || len(etapa.QuantidadeAnimal) != 1 || etapa.QuantidadeAnimal[0].QtdDenticao != 0 {
						t.Fatalf("id=%d etapa=%+v", id, etapa)
					}
					return nil
				},
			})
			body := fmt.Sprintf(`{"quantidade_animal":[{%s"qtd_animais":1}]}`, test.denticao)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "7"}}
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString(body))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.UpdateEtapaFazenda(c)

			if w.Code != test.wantStatus {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if called != (test.wantStatus == http.StatusOK) {
				t.Fatalf("use case chamado=%v", called)
			}
		})
	}
}

func TestAbateControllerRejectsInvertedDateFilter(t *testing.T) {
	controller := NewAbateController(abateControllerUseCaseStub{
		find: func(domain.FiltroAbate) ([]domain.Abate, error) {
			t.Fatal("use case não deveria ser chamado")
			return nil, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?data_inicio=2026-09-12&data_fim=2026-09-11", nil)
	controller.FindAbates(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAbateControllerAppliesBoundedPagination(t *testing.T) {
	controller := NewAbateController(abateControllerUseCaseStub{
		find: func(filtro domain.FiltroAbate) ([]domain.Abate, error) {
			if filtro.Limit != 25 || filtro.Offset != 50 {
				t.Fatalf("limit=%d offset=%d", filtro.Limit, filtro.Offset)
			}
			return []domain.Abate{}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?pagina=3&limite=25", nil)
	controller.FindAbates(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAbateControllerReadsMultipartPhoto(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("foto", "animal.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(png); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	controller := NewAbateController(abateControllerUseCaseStub{
		upload: func(_ context.Context, id int, etapa, nome string, data []byte) (*domain.FotoAbate, error) {
			if id != 3 || etapa != "fazenda" || nome != "animal.png" || !bytes.Equal(data, png) {
				t.Fatalf("upload: id=%d etapa=%s nome=%s data=%x", id, etapa, nome, data)
			}
			return &domain.FotoAbate{ID: 8, Etapa: "FAZENDA", NomeOriginal: nome, ContentType: "image/png", Tamanho: int64(len(data))}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}, {Key: "etapa", Value: "fazenda"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	controller.UploadFotoAbate(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAbateControllerGeneratesPDFReport(t *testing.T) {
	controller := NewAbateController(abateControllerUseCaseStub{
		generateReport: func(_ context.Context, ids []int) ([]byte, error) {
			if len(ids) != 2 || ids[0] != 4 || ids[1] != 7 {
				t.Fatalf("ids=%v", ids)
			}
			return []byte("%PDF-1.7"), nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"abate_ids":[4,7]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.GenerateAbateReport(c)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("status=%d content-type=%q body=%q", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	if disposition := w.Header().Get("Content-Disposition"); disposition != `attachment; filename=relatorio-abates.pdf` {
		t.Fatalf("content-disposition=%q", disposition)
	}
}

func TestAbateControllerRejectsEmptyReportIDs(t *testing.T) {
	controller := NewAbateController(abateControllerUseCaseStub{
		generateReport: func(context.Context, []int) ([]byte, error) {
			t.Fatal("use case não deveria ser chamado")
			return nil, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"abate_ids":[]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	controller.GenerateAbateReport(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
