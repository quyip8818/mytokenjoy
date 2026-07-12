package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestValidateGroupKeyBudget(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	groups := snapshot.BudgetGroups
	keys := snapshot.PlatformKeys

	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup1 {
			if msg := budget.ValidateGroupKeyBudget(group, keys, budgetfix.DisplayPoints(99999), ""); msg == nil {
				t.Fatal("expected validation error when quota exceeds group remaining")
			}
			return
		}
	}
	t.Fatal("bg-1 not found in seed")
}

func TestGetGroupBudgetRemaining(t *testing.T) {
	t.Parallel()
	snapshot := seed.Load(testutil.TestConfig())
	groups := snapshot.BudgetGroups
	keys := snapshot.PlatformKeys

	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup1 {
			allocated := budget.GetAllocatedGroupKeyBudget(keys, group.ID)
			want := group.Budget - group.Consumed - allocated
			if want < 0 {
				want = 0
			}
			remaining := budget.GetGroupBudgetRemaining(group, keys)
			if remaining != want {
				t.Fatalf("expected bg-1 remaining %v, got %v", want, remaining)
			}
			return
		}
	}
	t.Fatal("bg-1 not found in seed")
}
