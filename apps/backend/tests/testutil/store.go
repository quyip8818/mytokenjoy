package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
)

func NewMemoryStore(t *testing.T, cfg config.Config) store.Store {
	t.Helper()
	return store.NewMemory(seed.Load(cfg))
}

func NewMemoryStoreFromConfig(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
	t.Helper()
	cfg := TestConfig(opts...)
	return cfg, NewMemoryStore(t, cfg)
}
