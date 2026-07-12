package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestUpdateNodeSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	prepareDept3NodeUpdateFixture(t, st)
	reserved := budgetfix.DisplayPoints(1500)
	wantBudget := chooseValidDeptBudget(t, st, contract.IDDept3, reserved)
	updated, err := svc.UpdateNode(testutil.Ctx(), contract.IDDept3, wantBudget, &reserved)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Budget != wantBudget {
		t.Fatalf("expected budget %v, got %v", wantBudget, updated.Budget)
	}
	nodeTree, err := common.LoadBudgetTree(testutil.Ctx(), st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(nodeTree, contract.IDDept3)
	if node == nil || node.Budget != wantBudget {
		t.Fatalf("expected persisted budget %v, got %+v", wantBudget, node)
	}
	row, found, err := st.Budget().OrgNodeBudget().Get(testutil.Ctx(), contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("org_node_budget row missing: found=%v err=%v", found, err)
	}
	if row.Budget != wantBudget {
		t.Fatalf("org_node_budget budget: got %v want %v", row.Budget, wantBudget)
	}
	if row.ReservedPool == nil || *row.ReservedPool != reserved {
		t.Fatalf("org_node_budget reserved: got %+v want %v", row.ReservedPool, reserved)
	}
}

func TestUpdateNodeOversell(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	reserved := budgetfix.DisplayPoints(1500)
	_, err := svc.UpdateNode(testutil.Ctx(), contract.IDDept3, budgetfix.DisplayPoints(90000), &reserved)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberBudgetBelowAllocated(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	_, err := svc.UpdateMemberBudget(testutil.Ctx(), contract.IDMember1, 1000)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberBudgetSuccess(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	members, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	filtered := make([]types.Member, 0, len(members))
	for _, member := range members {
		if member.DepartmentID == contract.IDDept3 && member.ID != contract.IDMember1 {
			continue
		}
		filtered = append(filtered, member)
	}
	if err := st.Org().SetMembers(testutil.Ctx(), filtered); err != nil {
		t.Fatal(err)
	}
	svc := budget.NewService(cfg, st, common.NewDelayer(false), nil)

	wantQuota := budgetfix.DisplayPoints(15000)
	result, err := svc.UpdateMemberBudget(testutil.Ctx(), contract.IDMember1, wantQuota)
	if err != nil {
		t.Fatal(err)
	}
	if result.PersonalBudget != wantQuota {
		t.Fatalf("expected personal quota %v, got %v", wantQuota, result.PersonalBudget)
	}
	poolMap, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var pool float64
	for _, member := range poolMap {
		if member.ID == contract.IDMember1 {
			pool = member.PersonalBudget
			break
		}
	}
	if pool != wantQuota {
		t.Fatalf("expected member personal quota %v, got %v", wantQuota, pool)
	}
}

func TestListMemberBudgetsUnknownDept(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	_, err := svc.ListMemberBudgets(testutil.Ctx(), "dept-missing")
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestCreateProject(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	created, err := svc.CreateProject(testutil.Ctx(), types.Project{
		Name:              "Test Project",
		Budget:            5000,
		OwnerDepartmentID: contract.IDDept3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected created project id")
	}
	projects, err := st.Budget().Projects(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, project := range projects {
		if project.ID == created.ID && project.Name == "Test Project" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created project not found in store")
	}
}

func TestUpdateProjectMemberIDsPreservesBudget(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	created, err := svc.CreateProject(testutil.Ctx(), types.Project{
		Name:              "Patch Members",
		Budget:            5000,
		OwnerDepartmentID: contract.IDDept3,
		MemberIDs:         []string{contract.IDMember1},
	})
	if err != nil {
		t.Fatal(err)
	}
	memberIDs := []string{contract.IDMember1, contract.IDMember3}
	updated, err := svc.UpdateProject(testutil.Ctx(), created.ID, types.UpdateProjectInput{
		MemberIDs: &memberIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Budget != 5000 {
		t.Fatalf("expected budget preserved at 5000, got %v", updated.Budget)
	}
	if len(updated.MemberIDs) != 2 {
		t.Fatalf("expected 2 members, got %v", updated.MemberIDs)
	}
	stored, err := st.Budget().Projects(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range stored {
		if project.ID == created.ID && project.Budget != 5000 {
			t.Fatalf("expected persisted budget 5000, got %v", project.Budget)
		}
	}
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	if err := svc.DeleteProject(testutil.Ctx(), contract.IDProject4); err != nil {
		t.Fatal(err)
	}
	projects, err := st.Budget().Projects(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range projects {
		if project.ID == contract.IDProject4 {
			t.Fatal("expected proj-4 deleted")
		}
	}
}

func TestDeptRemainingAllocatableBudget(t *testing.T) {
	t.Parallel()
	_, st := newBudgetService(t)
	ctx := testutil.Ctx()
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	dept3 := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	if dept3 == nil {
		t.Fatal("dept-3 not found")
	}
	childrenSum := 0.0
	for _, child := range dept3.Children {
		childrenSum += child.Budget
	}
	reserved := 0.0
	if dept3.ReservedPool != nil {
		reserved = *dept3.ReservedPool
	}
	remaining := dept3.Budget - reserved - childrenSum
	if remaining <= 0 {
		t.Fatalf("expected positive remaining allocatable, got %f", remaining)
	}
}

func TestOrgSyncSetTreeDoesNotOverwriteBudget(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	prepareDept3NodeUpdateFixture(t, st)
	ctx := testutil.Ctx()
	reserved := budgetfix.DisplayPoints(1500)
	wantBudget := chooseValidDeptBudget(t, st, contract.IDDept3, reserved)
	if _, err := svc.UpdateNode(ctx, contract.IDDept3, wantBudget, &reserved); err != nil {
		t.Fatal(err)
	}
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var touch func([]types.OrgNode)
	touch = func(list []types.OrgNode) {
		for i := range list {
			if list[i].ID == contract.IDDept3 {
				list[i].Name = list[i].Name + " Synced"
				list[i].Budget = 0
			}
			if len(list[i].Children) > 0 {
				touch(list[i].Children)
			}
		}
	}
	touch(nodes)
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}
	limit, found, err := st.Org().Nodes().GetNodeBudget(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatalf("get budget: found=%v err=%v", found, err)
	}
	if limit != wantBudget {
		t.Fatalf("expected budget %v unchanged after org sync, got %v", wantBudget, limit)
	}
}
