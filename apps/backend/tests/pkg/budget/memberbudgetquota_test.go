package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestValidateMemberQuotaBelowAllocated(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	tree := types.OrgNodesToBudgetTree(snapshot.OrgNodes)
	members := snapshot.Members
	platformKeys := snapshot.PlatformKeys

	msg := budget.ValidateMemberBudgetUpdate(tree, members, platformKeys, contract.IDMember1, 1000)
	if msg == nil {
		t.Fatal("expected validation error when quota below allocated")
	}
}

func TestValidateMemberQuotaExceedsDeptCapacity(t *testing.T) {
	t.Parallel()
	tree := []types.BudgetNode{
		{ID: "dept-3", Budget: 20000, ReservedPool: budgetfix.FloatPtr(2000)},
	}
	members := []types.Member{
		{ID: "m-1", DepartmentID: "dept-3", PersonalBudget: 10000},
		{ID: "m-2", DepartmentID: "dept-3", PersonalBudget: 5000},
	}
	msg := budget.ValidateMemberBudgetUpdate(tree, members, nil, "m-2", 10000)
	if msg == nil {
		t.Fatal("expected validation error when exceeding dept capacity")
	}
}
