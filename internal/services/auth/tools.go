package auth

import (
	"auth-service/internal/models"
	"auth-service/pkg/lib"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

// HashPassword хэширует пароль
func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateToken создает jwt token
func (s *Service) CreateToken(email string, userID uuid.UUID) (string, error) {
	userRoles, err := models.GetUserRoles(context.Background(), userID)
	if err != nil {
		return "", err
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     email,
		"user_id": userID.String(),
		"iss":     "auth-service",
		"aud":     userRoles,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	// Sign the token with the secret key
	tokenString, err := claims.SignedString([]byte(s.cfg.Cert.Jwt))
	if err != nil {
		return "", err
	}

	s.log.Info("Generate Jwt-token",
		slog.String("userID", userID.String()),
		slog.String("email", email),
		slog.String("token", lib.MaskedText(tokenString, 3)),
	)

	return tokenString, nil
}
