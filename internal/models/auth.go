package models

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	dbTimeOut = 10 * time.Second

	roleReader = 1
)

// AuthRequest предназначена для объединения данных, получаемых во время регистрации
type AuthRequest struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}

type AuthResponse struct {
	JWTToken string `json:"jwt_token"`
}

type User struct {
	UserId       uuid.UUID `db:"user_id" json:"user_id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // Не включаем в JSON
}

// ---------------------------------------------------------------------------------------------------------------------

// CreateUser создает нового пользователя в базе данных с использованием sqlx
func CreateUser(ctx context.Context, userID uuid.UUID, email, passwordHash string) (*User, error) {
	// Текущее время для полей created_at и updated_at
	now := time.Now()

	// Начинаем транзакцию
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Ошибка при начале транзакции: %v", err)
	}
	defer tx.Rollback() // Откат в случае ошибки

	// SQL запрос для вставки нового пользователя
	query := `
		INSERT INTO auth.users (user_id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, username, email
	`

	// Создаем объект пользователя для возврата
	user := new(User)

	// Выполняем запрос с использованием sqlx
	err = tx.QueryRowxContext(ctx, query, userID, email, passwordHash, now, now).
		StructScan(user)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности (дублирование email)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, status.Error(codes.AlreadyExists, "Пользователь с таким email уже существует")
		}

		// Другие ошибки базы данных
		return nil, status.Errorf(codes.Internal, "Ошибка создания пользователя: %v", err)
	}

	// Назначаем роль "reader" напрямую
	_, err = tx.ExecContext(ctx, `
		INSERT INTO auth.user_roles (user_id, role_id)
		VALUES ($1, $2)
	`, userID, roleReader)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Ошибка при назначении роли по умолчанию: %v", err)
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "Ошибка при фиксации транзакции: %v", err)
	}

	return user, nil
}

// GetUserRoles возвращает все роли пользователя по его ID
func GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeOut)
	defer cancel()

	query := `
		SELECT r.role_name
		FROM auth.user_roles ur
		JOIN auth.roles r ON ur.role_id = r.role_id
		WHERE ur.user_id = $1
	`

	var roles []string
	err := db.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Ошибка при получении ролей пользователя: %v", err)
	}

	return roles, nil
}

// GetUserByEmail получает пользователя из базы данных по email
func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT user_id, username, email, password_hash
		FROM auth.users
		WHERE email = $1
	`

	user := new(User)

	err := db.GetContext(ctx, user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "пользователь не найден")
		}
		return nil, status.Errorf(codes.Internal, "ошибка при получении пользователя: %v", err)
	}

	return user, nil
}
