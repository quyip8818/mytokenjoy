package budgetutil

import "github.com/tokenjoy/backend/internal/domain/types"

func SumChildrenBudget(node types.BudgetNode) float64 {
	sum := 0.0
	for _, child := range node.Children {
		sum += child.Budget
	}
	return sum
}

func FindBudgetNode(nodes []types.BudgetNode, id string) *types.BudgetNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if len(nodes[i].Children) > 0 {
			if found := FindBudgetNode(nodes[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func UpdateBudgetNodeInTree(nodes []types.BudgetNode, id string, data types.BudgetNode) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Budget = data.Budget
			if data.ReservedPool != nil {
				nodes[i].ReservedPool = data.ReservedPool
			}
			return true
		}
		if len(nodes[i].Children) > 0 && UpdateBudgetNodeInTree(nodes[i].Children, id, data) {
			return true
		}
	}
	return false
}
