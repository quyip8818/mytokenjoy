package testutil

import "github.com/tokenjoy/backend/internal/config"

const defaultDemoToday = "2026-06-19"

type ConfigOption func(*config.Config)

func WithDemoToday(value string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.DemoToday = value
	}
}

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

func TestConfig(opts ...ConfigOption) config.Config {
	cfg := config.Config{
		DemoToday:     defaultDemoToday,
		SimulateDelay: false,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
