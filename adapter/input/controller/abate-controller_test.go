package controller

import (
	"bytes"
	"context"
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
