package validator

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"unicode"
)

// DefaultPasswordPolicy содержит стандартные требования к паролю
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	RequireUpper:   true,
	RequireLower:   true,
	RequireNumber:  true,
	RequireSpecial: true,
}

type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// ValidateRegisterRequest проверяет email и пароль
func (v *Validator) ValidateRegisterRequest(email, password string) error {
	if err := v.validateEmail(email); err != nil {
		return err
	}
	return v.validatePassword(password)
}

func (v *Validator) validateEmail(email string) error {
	if email == "" {
		return v.createError("email", "Email обязателен")
	}

	if !govalidator.IsEmail(email) {
		return v.createError("email", "Неверный формат email")
	}

	return nil
}

func (v *Validator) validatePassword(password string) error {
	if len(password) < v.passPolicy.MinLength {
		return v.createError("password",
			"Пароль должен содержать не менее %d символов", v.passPolicy.MinLength)
	}

	if v.passPolicy.RequireUpper && !govalidator.HasUpperCase(password) {
		return v.createError("password", "Пароль должен содержать заглавную букву")
	}

	if v.passPolicy.RequireLower && !govalidator.HasLowerCase(password) {
		return v.createError("password", "Пароль должен содержать строчную букву")
	}

	if v.passPolicy.RequireNumber && !hasNumber(password) {
		return v.createError("password", "Пароль должен содержать цифру")
	}

	if v.passPolicy.RequireSpecial && !hasSpecial(password) {
		return v.createError("password", "Пароль должен содержать специальный символ")
	}

	return nil
}

func (v *Validator) createError(field, format string, args ...interface{}) error {
	st := status.New(codes.InvalidArgument, "Ошибка валидации")
	desc := fmt.Sprintf(format, args...)
	st, _ = st.WithDetails(&errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: desc,
	})
	return st.Err()
}

// Вспомогательные функции для проверки пароля
func hasNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecial(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
