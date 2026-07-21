package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
)

func ApplyLocalEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DEPLOY_ENV", config.DeployEnvLocal)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "true")
	t.Setenv("NEW_API_GATEWAY_ENABLED", "false")
	t.Setenv("NEW_API_BASE_URL", "http://127.0.0.1:3000")
	t.Setenv("SESSION_SECRET", TestSessionSecret)
	t.Setenv("DATA_SOURCE_CREDENTIAL_KEY", DefaultTestCredentialKey)
}

func ApplyProductionEnv(t *testing.T) {
	t.Helper()
	ApplyLocalEnv(t)
	t.Setenv("DEPLOY_ENV", config.DeployEnvProduction)
	t.Setenv("BOOTSTRAP_MODE", config.BootstrapNone)
	t.Setenv("SECURE_COOKIE", "true")
	t.Setenv("NEW_API_GATEWAY_ENABLED", "true")
	t.Setenv("LOG_DATABASE_URL", "postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable")
	t.Setenv("NEW_API_WEBHOOK_SECRET", "webhook-secret")
	t.Setenv("SIMULATE_DELAY", "false")
	t.Setenv("CLOCK_ANCHOR", "")
}
