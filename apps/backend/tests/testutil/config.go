package testutil

import (
	"os"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/contract"
)

const (
	defaultDemoToday         = "2026-06-19"
	DefaultTestWebhookSecret = "test-webhook-secret"
)

type ConfigOption func(*config.Config)

func WithConfig(cfg config.Config) ConfigOption {
	return func(c *config.Config) {
		*c = cfg
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

func WithMinimalSeed() ConfigOption {
	return func(cfg *config.Config) {
		cfg.MinimalSeed = true
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

func WithIngestEnabled(enabled bool) ConfigOption {
	return func(cfg *config.Config) {
		if enabled {
			cfg.LogDatabaseURL = cfg.DatabaseURL
			cfg.LogSchemaIsolated = true
		} else {
			cfg.LogDatabaseURL = ""
			cfg.LogSchemaIsolated = false
		}
	}
}

func defaultTestDatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return config.DefaultDatabaseURL
}

func TestConfig(opts ...ConfigOption) config.Config {
	cfg := config.Config{
		DatabaseURL:           defaultTestDatabaseURL(),
		DemoToday:             defaultDemoToday,
		SimulateDelay:         false,
		CompanyName:           "Demo Company",
		TokenJoyCompanyID:     contract.TokenJoyCompanyID,
		LocalCompanyID:        contract.LocalCompanyID,
		DefaultCompanyID:      contract.DefaultCompanyID,
		SessionSecret:         TestSessionSecret,
		PlatformSessionSecret: TestSessionSecret,
		SessionTTLSec:         86400,
		StoreBootstrap: config.StoreBootstrap{
			SkipRuntimeSeed:     true,
			TestPartitionMonths: 12,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if !cfg.SupportSaas {
		cfg.DefaultCompanyID = cfg.LocalCompanyID
	}
	if cfg.IngestEnabled() && cfg.NewAPIWebhookSecret == "" {
		cfg.NewAPIWebhookSecret = DefaultTestWebhookSecret
	}
	return cfg
}
