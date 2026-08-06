package budget_test

import (
	"testing"

	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOpenBudgetPeriodAlignsTreeAndDepartmentFactory(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), contract.IDDept3, testutil.TestClock())
	if err != nil {
		t.Fatal(err)
	}
	if open.String() != "2026-06" {
		t.Fatalf("OpenDepartmentPeriod = %q, want 2026-06", open.String())
	}

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	if node == nil {
		t.Fatal("dept-3 missing from tree")
	}
	treePeriod := pkgbudget.OpenSnapshotKey(node.Period, testutil.TestClock())
	if treePeriod.String() != open.String() {
		t.Fatalf("tree open period %q != department open %q", treePeriod.String(), open.String())
	}
}
