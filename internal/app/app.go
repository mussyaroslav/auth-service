package app

import (
	"auth-service/config"
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/models"
	pgClient "auth-service/pkg/storage/pg-client"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	// создаем postgres клиента
	db, err := pgClient.NewDB(&cfg.Storage)
	if err != nil {
		fmt.Println("Failed connection to the DB. Check config.yml!")
		os.Exit(2)
	}
	log.Info("DB auth client ready",
		slog.String("address", cfg.Storage.Host+":"+strconv.Itoa(cfg.Storage.Port)),
	)
	models.SetDB(db)

	grpcApp := grpcapp.New(log, cfg.GRPC.Port, db)

	return &App{
		log:        log,
		GRPCServer: grpcApp,
	}
}

func (a *App) MustRun() {
	a.GRPCServer.MustRun()

	a.log.Info("Application is running")
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
}
