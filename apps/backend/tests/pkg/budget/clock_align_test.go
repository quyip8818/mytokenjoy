package budget_test

import (
	"testing"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOpenBudgetPeriodAlignsTreeAndDepartmentFactory(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithClockAnchor("2026-06-19"))
	ctx := testutil.Ctx()

	open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), contract.IDDept3, cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if open.String() != "2026-06" {
		t.Fatalf("OpenDepartmentPeriod = %q, want 2026-06", open.String())
	}

	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, st.BudgetSnapshots(), st.Org().Nodes(), cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	if node == nil {
		t.Fatal("dept-3 missing from tree")
	}
	treePeriod := pkgbudget.OpenSnapshotKey(node.Period, cfg.Clock())
	if treePeriod.String() != open.String() {
		t.Fatalf("tree open period %q != department open %q", treePeriod.String(), open.String())
	}
}
