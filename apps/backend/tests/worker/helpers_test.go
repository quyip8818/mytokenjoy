package worker_test

import (
	"log/slog"
	"os"
	"testing"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newWorkerRunner(t *testing.T, stub *mock.StubAdminClient) (*worker.Runner, store.Store, *relay.TokenLifecycle) {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://relay.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	orgSvc := testutil.NewOrgService(t, cfg, st)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	ingest := domainbudget.NewIngestService(cfg, st, lifecycle, notifier, logger)
	overrun := domainbudget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, st, stub, lifecycle)
	runner := worker.NewRunner(cfg, st, stub, lifecycle, ingest, overrun, rebalance, orgSvc, logger)
	return runner, st, lifecycle
}

func pendingRelayOutbox(st store.Store, kind string) int {
	entries, err := st.Relay().ClaimPendingRelayOutbox(testutil.Ctx(), 100)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

func pendingWebhookOutbox(st store.Store) int {
	entries, err := st.Relay().ClaimPendingWebhookOutbox(testutil.Ctx(), 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
