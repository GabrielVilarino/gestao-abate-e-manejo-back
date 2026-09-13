package route

import (
	"net/http"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/middleware"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
	"github.com/gin-gonic/gin"
)

func InitRoutes(
	r *gin.Engine,
	userController *controller.UserController,
	proprietarioController *controller.ProprietarioController,
	fazendaController *controller.FazendaController,
	abateController *controller.AbateController,
	agendaController *controller.AgendaController,
	tokenPort output.TokenPort,
	userPort output.UserPort,
) {

	api := r.Group("/api")

	// Rota de saúde para monitoramento
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	auth := middleware.Authenticate(tokenPort, userPort)
	admin := middleware.AuthorizeRoles(domain.RoleAdmin)
	adminOrUser := middleware.AuthorizeRoles(domain.RoleAdmin, domain.RoleUser)
	initRoutesv1(api, userController, proprietarioController, fazendaController, abateController, agendaController, auth, admin, adminOrUser)
}
