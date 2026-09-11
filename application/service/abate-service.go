package service

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type AbateService struct {
	AbatePort   output.AbatePort
	StoragePort output.StoragePort
}

func NewAbateService(
	abatePort output.AbatePort,
	storagePort output.StoragePort,
) *AbateService {
	return &AbateService{
		AbatePort:   abatePort,
		StoragePort: storagePort,
	}
}

func (a *AbateService) CreateAbate(abate *domain.Abate) error {
	return a.AbatePort.CreateAbate(abate)
}

func (a *AbateService) FindAbateByFazendaID(id int) (*domain.Abate, error) {
	return a.AbatePort.FindAbateByFazendaID(id)
}

func (a *AbateService) UpdateDadosGeraisAbate(abateID int, dadosGerais domain.DadosGeraisAbate) error {
	return a.AbatePort.UpdateDadosGeraisAbate(abateID, dadosGerais)
}

func (a *AbateService) UpdateEtapaFazenda(abateID int, etapa domain.EtapaFazenda) error {
	return a.AbatePort.UpdateEtapaFazenda(abateID, etapa)
}

func (a *AbateService) UpdateEtapaFrigorifico(abateID int, etapa domain.EtapaFrigorifico) error {
	return a.AbatePort.UpdateEtapaFrigorifico(abateID, etapa)
}

func (a *AbateService) Delete(abateID int) error {
	return a.AbatePort.Delete(abateID)
}

func (a *AbateService) UploadFotoAbate(data []byte) (string, error) {
	return a.StoragePort.Upload(data)
}

func (a *AbateService) DownloadFotoAbate(fotoURL string) ([]byte, error) {
	return a.StoragePort.Download(fotoURL)
}
