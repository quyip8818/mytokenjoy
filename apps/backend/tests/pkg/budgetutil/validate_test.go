package budgetutil_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
)

func TestValidateBudgetNodeUpdate(t *testing.T) {
	tree := []types.BudgetNode{
		{
			ID: "root", Budget: 100000, ReservedPool: floatPtr(10000),
			Children: []types.BudgetNode{
				{ID: "dept-a", Budget: 40000, ReservedPool: floatPtr(5000)},
				{ID: "dept-b", Budget: 40000},
			},
		},
	}

	if msg := budgetutil.ValidateBudgetNodeUpdate(tree, "dept-a", 45000, 5000); msg != nil {
		t.Fatalf("expected valid increase, got %s", *msg)
	}
	if msg := budgetutil.ValidateBudgetNodeUpdate(tree, "dept-a", 90000, 5000); msg == nil {
		t.Fatal("expected oversell against parent")
	}
}

func floatPtr(v float64) *float64 { return &v }
