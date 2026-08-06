package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/seed/contract"
)

func TestValidateProjectKeyBudget(t *testing.T) {
	t.Parallel()
	snapshot := cachedSnapshot()
	projects := snapshot.Projects
	keys := snapshot.PlatformKeys

	for _, project := range projects {
		if project.ID == contract.IDProject1 {
			if msg := budget.ValidateProjectKeyBudget(project, keys, 99999, uuid.Nil); msg == nil {
				t.Fatal("expected validation error when budget exceeds project remaining")
			}
			return
		}
	}
	t.Fatal("proj-1 not found in seed")
}

func TestGetProjectBudgetRemaining(t *testing.T) {
	t.Parallel()
	snapshot := cachedSnapshot()
	projects := snapshot.Projects
	keys := snapshot.PlatformKeys

	for _, project := range projects {
		if project.ID == contract.IDProject1 {
			allocated := budget.GetAllocatedProjectKeyBudget(keys, project.ID)
			want := project.Budget - project.Consumed - allocated
			if want < 0 {
				want = 0
			}
			remaining := budget.GetProjectBudgetRemaining(project, keys)
			if remaining != want {
				t.Fatalf("GetProjectBudgetRemaining() = %v, want %v", remaining, want)
			}
			return
		}
	}
	t.Fatal("proj-1 not found in seed")
}
