package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"auth-service/internal/models"
	"context"
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
	s.log.Debug("register received")

	if err := s.validator.ValidateRegisterRequest(req.Email, req.Password); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeOutAuth)
	defer cancel()

	reqRegister := &models.RegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	rsp, err := s.authApp.Register(ctx, reqRegister)
	if err != nil {
		return nil, err
	}

	return &apiAuthServices.RegisterResponse{
		UserId:   rsp.UserID.String(),
		JwtToken: rsp.JWTToken,
		Error:    nil,
	}, nil
}
