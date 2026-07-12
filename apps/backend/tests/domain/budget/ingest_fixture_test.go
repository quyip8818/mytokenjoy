//go:build testhook

package budget_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func newBudgetIngestFixture(t *testing.T, stub *mock.StubAdminClient) (config.Config, store.Store, *domainusage.IngestService, jobs.Enqueuer) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
		testutil.WithNewAPIEnabled(true),
		testutil.WithIngestEnabled(true),
		testutil.WithNewAPIWebhookSecret("secret"),
	)
	resetBudgetProjectorCursor(t, st)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, holder, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(stub))
	if err != nil {
		t.Fatal(err)
	}
	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)
	holder.Set(enqueuer)
	return cfg, st, reg.MustIngestService(), enqueuer
}

func defaultBudgetIngestStub() *mock.StubAdminClient {
	return &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
}

func resetBudgetProjectorCursor(t *testing.T, st store.Store) {
	t.Helper()
	pool := postgres.MainPool(st)
	if pool == nil {
		t.Fatal("expected postgres pool")
	}
	_, err := pool.Exec(testutil.Ctx(), `
		UPDATE budget_projection_progress
		SET last_occurred_at = NULL, last_ledger_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND stream = $2
	`, contract.DefaultCompanyID, store.BudgetProjectionStream)
	if err != nil {
		t.Fatalf("reset budget projection cursor: %v", err)
	}
}

type recordingEnqueuer struct {
	inner      jobs.Enqueuer
	rebalances int
	overruns   int
}

func (r *recordingEnqueuer) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error {
	switch args.Kind() {
	case jobs.KindRebalance:
		r.rebalances++
	case jobs.KindOverrun:
		r.overruns++
	}
	return r.inner.Insert(ctx, args, opts)
}

func (r *recordingEnqueuer) InsertInTx(ctx context.Context, tx store.Tx, args river.JobArgs, opts *river.InsertOpts) error {
	switch args.Kind() {
	case jobs.KindRebalance:
		r.rebalances++
	case jobs.KindOverrun:
		r.overruns++
	}
	return r.inner.InsertInTx(ctx, tx, args, opts)
}
