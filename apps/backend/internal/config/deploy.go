package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/clock"
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

func (c Config) Clock() clock.Clock {
	anchor := strings.TrimSpace(c.ClockAnchor)
	if anchor == "" {
		return clock.System()
	}
	parsed, err := time.Parse("2006-01-02", anchor)
	if err != nil {
		return clock.System()
	}
	return clock.Fixed(time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC))
}

func (c Config) validateProductionContract() error {
	if !c.BootstrapIsNone() {
		return fmt.Errorf("BOOTSTRAP_MODE must be none in production")
	}
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
	if strings.TrimSpace(c.ClockAnchor) != "" {
		return fmt.Errorf("CLOCK_ANCHOR must not be set in production")
	}
	return nil
}

func validateClockAnchor(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(raw)); err != nil {
		return fmt.Errorf("CLOCK_ANCHOR must be YYYY-MM-DD")
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
