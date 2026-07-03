package testutil

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/seed"
)

const defaultDemoToday = "2026-06-19"

type ConfigOption func(*config.Config)

func WithNewAPIEnabled(enabled bool) ConfigOption {
	return func(cfg *config.Config) {
		cfg.NewAPIEnabled = enabled
	}
}

func WithNewAPIWebhookSecret(secret string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.NewAPIWebhookSecret = secret
	}
}

func WithNewAPIBaseURL(baseURL string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.NewAPIBaseURL = baseURL
	}
}

func WithNewAPIAdminToken(token string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.NewAPIAdminToken = token
	}
}

func WithProfile(profile string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.Profile = profile
	}
}

func WithSupportSaas(enabled bool) ConfigOption {
	return func(cfg *config.Config) {
		cfg.SupportSaas = enabled
	}
}

func WithPlatformBootstrap(email, password string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.PlatformBootstrapEmail = email
		cfg.PlatformBootstrapPassword = password
	}
}

func TestConfig(opts ...ConfigOption) config.Config {
	cfg := config.Config{
		DemoToday:             defaultDemoToday,
		SimulateDelay:         false,
		CompanyName:           "Demo Company",
		DefaultCompanyID:      seed.DefaultCompanyID,
		SessionSecret:         TestSessionSecret,
		PlatformSessionSecret: TestSessionSecret,
		SessionTTLSec:         86400,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
