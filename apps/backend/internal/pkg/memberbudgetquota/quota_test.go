package memberbudgetquota_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/memberbudgetquota"
	"github.com/tokenjoy/backend/internal/seed"
)

func TestValidateMemberQuotaBelowAllocated(t *testing.T) {
	cfg, _ := config.Load()
	snapshot := seed.Load(cfg)
	tree := snapshot.BudgetTree
	members := snapshot.Members
	pools := snapshot.MemberQuotaPools
	platformKeys := snapshot.PlatformKeys

	msg := memberbudgetquota.ValidateMemberQuotaUpdate(tree, members, pools, platformKeys, "m-1", 1000)
	if msg == nil {
		t.Fatal("expected validation error when quota below allocated")
	}
}
