package app

import (
	"auth-service/config"
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/services/auth"
	"auth-service/internal/services/validator"
	"log/slog"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
	AuthApp    *auth.Service
}

func New(log *slog.Logger, cfg *config.Config) *App {
	authApp := auth.New(log, cfg)
	validatorApp := validator.New()
	grpcApp := grpcapp.New(log, cfg.GRPC.Port, authApp, validatorApp)

	return &App{
		log:        log,
		GRPCServer: grpcApp,
		AuthApp:    authApp,
	}
}

func (a *App) MustRun() {
	a.GRPCServer.MustRun()

	a.log.Info("Application is running")
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.AuthApp.Close()
	a.log.Info("Application is stopped")
}
