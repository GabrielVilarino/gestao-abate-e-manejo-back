package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/request"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/response"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/input"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
	"github.com/gin-gonic/gin"
)

const (
	dateLayout       = "2006-01-02"
	maxFotoUpload    = 10 << 20
	maxMultipartBody = maxFotoUpload + (2 << 20)
)

type AbateController struct {
	AbateUseCase input.AbateUseCase
}

func NewAbateController(abateUseCase input.AbateUseCase) *AbateController {
	return &AbateController{AbateUseCase: abateUseCase}
}

func (a *AbateController) CreateAbate(c *gin.Context) {
	requestData := &request.AbateCreateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	abate, err := requestToAbate(*requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: "data_abate deve usar o formato AAAA-MM-DD"})
		return
	}
	if err := a.AbateUseCase.CreateAbate(abate); err != nil {
		logger.Error("[ABATE] - CreateAbate", err)
		c.JSON(http.StatusInternalServerError, response.AbateErrorResponse{Error: "Erro ao criar abate"})
		return
	}
	c.JSON(http.StatusCreated, response.AbateSuccessResponse{ID: abate.ID, Message: "Abate criado com sucesso!"})
}

func (a *AbateController) FindAbateByID(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	abate, err := a.AbateUseCase.FindAbateByID(id)
	if err != nil {
		a.internalError(c, "FindAbateByID", err)
		return
	}
	if abate == nil {
		c.JSON(http.StatusNotFound, response.AbateErrorResponse{Error: domain.ErrAbateNaoEncontrado.Error()})
		return
	}
	c.JSON(http.StatusOK, abateResponse(*abate))
}

func (a *AbateController) FindAbates(c *gin.Context) {
	filtro := domain.FiltroAbate{}
	var err error
	if filtro.ProprietarioID, err = optionalPositiveInt(c, "proprietario_id"); err != nil {
		badFilter(c, err)
		return
	}
	if filtro.FazendaID, err = optionalPositiveInt(c, "fazenda_id"); err != nil {
		badFilter(c, err)
		return
	}
	if filtro.NumeroLote, err = optionalPositiveInt(c, "numero_lote"); err != nil {
		badFilter(c, err)
		return
	}
	if filtro.DataInicio, err = optionalDate(c, "data_inicio"); err != nil {
		badFilter(c, err)
		return
	}
	if filtro.DataFim, err = optionalDate(c, "data_fim"); err != nil {
		badFilter(c, err)
		return
	}
	pagina, err := optionalPositiveInt(c, "pagina")
	if err != nil {
		badFilter(c, err)
		return
	}
	limite, err := optionalPositiveInt(c, "limite")
	if err != nil {
		badFilter(c, err)
		return
	}
	filtro.Limit = 50
	paginaAtual := 1
	if pagina != nil {
		paginaAtual = *pagina
	}
	if limite != nil {
		if *limite > 100 {
			c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: "limite deve ser no máximo 100"})
			return
		}
		filtro.Limit = *limite
	}
	filtro.Offset = (paginaAtual - 1) * filtro.Limit
	if filtro.DataInicio != nil && filtro.DataFim != nil && filtro.DataInicio.After(*filtro.DataFim) {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: domain.ErrPeriodoInvalido.Error()})
		return
	}

	abates, err := a.AbateUseCase.FindAbates(filtro)
	if err != nil {
		if errors.Is(err, domain.ErrPeriodoInvalido) {
			c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
			return
		}
		a.internalError(c, "FindAbates", err)
		return
	}
	data := make([]response.AbateDataResponse, len(abates))
	for i, abate := range abates {
		data[i] = abateResponse(abate)
	}
	c.JSON(http.StatusOK, response.GetAbatesResponse{Abates: data, Pagina: paginaAtual, Limite: filtro.Limit})
}

func (a *AbateController) UpdateDadosGeraisAbate(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	requestData := &request.DadosGeraisAbateRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	dados, err := requestToDadosGerais(*requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: "data_abate deve usar o formato AAAA-MM-DD"})
		return
	}
	if err := a.AbateUseCase.UpdateDadosGeraisAbate(id, dados); err != nil {
		a.handleMutationError(c, "UpdateDadosGeraisAbate", err)
		return
	}
	c.JSON(http.StatusOK, response.AbateSuccessResponse{Message: "Dados gerais atualizados com sucesso!"})
}

func (a *AbateController) UpdateEtapaFazenda(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	requestData := &request.EtapaFazendaRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	if err := a.AbateUseCase.UpdateEtapaFazenda(id, requestToEtapaFazenda(*requestData)); err != nil {
		a.handleMutationError(c, "UpdateEtapaFazenda", err)
		return
	}
	c.JSON(http.StatusOK, response.AbateSuccessResponse{Message: "Etapa fazenda atualizada com sucesso!"})
}

func (a *AbateController) UpdateEtapaFrigorifico(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	requestData := &request.EtapaFrigorificoRequest{}
	if err := c.ShouldBindJSON(requestData); err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	if err := a.AbateUseCase.UpdateEtapaFrigorifico(id, requestToEtapaFrigorifico(*requestData)); err != nil {
		a.handleMutationError(c, "UpdateEtapaFrigorifico", err)
		return
	}
	c.JSON(http.StatusOK, response.AbateSuccessResponse{Message: "Etapa frigorífico atualizada com sucesso!"})
}

func (a *AbateController) DeleteAbate(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	if err := a.AbateUseCase.Delete(c.Request.Context(), id); err != nil {
		a.handleMutationError(c, "DeleteAbate", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *AbateController) UploadFotoAbate(c *gin.Context) {
	id, ok := positiveParam(c, "id")
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMultipartBody)
	fileHeader, err := c.FormFile("foto")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: "Envie a foto no campo foto"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		a.internalError(c, "UploadFotoAbate", err)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxFotoUpload+1))
	if err != nil {
		a.internalError(c, "UploadFotoAbate", err)
		return
	}
	foto, err := a.AbateUseCase.UploadFotoAbate(c.Request.Context(), id, c.Param("etapa"), fileHeader.Filename, data)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAbateNaoEncontrado):
			c.JSON(http.StatusNotFound, response.AbateErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrEtapaFotoInvalida),
			errors.Is(err, domain.ErrFormatoFotoInvalido),
			errors.Is(err, domain.ErrTamanhoFotoInvalido):
			c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
		default:
			a.internalError(c, "UploadFotoAbate", err)
		}
		return
	}
	c.JSON(http.StatusCreated, fotoResponse(*foto))
}

func (a *AbateController) DownloadFotoAbate(c *gin.Context) {
	id, ok := positiveParam(c, "fotoID")
	if !ok {
		return
	}
	foto, reader, err := a.AbateUseCase.DownloadFotoAbate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrFotoNaoEncontrada) {
			c.JSON(http.StatusNotFound, response.AbateErrorResponse{Error: err.Error()})
			return
		}
		a.internalError(c, "DownloadFotoAbate", err)
		return
	}
	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, foto.ContentType, reader, map[string]string{
		"Content-Disposition": mime.FormatMediaType("attachment", map[string]string{"filename": foto.NomeOriginal}),
	})
}

func (a *AbateController) DeleteFotoAbate(c *gin.Context) {
	id, ok := positiveParam(c, "fotoID")
	if !ok {
		return
	}
	if err := a.AbateUseCase.DeleteFotoAbate(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrFotoNaoEncontrada) {
			c.JSON(http.StatusNotFound, response.AbateErrorResponse{Error: err.Error()})
			return
		}
		a.internalError(c, "DeleteFotoAbate", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *AbateController) handleMutationError(c *gin.Context, operation string, err error) {
	if errors.Is(err, domain.ErrAbateNaoEncontrado) {
		c.JSON(http.StatusNotFound, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, domain.ErrDadosGeraisComFotos) {
		c.JSON(http.StatusConflict, response.AbateErrorResponse{Error: err.Error()})
		return
	}
	a.internalError(c, operation, err)
}

func (a *AbateController) internalError(c *gin.Context, operation string, err error) {
	logger.Error("[ABATE] - "+operation, err)
	c.JSON(http.StatusInternalServerError, response.AbateErrorResponse{Error: "Erro interno ao processar abate"})
}

func positiveParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: name + " inválido"})
		return 0, false
	}
	return id, true
}

func optionalPositiveInt(c *gin.Context, name string) (*int, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return nil, errors.New(name + " deve ser um inteiro positivo")
	}
	return &parsed, nil
}

func optionalDate(c *gin.Context, name string) (*time.Time, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil, errors.New(name + " deve usar o formato AAAA-MM-DD")
	}
	return &parsed, nil
}

func badFilter(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, response.AbateErrorResponse{Error: err.Error()})
}

func requestToAbate(value request.AbateCreateRequest) (*domain.Abate, error) {
	dados, err := requestToDadosGerais(value.DadosGeraisAbate)
	if err != nil {
		return nil, err
	}
	return &domain.Abate{
		DadosGeraisAbate: dados,
		EtapaFazenda:     requestToEtapaFazenda(value.EtapaFazenda),
		EtapaFrigorifico: requestToEtapaFrigorifico(value.EtapaFrigorifico),
	}, nil
}

func requestToDadosGerais(value request.DadosGeraisAbateRequest) (domain.DadosGeraisAbate, error) {
	data, err := time.Parse(dateLayout, value.DataAbate)
	if err != nil {
		return domain.DadosGeraisAbate{}, err
	}
	return domain.DadosGeraisAbate{
		DataAbate: data, FazendaID: value.FazendaID, NumeroLote: value.NumeroLote,
		NomeFrigorifico: value.NomeFrigorifico, DistanciaFrigorifico: value.DistanciaFrigorifico,
		CategoriaAnimal: value.CategoriaAnimal, PrecoFunrural: value.PrecoFunrural,
		PrecoSemFunrural: value.PrecoSemFunrural,
	}, nil
}

func requestToEtapaFazenda(value request.EtapaFazendaRequest) domain.EtapaFazenda {
	etapa := domain.EtapaFazenda{PesoTotal: value.PesoTotal, QuantidadeAnimal: make([]domain.QtdDenticao, len(value.QuantidadeAnimal))}
	for i, item := range value.QuantidadeAnimal {
		etapa.QuantidadeAnimal[i] = domain.QtdDenticao{QtdDenticao: item.QtdDenticao, QtdAnimais: item.QtdAnimais}
	}
	return etapa
}

func requestToEtapaFrigorifico(value request.EtapaFrigorificoRequest) domain.EtapaFrigorifico {
	etapa := domain.EtapaFrigorifico{
		PesoTotal: value.PesoTotal, Balancao: value.Balancao,
		AcabamentoCarcaca:        make([]domain.AcabamentoCarcaca, len(value.AcabamentoCarcaca)),
		ClassificacaoFrigorifico: make([]domain.ClassificacaoFrigorifico, len(value.ClassificacaoFrigorifico)),
		DistribuicaoPeso:         make([]domain.DistribuicaoPeso, len(value.DistribuicaoPeso)),
	}
	for i, item := range value.AcabamentoCarcaca {
		etapa.AcabamentoCarcaca[i] = domain.AcabamentoCarcaca{Acabamento: item.Acabamento, QtdAnimais: item.QtdAnimais}
	}
	for i, item := range value.ClassificacaoFrigorifico {
		etapa.ClassificacaoFrigorifico[i] = domain.ClassificacaoFrigorifico{Classificacao: item.Classificacao, QtdAnimais: item.QtdAnimais}
	}
	for i, item := range value.DistribuicaoPeso {
		etapa.DistribuicaoPeso[i] = domain.DistribuicaoPeso{Classificacao: item.Classificacao, QtdAnimais: item.QtdAnimais, PesoTotal: item.PesoTotal}
	}
	return etapa
}

func abateResponse(value domain.Abate) response.AbateDataResponse {
	result := response.AbateDataResponse{
		ID: value.ID, ProprietarioID: value.ProprietarioID,
		NomeProprietario: value.NomeProprietario, NomeFazenda: value.NomeFazenda,
		DadosGerais: response.DadosGeraisAbateResponse{
			DataAbate: value.DadosGeraisAbate.DataAbate.Format(dateLayout),
			FazendaID: value.DadosGeraisAbate.FazendaID, NumeroLote: value.DadosGeraisAbate.NumeroLote,
			NomeFrigorifico:      value.DadosGeraisAbate.NomeFrigorifico,
			DistanciaFrigorifico: value.DadosGeraisAbate.DistanciaFrigorifico,
			CategoriaAnimal:      value.DadosGeraisAbate.CategoriaAnimal,
			PrecoFunrural:        value.DadosGeraisAbate.PrecoFunrural,
			PrecoSemFunrural:     value.DadosGeraisAbate.PrecoSemFunrural,
		},
		EtapaFazenda: response.EtapaFazendaResponse{
			PesoTotal:        value.EtapaFazenda.PesoTotal,
			QuantidadeAnimal: make([]response.QtdDenticaoResponse, len(value.EtapaFazenda.QuantidadeAnimal)),
			Fotos:            make([]response.FotoAbateResponse, len(value.EtapaFazenda.Fotos)),
		},
		EtapaFrigorifico: response.EtapaFrigorificoResponse{
			PesoTotal: value.EtapaFrigorifico.PesoTotal, Balancao: value.EtapaFrigorifico.Balancao,
			AcabamentoCarcaca:        make([]response.AcabamentoCarcacaResponse, len(value.EtapaFrigorifico.AcabamentoCarcaca)),
			ClassificacaoFrigorifico: make([]response.ClassificacaoFrigorificoResponse, len(value.EtapaFrigorifico.ClassificacaoFrigorifico)),
			DistribuicaoPeso:         make([]response.DistribuicaoPesoResponse, len(value.EtapaFrigorifico.DistribuicaoPeso)),
			Fotos:                    make([]response.FotoAbateResponse, len(value.EtapaFrigorifico.Fotos)),
		},
	}
	for i, item := range value.EtapaFazenda.QuantidadeAnimal {
		result.EtapaFazenda.QuantidadeAnimal[i] = response.QtdDenticaoResponse{QtdDenticao: item.QtdDenticao, QtdAnimais: item.QtdAnimais}
	}
	for i, foto := range value.EtapaFazenda.Fotos {
		result.EtapaFazenda.Fotos[i] = fotoResponse(foto)
	}
	for i, item := range value.EtapaFrigorifico.AcabamentoCarcaca {
		result.EtapaFrigorifico.AcabamentoCarcaca[i] = response.AcabamentoCarcacaResponse{Acabamento: item.Acabamento, QtdAnimais: item.QtdAnimais}
	}
	for i, item := range value.EtapaFrigorifico.ClassificacaoFrigorifico {
		result.EtapaFrigorifico.ClassificacaoFrigorifico[i] = response.ClassificacaoFrigorificoResponse{Classificacao: item.Classificacao, QtdAnimais: item.QtdAnimais}
	}
	for i, item := range value.EtapaFrigorifico.DistribuicaoPeso {
		result.EtapaFrigorifico.DistribuicaoPeso[i] = response.DistribuicaoPesoResponse{Classificacao: item.Classificacao, QtdAnimais: item.QtdAnimais, PesoTotal: item.PesoTotal}
	}
	for i, foto := range value.EtapaFrigorifico.Fotos {
		result.EtapaFrigorifico.Fotos[i] = fotoResponse(foto)
	}
	return result
}

func fotoResponse(value domain.FotoAbate) response.FotoAbateResponse {
	return response.FotoAbateResponse{
		ID: value.ID, Etapa: value.Etapa, NomeOriginal: value.NomeOriginal,
		ContentType: value.ContentType, Tamanho: value.Tamanho, SHA256: value.SHA256,
	}
}
