//go:build testhook

package budget_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
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

type recordingBudgetEnqueuer struct {
	inner      budget.JobEnqueuer
	rebalances int
	overruns   int
}

func (r *recordingBudgetEnqueuer) InsertOverrun(ctx context.Context, companyID int64, payload []byte) error {
	r.overruns++
	return r.inner.InsertOverrun(ctx, companyID, payload)
}

func (r *recordingBudgetEnqueuer) InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error {
	r.rebalances++
	return r.inner.InsertRebalance(ctx, companyID, axisKind, axisID)
}

func (r *recordingBudgetEnqueuer) InsertBudgetReconcile(ctx context.Context, companyID int64) error {
	return r.inner.InsertBudgetReconcile(ctx, companyID)
}
