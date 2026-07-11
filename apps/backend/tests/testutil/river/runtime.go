//go:build testhook

package riverfix

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type TestRuntime struct {
	Client   *riverinfra.Client
	Enqueuer jobs.Enqueuer
	Registry app.ServiceRegistry
	Cfg      config.Config
	st       store.Store
}

func NewRuntime(t *testing.T, stub *mock.StubAdminClient) (*TestRuntime, store.Store) {
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
	pool := postgres.MainPool(st)
	budgetAsync := budget.NewAsync(cfg, st, holder, budgetcheck.Noop{}, logger)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{
		Billing:            reg.BillingSvc,
		Overrun:            reg.Overrun,
		Rebalance:          reg.Rebalance,
		NewAPISync:         reg.MustNewAPISync(),
		OrgSync:            reg.OrgSync,
		MonthlyRebalance:   budget.NewMonthlyRebalanceScheduler(cfg, st, holder),
		BudgetProjector:    budgetAsync.Projector,
		BudgetReconcile:    budgetAsync.Reconcile,
		DashboardProjector: domaindashboard.NewProjector(cfg, st, holder, logger),
		DashboardReconcile: domaindashboard.NewReconcileService(cfg, st, holder, logger),
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	holder.Set(client.Enqueuer)
	return &TestRuntime{Client: client, Enqueuer: holder, Registry: reg, Cfg: cfg, st: st}, st
}

func (r *TestRuntime) Start(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := r.Client.Start(ctx); err != nil {
		t.Fatal(err)
	}
}

func (r *TestRuntime) Stop(t *testing.T, ctx context.Context) {
	t.Helper()
	_ = r.Client.Stop(ctx)
}

func (r *TestRuntime) WorkOnce(t *testing.T, ctx context.Context) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if pendingActiveJobs(r.st) == 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for river jobs to complete; pending=%d", pendingActiveJobs(r.st))
}

func pendingActiveJobs(st store.Store) int {
	ctx := context.Background()
	pool := postgres.MainPool(st)
	if pool == nil {
		return 0
	}
	var count int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM river_job
		WHERE state IN ('available', 'retryable', 'scheduled', 'running')
	`).Scan(&count)
	return count
}

func PendingJobCount(st store.Store, kind string, companyID int64) int {
	ctx := context.Background()
	pool := postgres.MainPool(st)
	if pool == nil {
		return 0
	}
	var count int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND ($2 = 0 OR (args->>'company_id')::bigint = $2)
	`, kind, companyID).Scan(&count)
	return count
}

func HasPendingWalletSync(st store.Store, companyID int64) bool {
	return PendingJobCount(st, jobs.KindWalletSync, companyID) > 0
}

func ListPendingNewAPISync(st store.Store, subKind string, limit int) int {
	ctx := context.Background()
	pool := postgres.MainPool(st)
	if pool == nil {
		return 0
	}
	rows, err := pool.Query(ctx, `
		SELECT 1 FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND ($2 = '' OR args->>'sub_kind' = $2)
		LIMIT $3
	`, jobs.KindNewAPISync, subKind, limit)
	if err != nil {
		return 0
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n
}

func WaitForJobs(t *testing.T, ctx context.Context, r *TestRuntime, rounds int) {
	t.Helper()
	for i := 0; i < rounds; i++ {
		r.WorkOnce(t, ctx)
		time.Sleep(10 * time.Millisecond)
	}
}
