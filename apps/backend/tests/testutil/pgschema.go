//go:build testhook

package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil/pg"
)

func TestSchemaURL(t *testing.T) string {
	t.Helper()
	return openClonedTestSchema(t).URL
}

func openTestSchema(t *testing.T) pg.Handle {
	t.Helper()
	baseURL := defaultTestDatabaseURL()
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: docker compose -f docker-compose.test.yml up -d")
	}
	return pg.OpenSlow(t, baseURL)
}

func openClonedTestSchema(t *testing.T) pg.Handle {
	t.Helper()
	baseURL := defaultTestDatabaseURL()
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: docker compose -f docker-compose.test.yml up -d")
	}
	mode := CurrentTestMode()
	return pg.OpenCloned(t, baseURL, string(mode), templateStoreConfig(mode))
}

func openSharedClonedSchema(t *testing.T) pg.Handle {
	t.Helper()
	baseURL := defaultTestDatabaseURL()
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: docker compose -f docker-compose.test.yml up -d")
	}
	mode := CurrentTestMode()
	return pg.OpenClonedShared(t, baseURL, string(mode), templateStoreConfig(mode))
}

func templateStoreConfig(mode TestMode) config.Config {
	switch mode {
	case ModeSaaS:
		cfg := TestConfig(
			WithSupportSaas(true),
			WithBootstrapMode(config.BootstrapDemo),
			WithPlatformBootstrap("admin@tokenjoy.me", "admin1234"),
			WithIngestEnabled(true),
		)
		cfg.StoreBootstrap.TestPartitionMonths = 12
		return cfg
	default: // ModeLocal
		cfg := TestConfig(
			WithSupportSaas(false),
			WithBootstrapMode(config.BootstrapDemo),
			WithIngestEnabled(true),
		)
		cfg.StoreBootstrap.TestPartitionMonths = 12
		return cfg
	}
}
