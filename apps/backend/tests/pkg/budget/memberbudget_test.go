package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGetBudgetRemaining(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys

	remaining := budget.GetBudgetRemaining(members, keys, contract.IDMember1)
	if remaining != testutil.DisplayPoints(3000) {
		t.Fatalf("expected remaining %v (10000 personal - 7000 allocated), got %v", testutil.DisplayPoints(3000), remaining)
	}
}

func TestBuildBudgetSummary(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys
	reserved := budget.GetReservedPoolForMember(types.OrgNodesToBudgetTree(snapshot.OrgNodes), members, contract.IDMember1)

	summary := budget.BuildBudgetSummary(members, keys, contract.IDMember1, reserved)
	if summary.TotalBudget != testutil.DisplayPoints(10000) {
		t.Fatalf("expected total quota %v, got %v", testutil.DisplayPoints(10000), summary.TotalBudget)
	}
	if summary.Remaining != testutil.DisplayPoints(3000) {
		t.Fatalf("expected remaining %v, got %v", testutil.DisplayPoints(3000), summary.Remaining)
	}
	if summary.ReservedPool != testutil.DisplayPoints(1500) {
		t.Fatalf("expected reserved pool %v, got %v", testutil.DisplayPoints(1500), summary.ReservedPool)
	}
}
