package budgetlookup_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/budgetlookup"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGetReservedPoolForMember(t *testing.T) {
	snapshot := seed.Load(testutil.TestConfig())
	tree := snapshot.BudgetTree
	members := snapshot.Members

	if got := budgetlookup.GetReservedPoolForMember(tree, members, seed.IDMember1); got != 1500 {
		t.Fatalf("expected m-1 reserved pool 1500, got %v", got)
	}
	if got := budgetlookup.GetReservedPoolForMember(tree, members, "m-5"); got != 0 {
		t.Fatalf("expected m-5 reserved pool 0 (dept-4 has none), got %v", got)
	}
}
