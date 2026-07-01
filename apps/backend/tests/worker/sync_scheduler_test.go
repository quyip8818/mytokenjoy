package worker_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func syncLogCount(t *testing.T, env testutil.FeishuOrgEnv) int {
	t.Helper()
	logs, err := env.Store.Org().SyncLogs(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	return len(logs)
}

func TestScheduledSyncUsesLock(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	acquired, err := env.Store.SchedulerLock().TryAcquire(testutil.Ctx(), types.SchedulerLockOrgSync, "other", time.Minute)
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
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
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
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
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
