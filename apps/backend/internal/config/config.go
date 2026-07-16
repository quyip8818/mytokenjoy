package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/tokenjoy/backend/internal/pkg/baseurl"
)

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Port        string `env:"PORT" envDefault:"8080"`
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
}

// DeployConfig holds deployment and environment settings.
type DeployConfig struct {
	BootstrapMode string `env:"BOOTSTRAP_MODE" envDefault:"none"`
	SecureCookie  bool   `env:"SECURE_COOKIE" envDefault:"false"`
	ClockAnchor   string `env:"CLOCK_ANCHOR"`
	DeployEnv     string `env:"DEPLOY_ENV" envDefault:"local"`
	SimulateDelay bool   `env:"SIMULATE_DELAY" envDefault:"false"`
}

// DatabaseConfig holds primary and log database connection settings.
type DatabaseConfig struct {
	DatabaseURL       string `env:"DATABASE_URL"`
	LogDatabaseURL    string `env:"LOG_DATABASE_URL"`
	LogSchemaIsolated bool

	StoreBootstrap StoreBootstrap
}

// NewAPIConfig holds settings for the external NewAPI integration.
type NewAPIConfig struct {
	NewAPIEnabled       bool   `env:"NEW_API_ENABLED" envDefault:"true"`
	NewAPIBaseURL       string `env:"NEW_API_BASE_URL"`
	NewAPIAdminToken    string `env:"NEW_API_ADMIN_TOKEN"`
	NewAPIAdminUserID   int64  `env:"NEW_API_ADMIN_USER_ID" envDefault:"1"`
	NewAPIWebhookSecret string `env:"NEW_API_WEBHOOK_SECRET"`
	GatewayEnabled      bool   `env:"NEW_API_GATEWAY_ENABLED" envDefault:"false"`
}

// DataSourceConfig holds external data source integration settings.
type DataSourceConfig struct {
	SyncTriggerAPIKey       string `env:"SYNC_TRIGGER_API_KEY"`
	DataSourceCredentialKey string `env:"DATA_SOURCE_CREDENTIAL_KEY"`
	FeishuBaseURL           string `env:"FEISHU_BASE_URL" envDefault:"https://open.feishu.cn"`
}

// NotificationConfig holds notification channel settings (webhook, email, SMS).
type NotificationConfig struct {
	NotifyWebhookURL string `env:"NOTIFY_WEBHOOK_URL"`

	// SMTP / Email channel
	SMTPHost string `env:"SMTP_HOST"`
	SMTPPort int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser string `env:"SMTP_USER"`
	SMTPPass string `env:"SMTP_PASS"`
	SMTPFrom string `env:"SMTP_FROM"`

	// SMS channel (Twilio)
	TwilioAccountSID string `env:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `env:"TWILIO_AUTH_TOKEN"`
	TwilioFromNumber string `env:"TWILIO_FROM_NUMBER"`
}

// IngestConfig holds ingest worker and reconciliation settings.
type IngestConfig struct {
	WorkerPollIntervalSec      int `env:"WORKER_POLL_INTERVAL_SEC" envDefault:"1"` // ingest pending poll only
	IngestReconcileIntervalSec int `env:"INGEST_RECONCILE_INTERVAL_SEC" envDefault:"300"`
	IngestReconcileBatchSize   int `env:"INGEST_RECONCILE_BATCH_SIZE" envDefault:"500"`
	IngestReconcileMaxRounds   int `env:"INGEST_RECONCILE_MAX_ROUNDS" envDefault:"10"`
	IngestJobBatchSize         int `env:"INGEST_JOB_BATCH_SIZE" envDefault:"20"`
}

// WatchdogConfig holds watchdog/scheduler settings.
type WatchdogConfig struct {
	WatchdogIntervalSec      int `env:"WATCHDOG_INTERVAL_SEC" envDefault:"604800"`
	WatchdogBulkBatchSizeEnv int `env:"WATCHDOG_BULK_BATCH_SIZE" envDefault:"200"`
	WatchdogStartupDelaySec  int `env:"WATCHDOG_STARTUP_DELAY_SEC" envDefault:"5"`
}

// PlatformConfig holds multi-tenant platform and company settings.
type PlatformConfig struct {
	SupportSaas               bool   `env:"SUPPORT_SAAS" envDefault:"false"`
	CompanyName               string `env:"COMPANY_NAME"`
	TokenJoyCompanyID         int64  `env:"TOKENJOY_COMPANY_ID" envDefault:"1"`
	LocalCompanyID            int64  `env:"LOCAL_COMPANY_ID" envDefault:"2"`
	PlatformSharedNewAPIGroup string `env:"PLATFORM_SHARED_NEW_API_GROUP" envDefault:"platform_shared"`
	DefaultProviderDeptID     string `env:"DEFAULT_PROVIDER_DEPT_ID" envDefault:"dept-3"`
	CompanyWalletCacheTTLSec  int    `env:"COMPANY_WALLET_CACHE_TTL_SEC" envDefault:"30"`

	PlatformBootstrapEmail    string `env:"PLATFORM_BOOTSTRAP_EMAIL"`
	PlatformBootstrapPassword string `env:"PLATFORM_BOOTSTRAP_PASSWORD"`
}

// IdentityConfig holds authentication, session, and authorization settings.
type IdentityConfig struct {
	SessionSecret         string `env:"SESSION_SECRET"`
	SessionTTLSec         int    `env:"SESSION_TTL_SEC" envDefault:"86400"`
	PlatformSessionSecret string `env:"PLATFORM_SESSION_SECRET"`
	AuthzCacheSize        int    `env:"AUTHZ_CACHE_SIZE" envDefault:"4096"`
}

// CacheConfig holds Redis and gateway budget-check cache settings.
type CacheConfig struct {
	RedisURL                 string `env:"REDIS_URL"`
	GatewayBudgetCheckTTLSec int    `env:"GATEWAY_BUDGET_CHECK_TTL_SEC" envDefault:"600"`
}

// Config is the root application configuration.
// Sub-structs are embedded so field access remains flat (e.g. cfg.Port, cfg.DatabaseURL).
type Config struct {
	HTTPConfig
	DeployConfig
	DatabaseConfig
	NewAPIConfig
	DataSourceConfig
	NotificationConfig
	IngestConfig
	WatchdogConfig
	RiverConfig
	PlatformConfig
	IdentityConfig
	CacheConfig
}

func (c Config) GatewayBudgetCheckEnabled() bool {
	return strings.TrimSpace(c.RedisURL) != ""
}

func (c Config) GatewayBudgetCheckTTL() time.Duration {
	sec := c.GatewayBudgetCheckTTLSec
	if sec <= 0 {
		sec = 600
	}
	return time.Duration(sec) * time.Second
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.BootstrapMode = strings.ToLower(strings.TrimSpace(cfg.BootstrapMode))
	cfg.DeployEnv = strings.ToLower(strings.TrimSpace(cfg.DeployEnv))
	if err := cfg.normalize(); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) normalize() error {
	if !c.NewAPIEnabled || strings.TrimSpace(c.NewAPIBaseURL) == "" {
		return nil
	}
	origin, err := baseurl.Origin(c.NewAPIBaseURL)
	if err != nil {
		return fmt.Errorf("NEW_API_BASE_URL: %w", err)
	}
	c.NewAPIBaseURL = origin
	return nil
}

func (c Config) IngestEnabled() bool {
	return strings.TrimSpace(c.LogDatabaseURL) != ""
}

func (c Config) CORSOriginList() []string {
	parts := strings.Split(c.CORSOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
