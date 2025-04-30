package auth

import (
	"auth-service/internal/models"
	"auth-service/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

// Register выполняет регистрацию пользователя и возвращает JWT токен
func (s *Service) Register(ctx context.Context, request *models.AuthRequest) (*models.AuthResponse, error) {
	// Создаем логгер с хешированным email для повторного использования
	hashedEmail := s.HashEmail(request.Email)
	l := s.log.With(slog.String("email_hash", hashedEmail), slog.String("op", "register"))

	l.Info("начало регистрации пользователя")

	// 1. Хеширование пароля
	hashedPwd, err := s.HashPassword(request.Password)
	if err != nil {
		l.Error("ошибка хеширования пароля", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	// 2. Создание записи пользователя
	userID := uuid.New()
	user, err := models.CreateUser(ctx, userID, request.Email, hashedPwd)
	if err != nil {
		l.Error("ошибка создания пользователя", logger.Err(err))
		return nil, err
	}

	// 3. Генерация JWT токена
	token, err := s.CreateToken(user.Email, user.UserId)
	if err != nil {
		l.Error("ошибка создания токена", logger.Err(err))
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	l.Info("пользователь успешно зарегистрирован")

	return &models.AuthResponse{
		JWTToken: token,
	}, nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT токен
func (s *Service) Login(ctx context.Context, request *models.AuthRequest) (*models.AuthResponse, error) {
	hashedEmail := s.HashEmail(request.Email)
	l := s.log.With(slog.String("email_hash", hashedEmail), slog.String("op", "login"))

	l.Info("попытка входа в систему")

	// 1. Получаем пользователя по email
	user, err := models.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Debug("пользователь не найден")
			return nil, status.Error(codes.NotFound, "пользователь не найден")
		}
		l.Error("ошибка при поиске пользователя", logger.Err(err))
		return nil, status.Errorf(codes.Internal, "ошибка при поиске пользователя: %v", err)
	}

	// 2. Проверяем пароль
	if !s.CheckPasswordHash(request.Password, user.PasswordHash) {
		l.Debug("неверный пароль")
		return nil, status.Error(codes.Unauthenticated, "неверный пароль")
	}

	// 3. Создаем JWT токен (роли будут получены внутри CreateToken)
	token, err := s.CreateToken(user.Email, user.UserId)
	if err != nil {
		l.Error("ошибка при создании токена", logger.Err(err))
		return nil, status.Errorf(codes.Internal, "ошибка при создании токена: %v", err)
	}

	l.Info("успешный вход в систему")
	return &models.AuthResponse{
		JWTToken: token,
	}, nil
}
