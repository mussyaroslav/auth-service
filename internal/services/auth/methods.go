package auth

import (
	"auth-service/internal/models"
	"auth-service/pkg/logger"
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

func (s *Service) Register(ctx context.Context, request *models.RegisterRequest) (*models.RegisterResponse, error) {
	s.log.Info("register user",
		slog.String("email", request.Email),
	)

	// 1. Хеширование пароля
	hashedPwd, err := s.HashPassword(request.Password)
	if err != nil {
		s.log.Error("Hash password error",
			slog.String("email", request.Email),
			logger.Err(err),
		)
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	// 2. Создание записи пользователя
	userID := uuid.New()
	user, err := models.CreateUser(ctx, userID, request.Email, hashedPwd)
	if err != nil {
		s.log.Error("CreateUser",
			slog.String("email", request.Email),
			logger.Err(err),
		)
		return nil, err
	}

	// 3. Генерация JWT токена
	token, err := s.CreateToken(request.Email, userID)
	if err != nil {
		s.log.Error("CreateToken",
			slog.String("email", request.Email),
			logger.Err(err),
		)
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	return &models.RegisterResponse{
		UserID:   user.UserId,
		JWTToken: token,
	}, nil
}
