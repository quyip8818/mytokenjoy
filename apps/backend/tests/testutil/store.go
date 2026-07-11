//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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
	if !cfg.BootstrapIsMinimal() {
		resetRuntimeTables(t, st)
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
		UPDATE river_job
		SET state = 'completed', finalized_at = NOW()
		WHERE kind = $1
		  AND (args->>'company_id')::bigint = $2
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
	`, jobs.KindWalletSync, companyID)
	if err != nil {
		t.Fatalf("drain pending wallet sync: %v", err)
	}
}
