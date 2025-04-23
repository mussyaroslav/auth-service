package app

import (
	"auth-service/config"
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/services/auth"
	"log/slog"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	authApp := auth.New(log, cfg)
	grpcApp := grpcapp.New(log, cfg.GRPC.Port, authApp)

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
