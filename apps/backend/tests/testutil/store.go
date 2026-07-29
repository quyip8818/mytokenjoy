//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

func NewTestStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg := TestConfig(opts...)
	var schemaURL string
	if cfg.BootstrapIsMinimal() {
		schemaURL = openTestSchema(t).URL
	} else {
		schemaURL = openClonedTestSchema(t).URL
		cfg.StoreBootstrap.SkipSchema = true
		cfg.StoreBootstrap.SkipSeed = true
	}
	cfg.DatabaseURL = schemaURL
	if cfg.IngestEnabled() {
		cfg.LogDatabaseURL = schemaURL
	}
	st, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	t.Cleanup(func() {
		if pg, ok := st.(*postgres.Store); ok {
			pg.Close()
		}
	})
	return cfg, st
}

func PreparedConfig(schemaURL string) config.Config {
	cfg := TestConfig()
	cfg.DatabaseURL = schemaURL
	if cfg.IngestEnabled() {
		cfg.LogDatabaseURL = schemaURL
	}
	cfg.StoreBootstrap.SkipSchema = true
	cfg.StoreBootstrap.SkipSeed = true
	return cfg
}
