package route

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/gin-gonic/gin"
)

func initRoutesv1(
	r *gin.RouterGroup,
	userController *controller.UserController,
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
}
