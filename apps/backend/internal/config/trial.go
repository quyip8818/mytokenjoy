package config

// TrialConfig holds Trial/registration settings for SaaS free trial mode.
type TrialConfig struct {
	RegistrationEnabled bool  `env:"REGISTRATION_ENABLED" envDefault:"true"`
	TrialMemberLimit    int   `env:"TRIAL_MEMBER_LIMIT" envDefault:"50"`
	TrialCreditAmount   int64 `env:"TRIAL_CREDIT_AMOUNT" envDefault:"10000"`
}
