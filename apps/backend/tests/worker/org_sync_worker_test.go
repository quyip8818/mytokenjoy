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
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type recordingOrgSync struct {
	fanoutCalls    int
	scheduledCalls int
}

func (r *recordingOrgSync) GetSyncConfig(context.Context) (types.SyncConfig, error) {
	return types.SyncConfig{}, nil
}

func (r *recordingOrgSync) UpdateSyncConfig(context.Context, types.SyncConfig) error { return nil }

func (r *recordingOrgSync) TriggerSync(context.Context) (types.ImportResult, error) {
	return types.ImportResult{}, nil
}

func (r *recordingOrgSync) RunScheduledSync(context.Context) error {
	r.scheduledCalls++
	return nil
}

func (r *recordingOrgSync) FanoutScheduledSyncJobs(context.Context) error {
	r.fanoutCalls++
	return nil
}

func (r *recordingOrgSync) ListSyncLogs(context.Context, int, int) (types.PageResult[types.SyncLog], error) {
	return types.PageResult[types.SyncLog]{}, nil
}

func startOrgSyncRiver(t *testing.T, recording *recordingOrgSync) (*riverinfra.Client, jobs.Enqueuer) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, holder, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(&mock.StubAdminClient{}), app.WithOrgSync(recording))
	if err != nil {
		t.Fatal(err)
	}
	client, err := riverinfra.NewClient(cfg, postgres.MainPool(st), riverinfra.Deps{OrgSync: reg.OrgSync}, logger)
	if err != nil {
		t.Fatal(err)
	}
	holder.Set(client.Enqueuer)
	if err := client.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Stop(context.Background()) })
	return client, client.Enqueuer
}

func waitUntil(t *testing.T, deadline time.Duration, ok func() bool) {
	t.Helper()
	until := time.Now().Add(deadline)
	for time.Now().Before(until) {
		if ok() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}

func TestOrgSyncWorkerRoutesFanoutAndTenantJobs(t *testing.T) {
	t.Parallel()
	recording := &recordingOrgSync{}
	_, enqueuer := startOrgSyncRiver(t, recording)
	ctx := context.Background()

	if err := jobs.InsertOrgSyncFanout(ctx, enqueuer, nil); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, 5*time.Second, func() bool { return recording.fanoutCalls == 1 })

	if err := jobs.InsertOrgSync(ctx, enqueuer, nil, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, 5*time.Second, func() bool { return recording.scheduledCalls == 1 })
}
