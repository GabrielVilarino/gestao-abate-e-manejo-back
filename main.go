package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/route"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/notification"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/repository"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/security"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/output/storage"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
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
	tokenPort := security.NewTokenPort()
	userPort := repository.NewUserRepository(db)
	userController := initUserController(userPort, tokenPort)
	proprietarioController := initProprietarioController(db)
	fazendaController := initFazendaController(db)
	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	abateController, err := initAbateController(db, cleanupCtx)
	if err != nil {
		logger.Error("Erro ao configurar armazenamento R2", err)
		return
	}
	agendaController, err := initAgendaController(db, cleanupCtx)
	if err != nil {
		logger.Error("Erro ao configurar notificações de agenda", err)
		return
	}

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	route.InitRoutes(
		router,
		userController,
		proprietarioController,
		fazendaController,
		abateController,
		agendaController,
		tokenPort,
		userPort,
	)

	// Inicialização do Servidor
	if err := router.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		logger.Error("Erro ao iniciar o servidor", err)
		return
	}
}

func initAgendaController(db *sql.DB, workerCtx context.Context) (*controller.AgendaController, error) {
	agendaPort := repository.NewAgendaRepository(db)
	pushSender, err := notification.NewWebPushSender(
		os.Getenv("VAPID_PUBLIC_KEY"),
		os.Getenv("VAPID_PRIVATE_KEY"),
		os.Getenv("VAPID_SUBJECT"),
	)
	if err != nil {
		return nil, err
	}
	agendaService := service.NewAgendaNotificationService(agendaPort, agendaPort, pushSender)
	agendaService.StartNotificationWorker(workerCtx, time.Minute)
	return controller.NewAgendaController(agendaService), nil
}

func initAbateController(db *sql.DB, cleanupCtx context.Context) (*controller.AbateController, error) {
	storagePort, err := storage.NewR2Storage(
		os.Getenv("R2_BASE_URL"),
		os.Getenv("R2_ACCESS_KEY"),
		os.Getenv("R2_SECRET_KEY"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	if err != nil {
		return nil, err
	}
	abatePort := repository.NewAbateRepository(db)
	abateService := service.NewAbateService(abatePort, storagePort)
	abateService.StartStorageCleanupWorker(cleanupCtx, time.Minute)
	return controller.NewAbateController(abateService), nil
}

func initProprietarioController(db *sql.DB) *controller.ProprietarioController {
	proprietarioPort := repository.NewProprietarioRepository(db)
	proprietarioService := service.NewProprietarioService(proprietarioPort)
	return controller.NewProprietarioController(proprietarioService)
}

func initFazendaController(db *sql.DB) *controller.FazendaController {
	fazendaPort := repository.NewFazendaRepository(db)
	fazendaService := service.NewFazendaService(fazendaPort)
	return controller.NewFazendaController(fazendaService)
}

func initUserController(
	userPort output.UserPort,
	tokenPort output.TokenPort,
) *controller.UserController {
	hashPort := security.NewHashPort()

	service := service.NewUserService(
		userPort,
		hashPort,
		tokenPort,
	)

	return controller.NewUserController(service)
}
