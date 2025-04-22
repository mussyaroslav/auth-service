package pg_client

import (
	"auth-service/config"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL драйвер
)

// NewDB создает и возвращает новое соединение с базой данных, используя sqlx
func NewDB(cfg *config.StorageData) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable search_path=%s",
		cfg.Host,
		cfg.User,
		cfg.Pass,
		cfg.Database,
		cfg.Port,
		cfg.Schema,
	)

	// Создаем новое соединение с базой данных
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
