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
		TRUNCATE company_recharge_lots, company_recharge_orders, usage_buckets RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("clear demo runtime seed: %v", err)
	}
}

func DrainPendingWalletSync(t *testing.T, st store.Store, companyID int64) {
	t.Helper()
	entries, err := st.Relay().ClaimPendingWalletSync(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.CompanyID != companyID {
			continue
		}
		if err := st.Relay().MarkWalletSyncDone(context.Background(), entry.ID); err != nil {
			t.Fatal(err)
		}
	}
}
