package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestValidateBudgetNodeUpdateIncludesProjectsAndMembers(t *testing.T) {
	t.Parallel()
	deptA := uuid.MustParse("00000000-0000-7000-0000-00000000da01")
	tree := []types.BudgetNode{
		{ID: deptA, Budget: 100000},
	}
	groups := []types.Project{
		{ID: uuid.MustParse("00000000-0000-7000-0000-000000000a01"), Budget: 10000, OwnerDepartmentID: deptA},
	}
	members := []types.Member{
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000ee01"), DepartmentID: deptA, PersonalBudget: 20000},
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, deptA, 25000, 5000, groups, members); msg == nil {
		t.Fatal("expected budget below project+member+reserved allocation to fail")
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, deptA, 35000, 5000, groups, members); msg != nil {
		t.Fatalf("expected budget covering allocations to pass, got %s", *msg)
	}
}

func TestValidateBudgetNodeUpdate(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{
			ID: uuid.MustParse("00000000-0000-7000-0000-00000000ff01"), Budget: 100000, ReservedPool: budgetfix.Int64Ptr(10000),
			Children: []types.BudgetNode{
				{ID: uuid.MustParse("00000000-0000-7000-0000-00000000da01"), Budget: 40000, ReservedPool: budgetfix.Int64Ptr(5000)},
				{ID: uuid.MustParse("00000000-0000-7000-0000-00000000da02"), Budget: 40000},
			},
		},
	}

	if msg := budget.ValidateBudgetNodeUpdate(tree, uuid.MustParse("00000000-0000-7000-0000-00000000da01"), 45000, 5000, nil, nil); msg != nil {
		t.Fatalf("expected valid increase, got %s", *msg)
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, uuid.MustParse("00000000-0000-7000-0000-00000000da01"), 90000, 5000, nil, nil); msg == nil {
		t.Fatal("expected oversell against parent")
	}
}

func TestValidateBudgetNodeUpdateSiblingOversell(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{
			ID: uuid.MustParse("00000000-0000-7000-0000-00000000ff01"), Budget: 100000, ReservedPool: budgetfix.Int64Ptr(10000),
			Children: []types.BudgetNode{
				{ID: uuid.MustParse("00000000-0000-7000-0000-00000000da01"), Budget: 40000, ReservedPool: budgetfix.Int64Ptr(5000)},
				{ID: uuid.MustParse("00000000-0000-7000-0000-00000000da02"), Budget: 40000},
			},
		},
	}
	if msg := budget.ValidateBudgetNodeUpdate(tree, uuid.MustParse("00000000-0000-7000-0000-00000000da02"), 55000, 0, nil, nil); msg == nil {
		t.Fatal("expected sibling sum to exceed parent capacity")
	}
}

func TestGetMemberBudgetCapacity(t *testing.T) {
	t.Parallel()
	reserved := int64(2000)
	node := types.BudgetNode{
		ID: uuid.MustParse("00000000-0000-7000-0000-00000000da05"), Budget: 20000, ReservedPool: &reserved,
		Children: []types.BudgetNode{
			{ID: uuid.MustParse("00000000-0000-7000-0000-00000000ca01"), Budget: 8000},
			{ID: uuid.MustParse("00000000-0000-7000-0000-00000000ca02"), Budget: 5000},
		},
	}
	capacity := budget.GetMemberBudgetCapacity(node)
	want := int64(20000 - 2000 - 8000 - 5000)
	if capacity != want {
		t.Fatalf("expected capacity %d, got %d", want, capacity)
	}
}
