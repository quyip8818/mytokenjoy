package treeutil

import "github.com/tokenjoy/backend/internal/domain/types"

func FlattenBudgetTree(nodes []types.BudgetNode) []types.BudgetNode {
	result := make([]types.BudgetNode, 0)
	for _, node := range nodes {
		cloned := node
		cloned.Children = nil
		result = append(result, cloned)
		if len(node.Children) > 0 {
			result = append(result, FlattenBudgetTree(node.Children)...)
		}
	}
	return result
}
