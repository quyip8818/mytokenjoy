package budget_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/budget"
	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/support/clock"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestRotatePeriodAppliesSnapshot(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx()
	cfg, st := testutil.NewTestStore(t)

	// Use a clock at June 19 — current period is 2026-06.
	juneClock := clock.Fixed(time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC))
	svc := budget.NewService(cfg, st, common.NewDelayer(false), nil, juneClock)

	// Mark current period as already rotated (simulate normal state mid-month).
	junePeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC))
	if err := st.TenantBackgroundState().EnsureRow(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().SetLastRebalancedPeriod(ctx, contract.DefaultCompanyID, junePeriod); err != nil {
		t.Fatal(err)
	}

	// Pre-configure July (next month) snapshot with a modified dept budget.
	julyPeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if err := svc.CopyPeriod(ctx, julyPeriod); err != nil {
		t.Fatal(err)
	}
	newDeptBudget := 77777.0
	if err := svc.UpdateSnapshotNode(ctx, julyPeriod, contract.IDDept3, newDeptBudget, nil); err != nil {
		t.Fatal(err)
	}

	// Advance clock to July — now RotatePeriod should apply the snapshot.
	julyClock := clock.Fixed(time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC))
	svcJuly := budget.NewService(cfg, st, common.NewDelayer(false), nil, julyClock)

	if err := svcJuly.RotatePeriod(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify: live budget for dept3 should now be 77777.
	tree, err := svcJuly.GetTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	node := findNodeInPayload(tree, contract.IDDept3)
	if node == nil {
		t.Fatal("dept-3 not found in tree after rotation")
	}
	if node.Budget != newDeptBudget {
		t.Fatalf("expected dept budget %v after rotation, got %v", newDeptBudget, node.Budget)
	}

	// Verify: previous period (June) was archived.
	snap, err := st.BudgetSnapshot().Get(ctx, junePeriod)
	if err != nil {
		t.Fatal(err)
	}
	if snap == nil {
		t.Fatal("expected previous period snapshot to be archived")
	}

	// Verify: TBS marked as rotated to July.
	tbs, err := st.TenantBackgroundState().Get(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if tbs.LastRebalancedPeriod != julyPeriod {
		t.Fatalf("expected lastRebalancedPeriod=%s, got %s", julyPeriod, tbs.LastRebalancedPeriod)
	}
}

func TestRotatePeriodIdempotent(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx()
	cfg, st := testutil.NewTestStore(t)

	// Start at July 1, with lastRebalancedPeriod = June (needs rotation).
	julyClock := clock.Fixed(time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC))
	svc := budget.NewService(cfg, st, common.NewDelayer(false), nil, julyClock)

	junePeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if err := st.TenantBackgroundState().EnsureRow(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if err := st.TenantBackgroundState().SetLastRebalancedPeriod(ctx, contract.DefaultCompanyID, junePeriod); err != nil {
		t.Fatal(err)
	}

	// First call: should rotate.
	if err := svc.RotatePeriod(ctx); err != nil {
		t.Fatal("first RotatePeriod failed:", err)
	}

	// Capture tree state.
	treeAfterFirst, err := svc.GetTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	treeJSON1, _ := json.Marshal(treeAfterFirst)

	// Second call: should be no-op (already rotated).
	if err := svc.RotatePeriod(ctx); err != nil {
		t.Fatal("second RotatePeriod failed:", err)
	}

	// Tree should be identical.
	treeAfterSecond, err := svc.GetTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	treeJSON2, _ := json.Marshal(treeAfterSecond)

	if string(treeJSON1) != string(treeJSON2) {
		t.Fatal("expected tree unchanged after idempotent RotatePeriod")
	}
}
