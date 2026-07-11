package worker_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type recordingOrgSync struct {
	syncAllCalls int
}

func (r *recordingOrgSync) GetSyncConfig(context.Context) (types.SyncConfig, error) {
	return types.SyncConfig{}, nil
}

func (r *recordingOrgSync) UpdateSyncConfig(context.Context, types.SyncConfig) error { return nil }

func (r *recordingOrgSync) TriggerSync(context.Context) (types.ImportResult, error) {
	return types.ImportResult{}, nil
}

func (r *recordingOrgSync) RunScheduledSync(context.Context) error { return nil }

func (r *recordingOrgSync) RunScheduledSyncAll(context.Context) error {
	r.syncAllCalls++
	return nil
}

func (r *recordingOrgSync) ListSyncLogs(context.Context, int, int) (types.PageResult[types.SyncLog], error) {
	return types.PageResult[types.SyncLog]{}, nil
}

func TestOrgSyncWorkerInvokesRunScheduledSyncAll(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := context.Background()
	recording := &recordingOrgSync{}
	stub := &mock.StubAdminClient{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, holder, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(stub), app.WithOrgSync(recording))
	if err != nil {
		t.Fatal(err)
	}
	pool := postgres.MainPool(st)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{OrgSync: reg.OrgSync}, logger)
	if err != nil {
		t.Fatal(err)
	}
	holder.Set(client.Enqueuer)
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Stop(ctx) })

	if err := jobs.InsertOrgSync(ctx, client.Enqueuer, nil); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && recording.syncAllCalls == 0 {
		time.Sleep(50 * time.Millisecond)
	}
	if recording.syncAllCalls != 1 {
		t.Fatalf("expected RunScheduledSyncAll once, got %d", recording.syncAllCalls)
	}
}
