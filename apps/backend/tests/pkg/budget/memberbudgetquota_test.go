package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestValidateMemberQuotaBelowAllocated(t *testing.T) {
	snapshot := seed.Load(testutil.TestConfig())
	tree := types.OrgNodesToBudgetTree(snapshot.OrgNodes)
	members := snapshot.Members
	platformKeys := snapshot.PlatformKeys

	msg := budget.ValidateMemberQuotaUpdate(tree, members, platformKeys, seed.IDMember1, 1000)
	if msg == nil {
		t.Fatal("expected validation error when quota below allocated")
	}
}
