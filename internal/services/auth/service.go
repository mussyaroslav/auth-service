package auth

import (
	"auth-service/config"
	"auth-service/internal/models"
	pgClient "auth-service/pkg/storage/pg-client"
	"log/slog"
	"os"
	"strconv"
)

type Service struct {
	log *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) *Service {
	// создаем postgres клиента auth
	db, err := pgClient.NewDB(&cfg.Storage)
	if err != nil {
		log.Warn("Failed connection to the auth DB. Check config.yml!")
		os.Exit(2)
	}
	log.Info("DB auth client ready",
		slog.String("address", cfg.Storage.Host+":"+strconv.Itoa(cfg.Storage.Port)),
	)
	models.SetDB(db)

	return &Service{log: log}
}
