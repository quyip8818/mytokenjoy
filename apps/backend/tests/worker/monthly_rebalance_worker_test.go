package worker_test

import (
	"context"
	"testing"
	"time"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestMonthlyRebalanceWorkerEnqueuesCompanyRebalance(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := context.Background()
	enqueuer := riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st)
	scheduler := domainbudget.NewMonthlyRebalanceScheduler(cfg, st, enqueuer)
	domainbudget.SetLastMonthForTest(scheduler, "2020-01")

	pool := postgres.MainPool(st)
	client, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{MonthlyRebalance: scheduler}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Stop(ctx) })

	if err := jobs.InsertMonthlyRebalance(ctx, client.Enqueuer, nil); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected company rebalance after monthly_rebalance worker")
	}
}
