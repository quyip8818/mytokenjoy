package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port          string `env:"PORT" envDefault:"8080"`
	CORSOrigins   string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
	SimulateDelay bool   `env:"SIMULATE_DELAY" envDefault:"false"`

	BootstrapMode string `env:"BOOTSTRAP_MODE" envDefault:"none"`
	SecureCookie  bool   `env:"SECURE_COOKIE" envDefault:"false"`
	ClockAnchor   string `env:"CLOCK_ANCHOR"`
	DeployEnv     string `env:"DEPLOY_ENV" envDefault:"local"`

	StoreBootstrap StoreBootstrap

	DatabaseURL         string `env:"DATABASE_URL"`
	LogDatabaseURL      string `env:"LOG_DATABASE_URL"`
	LogSchemaIsolated   bool
	NewAPIEnabled       bool   `env:"NEW_API_ENABLED" envDefault:"false"`
	NewAPIBaseURL       string `env:"NEW_API_BASE_URL"`
	NewAPIAdminToken    string `env:"NEW_API_ADMIN_TOKEN"`
	NewAPIWebhookSecret string `env:"NEW_API_WEBHOOK_SECRET"`

	SyncTriggerAPIKey string `env:"SYNC_TRIGGER_API_KEY"`

	DataSourceCredentialKey string `env:"DATA_SOURCE_CREDENTIAL_KEY"`
	FeishuBaseURL           string `env:"FEISHU_BASE_URL" envDefault:"https://open.feishu.cn"`

	NotifyWebhookURL string `env:"NOTIFY_WEBHOOK_URL"`

	WorkerPollIntervalSec    int `env:"WORKER_POLL_INTERVAL_SEC" envDefault:"5"`
	WorkerOrgSyncIntervalSec int `env:"WORKER_ORG_SYNC_INTERVAL_SEC" envDefault:"60"`

	IngestReconcileIntervalSec int `env:"INGEST_RECONCILE_INTERVAL_SEC" envDefault:"300"`
	IngestReconcileBatchSize   int `env:"INGEST_RECONCILE_BATCH_SIZE" envDefault:"500"`
	IngestReconcileMaxRounds   int `env:"INGEST_RECONCILE_MAX_ROUNDS" envDefault:"10"`
	IngestJobBatchSize         int `env:"INGEST_JOB_BATCH_SIZE" envDefault:"20"`

	SupportSaas              bool   `env:"SUPPORT_SAAS" envDefault:"false"`
	CompanyName              string `env:"COMPANY_NAME"`
	TokenJoyCompanyID        int64  `env:"TOKENJOY_COMPANY_ID" envDefault:"1"`
	LocalCompanyID           int64  `env:"LOCAL_COMPANY_ID" envDefault:"2"`
	DefaultCompanyID         int64  `env:"DEFAULT_COMPANY_ID" envDefault:"2"`
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
	cfg.BootstrapMode = strings.ToLower(strings.TrimSpace(cfg.BootstrapMode))
	cfg.DeployEnv = strings.ToLower(strings.TrimSpace(cfg.DeployEnv))
	if !cfg.SupportSaas {
		cfg.DefaultCompanyID = cfg.LocalCompanyID
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
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

func (c Config) ResolvedPlatformSessionSecret() string {
	return c.PlatformSessionSecret
}
