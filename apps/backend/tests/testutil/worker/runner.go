//go:build testhook

package workerfix

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	newapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(stub))
	if err != nil {
		t.Fatal(err)
	}
	return reg.WorkerRunner(logger), st, reg.MustNewAPISync(), reg.MustIngestService()
}

func PendingNewAPISyncOutbox(st store.Store, kind string) int {
	ctx := testutil.Ctx()
	entries, err := postgres.ListPendingNewAPISyncOutbox(ctx, postgres.MainPool(st), kind, 100)
	if err != nil {
		return 0
	}
	return len(entries)
}
