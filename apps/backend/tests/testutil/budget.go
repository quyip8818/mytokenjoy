package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/seed/contract"
)

func DeptConsumed(t *testing.T, tree []types.BudgetNode, deptID string) float64 {
	t.Helper()
	node := budget.FindBudgetNode(tree, deptID)
	if node == nil {
		t.Fatalf("department %s not found in budget tree", deptID)
	}
	return node.Consumed
}

func Dept3Consumed(t *testing.T, tree []types.BudgetNode) float64 {
	t.Helper()
	return DeptConsumed(t, tree, contract.IDDept3)
}

func SetDeptConsumed(t *testing.T, tree []types.BudgetNode, deptID string, consumed float64) {
	t.Helper()
	if !inflateConsumed(tree, deptID, consumed) {
		t.Fatalf("department %s not found in budget tree", deptID)
	}
}

func inflateConsumed(nodes []types.BudgetNode, id string, value float64) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Consumed = value
			return true
		}
		if len(nodes[i].Children) > 0 && inflateConsumed(nodes[i].Children, id, value) {
			return true
		}
	}
	return false
}
