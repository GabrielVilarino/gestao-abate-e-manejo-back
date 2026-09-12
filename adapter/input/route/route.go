package route

import (
	"net/http"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/gin-gonic/gin"
)

func InitRoutes(
	r *gin.Engine,
	userController *controller.UserController,
	proprietarioController *controller.ProprietarioController,
	fazendaController *controller.FazendaController,
) {

	api := r.Group("/api")

	// Rota de saúde para monitoramento
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	initRoutesv1(api, userController, proprietarioController, fazendaController)
}
