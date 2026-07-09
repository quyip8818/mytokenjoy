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

func WithRuntimeSeed() ConfigOption {
	return func(cfg *config.Config) {
		cfg.StoreBootstrap.RuntimeSeed = true
	}
}

func NewTestStoreWithRuntimeSeed(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	opts = append([]ConfigOption{WithRuntimeSeed()}, opts...)
	return NewTestStore(t, opts...)
}

func resetRuntimeTables(t *testing.T, st store.Store) {
	t.Helper()
	pool := postgres.MainPool(st)
	_, err := pool.Exec(context.Background(), `
		TRUNCATE company_recharge_lots, company_recharge_orders, usage_buckets, usage_ledger RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("reset runtime tables: %v", err)
	}
}

func applyDemoRuntime(t *testing.T, st store.Store, cfg config.Config) {
	t.Helper()
	ctx := company.WithContext(context.Background(), company.Context{CompanyID: contract.DefaultCompanyID})
	if err := runtime.ApplyUsageBuckets(ctx, st, cfg); err != nil {
		t.Fatalf("apply usage buckets: %v", err)
	}
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatalf("apply recharge orders: %v", err)
	}
	if err := runtime.ApplyUsageLedger(ctx, st, cfg); err != nil {
		t.Fatalf("apply usage ledger: %v", err)
	}
}
