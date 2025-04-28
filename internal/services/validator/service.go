package validator

import "log/slog"

type Validator struct {
	log        *slog.Logger
	passPolicy PasswordPolicy
}

func New(log *slog.Logger) *Validator {
	return &Validator{
		log:        log,
		passPolicy: DefaultPasswordPolicy,
	}
}
