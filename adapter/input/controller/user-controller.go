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

type UserController struct {
	UserUseCase input.UserUseCase
}

func NewUserController(
	userUseCase input.UserUseCase,
) *UserController {
	return &UserController{
		UserUseCase: userUseCase,
	}
}

func (u *UserController) Login(c *gin.Context) {
	request := &request.UserLoginRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{Error: err.Error()})
		return
	}

	token, err := u.UserUseCase.Login(request.Email, request.Password)
	if err != nil {
		logger.Error("[USER] - Login", err)
		c.JSON(http.StatusUnauthorized, response.UserErrorResponse{Error: err.Error()})
		return
	}

	c.SetCookie(
		"auth",
		*token,
		259200,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Login realizado com sucesso!"})
}

func (u *UserController) Logout(c *gin.Context) {
	c.SetCookie(
		"auth",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso!"})
}

func (u *UserController) CreateUser(c *gin.Context) {
	request := &request.UserCreateRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{Error: err.Error()})
		return
	}

	user := &domain.User{
		Nome:     request.Nome,
		Email:    request.Email,
		Password: request.Password,
		Role:     request.Role,
		Ativo:    true,
	}

	if err := u.UserUseCase.CreateUser(user); err != nil {
		logger.Error("[USER] - CreateUser", err)
		c.JSON(http.StatusInternalServerError, response.UserErrorResponse{Error: domain.ErrCreateUser.Error()})
		return
	}

	c.JSON(http.StatusOK, response.UserSuccessResponse{Message: "Usuário criado com sucesso!"})
}

func (u *UserController) GetUsers(c *gin.Context) {
	users, err := u.UserUseCase.GetUsers()
	if err != nil {
		logger.Error("[USER] - GetUsers", err)
		c.JSON(http.StatusInternalServerError, response.UserErrorResponse{Error: err.Error()})
		return
	}

	if users == nil || len(*users) == 0 {
		c.JSON(http.StatusOK, response.GetUsersResponse{Users: []response.UserDataResponse{}})
		return
	}

	usersData := make([]response.UserDataResponse, len(*users))
	for i, user := range *users {
		usersData[i] = response.UserDataResponse{
			Nome:  user.Nome,
			Email: user.Email,
			Role:  user.Role,
			Ativo: user.Ativo,
		}
	}

	c.JSON(http.StatusOK, response.GetUsersResponse{Users: usersData})
}

func (u *UserController) UpdateUser(c *gin.Context) {
	request := &request.UserUpdateRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		logger.Error("[USER] - UpdateUser", err)
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{Error: domain.ErrUpdateUser.Error()})
		return
	}

	userDomain := &domain.User{
		ID:    request.ID,
		Nome:  request.Nome,
		Email: request.Email,
		Role:  request.Role,
	}

	if err := u.UserUseCase.UpdateUser(userDomain); err != nil {
		logger.Error("[USER] - UpdateUser", err)
		c.JSON(http.StatusInternalServerError, response.UserErrorResponse{Error: domain.ErrUpdateUser.Error()})
		return
	}

	c.JSON(http.StatusOK, response.UserSuccessResponse{Message: "Usuário atualizado com sucesso!"})
}

func (u *UserController) ActivateUser(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		logger.Error("[USER] - ActivateUser", errors.New("ID do usuário é obrigatório"))
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{Error: domain.ErrActivateUser.Error()})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("[USER] - ActivateUser", err)
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{
			Error: domain.ErrActivateUser.Error(),
		})
		return
	}

	if err := u.UserUseCase.ActivateUser(id); err != nil {
		logger.Error("[USER] - ActivateUser", err)
		c.JSON(http.StatusInternalServerError, response.UserErrorResponse{Error: domain.ErrActivateUser.Error()})
		return
	}
	c.JSON(http.StatusOK, response.UserSuccessResponse{Message: "Usuário ativado com sucesso!"})
}

func (u *UserController) DeactivateUser(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		logger.Error("[USER] - DeactivateUser", errors.New("ID do usuário é obrigatório"))
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{Error: domain.ErrDeactivateUser.Error()})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("[USER] - DeactivateUser", err)
		c.JSON(http.StatusBadRequest, response.UserErrorResponse{
			Error: domain.ErrDeactivateUser.Error(),
		})
		return
	}

	if err := u.UserUseCase.DeactivateUser(id); err != nil {
		logger.Error("[USER] - DeactivateUser", err)
		c.JSON(http.StatusInternalServerError, response.UserErrorResponse{Error: domain.ErrDeactivateUser.Error()})
		return
	}

	c.JSON(http.StatusOK, response.UserSuccessResponse{Message: "Usuário desativado com sucesso!"})
}
