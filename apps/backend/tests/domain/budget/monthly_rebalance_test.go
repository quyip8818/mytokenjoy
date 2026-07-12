package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestMonthlyRebalanceEnqueuesOnMonthChange(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	enqueuer := riverfix.NewBudgetInsertOnlyEnqueuer(t, cfg, st)

	scheduler := budget.NewMonthlyRebalanceScheduler(cfg, st, enqueuer)
	budget.SetLastMonthForTest(scheduler, "2020-01")

	if err := scheduler.EnqueueMonthlyRebalanceAll(ctx); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected company rebalance after month change")
	}

	before := riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID)
	if err := scheduler.EnqueueMonthlyRebalanceAll(ctx); err != nil {
		t.Fatal(err)
	}
	if riverfix.PendingRebalanceCount(st, contract.DefaultCompanyID) != before {
		t.Fatal("expected no duplicate monthly rebalance within same month")
	}
}

func TestMonthlyRebalanceWorkerKind(t *testing.T) {
	t.Parallel()
	var args jobs.MonthlyRebalanceArgs
	if args.Kind() != jobs.KindMonthlyRebalance {
		t.Fatalf("expected kind %q, got %q", jobs.KindMonthlyRebalance, args.Kind())
	}
}
