package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGetQuotaRemaining(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys

	remaining := budget.GetQuotaRemaining(members, keys, seed.IDMember1)
	if remaining != 3000 {
		t.Fatalf("expected remaining 3000 (10000 personal - 7000 allocated), got %v", remaining)
	}
}

func TestBuildQuotaSummary(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	members := snapshot.Members
	keys := snapshot.PlatformKeys
	reserved := budget.GetReservedPoolForMember(types.OrgNodesToBudgetTree(snapshot.OrgNodes), members, seed.IDMember1)

	summary := budget.BuildQuotaSummary(members, keys, seed.IDMember1, reserved)
	if summary.TotalQuota != 10000 {
		t.Fatalf("expected total quota 10000, got %v", summary.TotalQuota)
	}
	if summary.Remaining != 3000 {
		t.Fatalf("expected remaining 3000, got %v", summary.Remaining)
	}
	if summary.ReservedPool != 1500 {
		t.Fatalf("expected reserved pool 1500, got %v", summary.ReservedPool)
	}
}
