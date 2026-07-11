//go:build testhook

package workerfix

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingest"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type TestRuntime struct {
	*riverfix.TestRuntime
	t *testing.T
}

func (r *TestRuntime) RunOnce(ctx context.Context) {
	r.WorkOnce(r.t, ctx)
}

func NewRuntime(t *testing.T, stub *mock.StubAdminClient) (*TestRuntime, store.Store, *domainusage.IngestService) {
	t.Helper()
	rt, st := riverfix.NewRuntime(t, stub)
	wrapped := &TestRuntime{TestRuntime: rt, t: t}
	ctx := context.Background()
	rt.Start(t, ctx)
	t.Cleanup(func() { rt.Stop(t, ctx) })
	return wrapped, st, rt.Registry.MustIngestService()
}

func NewIngestOnlyRunner(t *testing.T) (*ingest.Worker, store.Store, *ingest.Worker) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, _, err := app.BuildRegistry(cfg, logger, st)
	if err != nil {
		t.Fatal(err)
	}
	w := reg.IngestWorker(cfg, logger)
	return w, st, w
}
