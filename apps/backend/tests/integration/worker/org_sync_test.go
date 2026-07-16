//go:build testhook && integration

package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type recordingOrgSync struct {
	scheduledCalls int
}

func (r *recordingOrgSync) GetSyncConfig(context.Context) (types.SyncConfig, error) {
	return types.SyncConfig{Enabled: true, FrequencyHours: 24}, nil
}

func (r *recordingOrgSync) UpdateSyncConfig(context.Context, types.SyncConfig) error { return nil }

func (r *recordingOrgSync) TriggerSync(context.Context) (types.ImportResult, error) {
	return types.ImportResult{}, nil
}

func (r *recordingOrgSync) RunScheduledSync(context.Context) error {
	r.scheduledCalls++
	return nil
}

func (r *recordingOrgSync) ListSyncLogs(context.Context, int, int) (types.PageResult[types.SyncLog], error) {
	return types.PageResult[types.SyncLog]{}, nil
}

func TestOrgSyncWorkerRunsTenantJob(t *testing.T) {
	riverfix.TestMu.Lock()
	defer riverfix.TestMu.Unlock()

	recording := &recordingOrgSync{}
	rt, _ := riverfix.NewRuntimeWithOrgSync(t, &mock.StubAdminClient{}, recording)
	ctx := testutil.Ctx()
	t.Cleanup(func() { rt.Stop(t, ctx) })

	if err := jobs.InsertOrgSync(ctx, rt.Enqueuer, nil, contract.DefaultCompanyID, nil); err != nil {
		t.Fatal(err)
	}
	rt.Start(t, ctx)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if recording.scheduledCalls == 1 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("expected RunScheduledSync once, got %d", recording.scheduledCalls)
}
