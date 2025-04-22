package models

import (
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// SetDB устанавливает соединение с базой данных
func SetDB(newDB *sqlx.DB) {
	db = newDB
}
