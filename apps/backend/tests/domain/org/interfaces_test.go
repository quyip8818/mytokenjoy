package org_test

import (
	"log/slog"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newTestService(t *testing.T) org.Service {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	factory := datasource.NewFactory(cfg)
	lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
	notifier := notification.NewService(cfg, st, slog.Default())
	return org.NewService(cfg, st, factory, lifecycle, notifier, common.NewDelayer(false), slog.Default())
}

func TestServiceImplementsSubInterfaces(t *testing.T) {
	svc := newTestService(t)
	var (
		_ org.Service           = svc
		_ org.DataSourceService = svc
		_ org.SyncService       = svc
		_ org.DepartmentService = svc
		_ org.MemberService     = svc
		_ org.RoleService       = svc
	)
}
