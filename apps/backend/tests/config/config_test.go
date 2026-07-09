package config_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestProductionRequiresDatabaseURL(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production deploy has no DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLocalRequiresDatabaseURL(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when local deploy has no DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestLocalLoadsWithDatabaseURL(t *testing.T) {
	testutil.ApplyLocalEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected local config to load, got %v", err)
	}
	if cfg.DeployEnv != config.DeployEnvLocal {
		t.Fatalf("expected local deploy env, got %q", cfg.DeployEnv)
	}
	if cfg.ResolvedCompanyName() != "Acme Corp" {
		t.Fatalf("expected resolved company name Acme Corp, got %q", cfg.ResolvedCompanyName())
	}
}

func TestIngestRequiresWebhookSecret(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("LOG_DATABASE_URL", "postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable")
	t.Setenv("NEW_API_WEBHOOK_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when ingest enabled without webhook secret")
	}
	if !strings.Contains(err.Error(), "NEW_API_WEBHOOK_SECRET") {
		t.Fatalf("expected webhook secret error, got %v", err)
	}
}

func TestPrivateRequiresCompanyName(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("COMPANY_NAME", "")
	t.Setenv("SUPPORT_SAAS", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when private deployment has no COMPANY_NAME")
	}
	if !strings.Contains(err.Error(), "COMPANY_NAME") {
		t.Fatalf("expected COMPANY_NAME error, got %v", err)
	}
}

func TestProductionRequiresRelayGatewayEnabled(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("RELAY_GATEWAY_ENABLED", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production disables relay gateway")
	}
	if !strings.Contains(err.Error(), "RELAY_GATEWAY_ENABLED") {
		t.Fatalf("expected RELAY_GATEWAY_ENABLED error, got %v", err)
	}
}

func TestProductionRequiresNewAPIEnabled(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("NEW_API_ENABLED", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production disables new api")
	}
	if !strings.Contains(err.Error(), "NEW_API_ENABLED") {
		t.Fatalf("expected NEW_API_ENABLED error, got %v", err)
	}
}

func TestProductionRejectsNewAPIBaseURLWithPath(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("NEW_API_BASE_URL", "http://127.0.0.1:3000/v1")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when NEW_API_BASE_URL includes a path")
	}
	if !strings.Contains(err.Error(), "path") {
		t.Fatalf("expected path error, got %v", err)
	}
}

func TestProductionLoadsWithRelayEnabled(t *testing.T) {
	testutil.ApplyProductionEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected production relay config to load, got %v", err)
	}
	if !cfg.NewAPIEnabled || !cfg.RelayGatewayEnabled {
		t.Fatalf("expected relay and gateway enabled, got relay=%v gateway=%v", cfg.NewAPIEnabled, cfg.RelayGatewayEnabled)
	}
}

func TestLocalAllowsMissingRelayStack(t *testing.T) {
	testutil.ApplyLocalEnv(t)

	_, err := config.Load()
	if err != nil {
		t.Fatalf("expected local config without relay to load, got %v", err)
	}
}

func TestStagingAllowsMissingRelayStack(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("DEPLOY_ENV", config.DeployEnvStaging)

	_, err := config.Load()
	if err != nil {
		t.Fatalf("expected staging config without relay to load, got %v", err)
	}
}

func TestRelayGatewayRequiresNewAPIEnabled(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("RELAY_GATEWAY_ENABLED", "true")
	t.Setenv("NEW_API_ENABLED", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when gateway enabled without new api")
	}
	if !strings.Contains(err.Error(), "NEW_API_ENABLED") {
		t.Fatalf("expected NEW_API_ENABLED error, got %v", err)
	}
}

func TestInvalidBootstrapModeFails(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("BOOTSTRAP_MODE", "invalid")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid bootstrap mode")
	}
	if !strings.Contains(err.Error(), "BOOTSTRAP_MODE") {
		t.Fatalf("expected BOOTSTRAP_MODE error, got %v", err)
	}
}

func TestInvalidDeployEnvFails(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("DEPLOY_ENV", "invalid")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid deploy env")
	}
	if !strings.Contains(err.Error(), "DEPLOY_ENV") {
		t.Fatalf("expected DEPLOY_ENV error, got %v", err)
	}
}

func TestMissingCredentialKeyFails(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("DATA_SOURCE_CREDENTIAL_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when credential key is missing")
	}
	if !strings.Contains(err.Error(), "DATA_SOURCE_CREDENTIAL_KEY") {
		t.Fatalf("expected credential key error, got %v", err)
	}
}

func TestSaaSResolvesTestCompanyName(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("COMPANY_NAME", "")
	t.Setenv("SUPPORT_SAAS", "true")
	t.Setenv("PLATFORM_SESSION_SECRET", "test-platform-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected saas config to load, got %v", err)
	}
	if cfg.ResolvedCompanyName() != config.SaaSDefaultCompanyName {
		t.Fatalf("expected %q, got %q", config.SaaSDefaultCompanyName, cfg.ResolvedCompanyName())
	}
}

func TestProductionRequiresSecureCookie(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("SECURE_COOKIE", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production disables secure cookie")
	}
	if !strings.Contains(err.Error(), "SECURE_COOKIE") {
		t.Fatalf("expected SECURE_COOKIE error, got %v", err)
	}
}

func TestProductionRequiresLogDatabaseURL(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("LOG_DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production has no LOG_DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "LOG_DATABASE_URL") {
		t.Fatalf("expected LOG_DATABASE_URL error, got %v", err)
	}
}

func TestProductionRejectsSimulateDelay(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("SIMULATE_DELAY", "true")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production enables simulate delay")
	}
	if !strings.Contains(err.Error(), "SIMULATE_DELAY") {
		t.Fatalf("expected SIMULATE_DELAY error, got %v", err)
	}
}

func TestProductionRejectsClockAnchor(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("CLOCK_ANCHOR", "2026-06-19")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production sets CLOCK_ANCHOR")
	}
	if !strings.Contains(err.Error(), "CLOCK_ANCHOR") {
		t.Fatalf("expected CLOCK_ANCHOR error, got %v", err)
	}
}

func TestProductionRejectsDemoBootstrap(t *testing.T) {
	testutil.ApplyProductionEnv(t)
	t.Setenv("BOOTSTRAP_MODE", config.BootstrapDemo)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when production uses demo bootstrap")
	}
	if !strings.Contains(err.Error(), "BOOTSTRAP_MODE") {
		t.Fatalf("expected BOOTSTRAP_MODE error, got %v", err)
	}
}

func TestInvalidClockAnchorFails(t *testing.T) {
	testutil.ApplyLocalEnv(t)
	t.Setenv("CLOCK_ANCHOR", "not-a-date")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid CLOCK_ANCHOR")
	}
	if !strings.Contains(err.Error(), "CLOCK_ANCHOR") {
		t.Fatalf("expected CLOCK_ANCHOR error, got %v", err)
	}
}
