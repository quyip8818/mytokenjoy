//go:build testhook

package testutil

import (
	"context"
	"sync"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

// SharedTestStore returns a single cloned schema + store shared across all tests
// in the calling package. Use for read-only or append-only tests where isolation
// between test functions is not required (e.g. GET handlers, query tests).
//
// The store is created once (per process) and never reset between tests.
// Do NOT use this if your test mutates shared seed data that other tests observe.
var (
	sharedOnce  sync.Once
	sharedCfg   config.Config
	sharedStore store.Store
	sharedErr   error
)

func SharedTestStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	sharedOnce.Do(func() {
		cfg := TestConfig(opts...)
		schemaURL := openClonedTestSchema(t).URL
		cfg.DatabaseURL = schemaURL
		if cfg.IngestEnabled() {
			cfg.LogDatabaseURL = schemaURL
		}
		cfg.StoreBootstrap.SchemaPrepared = true
		st, err := postgres.New(context.Background(), cfg)
		if err != nil {
			sharedErr = err
			return
		}
		sharedCfg = cfg
		sharedStore = st
	})
	if sharedErr != nil {
		t.Fatalf("shared test store: %v", sharedErr)
	}
	return sharedCfg, sharedStore
}

// SharedTestStoreWithDemoRuntime is like SharedTestStore but also applies demo
// runtime seed data (usage ledger, recharge lots, etc). Use for tests that only
// read the demo runtime data.
var sharedDemoOnce sync.Once

func SharedTestStoreWithDemoRuntime(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg, st := SharedTestStore(t, opts...)
	sharedDemoOnce.Do(func() {
		ApplyDemoRuntime(t, st, cfg)
	})
	return cfg, st
}
