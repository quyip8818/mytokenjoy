package config_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
)

func TestProdRequiresDatabaseURL(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileProd)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when prod profile has no DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestDemoRequiresDatabaseURL(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileDemo)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when demo profile has no DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}

func TestDemoLoadsWithDatabaseURL(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileDemo)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected demo config to load, got %v", err)
	}
	if cfg.IsProdProfile() {
		t.Fatal("expected demo profile")
	}
	if cfg.ResolvedCompanyName() != "Acme Corp" {
		t.Fatalf("expected resolved company name Acme Corp, got %q", cfg.ResolvedCompanyName())
	}
	if cfg.DatabaseURL != config.DefaultDatabaseURL {
		t.Fatalf("expected default database url, got %q", cfg.DatabaseURL)
	}
}

func TestIngestRequiresWebhookSecret(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileDemo)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("LOG_DATABASE_URL", "postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable")
	t.Setenv("NEW_API_WEBHOOK_SECRET", "")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when ingest enabled without webhook secret")
	}
	if !strings.Contains(err.Error(), "NEW_API_WEBHOOK_SECRET") {
		t.Fatalf("expected webhook secret error, got %v", err)
	}
}

func TestPrivateRequiresCompanyName(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileDemo)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "")
	t.Setenv("SUPPORT_SAAS", "false")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when private deployment has no COMPANY_NAME")
	}
	if !strings.Contains(err.Error(), "COMPANY_NAME") {
		t.Fatalf("expected COMPANY_NAME error, got %v", err)
	}
}

func setProdRelayEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_PROFILE", config.ProfileProd)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Acme Corp")
	t.Setenv("NEW_API_ENABLED", "true")
	t.Setenv("RELAY_GATEWAY_ENABLED", "true")
	t.Setenv("NEW_API_BASE_URL", "http://127.0.0.1:3000")
	t.Setenv("NEW_API_ADMIN_TOKEN", "admin-token")
	t.Setenv("LOG_DATABASE_URL", "postgres://tokenjoy:tokenjoy@127.0.0.1:5432/logs?sslmode=disable")
	t.Setenv("NEW_API_WEBHOOK_SECRET", "webhook-secret")
	t.Setenv("SESSION_SECRET", "test-session-secret")
}

func TestProdRequiresRelayGatewayEnabled(t *testing.T) {
	setProdRelayEnv(t)
	t.Setenv("RELAY_GATEWAY_ENABLED", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when prod profile disables relay gateway")
	}
	if !strings.Contains(err.Error(), "RELAY_GATEWAY_ENABLED") {
		t.Fatalf("expected RELAY_GATEWAY_ENABLED error, got %v", err)
	}
}

func TestProdRejectsNewAPIBaseURLWithPath(t *testing.T) {
	setProdRelayEnv(t)
	t.Setenv("NEW_API_BASE_URL", "http://127.0.0.1:3000/v1")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when NEW_API_BASE_URL includes a path")
	}
	if !strings.Contains(err.Error(), "path") {
		t.Fatalf("expected path error, got %v", err)
	}
}

func TestProdLoadsWithRelayEnabled(t *testing.T) {
	setProdRelayEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected prod relay config to load, got %v", err)
	}
	if !cfg.NewAPIEnabled || !cfg.RelayGatewayEnabled {
		t.Fatalf("expected relay and gateway enabled, got relay=%v gateway=%v", cfg.NewAPIEnabled, cfg.RelayGatewayEnabled)
	}
}

func TestSaaSResolvesTestCompanyName(t *testing.T) {
	t.Setenv("APP_PROFILE", config.ProfileDemo)
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "")
	t.Setenv("SUPPORT_SAAS", "true")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")
	t.Setenv("PLATFORM_SESSION_SECRET", "test-platform-secret")
	t.Setenv("SESSION_SECRET", "test-session-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected saas config to load, got %v", err)
	}
	if cfg.ResolvedCompanyName() != config.SaaSDefaultCompanyName {
		t.Fatalf("expected %q, got %q", config.SaaSDefaultCompanyName, cfg.ResolvedCompanyName())
	}
}
