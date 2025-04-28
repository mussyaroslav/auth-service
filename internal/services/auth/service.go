package auth

import (
	"auth-service/config"
	"auth-service/internal/models"
	"auth-service/pkg/logger"
	pgClient "auth-service/pkg/storage/pg-client"
	"log/slog"
	"os"
	"strconv"
)

type Service struct {
	log *slog.Logger
	cfg *config.Config
}

func New(log *slog.Logger, cfg *config.Config) *Service {
	// создаем postgres клиента auth
	db, err := pgClient.NewDB(&cfg.Storage)
	if err != nil {
		log.Warn("Failed connection to the auth DB. Check config.yaml!")
		os.Exit(2)
	}
	log.Info("DB auth client ready",
		slog.String("address", cfg.Storage.Host+":"+strconv.Itoa(cfg.Storage.Port)),
	)
	models.SetDB(db)

	return &Service{log: log.With("proc", "auth"), cfg: cfg}
}

// Start запускает службы
func (s *Service) Start() {}

// Close закрывает службы
func (s *Service) Close() {
	if models.GetDB() != nil {
		err := models.CloseDB()
		if err != nil {
			s.log.Error("Failed to close DB", logger.Err(err))
		}
		s.log.Info("DB client closed")
	}

	s.log.Info("Service is stopped")
}
