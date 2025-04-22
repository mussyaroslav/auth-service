package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"google.golang.org/grpc"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type serverAPI struct {
	log *slog.Logger
	db  *sqlx.DB
	apiAuthServices.UnimplementedAuthServiceServer
}

// Register регистрирует GRPC-сервис
func Register(
	gRPC *grpc.Server,
	log *slog.Logger,
	db *sqlx.DB,
) {
	apiAuthServices.RegisterAuthServiceServer(gRPC, &serverAPI{log: log, db: db})
}
