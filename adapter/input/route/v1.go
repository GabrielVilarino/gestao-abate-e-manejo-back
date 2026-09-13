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
	abateController *controller.AbateController,
	agendaController *controller.AgendaController,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
	adminOrUserMiddleware gin.HandlerFunc,
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
			authMiddleware,
			adminOrUserMiddleware,
			userController.Logout,
		)

		users := v1.Group("")
		users.Use(authMiddleware, adminMiddleware)
		users.POST(
			"/user",
			userController.CreateUser,
		)

		users.GET(
			"/users",
			userController.GetUsers,
		)

		users.PUT(
			"/user",
			userController.UpdateUser,
		)

		users.PUT(
			"/user/activate/:id",
			userController.ActivateUser,
		)

		users.PUT(
			"/user/deactivate/:id",
			userController.DeactivateUser,
		)
	}

	// Rotas de Proprietario
	{
		proprietarios := v1.Group("")
		proprietarios.Use(authMiddleware, adminOrUserMiddleware)
		proprietarios.POST(
			"/proprietario",
			proprietarioController.CreateProprietario,
		)
		proprietarios.GET(
			"/proprietarios",
			proprietarioController.GetProprietarios,
		)
		proprietarios.PUT(
			"/proprietario",
			proprietarioController.UpdateProprietario,
		)
		proprietarios.PUT(
			"/proprietario/activate/:id",
			proprietarioController.ActivateProprietario,
		)
		proprietarios.PUT(
			"/proprietario/deactivate/:id",
			proprietarioController.DeactivateProprietario,
		)
	}

	// Rotas de Fazenda
	{
		fazendas := v1.Group("")
		fazendas.Use(authMiddleware, adminOrUserMiddleware)
		fazendas.POST(
			"/fazenda",
			fazendaController.CreateFazenda,
		)
		fazendas.GET(
			"/fazendas/:idProprietario",
			fazendaController.GetFazendas,
		)
		fazendas.PUT(
			"/fazenda",
			fazendaController.UpdateFazenda,
		)
		fazendas.PUT(
			"/fazenda/activate/:id",
			fazendaController.ActivateFazenda,
		)
		fazendas.PUT(
			"/fazenda/deactivate/:id",
			fazendaController.DeactivateFazenda,
		)
	}

	// Rotas de Abate
	{
		abate := v1.Group("")
		abate.Use(authMiddleware, adminOrUserMiddleware)
		abate.POST(
			"/abate",
			abateController.CreateAbate,
		)
		abate.GET(
			"/abate/:id",
			abateController.FindAbateByID,
		)
		abate.GET(
			"/abates",
			abateController.FindAbates,
		)
		abate.PUT(
			"/abate/:id/dados-gerais",
			abateController.UpdateDadosGeraisAbate,
		)
		abate.PUT(
			"/abate/:id/etapa-fazenda",
			abateController.UpdateEtapaFazenda,
		)
		abate.PUT(
			"/abate/:id/etapa-frigorifico",
			abateController.UpdateEtapaFrigorifico,
		)
		abate.DELETE(
			"/abate/:id",
			adminMiddleware,
			abateController.DeleteAbate,
		)
		abate.POST(
			"/abate/:id/fotos/:etapa",
			abateController.UploadFotoAbate,
		)
		abate.GET(
			"/abate/fotos/:fotoID",
			abateController.DownloadFotoAbate,
		)
		abate.DELETE(
			"/abate/fotos/:fotoID",
			adminMiddleware,
			abateController.DeleteFotoAbate,
		)
	}

	// Rotas de Agenda e assinaturas Web Push
	{
		agenda := v1.Group("")
		agenda.Use(authMiddleware, adminOrUserMiddleware)
		agenda.POST(
			"/agenda",
			agendaController.CreateAgenda,
		)
		agenda.GET(
			"/agenda/:id",
			agendaController.GetAgenda,
		)
		agenda.GET(
			"/agendas",
			agendaController.FindAgendas,
		)
		agenda.PUT(
			"/agenda/:id",
			agendaController.UpdateAgenda,
		)
		agenda.DELETE(
			"/agenda/:id",
			adminMiddleware,
			agendaController.DeleteAgenda,
		)
		agenda.POST(
			"/push/subscriptions",
			agendaController.CreatePushSubscription,
		)
		agenda.DELETE(
			"/push/subscriptions/:id",
			adminMiddleware,
			agendaController.DeletePushSubscription,
		)
	}
}
