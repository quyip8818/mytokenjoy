package testutil

import (
	"log/slog"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func NewOrgService(t *testing.T, cfg config.Config, st store.Store) org.Service {
	t.Helper()
	factory := datasource.NewFactory(cfg)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil, nil, relay.NewChannelPolicy(cfg))
	notifier := notification.NewService(cfg, st, slog.Default())
	return org.NewService(cfg, st, factory, lifecycle, notifier, common.NewDelayer(false), slog.Default())
}

func NewOrgServiceFromStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store, org.Service) {
	t.Helper()
	cfg, st := NewMemoryStoreFromConfig(t, opts...)
	return cfg, st, NewOrgService(t, cfg, st)
}
