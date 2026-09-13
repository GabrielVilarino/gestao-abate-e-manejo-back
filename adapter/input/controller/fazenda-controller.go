package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/request"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/response"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/input"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
	"github.com/gin-gonic/gin"
)

type FazendaController struct {
	FazendaUseCase input.FazendaUseCase
}

func NewFazendaController(fazendaUseCase input.FazendaUseCase) *FazendaController {
	return &FazendaController{FazendaUseCase: fazendaUseCase}
}

func (f *FazendaController) CreateFazenda(c *gin.Context) {
	requestData := &request.FazendaCreateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.FazendaErrorResponse{Error: err.Error()})
		return
	}

	fazenda := &domain.Fazenda{
		Nome:           requestData.Nome,
		Cidade:         requestData.Cidade,
		InscricaoRural: requestData.InscricaoRural,
		Observacao:     requestData.Observacao,
		IDProprietario: requestData.IDProprietario,
		Ativo:          true,
	}
	if err := f.FazendaUseCase.CreateFazenda(fazenda); err != nil {
		logger.Error("[FAZENDA] - CreateFazenda", err)
		c.JSON(http.StatusInternalServerError, response.FazendaErrorResponse{Error: domain.ErrCreateFazenda.Error()})
		return
	}
	c.JSON(http.StatusOK, response.FazendaSuccessResponse{Message: "Fazenda criada com sucesso!"})
}

func (f *FazendaController) GetFazendas(c *gin.Context) {
	idProprietario, err := strconv.Atoi(c.Param("idProprietario"))
	if err != nil || idProprietario <= 0 {
		if err == nil {
			err = errors.New("ID do proprietário inválido")
		}
		logger.Error("[FAZENDA] - GetFazendas", err)
		c.JSON(http.StatusBadRequest, response.FazendaErrorResponse{Error: "ID do proprietário inválido"})
		return
	}

	fazendas, err := f.FazendaUseCase.GetFazendas(idProprietario)
	if err != nil {
		logger.Error("[FAZENDA] - GetFazendas", err)
		c.JSON(http.StatusInternalServerError, response.FazendaErrorResponse{Error: err.Error()})
		return
	}
	if fazendas == nil || len(*fazendas) == 0 {
		c.JSON(http.StatusOK, response.GetFazendasResponse{Fazendas: []response.FazendaDataResponse{}})
		return
	}

	fazendasData := make([]response.FazendaDataResponse, len(*fazendas))
	for i, fazenda := range *fazendas {
		fazendasData[i] = response.FazendaDataResponse{
			ID:             fazenda.ID,
			Nome:           fazenda.Nome,
			Cidade:         fazenda.Cidade,
			InscricaoRural: fazenda.InscricaoRural,
			Observacao:     fazenda.Observacao,
			IDProprietario: fazenda.IDProprietario,
			Ativo:          fazenda.Ativo,
		}
	}
	c.JSON(http.StatusOK, response.GetFazendasResponse{Fazendas: fazendasData})
}

func (f *FazendaController) UpdateFazenda(c *gin.Context) {
	requestData := &request.FazendaUpdateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		logger.Error("[FAZENDA] - UpdateFazenda", err)
		c.JSON(http.StatusBadRequest, response.FazendaErrorResponse{Error: domain.ErrUpdateFazenda.Error()})
		return
	}

	fazenda := &domain.Fazenda{
		ID:             requestData.ID,
		Nome:           requestData.Nome,
		Cidade:         requestData.Cidade,
		InscricaoRural: requestData.InscricaoRural,
		Observacao:     requestData.Observacao,
	}
	if err := f.FazendaUseCase.UpdateFazenda(fazenda); err != nil {
		logger.Error("[FAZENDA] - UpdateFazenda", err)
		c.JSON(http.StatusInternalServerError, response.FazendaErrorResponse{Error: domain.ErrUpdateFazenda.Error()})
		return
	}
	c.JSON(http.StatusOK, response.FazendaSuccessResponse{Message: "Fazenda atualizada com sucesso!"})
}

func (f *FazendaController) ActivateFazenda(c *gin.Context) {
	id, ok := fazendaID(c, domain.ErrActivateFazenda)
	if !ok {
		return
	}
	if err := f.FazendaUseCase.ActivateFazenda(id); err != nil {
		logger.Error("[FAZENDA] - ActivateFazenda", err)
		if errors.Is(err, domain.ErrProprietarioInativo) {
			c.JSON(http.StatusConflict, response.FazendaErrorResponse{Error: domain.ErrProprietarioInativo.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, response.FazendaErrorResponse{Error: domain.ErrActivateFazenda.Error()})
		return
	}
	c.JSON(http.StatusOK, response.FazendaSuccessResponse{Message: "Fazenda ativada com sucesso!"})
}

func (f *FazendaController) DeactivateFazenda(c *gin.Context) {
	id, ok := fazendaID(c, domain.ErrDeactivateFazenda)
	if !ok {
		return
	}
	if err := f.FazendaUseCase.DeactivateFazenda(id); err != nil {
		logger.Error("[FAZENDA] - DeactivateFazenda", err)
		c.JSON(http.StatusInternalServerError, response.FazendaErrorResponse{Error: domain.ErrDeactivateFazenda.Error()})
		return
	}
	c.JSON(http.StatusOK, response.FazendaSuccessResponse{Message: "Fazenda desativada com sucesso!"})
}

func fazendaID(c *gin.Context, responseError error) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		if err == nil {
			err = responseError
		}
		logger.Error("[FAZENDA] - ID inválido", err)
		c.JSON(http.StatusBadRequest, response.FazendaErrorResponse{Error: responseError.Error()})
		return 0, false
	}
	return id, true
}
