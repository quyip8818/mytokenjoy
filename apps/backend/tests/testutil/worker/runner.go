//go:build testhook

package workerfix

import (
	"context"
	"log/slog"
	"os"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
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

func NewRunner(t *testing.T, stub *mock.StubAdminClient) (*worker.Runner, store.Store, *relay.TokenLifecycle, *domainusage.IngestService) {
	t.Helper()
	return newRunner(t, stub, true, true)
}

func NewIngestOnlyRunner(t *testing.T) (*worker.Runner, store.Store, *domainusage.IngestService) {
	t.Helper()
	runner, st, _, ingest := newRunner(t, &mock.StubAdminClient{}, false, true)
	return runner, st, ingest
}

func newRunner(t *testing.T, stub *mock.StubAdminClient, newAPIEnabled, ingestEnabled bool) (*worker.Runner, store.Store, *relay.TokenLifecycle, *domainusage.IngestService) {
	t.Helper()
	opts := []testutil.ConfigOption{
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
	}
	if newAPIEnabled {
		opts = append(opts, testutil.WithNewAPIEnabled(true))
	}
	if ingestEnabled {
		opts = append(opts, testutil.WithIngestEnabled(true), testutil.WithNewAPIWebhookSecret("secret"))
	}
	cfg, st := testutil.NewTestStore(t, opts...)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	orgSvc := orgfix.NewService(t, cfg, st)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	enqueueWalletSync := func(ctx context.Context, companyID int64) error {
		return st.Relay().EnqueueWalletSync(company.WithContext(ctx, company.Context{CompanyID: companyID}), companyID)
	}
	ingest := domainusage.NewIngestService(cfg, st, st.Logs(), notifier, logger, enqueueWalletSync)
	failureRecorder := domainusage.NewFailureRecorder(st.Logs(), logger)
	overrun := domainbudget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, st, stub)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	billingSvc := domainbilling.NewService(cfg, st, reader, stub, nil, nil, enqueueWalletSync)
	runner := worker.NewRunner(cfg, st.Relay(), st.SchedulerLock(), st.Logs(), ingestmetrics.NewCollector(), lifecycle, ingest, failureRecorder, overrun, rebalance, billingSvc, orgSvc, logger)
	return runner, st, lifecycle, ingest
}

func PendingRelayOutbox(st store.Store, kind string) int {
	ctx := testutil.Ctx()
	entries, err := postgres.ListPendingRelayOutbox(ctx, postgres.MainPool(st), kind, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
