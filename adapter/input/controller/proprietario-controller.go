package controller

import (
	"net/http"
	"strconv"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/request"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/response"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/input"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
	"github.com/gin-gonic/gin"
)

type ProprietarioController struct {
	ProprietarioUseCase input.ProprietarioUseCase
}

func NewProprietarioController(proprietarioUseCase input.ProprietarioUseCase) *ProprietarioController {
	return &ProprietarioController{ProprietarioUseCase: proprietarioUseCase}
}

func (p *ProprietarioController) CreateProprietario(c *gin.Context) {
	requestData := &request.ProprietarioCreateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.ProprietarioErrorResponse{Error: err.Error()})
		return
	}

	proprietario := &domain.Proprietario{
		Nome:       requestData.Nome,
		CPF:        requestData.CPF,
		Observacao: requestData.Observacao,
		Ativo:      true,
	}
	if err := p.ProprietarioUseCase.CreateProprietario(proprietario); err != nil {
		logger.Error("[PROPRIETARIO] - CreateProprietario", err)
		c.JSON(http.StatusInternalServerError, response.ProprietarioErrorResponse{Error: domain.ErrCreateProprietario.Error()})
		return
	}

	c.JSON(http.StatusOK, response.ProprietarioSuccessResponse{Message: "Proprietário criado com sucesso!"})
}

func (p *ProprietarioController) GetProprietarios(c *gin.Context) {
	proprietarios, err := p.ProprietarioUseCase.GetProprietarios()
	if err != nil {
		logger.Error("[PROPRIETARIO] - GetProprietarios", err)
		c.JSON(http.StatusInternalServerError, response.ProprietarioErrorResponse{Error: err.Error()})
		return
	}

	if proprietarios == nil || len(*proprietarios) == 0 {
		c.JSON(http.StatusOK, response.GetProprietariosResponse{Proprietarios: []response.ProprietarioDataResponse{}})
		return
	}

	proprietariosData := make([]response.ProprietarioDataResponse, len(*proprietarios))
	for i, proprietario := range *proprietarios {
		proprietariosData[i] = response.ProprietarioDataResponse{
			ID:         proprietario.ID,
			Nome:       proprietario.Nome,
			CPF:        proprietario.CPF,
			Observacao: proprietario.Observacao,
			Ativo:      proprietario.Ativo,
		}
	}
	c.JSON(http.StatusOK, response.GetProprietariosResponse{Proprietarios: proprietariosData})
}

func (p *ProprietarioController) UpdateProprietario(c *gin.Context) {
	requestData := &request.ProprietarioUpdateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		logger.Error("[PROPRIETARIO] - UpdateProprietario", err)
		c.JSON(http.StatusBadRequest, response.ProprietarioErrorResponse{Error: domain.ErrUpdateProprietario.Error()})
		return
	}

	proprietario := &domain.Proprietario{
		ID:         requestData.ID,
		Nome:       requestData.Nome,
		CPF:        requestData.CPF,
		Observacao: requestData.Observacao,
	}
	if err := p.ProprietarioUseCase.UpdateProprietario(proprietario); err != nil {
		logger.Error("[PROPRIETARIO] - UpdateProprietario", err)
		c.JSON(http.StatusInternalServerError, response.ProprietarioErrorResponse{Error: domain.ErrUpdateProprietario.Error()})
		return
	}
	c.JSON(http.StatusOK, response.ProprietarioSuccessResponse{Message: "Proprietário atualizado com sucesso!"})
}

func (p *ProprietarioController) ActivateProprietario(c *gin.Context) {
	id, ok := proprietarioID(c, domain.ErrActivateProprietario)
	if !ok {
		return
	}
	if err := p.ProprietarioUseCase.ActivateProprietario(id); err != nil {
		logger.Error("[PROPRIETARIO] - ActivateProprietario", err)
		c.JSON(http.StatusInternalServerError, response.ProprietarioErrorResponse{Error: domain.ErrActivateProprietario.Error()})
		return
	}
	c.JSON(http.StatusOK, response.ProprietarioSuccessResponse{Message: "Proprietário ativado com sucesso!"})
}

func (p *ProprietarioController) DeactivateProprietario(c *gin.Context) {
	id, ok := proprietarioID(c, domain.ErrDeactivateProprietario)
	if !ok {
		return
	}
	if err := p.ProprietarioUseCase.DeactivateProprietario(id); err != nil {
		logger.Error("[PROPRIETARIO] - DeactivateProprietario", err)
		c.JSON(http.StatusInternalServerError, response.ProprietarioErrorResponse{Error: domain.ErrDeactivateProprietario.Error()})
		return
	}
	c.JSON(http.StatusOK, response.ProprietarioSuccessResponse{Message: "Proprietário desativado com sucesso!"})
}

func proprietarioID(c *gin.Context, responseError error) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		if err == nil {
			err = responseError
		}
		logger.Error("[PROPRIETARIO] - ID inválido", err)
		c.JSON(http.StatusBadRequest, response.ProprietarioErrorResponse{Error: responseError.Error()})
		return 0, false
	}
	return id, true
}
