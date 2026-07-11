package worker_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	newapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/ingest"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type workerFixture struct {
	rt           *riverfix.TestRuntime
	st           store.Store
	newAPISync   *newapisync.NewAPISync
	ingestWorker *ingest.Worker
	ctx          context.Context
}

func newWorkerFixture(t *testing.T, stub *mock.StubAdminClient) workerFixture {
	t.Helper()
	rt, st := riverfix.NewRuntime(t, stub)
	nas := rt.Registry.MustNewAPISync()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	rt.Start(t, ctx)
	t.Cleanup(func() { rt.Stop(t, ctx) })
	return workerFixture{
		rt:           rt,
		st:           st,
		newAPISync:   nas,
		ingestWorker: rt.Registry.IngestWorker(rt.Registry.Config, logger),
		ctx:          ctx,
	}
}

func (f workerFixture) runRiver(t *testing.T) {
	t.Helper()
	f.rt.WorkOnce(t, f.ctx)
}
