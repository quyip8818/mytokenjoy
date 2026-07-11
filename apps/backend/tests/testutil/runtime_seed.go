//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/runtime"
)

func NewTestStoreWithDemoRuntime(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg, st := NewTestStore(t, opts...)
	ApplyDemoRuntime(t, st, cfg)
	return cfg, st
}

func resetRuntimeTables(t *testing.T, cfg config.Config, st store.Store) {
	t.Helper()
	ctx := context.Background()
	pool := postgres.MainPool(st)
	_, err := pool.Exec(ctx, `
		TRUNCATE company_recharge_lots, company_recharge_orders, usage_buckets, usage_ledger RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("reset runtime tables: %v", err)
	}
	if cfg.IngestEnabled() {
		logPool := postgres.LogPool(st)
		if _, err := logPool.Exec(ctx, `
			TRUNCATE logs, ingest_jobs RESTART IDENTITY;
			UPDATE reconcile_cursors SET last_log_id = 0, updated_at = NOW();
		`); err != nil {
			t.Fatalf("reset ingest runtime tables: %v", err)
		}
	}
}

func ApplyDemoRuntime(t *testing.T, st store.Store, cfg config.Config) {
	t.Helper()
	ctx := company.WithContext(context.Background(), company.Context{CompanyID: contract.DefaultCompanyID})
	if err := runtime.ApplyDemo(ctx, st, cfg); err != nil {
		t.Fatalf("apply demo runtime: %v", err)
	}
}
