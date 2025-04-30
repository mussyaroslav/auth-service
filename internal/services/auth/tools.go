package auth

import (
	"auth-service/internal/models"
	"auth-service/pkg/lib"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
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
		slog.String("token", lib.MaskedText(tokenString, 3)),
	)

	return tokenString, nil
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
