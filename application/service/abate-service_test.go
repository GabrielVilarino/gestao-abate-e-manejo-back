package service

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type abatePortStub struct {
	create                 func(*domain.Abate) error
	findByFazendaID        func(int) (*domain.Abate, error)
	updateDadosGerais      func(int, domain.DadosGeraisAbate) error
	updateEtapaFazenda     func(int, domain.EtapaFazenda) error
	updateEtapaFrigorifico func(int, domain.EtapaFrigorifico) error
	delete                 func(int) error
}

func (s abatePortStub) CreateAbate(abate *domain.Abate) error { return s.create(abate) }
func (s abatePortStub) FindAbateByFazendaID(id int) (*domain.Abate, error) {
	return s.findByFazendaID(id)
}
func (s abatePortStub) UpdateDadosGeraisAbate(id int, dados domain.DadosGeraisAbate) error {
	return s.updateDadosGerais(id, dados)
}
func (s abatePortStub) UpdateEtapaFazenda(id int, etapa domain.EtapaFazenda) error {
	return s.updateEtapaFazenda(id, etapa)
}
func (s abatePortStub) UpdateEtapaFrigorifico(id int, etapa domain.EtapaFrigorifico) error {
	return s.updateEtapaFrigorifico(id, etapa)
}
func (s abatePortStub) Delete(id int) error { return s.delete(id) }

type storagePortStub struct {
	upload   func([]byte) (string, error)
	download func(string) ([]byte, error)
}

func (s storagePortStub) Upload(data []byte) (string, error)  { return s.upload(data) }
func (s storagePortStub) Download(url string) ([]byte, error) { return s.download(url) }

func TestAbateServiceCreateAbate(t *testing.T) {
	abate := &domain.Abate{ID: 1}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			create: func(got *domain.Abate) error {
				if got != abate {
					t.Fatal("CreateAbate recebeu outro ponteiro")
				}
				return expectedErr
			},
		}, storagePortStub{})
		return service.CreateAbate(abate)
	})
}

func TestAbateServiceFindAbateByFazendaID(t *testing.T) {
	abate := &domain.Abate{ID: 1}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			findByFazendaID: func(id int) (*domain.Abate, error) {
				if id != 27 {
					t.Fatalf("id da fazenda = %d", id)
				}
				return abate, expectedErr
			},
		}, storagePortStub{})

		got, err := service.FindAbateByFazendaID(27)
		if got != abate {
			t.Fatalf("abate recebido = %v", got)
		}
		return err
	})
}

func TestAbateServiceUpdateDadosGeraisAbate(t *testing.T) {
	dados := domain.DadosGeraisAbate{
		DataAbate:       "2026-09-10",
		FazendaID:       27,
		NomeFrigorifico: "Frigorifico Um",
	}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			updateDadosGerais: func(id int, got domain.DadosGeraisAbate) error {
				if id != 42 {
					t.Fatalf("id do abate = %d", id)
				}
				if !reflect.DeepEqual(got, dados) {
					t.Fatalf("dados recebidos = %#v, esperados %#v", got, dados)
				}
				return expectedErr
			},
		}, storagePortStub{})
		return service.UpdateDadosGeraisAbate(42, dados)
	})
}

func TestAbateServiceUpdateEtapaFazenda(t *testing.T) {
	etapa := domain.EtapaFazenda{
		QuantidadeAnimal: []domain.QtdDenticao{{QtdDenticao: 2, QtdAnimais: 15}},
		PesoTotal:        5000,
		Fotos:            []string{"foto-1"},
	}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			updateEtapaFazenda: func(id int, got domain.EtapaFazenda) error {
				if id != 42 {
					t.Fatalf("id do abate = %d", id)
				}
				if !reflect.DeepEqual(got, etapa) {
					t.Fatalf("etapa recebida = %#v, esperada %#v", got, etapa)
				}
				return expectedErr
			},
		}, storagePortStub{})
		return service.UpdateEtapaFazenda(42, etapa)
	})
}

func TestAbateServiceUpdateEtapaFrigorifico(t *testing.T) {
	etapa := domain.EtapaFrigorifico{
		PesoTotal:         4900,
		Balancao:          100,
		AcabamentoCarcaca: []domain.AcabamentoCarcaca{{Acabamento: "uniforme", QtdAnimais: 15}},
		Fotos:             []string{"foto-2"},
	}
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			updateEtapaFrigorifico: func(id int, got domain.EtapaFrigorifico) error {
				if id != 42 {
					t.Fatalf("id do abate = %d", id)
				}
				if !reflect.DeepEqual(got, etapa) {
					t.Fatalf("etapa recebida = %#v, esperada %#v", got, etapa)
				}
				return expectedErr
			},
		}, storagePortStub{})
		return service.UpdateEtapaFrigorifico(42, etapa)
	})
}

func TestAbateServiceDelete(t *testing.T) {
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{
			delete: func(id int) error {
				if id != 42 {
					t.Fatalf("id recebido = %d", id)
				}
				return expectedErr
			},
		}, storagePortStub{})
		return service.Delete(42)
	})
}

func TestAbateServiceUploadFotoAbate(t *testing.T) {
	data := []byte("conteudo da foto")
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{}, storagePortStub{
			upload: func(got []byte) (string, error) {
				if !bytes.Equal(got, data) {
					t.Fatalf("dados recebidos = %q", got)
				}
				return "https://storage/foto.jpg", expectedErr
			},
		})

		url, err := service.UploadFotoAbate(data)
		if url != "https://storage/foto.jpg" {
			t.Fatalf("url = %q", url)
		}
		return err
	})
}

func TestAbateServiceDownloadFotoAbate(t *testing.T) {
	expectedData := []byte("conteudo da foto")
	testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
		service := NewAbateService(abatePortStub{}, storagePortStub{
			download: func(url string) ([]byte, error) {
				if url != "https://storage/foto.jpg" {
					t.Fatalf("url recebida = %q", url)
				}
				return expectedData, expectedErr
			},
		})

		data, err := service.DownloadFotoAbate("https://storage/foto.jpg")
		if !bytes.Equal(data, expectedData) {
			t.Fatalf("dados recebidos = %q", data)
		}
		return err
	})
}
