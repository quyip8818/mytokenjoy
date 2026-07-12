package worker_test

import (
	"testing"
	"time"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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
	t.Parallel()
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
	t.Parallel()
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

func TestScheduledSyncSkipsDuplicateWithinFrequency(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if err := env.Svc.RunScheduledSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	afterFirst := syncLogCount(t, env)
	if afterFirst == 0 {
		t.Fatal("expected first scheduled sync log")
	}
	if err := env.Svc.RunScheduledSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	if syncLogCount(t, env) != afterFirst {
		t.Fatalf("expected no duplicate scheduled log within frequency, first=%d after=%d", afterFirst, syncLogCount(t, env))
	}
}

func TestFanoutScheduledSyncEnqueuesDueTenant(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})
	svc := orgfix.NewServiceWithEnqueuer(t, env.Cfg, env.Store, riverfix.NewInsertOnlyEnqueuer(t, env.Cfg, env.Store))

	if err := svc.FanoutScheduledSyncJobs(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingJobCount(env.Store, jobs.KindOrgSync, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected org_sync job for due tenant")
	}
}
