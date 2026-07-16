//go:build testhook

package orgfix

import (
	"log/slog"
	"testing"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func NewService(t *testing.T, cfg config.Config, st store.Store) org.Service {
	t.Helper()
	return NewServiceWithEnqueuer(t, cfg, st, jobs.NoopEnqueuer{})
}

func NewServiceWithEnqueuer(t *testing.T, cfg config.Config, st store.Store, enqueuer jobs.Enqueuer) org.Service {
	t.Helper()
	factory := datasource.NewFactory(cfg)
	newAPISync := newapisync.New(cfg, st, nil, nil, policy.NewChannelPolicy(cfg), adapter.NewNewAPISyncEnqueuer(jobs.NoopEnqueuer{}))
	notifier := notification.NewService(cfg, st, slog.Default())
	return org.NewService(cfg, st, factory, newAPISync, notifier, common.NewDelayer(false), slog.Default(), permission.NewGrantNormalizer(), adapter.NewOrgEnqueuer(enqueuer, adapter.NewOrgRiverAdminHolder(nil)))
}

func NewServiceFromStore(t *testing.T, opts ...testutil.ConfigOption) (config.Config, store.Store, org.Service) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, opts...)
	return cfg, st, NewService(t, cfg, st)
}
