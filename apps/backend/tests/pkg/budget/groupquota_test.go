package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestValidateGroupKeyQuota(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	groups := snapshot.BudgetGroups
	keys := snapshot.PlatformKeys

	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup1 {
			if msg := budget.ValidateGroupKeyQuota(group, keys, 99999, ""); msg == nil {
				t.Fatal("expected validation error when quota exceeds group remaining")
			}
			return
		}
	}
	t.Fatal("bg-1 not found in seed")
}

func TestGetGroupQuotaRemaining(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	groups := snapshot.BudgetGroups
	keys := snapshot.PlatformKeys

	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup1 {
			remaining := budget.GetGroupQuotaRemaining(group, keys)
			if remaining != 3500 {
				t.Fatalf("expected bg-1 remaining 3500 (30000-18500-8000), got %v", remaining)
			}
			return
		}
	}
	t.Fatal("bg-1 not found in seed")
}
