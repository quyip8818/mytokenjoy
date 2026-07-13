//go:build testhook

package scheduler_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func enableOrgSyncSchedule(t *testing.T, st store.Store) {
	t.Helper()
	ctx := testutil.Ctx()
	integration, err := st.Org().Integration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	integration.ApplySyncConfig(types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 24,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})
	if err := st.Org().SetIntegration(ctx, integration); err != nil {
		t.Fatal(err)
	}
}

func TestCollectDueMonthRebalanceWhenPeriodMissing(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	svc := scheduler.NewService(cfg, st)

	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsMonthRebalance {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected month rebalance due for default company without last_rebalanced_period")
	}
}

func TestCollectDueSkipsMonthRebalanceAfterPeriodSet(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, cfg.Clock()).String()
	if err := st.TenantBackgroundState().EnsureRow(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().SetLastRebalancedPeriod(ctx, contract.DefaultCompanyID, current); err != nil {
		t.Fatal(err)
	}

	svc := scheduler.NewService(cfg, st)
	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsMonthRebalance {
			t.Fatalf("expected month rebalance skipped for period %q", current)
		}
	}
}

func TestCollectDueOrgSyncWhenEnabledWithoutSchedule(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	enableOrgSyncSchedule(t, st)

	svc := scheduler.NewService(cfg, st)
	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsOrgSync {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected org_sync due when org enabled but next_org_sync_at is missing")
	}
}

func TestCollectDueSkipsOrgSyncWhenDisabledWithoutSchedule(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	svc := scheduler.NewService(cfg, st)
	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsOrgSync {
			t.Fatal("expected org_sync skipped when org sync disabled and schedule missing")
		}
	}
}

func TestCollectDueOrgSyncWhenNextAtPassed(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	past := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	if err := st.TenantBackgroundState().EnsureRow(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().UpsertOrgSchedule(ctx, contract.DefaultCompanyID, past, nil); err != nil {
		t.Fatal(err)
	}

	svc := scheduler.NewService(cfg, st)
	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsOrgSync {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected org_sync due when next_org_sync_at is in the past")
	}
}

func TestCollectDueSkipsOrgSyncWhenPendingJobExists(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	past := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	if err := st.TenantBackgroundState().EnsureRow(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().UpsertOrgSchedule(ctx, contract.DefaultCompanyID, past, nil); err != nil {
		t.Fatal(err)
	}

	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)
	if err := jobs.InsertOrgSync(ctx, enqueuer, nil, contract.DefaultCompanyID, &past); err != nil {
		t.Fatal(err)
	}

	svc := scheduler.NewService(cfg, st)
	due, err := svc.CollectDue(ctx, time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range due {
		if item.CompanyID == contract.DefaultCompanyID && item.NeedsOrgSync {
			t.Fatal("expected org_sync skipped when active river job exists")
		}
	}
}
