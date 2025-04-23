package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"auth-service/internal/services/auth"
	"google.golang.org/grpc"
	"log/slog"
)

type serverAPI struct {
	log     *slog.Logger
	authApp *auth.Service
	apiAuthServices.UnimplementedAuthServiceServer
}

// Register регистрирует GRPC-сервис
func Register(
	gRPC *grpc.Server,
	log *slog.Logger,
	authApp *auth.Service,
) {
	apiAuthServices.RegisterAuthServiceServer(gRPC, &serverAPI{log: log, authApp: authApp})
}
