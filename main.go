package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/route"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/repository"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/security"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/service"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/database"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/configuration/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	logger.Info("==> Iniciando Servidor <==")

	err := godotenv.Overload()
	if err != nil {
		logger.Error("Erro ao carregar o arquivo .env", err)
		return
	}

	db, err := database.Connect()
	if err != nil {
		logger.Error("Erro ao conectar com o banco de dados", err)
		return
	}
	defer db.Close()

	// Inicialização dos Controladores
	userController := initUserController(db)

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	route.InitRoutes(
		router,
		userController,
	)

	// Inicialização do Servidor
	if err := router.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		logger.Error("Erro ao iniciar o servidor", err)
		return
	}
}

func initUserController(
	db *sql.DB,
) *controller.UserController {
	userPort := repository.NewUserRepository(db)
	hashPort := security.NewHashPort()
	tokenPort := security.NewTokenPort()

	service := service.NewUserService(
		userPort,
		hashPort,
		tokenPort,
	)

	return controller.NewUserController(service)
}
