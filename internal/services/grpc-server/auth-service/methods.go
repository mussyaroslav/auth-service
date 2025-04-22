package auth_service

import (
	apiAuthServices "auth-service/generate/auth-service"
	"context"
)

// Ping получает пинок от других сервисов
func (s *serverAPI) Ping(
	_ context.Context,
	_ *apiAuthServices.PingRequest,
) (*apiAuthServices.PingResponse, error) {
	s.log.Debug("ping received")
	return &apiAuthServices.PingResponse{Ok: true}, nil
}
