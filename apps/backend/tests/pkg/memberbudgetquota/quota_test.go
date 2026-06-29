package memberbudgetquota_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/memberbudgetquota"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestValidateMemberQuotaBelowAllocated(t *testing.T) {
	snapshot := seed.Load(testutil.TestConfig())
	tree := snapshot.BudgetTree
	members := snapshot.Members
	pools := snapshot.MemberQuotaPools
	platformKeys := snapshot.PlatformKeys

	msg := memberbudgetquota.ValidateMemberQuotaUpdate(tree, members, pools, platformKeys, seed.IDMember1, 1000)
	if msg == nil {
		t.Fatal("expected validation error when quota below allocated")
	}
}
