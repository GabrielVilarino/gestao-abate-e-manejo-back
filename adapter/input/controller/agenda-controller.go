package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/request"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/model/response"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/input"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
	"github.com/gin-gonic/gin"
)

type AgendaController struct {
	AgendaUseCase input.AgendaUseCase
}

func NewAgendaController(agendaUseCase input.AgendaUseCase) *AgendaController {
	return &AgendaController{AgendaUseCase: agendaUseCase}
}

func (a *AgendaController) CreateAgenda(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	var req request.AgendaCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: err.Error()})
		return
	}
	dataHora, err := parseAgendaDateTime(req.DataHora)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrHorarioAgendaInvalido.Error()})
		return
	}
	agenda := &domain.Agenda{UserID: user.ID, FazendaID: req.FazendaID, DataHora: dataHora, Observacao: req.Observacao}
	if err := a.AgendaUseCase.CreateAgenda(agenda); err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.AgendaSuccessResponse{ID: agenda.ID, Message: "Agendamento criado com sucesso!"})
}

func (a *AgendaController) GetAgenda(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	id, valid := agendaID(c)
	if !valid {
		return
	}
	agenda, err := a.AgendaUseCase.GetAgenda(id, user.ID)
	if err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusOK, agendaResponse(*agenda))
}

func (a *AgendaController) FindAgendas(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	filter := domain.AgendaFilter{UserID: user.ID, Pagina: 1, Limite: 20}
	var err error
	if raw := c.Query("fazenda_id"); raw != "" {
		value, parseErr := strconv.Atoi(raw)
		if parseErr != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrFazendaAgendaInvalida.Error()})
			return
		}
		filter.FazendaID = &value
	}
	if raw := c.Query("data_inicio"); raw != "" {
		value, parseErr := parseAgendaDateTime(raw)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrHorarioAgendaInvalido.Error()})
			return
		}
		filter.DataInicio = &value
	}
	if raw := c.Query("data_fim"); raw != "" {
		value, parseErr := parseAgendaDateTime(raw)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrHorarioAgendaInvalido.Error()})
			return
		}
		filter.DataFim = &value
	}
	if raw := c.Query("pagina"); raw != "" {
		filter.Pagina, err = strconv.Atoi(raw)
		if err != nil || filter.Pagina <= 0 {
			c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: "Página inválida"})
			return
		}
	}
	if raw := c.Query("limite"); raw != "" {
		filter.Limite, err = strconv.Atoi(raw)
		if err != nil || filter.Limite <= 0 || filter.Limite > 100 {
			c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: "Limite inválido"})
			return
		}
	}
	agendas, total, err := a.AgendaUseCase.FindAgenda(filter)
	if err != nil {
		agendaError(c, err)
		return
	}
	items := make([]response.AgendaDataResponse, 0, len(*agendas))
	for _, agenda := range *agendas {
		items = append(items, agendaResponse(agenda))
	}
	c.JSON(http.StatusOK, response.GetAgendasResponse{Agendas: items, Pagina: filter.Pagina, Limite: filter.Limite, Total: total})
}

func (a *AgendaController) UpdateAgenda(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	id, valid := agendaID(c)
	if !valid {
		return
	}
	var req request.AgendaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: err.Error()})
		return
	}
	dataHora, err := parseAgendaDateTime(req.DataHora)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrHorarioAgendaInvalido.Error()})
		return
	}
	agenda := &domain.Agenda{ID: id, UserID: user.ID, FazendaID: req.FazendaID, DataHora: dataHora, Observacao: req.Observacao}
	if err := a.AgendaUseCase.UpdateAgenda(agenda); err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.AgendaSuccessResponse{Message: "Agendamento atualizado com sucesso!"})
}

func (a *AgendaController) DeleteAgenda(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	id, valid := agendaID(c)
	if !valid {
		return
	}
	if err := a.AgendaUseCase.DeleteAgenda(id, user.ID); err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.AgendaSuccessResponse{Message: "Agendamento excluído com sucesso!"})
}

func (a *AgendaController) CreatePushSubscription(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	var req request.PushSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: domain.ErrAssinaturaPushInvalida.Error()})
		return
	}
	subscription := &domain.PushSubscription{UserID: user.ID, Endpoint: req.Endpoint, P256DH: req.Keys.P256DH, Auth: req.Keys.Auth}
	if req.ExpirationTime != nil {
		expiresAt := time.UnixMilli(*req.ExpirationTime).UTC()
		subscription.ExpiresAt = &expiresAt
	}
	if err := a.AgendaUseCase.CreatePushSubscription(subscription); err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.PushSubscriptionDataResponse{ID: subscription.ID, Message: "Assinatura registrada com sucesso!"})
}

func (a *AgendaController) DeletePushSubscription(c *gin.Context) {
	user, ok := currentAgendaUser(c)
	if !ok {
		return
	}
	id, valid := agendaID(c)
	if !valid {
		return
	}
	if err := a.AgendaUseCase.DeletePushSubscription(id, user.ID); err != nil {
		agendaError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.AgendaSuccessResponse{Message: "Assinatura removida com sucesso!"})
}

func currentAgendaUser(c *gin.Context) (*domain.User, bool) {
	value, exists := c.Get("user")
	user, ok := value.(*domain.User)
	if !exists || !ok || user == nil {
		c.JSON(http.StatusUnauthorized, response.AgendaErrorResponse{Error: "Autenticação obrigatória"})
		return nil, false
	}
	return user, true
}

func agendaID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.AgendaErrorResponse{Error: "ID inválido"})
		return 0, false
	}
	return id, true
}

func parseAgendaDateTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func agendaResponse(agenda domain.Agenda) response.AgendaDataResponse {
	return response.AgendaDataResponse{ID: agenda.ID, FazendaID: agenda.FazendaID, DataHora: agenda.DataHora.UTC().Format(time.RFC3339), Observacao: agenda.Observacao}
}

func agendaError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, domain.ErrAgendaNaoEncontrada) || errors.Is(err, domain.ErrAssinaturaNaoEncontrada) {
		status = http.StatusNotFound
	} else if errors.Is(err, domain.ErrAgendaEmProcessamento) {
		status = http.StatusConflict
	} else if errors.Is(err, domain.ErrPeriodoAgendaInvalido) || errors.Is(err, domain.ErrHorarioAgendaInvalido) || errors.Is(err, domain.ErrFazendaAgendaInvalida) || errors.Is(err, domain.ErrAssinaturaPushInvalida) {
		status = http.StatusBadRequest
	}
	logger.Error("[AGENDA] - operação", err)
	c.JSON(status, response.AgendaErrorResponse{Error: err.Error()})
}
