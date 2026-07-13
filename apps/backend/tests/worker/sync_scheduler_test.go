package worker_test

import (
	"testing"
	"time"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func syncLogCount(t *testing.T, env orgfix.FeishuOrgEnv) int {
	t.Helper()
	logs, err := env.Store.Org().SyncLogs(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	return len(logs)
}

func TestScheduledSyncUsesLock(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	acquired, err := env.Store.SchedulerLock().TryAcquire(testutil.Ctx(), types.OrgSyncLockName(contract.DefaultCompanyID), "other", time.Minute)
	if err != nil || !acquired {
		t.Fatalf("expected lock acquired, err=%v acquired=%v", err, acquired)
	}

	before := syncLogCount(t, env)
	if err := env.Svc.RunScheduledSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	if syncLogCount(t, env) != before {
		t.Fatalf("expected no new sync log while lock held, before=%d after=%d", before, syncLogCount(t, env))
	}
}

func TestScheduledSyncRunsWhenEnabled(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if err := env.Svc.RunScheduledSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	if syncLogCount(t, env) == 0 {
		t.Fatal("expected scheduled sync log")
	}
}

func TestScheduledSyncSkippedWhenDisabled(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: false, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	before := syncLogCount(t, env)
	if err := env.Svc.RunScheduledSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	if syncLogCount(t, env) != before {
		t.Fatalf("expected no sync log when disabled, before=%d after=%d", before, syncLogCount(t, env))
	}
}

func TestUpdateSyncConfigSchedulesOrgSync(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	svc := orgfix.NewServiceWithEnqueuer(t, env.Cfg, env.Store, riverfix.NewInsertOnlyEnqueuer(t, env.Cfg, env.Store))
	ctx := testutil.Ctx()

	if err := svc.UpdateSyncConfig(ctx, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 24,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	}); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingJobCount(env.Store, jobs.KindOrgSync, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected org_sync job after UpdateSyncConfig")
	}
	tbs, err := env.Store.TenantBackgroundState().Get(ctx, contract.DefaultCompanyID)
	if err != nil || tbs == nil || tbs.NextOrgSyncAt == nil {
		t.Fatal("expected tenant_background_state next_org_sync_at")
	}
}

func TestScheduledSyncSelfHealsExpiredScheduleWithoutPendingJob(t *testing.T) {
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	past := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)

	svc := orgfix.NewServiceWithEnqueuer(t, env.Cfg, env.Store, riverfix.NewInsertOnlyEnqueuer(t, env.Cfg, env.Store))
	if err := svc.UpdateSyncConfig(ctx, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 24,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	}); err != nil {
		t.Fatal(err)
	}
	if err := env.Store.TenantBackgroundState().UpsertOrgSchedule(ctx, contract.DefaultCompanyID, past, nil); err != nil {
		t.Fatal(err)
	}
	pool := postgres.MainPool(env.Store)
	if pool == nil {
		t.Fatal("expected postgres pool")
	}
	if _, err := pool.Exec(ctx, `
		UPDATE river_job
		SET state = 'completed', finalized_at = NOW()
		WHERE kind = $1
		  AND (args->>'company_id')::bigint = $2
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
	`, jobs.KindOrgSync, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingJobCount(env.Store, jobs.KindOrgSync, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected no pending org_sync before self-heal")
	}

	acquired, err := env.Store.SchedulerLock().TryAcquire(ctx, types.OrgSyncLockName(contract.DefaultCompanyID), "test-holder", time.Minute)
	if err != nil || !acquired {
		t.Fatalf("expected lock acquired, err=%v acquired=%v", err, acquired)
	}

	if err := svc.RunScheduledSync(ctx); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingJobCount(env.Store, jobs.KindOrgSync, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected org_sync job after ensureScheduledOrgSync self-heal")
	}
	tbs, err := env.Store.TenantBackgroundState().Get(ctx, contract.DefaultCompanyID)
	if err != nil || tbs == nil || tbs.NextOrgSyncAt == nil {
		t.Fatal("expected tenant_background_state next_org_sync_at after self-heal")
	}
}
