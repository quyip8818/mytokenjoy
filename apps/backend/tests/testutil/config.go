//go:build testhook

package testutil

import (
	"os"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/contract"
)

const (
	defaultClockAnchor = "2026-06-19"
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

func WithBootstrapMode(mode string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.BootstrapMode = mode
	}
}

func WithClockAnchor(date string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.ClockAnchor = date
	}
}

func WithDeployEnv(env string) ConfigOption {
	return func(cfg *config.Config) {
		cfg.DeployEnv = env
	}
}

func WithSecureCookie(enabled bool) ConfigOption {
	return func(cfg *config.Config) {
		cfg.SecureCookie = enabled
	}
}

func WithProductionContract() ConfigOption {
	return func(cfg *config.Config) {
		WithDeployEnv(config.DeployEnvProduction)(cfg)
		WithBootstrapMode(config.BootstrapNone)(cfg)
		WithSecureCookie(true)(cfg)
		WithNewAPIEnabled(true)(cfg)
		cfg.GatewayEnabled = true
		WithNewAPIBaseURL("http://127.0.0.1:3000")(cfg)
		WithNewAPIAdminToken("admin-token")(cfg)
		if cfg.DatabaseURL == "" {
			cfg.DatabaseURL = config.DefaultDatabaseURL
		}
		cfg.LogDatabaseURL = cfg.DatabaseURL
		WithNewAPIWebhookSecret(DefaultTestWebhookSecret)(cfg)
		cfg.SimulateDelay = false
		cfg.ClockAnchor = ""
		if cfg.DataSourceCredentialKey == "" {
			cfg.DataSourceCredentialKey = DefaultTestCredentialKey
		}
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
		DeployEnv:               config.DeployEnvLocal,
		BootstrapMode:           config.BootstrapNone,
		DatabaseURL:             defaultTestDatabaseURL(),
		ClockAnchor:             defaultClockAnchor,
		SimulateDelay:           false,
		CompanyName:             "Demo Company",
		TokenJoyCompanyID:       contract.TokenJoyCompanyID,
		LocalCompanyID:          contract.LocalCompanyID,
		NewAPIEnabled:           true,
		NewAPIBaseURL:           "http://newapi.test",
		NewAPIAdminToken:        "token",
		SessionSecret:           TestSessionSecret,
		PlatformSessionSecret:   TestSessionSecret,
		SessionTTLSec:           86400,
		DataSourceCredentialKey: DefaultTestCredentialKey,
		StoreBootstrap: config.StoreBootstrap{
			TestPartitionMonths: 12,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.IngestEnabled() && cfg.NewAPIWebhookSecret == "" {
		cfg.NewAPIWebhookSecret = DefaultTestWebhookSecret
	}
	return cfg
}
