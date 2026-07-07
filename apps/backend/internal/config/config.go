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
	MinimalSeed   bool

	DatabaseURL         string `env:"DATABASE_URL"`
	LogDatabaseURL      string `env:"LOG_DATABASE_URL"`
	LogSchemaIsolated   bool
	NewAPIEnabled       bool   `env:"NEW_API_ENABLED" envDefault:"false"`
	NewAPIBaseURL       string `env:"NEW_API_BASE_URL"`
	NewAPIAdminToken    string `env:"NEW_API_ADMIN_TOKEN"`
	NewAPIWebhookSecret string `env:"NEW_API_WEBHOOK_SECRET"`
	NewAPIPublicURL     string `env:"NEW_API_PUBLIC_URL"`

	SyncTriggerAPIKey string `env:"SYNC_TRIGGER_API_KEY"`

	DataSourceCredentialKey string `env:"DATA_SOURCE_CREDENTIAL_KEY"`
	FeishuBaseURL           string `env:"FEISHU_BASE_URL" envDefault:"https://open.feishu.cn"`

	NotifyWebhookURL string `env:"NOTIFY_WEBHOOK_URL"`

	WorkerPollIntervalSec    int `env:"WORKER_POLL_INTERVAL_SEC" envDefault:"5"`
	WorkerOrgSyncIntervalSec int `env:"WORKER_ORG_SYNC_INTERVAL_SEC" envDefault:"60"`

	IngestReconcileIntervalSec int `env:"INGEST_RECONCILE_INTERVAL_SEC" envDefault:"300"`
	IngestReconcileBatchSize   int `env:"INGEST_RECONCILE_BATCH_SIZE" envDefault:"500"`
	IngestReconcileMaxRounds   int `env:"INGEST_RECONCILE_MAX_ROUNDS" envDefault:"10"`
	IngestFailureBatchSize     int `env:"INGEST_FAILURE_BATCH_SIZE" envDefault:"20"`

	SupportSaas              bool   `env:"SUPPORT_SAAS" envDefault:"false"`
	CompanyName              string `env:"COMPANY_NAME"`
	DefaultCompanyID         int64  `env:"DEFAULT_COMPANY_ID" envDefault:"1"`
	PlatformSharedRelayGroup string `env:"PLATFORM_SHARED_RELAY_GROUP" envDefault:"platform_shared"`
	RelayGatewayEnabled      bool   `env:"RELAY_GATEWAY_ENABLED" envDefault:"false"`
	CompanyWalletCacheTTLSec int    `env:"COMPANY_WALLET_CACHE_TTL_SEC" envDefault:"30"`

	PlatformBootstrapEmail    string `env:"PLATFORM_BOOTSTRAP_EMAIL"`
	PlatformBootstrapPassword string `env:"PLATFORM_BOOTSTRAP_PASSWORD"`

	SessionSecret         string `env:"SESSION_SECRET"`
	SessionTTLSec         int    `env:"SESSION_TTL_SEC" envDefault:"86400"`
	PlatformSessionSecret string `env:"PLATFORM_SESSION_SECRET"`
	AuthzCacheSize        int    `env:"AUTHZ_CACHE_SIZE" envDefault:"4096"`
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
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if !c.SupportSaas && strings.TrimSpace(c.CompanyName) == "" {
		return fmt.Errorf("COMPANY_NAME is required when SUPPORT_SAAS=false")
	}
	if strings.TrimSpace(c.SessionSecret) == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	if c.SupportSaas && strings.TrimSpace(c.PlatformSessionSecret) == "" {
		return fmt.Errorf("PLATFORM_SESSION_SECRET is required when SUPPORT_SAAS=true")
	}
	if c.LogSchemaIsolated && !c.IngestEnabled() {
		return fmt.Errorf("log schema isolation requires LOG_DATABASE_URL")
	}
	if c.LogSchemaIsolated && c.IsProdProfile() {
		return fmt.Errorf("log schema isolation is not allowed in prod profile")
	}
	if c.IngestEnabled() && strings.TrimSpace(c.NewAPIWebhookSecret) == "" {
		return fmt.Errorf("NEW_API_WEBHOOK_SECRET is required when LOG_DATABASE_URL is set")
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
	return nil
}

func (c Config) IngestEnabled() bool {
	return strings.TrimSpace(c.LogDatabaseURL) != ""
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

func (c Config) ResolvedPlatformSessionSecret() string {
	return c.PlatformSessionSecret
}
