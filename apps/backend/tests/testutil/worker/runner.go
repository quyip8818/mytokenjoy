//go:build testhook

package workerfix

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	newapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

func NewRunner(t *testing.T, stub *mock.StubAdminClient) (*worker.Runner, store.Store, *newapisync.NewAPISync, *domainusage.IngestService) {
	t.Helper()
	return newRunner(t, stub, true)
}

func NewIngestOnlyRunner(t *testing.T) (*worker.Runner, store.Store, *domainusage.IngestService) {
	t.Helper()
	runner, st, _, ingest := newRunner(t, &mock.StubAdminClient{}, true)
	return runner, st, ingest
}

func newRunner(t *testing.T, stub *mock.StubAdminClient, ingestEnabled bool) (*worker.Runner, store.Store, *newapisync.NewAPISync, *domainusage.IngestService) {
	t.Helper()
	opts := []testutil.ConfigOption{
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
		testutil.WithNewAPIEnabled(true),
	}
	if ingestEnabled {
		opts = append(opts, testutil.WithIngestEnabled(true), testutil.WithNewAPIWebhookSecret("secret"))
	}
	cfg, st := testutil.NewTestStore(t, opts...)
	newAPISync := newapisync.New(cfg, st, stub, nil, newapisync.NewChannelPolicy(cfg))
	orgSvc := orgfix.NewService(t, cfg, st)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	enqueueWalletSync := app.EnqueueWalletSync(st)
	ingest := domainusage.NewIngestService(cfg, st, st.Logs(), notifier, logger, enqueueWalletSync, app.EnqueueRebalanceAxis(st))
	ingestQueue := domainusage.NewQueue(st.Logs())
	overrun := domainbudget.NewOverrunService(cfg, st, newAPISync, notifier, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, st, stub)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	billingSvc := domainbilling.NewService(cfg, st, reader, stub, nil, nil, enqueueWalletSync)
	runner := worker.NewRunner(cfg, st.AsyncJobs(), st.SchedulerLock(), st.Company(), st.Logs(), ingestmetrics.NewCollector(), newAPISync, ingest, ingestQueue, overrun, rebalance, billingSvc, orgSvc, logger)
	return runner, st, newAPISync, ingest
}

func PendingNewAPISyncOutbox(st store.Store, kind string) int {
	ctx := testutil.Ctx()
	entries, err := postgres.ListPendingNewAPISyncOutbox(ctx, postgres.MainPool(st), kind, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
