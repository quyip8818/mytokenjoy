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
	_, schemaURL := openTestSchema(t)
	cfg.DatabaseURL = schemaURL
	if cfg.IngestEnabled() {
		cfg.LogDatabaseURL = schemaURL
	}
	st, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	clearDemoRuntimeSeed(t, st)
	t.Cleanup(func() {
		if pg, ok := st.(*postgres.Store); ok {
			pg.Close()
		}
	})
	return cfg, st
}

func clearDemoRuntimeSeed(t *testing.T, st store.Store) {
	t.Helper()
	pool := postgres.MainPool(st)
	_, err := pool.Exec(context.Background(), `
		TRUNCATE usage_buckets, company_recharge_orders RESTART IDENTITY
	`)
	if err != nil {
		t.Fatalf("clear demo runtime seed: %v", err)
	}
}
