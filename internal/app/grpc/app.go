package grpcapp

import (
	"auth-service/internal/services/auth"
	AuthServices "auth-service/internal/services/grpc-server/auth-service"
	"auth-service/internal/services/validator"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

// New creates new gRPC server application
func New(log *slog.Logger, port int, authApp *auth.Service, validator *validator.Validator) *App {
	gRPCServer := grpc.NewServer()
	AuthServices.Register(gRPCServer, log, authApp, validator)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun runs gRPC server and panics if any error occurs
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server
func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		log.Info("gRPC server start..", slog.String("addr", l.Addr().String()))
		if err := a.gRPCServer.Serve(l); err != nil {
			log.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}()

	return nil
}

// Stop stops gRPC server
func (a *App) Stop() {
	const op = "grpcapp.Stop"
	log := a.log.With(
		slog.String("op", op),
	)
	log.Info("graceful stopping gRPC server...")
	a.gRPCServer.GracefulStop()
}
