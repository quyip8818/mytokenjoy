package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestValidateBudgetNodeUpdateIncludesProjectsAndMembers(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{ID: "dept-a", Budget: 100000},
	}
	groups := []types.Project{
		{ID: "proj-1", Budget: 10000, OwnerDepartmentID: "dept-a"},
	}
	members := []types.Member{
		{ID: "m-1", DepartmentID: "dept-a", PersonalBudget: 20000},
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, "dept-a", 25000, 5000, groups, members); msg == nil {
		t.Fatal("expected budget below project+member+reserved allocation to fail")
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, "dept-a", 35000, 5000, groups, members); msg != nil {
		t.Fatalf("expected budget covering allocations to pass, got %s", *msg)
	}
}

func TestValidateBudgetNodeUpdate(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{
			ID: "root", Budget: 100000, ReservedPool: budgetfix.FloatPtr(10000),
			Children: []types.BudgetNode{
				{ID: "dept-a", Budget: 40000, ReservedPool: budgetfix.FloatPtr(5000)},
				{ID: "dept-b", Budget: 40000},
			},
		},
	}

	if msg := budget.ValidateBudgetNodeUpdate(tree, "dept-a", 45000, 5000, nil, nil); msg != nil {
		t.Fatalf("expected valid increase, got %s", *msg)
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, "dept-a", 90000, 5000, nil, nil); msg == nil {
		t.Fatal("expected oversell against parent")
	}
}

func TestValidateBudgetNodeUpdateSiblingOversell(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{
			ID: "root", Budget: 100000, ReservedPool: budgetfix.FloatPtr(10000),
			Children: []types.BudgetNode{
				{ID: "dept-a", Budget: 40000, ReservedPool: budgetfix.FloatPtr(5000)},
				{ID: "dept-b", Budget: 40000},
			},
		},
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, "dept-b", 55000, 0, nil, nil); msg == nil {
		t.Fatal("expected sibling sum to exceed parent capacity")
	}
}

func TestGetMemberBudgetCapacity(t *testing.T) {
	t.Parallel()
	reserved := 2000.0
	node := types.BudgetNode{
		ID: "dept", Budget: 20000, ReservedPool: &reserved,
		Children: []types.BudgetNode{
			{ID: "child-a", Budget: 8000},
			{ID: "child-b", Budget: 5000},
		},
	}
	capacity := budget.GetMemberBudgetCapacity(node)
	want := 20000.0 - 2000.0 - 8000.0 - 5000.0
	if capacity != want {
		t.Fatalf("expected capacity %f, got %f", want, capacity)
	}
}
