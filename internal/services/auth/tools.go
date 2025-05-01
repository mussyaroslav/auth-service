package auth

import (
	"auth-service/internal/models"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mussyaroslav/libs/helper"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strings"
	"time"
)

// HashPassword хэширует пароль
func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *Service) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashEmail возвращает безопасный хеш email для логирования
// Формат: первые_3_символа@хеш_домена
func (s *Service) HashEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid_email"
	}

	localPart := parts[0]
	domain := parts[1]

	// Хешируем домен с перцем
	domainHash := sha256.Sum256([]byte(domain + s.cfg.Cert.EmailPepper))
	encodedDomain := base64.URLEncoding.EncodeToString(domainHash[:])[:8]

	// Сохраняем часть информации для отладки
	var prefix string
	if len(localPart) > 3 {
		prefix = localPart[:3]
	} else if len(localPart) > 0 {
		prefix = localPart
	} else {
		prefix = "xxx"
	}

	return prefix + "@" + encodedDomain
}

// CreateToken создает jwt token
func (s *Service) CreateToken(user *models.User) (string, error) {
	userRoles, err := models.GetUserRoles(context.Background(), user.UserId)
	if err != nil {
		return "", err
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.UserId,
		"email": user.Email,
		"iss":   "auth-service",
		"aud":   "chef-app-services",
		"roles": userRoles,
		"exp":   time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})

	// Sign the token with the secret key
	tokenString, err := claims.SignedString([]byte(s.cfg.Cert.Jwt))
	if err != nil {
		return "", err
	}

	s.log.Info("Generate Jwt-token",
		slog.String("email", s.HashEmail(user.Email)),
		slog.String("token", helper.MaskedText(tokenString, 3)),
	)

	return tokenString, nil
}
