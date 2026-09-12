package route

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/gin-gonic/gin"
)

func initRoutesv1(
	r *gin.RouterGroup,
	userController *controller.UserController,
	proprietarioController *controller.ProprietarioController,
	fazendaController *controller.FazendaController,
) {
	// Rotas v1
	v1 := r.Group("/v1")

	// Rotas de Usuario
	{
		v1.POST(
			"/user/login",
			userController.Login,
		)

		v1.POST(
			"/user/logout",
			userController.Logout,
		)

		v1.POST(
			"/user",
			userController.CreateUser,
		)

		v1.GET(
			"/users",
			userController.GetUsers,
		)

		v1.PUT(
			"/user",
			userController.UpdateUser,
		)

		v1.PUT(
			"/user/activate/:id",
			userController.ActivateUser,
		)

		v1.PUT(
			"/user/deactivate/:id",
			userController.DeactivateUser,
		)
	}

	// Rotas de Proprietario
	{
		v1.POST(
			"/proprietario",
			proprietarioController.CreateProprietario,
		)
		v1.GET(
			"/proprietarios",
			proprietarioController.GetProprietarios,
		)
		v1.PUT(
			"/proprietario",
			proprietarioController.UpdateProprietario,
		)
		v1.PUT(
			"/proprietario/activate/:id",
			proprietarioController.ActivateProprietario,
		)
		v1.PUT(
			"/proprietario/deactivate/:id",
			proprietarioController.DeactivateProprietario,
		)
	}

	// Rotas de Fazenda
	{
		v1.POST(
			"/fazenda",
			fazendaController.CreateFazenda,
		)
		v1.GET(
			"/fazendas/:idProprietario",
			fazendaController.GetFazendas,
		)
		v1.PUT(
			"/fazenda",
			fazendaController.UpdateFazenda,
		)
		v1.PUT(
			"/fazenda/activate/:id",
			fazendaController.ActivateFazenda,
		)
		v1.PUT(
			"/fazenda/deactivate/:id",
			fazendaController.DeactivateFazenda,
		)
	}
}
