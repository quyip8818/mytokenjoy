//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

// NewTestStore creates a test store using the shared cloned template (with seed data).
// For tests that need a blank DB (no seed data), use NewFreshTestStore instead.
func NewTestStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg := TestConfig(opts...)
	schemaURL := openClonedTestSchema(t).URL
	cfg.StoreBootstrap.SkipSchema = true
	cfg.StoreBootstrap.SkipSeed = true
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

// NewFreshTestStore creates a test store with a fresh schema (bootstrap + seed applied from scratch).
// Use for tests that need zero-state scenarios (e.g. no billing data).
func NewFreshTestStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg := TestConfig(opts...)
	schemaURL := openTestSchema(t).URL
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
