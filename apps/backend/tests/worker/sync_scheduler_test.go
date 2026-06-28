package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestScheduledSyncUsesLock(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	acquired, err := env.Store.SchedulerLock().TryAcquire(context.Background(), types.SchedulerLockOrgSync, "other", time.Minute)
	if err != nil || !acquired {
		t.Fatalf("expected lock acquired, err=%v acquired=%v", err, acquired)
	}

	before := len(env.Store.Org().SyncLogs())
	if err := env.Svc.RunScheduledSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(env.Store.Org().SyncLogs()) != before {
		t.Fatalf("expected no new sync log while lock held, before=%d after=%d", before, len(env.Store.Org().SyncLogs()))
	}
}

func TestScheduledSyncRunsWhenEnabled(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if err := env.Svc.RunScheduledSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(env.Store.Org().SyncLogs()) == 0 {
		t.Fatal("expected scheduled sync log")
	}
}

func TestScheduledSyncSkipsDuplicateWithinFrequency(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if err := env.Svc.RunScheduledSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	afterFirst := len(env.Store.Org().SyncLogs())
	if afterFirst == 0 {
		t.Fatal("expected first scheduled sync log")
	}
	if err := env.Svc.RunScheduledSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(env.Store.Org().SyncLogs()) != afterFirst {
		t.Fatalf("expected no duplicate scheduled log within frequency, first=%d after=%d", afterFirst, len(env.Store.Org().SyncLogs()))
	}
}
