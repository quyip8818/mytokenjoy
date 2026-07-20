package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestGetBudgetRemaining(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys

	remaining := budget.GetBudgetRemaining(members, keys, contract.IDMember1)
	if remaining != budgetfix.DisplayPoints(3000) {
		t.Fatalf("expected remaining %v (10000 personal - 7000 allocated), got %v", budgetfix.DisplayPoints(3000), remaining)
	}
}

func TestBuildBudgetSummary(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys
	reserved := budget.GetReservedPoolForMember(types.OrgNodesToBudgetTree(snapshot.OrgNodes), members, contract.IDMember1)

	summary := budget.BuildBudgetSummary(members, keys, contract.IDMember1, reserved)
	if summary.TotalBudget != budgetfix.DisplayPoints(10000) {
		t.Fatalf("expected total quota %v, got %v", budgetfix.DisplayPoints(10000), summary.TotalBudget)
	}
	if summary.Remaining != budgetfix.DisplayPoints(3000) {
		t.Fatalf("expected remaining %v, got %v", budgetfix.DisplayPoints(3000), summary.Remaining)
	}
	if summary.ReservedPool != budgetfix.DisplayPoints(1500) {
		t.Fatalf("expected reserved pool %v, got %v", budgetfix.DisplayPoints(1500), summary.ReservedPool)
	}
}

func TestValidateMemberBudgetBelowAllocated(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	tree := types.OrgNodesToBudgetTree(snapshot.OrgNodes)
	members := snapshot.Members
	platformKeys := snapshot.PlatformKeys

	msg := budget.ValidateMemberBudgetUpdate(tree, members, platformKeys, contract.IDMember1, 1000)
	if msg == nil {
		t.Fatal("expected validation error when budget below allocated")
	}
}

func TestValidateMemberBudgetExceedsDeptCapacity(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000dd03"), Budget: 20000, ReservedPool: budgetfix.Int64Ptr(2000)},
	}
	members := []types.Member{
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000ee01"), DepartmentID: uuid.MustParse("00000000-0000-7000-0000-00000000dd03"), PersonalBudget: 10000},
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000ee02"), DepartmentID: uuid.MustParse("00000000-0000-7000-0000-00000000dd03"), PersonalBudget: 5000},
	}
	msg := budget.ValidateMemberBudgetUpdate(tree, members, nil, uuid.MustParse("00000000-0000-7000-0000-00000000ee02"), 10000)
	if msg == nil {
		t.Fatal("expected validation error when exceeding dept capacity")
	}
}
