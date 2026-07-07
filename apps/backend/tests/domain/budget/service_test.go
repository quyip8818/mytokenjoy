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
)

func TestUpdateNodeSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	reserved := 1500.0
	updated, err := svc.UpdateNode(testutil.Ctx(), contract.IDDept3, 21000, &reserved)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Budget != 21000 {
		t.Fatalf("expected budget 21000, got %v", updated.Budget)
	}
	nodeTree, err := common.LoadBudgetTree(testutil.Ctx(), st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := pkgbudget.FindBudgetNode(nodeTree, contract.IDDept3)
	if node == nil || node.Budget != 21000 {
		t.Fatalf("expected persisted budget 21000, got %+v", node)
	}
}

func TestUpdateNodeOversell(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	reserved := 1500.0
	_, err := svc.UpdateNode(testutil.Ctx(), contract.IDDept3, 90000, &reserved)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberQuotaBelowAllocated(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	_, err := svc.UpdateMemberQuota(testutil.Ctx(), contract.IDMember1, 1000)
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberQuotaSuccess(t *testing.T) {
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
	svc := budget.NewService(cfg, st, common.NewDelayer(false))

	result, err := svc.UpdateMemberQuota(testutil.Ctx(), contract.IDMember1, 15000)
	if err != nil {
		t.Fatal(err)
	}
	if result.PersonalQuota != 15000 {
		t.Fatalf("expected personal quota 15000, got %v", result.PersonalQuota)
	}
	poolMap, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var pool float64
	for _, member := range poolMap {
		if member.ID == contract.IDMember1 {
			pool = member.PersonalQuota
			break
		}
	}
	if pool != 15000 {
		t.Fatalf("expected member personal quota 15000, got %v", pool)
	}
}

func TestListMemberQuotasUnknownDept(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	_, err := svc.ListMemberQuotas(testutil.Ctx(), "dept-missing")
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestCreateGroup(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	created, err := svc.CreateGroup(testutil.Ctx(), types.BudgetGroup{
		Name: "Test Group", Budget: 5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected created group id")
	}
	groups, err := st.Budget().Groups(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, group := range groups {
		if group.ID == created.ID && group.Name == "Test Group" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created group not found in store")
	}
}

func TestDeleteGroup(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	if err := svc.DeleteGroup(testutil.Ctx(), contract.IDBudgetGroup4); err != nil {
		t.Fatal(err)
	}
	groups, err := st.Budget().Groups(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup4 {
			t.Fatal("expected bg-4 deleted")
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
