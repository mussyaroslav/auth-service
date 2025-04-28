package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"auth-service/internal/services/auth"
	"auth-service/internal/services/validator"
	"google.golang.org/grpc"
	"log/slog"
)

type serverAPI struct {
	log       *slog.Logger
	authApp   *auth.Service
	validator *validator.Validator
	apiAuthServices.UnimplementedAuthServiceServer
}

// Register регистрирует GRPC-сервис
func Register(
	gRPC *grpc.Server,
	log *slog.Logger,
	authApp *auth.Service,
	validator *validator.Validator,
) {
	apiAuthServices.RegisterAuthServiceServer(gRPC, &serverAPI{log: log.With("proc", "gRPC server"), authApp: authApp, validator: validator})
}
