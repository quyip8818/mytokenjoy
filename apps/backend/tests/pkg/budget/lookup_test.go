package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/seed/contract"
)

func TestGetReservedPoolForMember(t *testing.T) {
	t.Parallel()
	snapshot := cachedSnapshot()
	tree := types.OrgNodesToBudgetTree(snapshot.OrgNodes)
	members := snapshot.Members

	if got := budget.GetReservedPoolForMember(tree, members, contract.IDMember1); got != 1500 {
		t.Fatalf("expected m-1 reserved pool %v, got %v", 1500, got)
	}
	if got := budget.GetReservedPoolForMember(tree, members, contract.IDMember5); got != 0 {
		t.Fatalf("expected m-5 reserved pool 0 (dept-4 has none), got %v", got)
	}
}
