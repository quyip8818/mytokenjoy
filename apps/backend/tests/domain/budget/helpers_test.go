package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func newBudgetService(t *testing.T) (budget.Service, store.Store) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	return budget.NewService(cfg, st, common.NewDelayer(false), nil), st
}

type deptBudgetInputs struct {
	tree     []types.BudgetNode
	projects []types.Project
	members  []types.Member
}

func loadDeptBudgetInputs(t *testing.T, st store.Store, deptID string) deptBudgetInputs {
	t.Helper()
	ctx := testutil.Ctx()
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tree := types.OrgNodesToBudgetTree(nodes)
	if pkgbudget.FindBudgetNode(tree, deptID) == nil {
		t.Fatalf("department %q not found in budget tree", deptID)
	}
	projects, err := st.Budget().Projects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return deptBudgetInputs{tree: tree, projects: projects, members: members}
}

// prepareDept3NodeUpdateFixture relaxes demo-seed allocations on dept-3 so UpdateNode
// tests exercise persistence rather than demo oversubscription (members + projects
// exceed what parent dept-2 can host under ValidateBudgetNodeUpdate).
func prepareDept3NodeUpdateFixture(t *testing.T, st store.Store) {
	t.Helper()
	ctx := testutil.Ctx()
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range members {
		if members[i].DepartmentID == contract.IDDept3 {
			members[i].PersonalBudget = 0
		}
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	projects, err := st.Budget().Projects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range projects {
		if projects[i].OwnerDepartmentID == contract.IDDept3 {
			projects[i].OwnerDepartmentID = contract.IDDept2
		}
	}
	if err := st.Budget().SetProjects(ctx, projects); err != nil {
		t.Fatal(err)
	}
}

// chooseValidDeptBudget picks the smallest budget >= allocated obligations that passes
// the same validation as domain UpdateNode, optionally bumped when parent headroom allows.
func chooseValidDeptBudget(t *testing.T, st store.Store, deptID string, reserved float64) float64 {
	t.Helper()
	inputs := loadDeptBudgetInputs(t, st, deptID)
	node := pkgbudget.FindBudgetNode(inputs.tree, deptID)
	floor := pkgbudget.SumChildrenBudget(*node) +
		reserved +
		pkgbudget.ProjectsBudgetForDept(inputs.projects, deptID) +
		pkgbudget.MemberBudgetSumForDept(inputs.members, deptID)

	candidates := []float64{floor + budgetfix.DisplayPoints(1000), floor}
	for _, budget := range candidates {
		if msg := pkgbudget.ValidateBudgetNodeUpdate(
			inputs.tree, deptID, budget, reserved, inputs.projects, inputs.members,
		); msg == nil {
			return budget
		}
	}
	t.Fatalf("no valid budget for %q with reserved %v (floor=%v)", deptID, reserved, floor)
	return 0
}
