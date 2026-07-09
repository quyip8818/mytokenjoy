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
	if cfg.MinimalSeed {
		schemaURL = openTestSchema(t).URL
	} else {
		schemaURL = openClonedTestSchema(t).URL
		cfg.StoreBootstrap.SchemaPrepared = true
	}
	cfg.DatabaseURL = schemaURL
	if cfg.IngestEnabled() {
		cfg.LogDatabaseURL = schemaURL
	}
	st, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("create postgres store: %v", err)
	}
	if !cfg.MinimalSeed {
		resetRuntimeTables(t, st)
	}
	if cfg.StoreBootstrap.RuntimeSeed {
		applyDemoRuntime(t, st, cfg)
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
	cfg.StoreBootstrap.SchemaPrepared = true
	return cfg
}

func DrainPendingWalletSync(t *testing.T, st store.Store, companyID int64) {
	t.Helper()
	pool := postgres.MainPool(st)
	_, err := pool.Exec(context.Background(), `
		UPDATE async_jobs
		SET status = $1, updated_at = NOW()
		WHERE channel = $2 AND company_id = $3 AND status = $4
	`, store.JobStatusDone, store.JobChannelWalletSync, companyID, store.JobStatusPending)
	if err != nil {
		t.Fatalf("drain pending wallet sync: %v", err)
	}
}
