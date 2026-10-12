package controller

import (
	"errors"
	"net/http"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/response"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/input"
	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	DashboardUseCase input.DashboardUseCase
}

func NewDashboardController(dashboardUseCase input.DashboardUseCase) *DashboardController {
	return &DashboardController{DashboardUseCase: dashboardUseCase}
}

func (d *DashboardController) GetDashboardAbates(c *gin.Context) {
	proprietarioID, err := optionalPositiveInt(c, "proprietario_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	fazendaID, err := optionalPositiveInt(c, "fazenda_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	dataInicio, err := optionalDate(c, "data_inicio")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	dataFim, err := optionalDate(c, "data_fim")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	filtro := domain.FiltroDashboardAbates{
		ProprietarioID: proprietarioID,
		FazendaID:      fazendaID,
		DataInicio:     dataInicio,
		DataFim:        dataFim,
	}
	dashboard, err := d.DashboardUseCase.GetDashboardAbates(c.Request.Context(), filtro)
	if err != nil {
		if errors.Is(err, domain.ErrPeriodoInvalido) {
			c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, response.AbateErrorResponse{Error: "Erro ao consultar dashboard"})
		return
	}
	c.JSON(http.StatusOK, dashboardResponse(*dashboard))
}

func dashboardResponse(value domain.DashboardAbates) response.DashboardAbatesResponse {
	result := response.DashboardAbatesResponse{
		Resumo: response.DashboardResumoAbatesResponse{
			TotalAbates: value.Resumo.TotalAbates, TotalCabecas: value.Resumo.TotalCabecas,
			MediaCabecasPorAbate: value.Resumo.MediaCabecasPorAbate,
			PesoMedioFazenda:     value.Resumo.PesoMedioFazenda, PesoMedioFrigorifico: value.Resumo.PesoMedioFrigorifico,
			RendimentoMedioCarcaca: value.Resumo.RendimentoMedioCarcaca, DistanciaMediaFrigorifico: value.Resumo.DistanciaMediaFrigorifico,
		},
		PorFazenda:         make([]response.DashboardPorFazendaResponse, len(value.PorFazenda)),
		Evolucao:           make([]response.DashboardEvolucaoResponse, len(value.Evolucao)),
		PorCategoriaAnimal: make([]response.DashboardPorCategoriaResponse, len(value.PorCategoriaAnimal)),
	}
	for i, item := range value.PorFazenda {
		result.PorFazenda[i] = response.DashboardPorFazendaResponse{
			FazendaID: item.FazendaID, NomeFazenda: item.NomeFazenda, TotalAbates: item.TotalAbates,
			TotalCabecas: item.TotalCabecas, MediaCabecasPorAbate: item.MediaCabecasPorAbate,
		}
	}
	for i, item := range value.Evolucao {
		result.Evolucao[i] = response.DashboardEvolucaoResponse{Data: item.Data.Format(dateLayout), TotalAbates: item.TotalAbates, TotalCabecas: item.TotalCabecas}
	}
	for i, item := range value.PorCategoriaAnimal {
		result.PorCategoriaAnimal[i] = response.DashboardPorCategoriaResponse{CategoriaAnimal: item.CategoriaAnimal, TotalAbates: item.TotalAbates, TotalCabecas: item.TotalCabecas}
	}
	return result
}
