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
