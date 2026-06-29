package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	ProfileDemo = "demo"
	ProfileProd = "prod"
)

type Config struct {
	Port          string `env:"PORT" envDefault:"8080"`
	CORSOrigins   string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
	SimulateDelay bool   `env:"SIMULATE_DELAY" envDefault:"true"`
	DemoToday     string `env:"DEMO_TODAY" envDefault:"2026-06-19"`
	Profile       string `env:"APP_PROFILE" envDefault:"demo"`

	DatabaseURL         string `env:"DATABASE_URL"`
	NewAPIEnabled       bool   `env:"NEW_API_ENABLED" envDefault:"false"`
	NewAPIBaseURL       string `env:"NEW_API_BASE_URL"`
	NewAPIAdminToken    string `env:"NEW_API_ADMIN_TOKEN"`
	NewAPIWebhookSecret string `env:"NEW_API_WEBHOOK_SECRET"`
	NewAPIPublicURL     string `env:"NEW_API_PUBLIC_URL"`

	SyncTriggerAPIKey string `env:"SYNC_TRIGGER_API_KEY"`

	DataSourceCredentialKey string `env:"DATA_SOURCE_CREDENTIAL_KEY"`
	FeishuBaseURL           string `env:"FEISHU_BASE_URL" envDefault:"https://open.feishu.cn"`

	NotifyWebhookURL string `env:"NOTIFY_WEBHOOK_URL"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.IsProdProfile() {
		cfg.SimulateDelay = false
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) IsDemoProfile() bool {
	return c.Profile != ProfileProd
}

func (c Config) IsProdProfile() bool {
	return c.Profile == ProfileProd
}

func (c Config) validate() error {
	if c.IsProdProfile() && strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required when APP_PROFILE=prod")
	}
	if !c.NewAPIEnabled {
		return nil
	}
	if strings.TrimSpace(c.NewAPIBaseURL) == "" {
		return fmt.Errorf("NEW_API_BASE_URL is required when NEW_API_ENABLED=true")
	}
	if strings.TrimSpace(c.NewAPIAdminToken) == "" {
		return fmt.Errorf("NEW_API_ADMIN_TOKEN is required when NEW_API_ENABLED=true")
	}
	if strings.TrimSpace(c.NewAPIWebhookSecret) == "" {
		return fmt.Errorf("NEW_API_WEBHOOK_SECRET is required when NEW_API_ENABLED=true")
	}
	return nil
}

func (c Config) CORSOriginList() []string {
	parts := strings.Split(c.CORSOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
