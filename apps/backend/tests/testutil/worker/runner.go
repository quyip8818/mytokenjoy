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
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type TestRuntime struct {
	*riverfix.TestRuntime
	t       *testing.T
	started bool
}

func (r *TestRuntime) RunOnce(ctx context.Context) {
	riverfix.TestMu.Lock()
	defer riverfix.TestMu.Unlock()
	if !r.started {
		r.Start(r.t, ctx)
		r.started = true
	}
	r.WorkOnce(r.t, ctx)
}

func NewRuntime(t *testing.T, stub *mock.StubAdminClient) (*TestRuntime, store.Store, *domainusage.IngestService) {
	t.Helper()
	rt, st := riverfix.NewRuntime(t, stub)
	budgetfix.EnsureMonthRebalanceCurrent(t, testutil.Ctx(), rt.Cfg, st, contract.DefaultCompanyID)
	wrapped := &TestRuntime{TestRuntime: rt, t: t}
	t.Cleanup(func() { rt.Stop(t, context.Background()) })
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
