package validator

type Validator struct {
	passPolicy PasswordPolicy
}

func New() *Validator {
	return &Validator{
		passPolicy: DefaultPasswordPolicy,
	}
}
