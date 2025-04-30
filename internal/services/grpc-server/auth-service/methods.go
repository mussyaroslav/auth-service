package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"auth-service/internal/models"
	"auth-service/pkg/logger"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	timeOutAuth = 30 * time.Second
)

// Ping получает пинок от других сервисов
func (s *serverAPI) Ping(
	_ context.Context,
	_ *apiAuthServices.PingRequest,
) (*apiAuthServices.PingResponse, error) {
	s.log.Debug("ping received")
	return &apiAuthServices.PingResponse{Ok: true}, nil
}

// Register регистрирует пользователя в системе
func (s *serverAPI) Register(
	ctx context.Context,
	req *apiAuthServices.RegisterRequest,
) (*apiAuthServices.RegisterResponse, error) {
	// Создаем логгер с хешированным email для повторного использования
	hashedEmail := s.authApp.HashEmail(req.Email)
	l := s.log.With("email_hash", hashedEmail, "op", "api_register")

	// Валидация запроса
	if err := s.validator.ValidateRegisterRequest(req.Email, req.Password); err != nil {
		l.Debug("ошибка валидации", logger.Err(err))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeOutAuth)
	defer cancel()

	reqRegister := &models.AuthRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	// Выполнение регистрации через сервис
	rsp, err := s.authApp.Register(ctx, reqRegister)
	if err != nil {
		l.Debug("ошибка регистрации", logger.Err(err))
		return nil, err
	}

	l.Debug("регистрация успешно завершена")
	return &apiAuthServices.RegisterResponse{
		JwtToken: rsp.JWTToken,
		Error:    nil,
	}, nil
}

// Login логинет пользователя в системе
func (s *serverAPI) Login(
	ctx context.Context,
	req *apiAuthServices.LoginRequest,
) (*apiAuthServices.LoginResponse, error) {
	// Создаем логгер с хешированным email для повторного использования
	hashedEmail := s.authApp.HashEmail(req.Email)
	l := s.log.With("email_hash", hashedEmail, "op", "api_login")

	l.Debug("попытка входа в систему")

	// Валидация входных данных
	if req.GetEmail() == "" {
		l.Debug("ошибка валидации: пустой email")
		return nil, status.Error(codes.InvalidArgument, "empty email")
	}
	if req.GetPassword() == "" {
		l.Debug("ошибка валидации: пустой пароль")
		return nil, status.Error(codes.InvalidArgument, "empty password")
	}

	// Установка таймаута для контекста
	ctx, cancel := context.WithTimeout(ctx, timeOutAuth)
	defer cancel()

	reqLogin := &models.AuthRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	// Выполнение входа через сервис
	rsp, err := s.authApp.Login(ctx, reqLogin)
	if err != nil {
		// Используем уровень Warn для неудачных попыток входа,
		// так как это может быть признаком попытки подбора пароля
		l.Warn("неудачная попытка входа", logger.Err(err))
		return nil, err
	}

	l.Info("успешный вход в систему")
	return &apiAuthServices.LoginResponse{
		JwtToken: rsp.JWTToken,
	}, nil
}
