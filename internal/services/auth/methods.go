package auth

import (
	"auth-service/internal/models"
	"auth-service/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
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

	l.Debug("начало регистрации пользователя")

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
	token, err := s.CreateToken(user)
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

	l.Debug("попытка входа в систему")

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
	token, err := s.CreateToken(user)
	if err != nil {
		l.Error("ошибка при создании токена", logger.Err(err))
		return nil, status.Errorf(codes.Internal, "ошибка при создании токена: %v", err)
	}

	l.Info("успешный вход в систему")
	return &models.AuthResponse{
		JWTToken: token,
	}, nil
}

// VerifyToken проверяет JWT токен и извлекает данные пользователя
func (s *Service) VerifyToken(_ context.Context, tokenString string) (*models.TokenInfo, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что используется правильный алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Cert.Jwt), nil
	})

	// Обрабатываем ошибки парсинга
	if err != nil {
		s.log.Warn("Ошибка при парсинге токена", slog.String("ошибка", err.Error()))
		return nil, fmt.Errorf("недействительный токен: %w", err)
	}

	// Проверяем валидность токена
	if !token.Valid {
		s.log.Warn("Токен недействителен")
		return nil, errors.New("токен недействителен")
	}

	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.log.Warn("Не удалось извлечь claims из токена")
		return nil, errors.New("не удалось извлечь данные из токена")
	}

	// Проверяем issuer
	if iss, ok := claims["iss"].(string); !ok || iss != "auth-service" {
		s.log.Warn("Недействительный издатель токена", slog.String("издатель", fmt.Sprintf("%v", claims["iss"])))
		return nil, errors.New("недействительный издатель токена")
	}

	// Проверяем audience
	if aud, ok := claims["aud"].(string); !ok || aud != "chef-app-services" {
		s.log.Warn("Недействительная аудитория токена", slog.String("аудитория", fmt.Sprintf("%v", claims["aud"])))
		return nil, errors.New("недействительная аудитория токена")
	}

	// Извлекаем userId
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		s.log.Warn("Недействительный или отсутствующий ID пользователя в токене")
		return nil, errors.New("недействительный ID пользователя в токене")
	}

	// Извлекаем email
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		s.log.Warn("Недействительный или отсутствующий email в токене")
		return nil, errors.New("недействительный email в токене")
	}

	// Извлекаем роли
	var roles []string
	if rolesInterface, ok := claims["roles"]; ok {
		if rolesArray, ok := rolesInterface.([]interface{}); ok {
			for _, role := range rolesArray {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	// Логируем успешную проверку
	s.log.Info("Токен успешно проверен",
		slog.String("email", s.HashEmail(email)),
	)

	// Возвращаем информацию о токене
	return &models.TokenInfo{
		UserID:  userID,
		Email:   email,
		Roles:   roles,
		IsValid: true,
	}, nil
}
