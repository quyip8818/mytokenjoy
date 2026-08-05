package config

import (
	"fmt"
	"strings"
)

const (
	DeployEnvLocal      = "local"
	DeployEnvStaging    = "staging"
	DeployEnvProduction = "production"
)

func (c Config) IsProductionDeploy() bool { return c.DeployEnv == DeployEnvProduction }

// AllowsDevHTTPRoutes is the single gate for /api/dev/* (including
// GET /api/dev/platform-keys/{id}/bearer). True only when DEPLOY_ENV=local.
//
// Do not register or serve dev HTTP routes under staging/production, and do not
// add alternate env flags or feature toggles for this surface.
func (c Config) AllowsDevHTTPRoutes() bool { return c.DeployEnv == DeployEnvLocal }

func (c Config) validateProductionContract() error {
	if !c.SecureCookie {
		return fmt.Errorf("SECURE_COOKIE must be true in production")
	}
	if !c.NewAPIEnabled {
		return fmt.Errorf("NEW_API_ENABLED must be true in production")
	}
	if !c.GatewayEnabled {
		return fmt.Errorf("NEW_API_GATEWAY_ENABLED must be true in production")
	}
	if strings.TrimSpace(c.LogDatabaseURL) == "" {
		return fmt.Errorf("LOG_DATABASE_URL is required in production")
	}
	if strings.TrimSpace(c.NewAPIWebhookSecret) == "" {
		return fmt.Errorf("NEW_API_WEBHOOK_SECRET is required in production")
	}
	if c.SimulateDelay {
		return fmt.Errorf("SIMULATE_DELAY must be false in production")
	}
	if c.SkipVerifyCode {
		return fmt.Errorf("SKIP_VERIFY_CODE must be false in production")
	}
	return nil
}

func validateDeployEnv(env string) error {
	switch env {
	case DeployEnvLocal, DeployEnvStaging, DeployEnvProduction:
		return nil
	default:
		return fmt.Errorf("DEPLOY_ENV must be local, staging, or production")
	}
}
