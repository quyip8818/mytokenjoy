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
		t.Fatal("DATABASE_URL required; run: pnpm start:postgres")
	}
	return pg.OpenSlow(t, baseURL)
}

func openClonedTestSchema(t *testing.T) pg.Handle {
	t.Helper()
	baseURL := defaultTestDatabaseURL()
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: pnpm start:postgres")
	}
	return pg.OpenCloned(t, baseURL, templateStoreConfig(baseURL))
}

func templateStoreConfig(baseURL string) config.Config {
	cfg := TestConfig(WithIngestEnabled(true))
	templateURL := pg.WithSearchPath(baseURL, "test_template")
	cfg.DatabaseURL = templateURL
	cfg.LogDatabaseURL = templateURL
	cfg.LogSchemaIsolated = true
	cfg.StoreBootstrap.SkipRuntimeSeed = true
	cfg.StoreBootstrap.TestPartitionMonths = 12
	return cfg
}
