//go:build testhook

package riverfix

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type TestRuntime struct {
	Client   *riverinfra.Client
	Enqueuer jobs.Enqueuer
	Registry app.ServiceRegistry
	Cfg      config.Config
	st       store.Store
	started  bool
}

func NewRuntime(t *testing.T, stub *mock.StubAdminClient) (*TestRuntime, store.Store) {
	return newRuntime(t, stub, nil)
}

func NewRuntimeWithOrgSync(t *testing.T, stub *mock.StubAdminClient, orgSync domainorg.SyncService) (*TestRuntime, store.Store) {
	return newRuntime(t, stub, orgSync)
}

// NewIngestRuntime creates a full River runtime with budget rebalance pre-seeded,
// registers a cleanup, and returns the runtime, store, and IngestService.
// This is the standard entry point for ingest integration tests.
func NewIngestRuntime(t *testing.T, stub *mock.StubAdminClient) (*TestRuntime, store.Store, *domainusage.IngestService) {
	t.Helper()
	rt, st := NewRuntime(t, stub)
	budgetfix.EnsureMonthRebalanceCurrent(t, testutil.Ctx(), rt.Cfg, st, contract.DefaultCompanyID)
	t.Cleanup(func() { rt.Stop(t, context.Background()) })
	return rt, st, rt.Registry.MustIngestService()
}

func newRuntime(t *testing.T, stub *mock.StubAdminClient, orgSync domainorg.SyncService) (*TestRuntime, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
		testutil.WithNewAPIEnabled(true),
		testutil.WithIngestEnabled(true),
		testutil.WithNewAPIWebhookSecret("secret"),
	)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	opts := []app.Option{app.WithAdminClient(stub)}
	if orgSync != nil {
		opts = append(opts, app.WithOrgSync(orgSync))
	}
	reg, holder, err := app.BuildRegistry(cfg, logger, st, opts...)
	if err != nil {
		t.Fatal(err)
	}
	pool := postgres.MainPool(st)
	budgetEnqueuer := budgetEnqueuerFromHolder(holder)
	budgetReconcile := budget.NewReconcileService(cfg, st, budgetEnqueuer, budgetcheck.WrapStore(budgetcheck.Noop{}), logger)
	dashboardEnqueuer := adapter.NewDashboardEnqueuer(holder)
	sched := scheduler.NewService(cfg, st)
	bulk := scheduler.NewBulkEnqueuer(cfg, holder)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{
		Cfg:                cfg,
		Store:              st,
		Billing:            reg.BillingSvc,
		Overrun:            reg.Overrun,
		Rebalance:          reg.Rebalance,
		NewAPISync:         reg.MustNewAPISync(),
		OrgSync:            reg.OrgSync,
		BudgetReconcile:    budgetReconcile,
		DashboardProjector: domaindashboard.NewProjector(cfg, st, dashboardEnqueuer, logger),
		DashboardReconcile: domaindashboard.NewReconcileService(cfg, st, dashboardEnqueuer, logger),
		Scheduler:          sched,
		BulkEnqueuer:       bulk,
		DisablePeriodic:    true,
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

// RunOnce acquires the test mutex, lazily starts the River client, and drains
// all pending jobs. This is the standard way to execute enqueued jobs in tests.
func (r *TestRuntime) RunOnce(t *testing.T, ctx context.Context) {
	t.Helper()
	TestMu.Lock()
	defer TestMu.Unlock()
	if !r.started {
		r.Start(t, ctx)
		r.started = true
	}
	r.WorkOnce(t, ctx)
}

func (r *TestRuntime) WorkOnce(t *testing.T, ctx context.Context) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var sawPending bool
	for time.Now().Before(deadline) {
		n := pendingActiveJobs(r.st)
		if n > 0 {
			sawPending = true
		}
		if sawPending && n == 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for river jobs to complete; pending=%d sawPending=%v", pendingActiveJobs(r.st), sawPending)
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

func PendingRebalanceCount(st store.Store, companyID int64) int {
	return PendingJobCount(st, jobs.KindRebalance, companyID)
}

func PendingOverrunCount(st store.Store, companyID int64) int {
	return PendingJobCount(st, jobs.KindOverrun, companyID)
}

func PendingWalletSyncCount(st store.Store, companyID int64) int {
	return PendingJobCount(st, jobs.KindWalletSync, companyID)
}

func PendingDashboardProjectCount(st store.Store, companyID int64) int {
	return PendingJobCount(st, jobs.KindDashboardProject, companyID)
}

func ListPendingNewAPISync(t testing.TB, st store.Store, subKind string, limit int) int {
	t.Helper()
	ctx := context.Background()
	pool := postgres.MainPool(st)
	if pool == nil {
		t.Fatal("main pool is nil")
	}
	rows, err := pool.Query(ctx, `
		SELECT 1 FROM river_job
		WHERE kind = $1
		  AND state IN ('available', 'retryable', 'scheduled', 'running')
		  AND ($2 = '' OR args->>'sub_kind' = $2)
		LIMIT $3
	`, jobs.KindNewAPISync, subKind, limit)
	if err != nil {
		t.Fatalf("list pending newapi_sync: %v", err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("list pending newapi_sync: %v", err)
	}
	return n
}
