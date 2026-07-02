package budget

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/tree"
)

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

func InsertBudgetNode(nodes []types.BudgetNode, parentID string, node types.BudgetNode) bool {
	for i := range nodes {
		if nodes[i].ID == parentID {
			nodes[i].Children = append(nodes[i].Children, node)
			return true
		}
		if len(nodes[i].Children) > 0 && InsertBudgetNode(nodes[i].Children, parentID, node) {
			return true
		}
	}
	return false
}

func RemoveBudgetNode(nodes []types.BudgetNode, id string) ([]types.BudgetNode, bool) {
	filtered := make([]types.BudgetNode, 0, len(nodes))
	removed := false
	for _, node := range nodes {
		if node.ID == id {
			removed = true
			continue
		}
		cloned := node
		if len(node.Children) > 0 {
			var childRemoved bool
			cloned.Children, childRemoved = RemoveBudgetNode(node.Children, id)
			removed = removed || childRemoved
		}
		filtered = append(filtered, cloned)
	}
	return filtered, removed
}

func UpdateBudgetNodeName(nodes []types.BudgetNode, id, name string) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Name = name
			return true
		}
		if len(nodes[i].Children) > 0 && UpdateBudgetNodeName(nodes[i].Children, id, name) {
			return true
		}
	}
	return false
}

func FlattenBudgetTree(nodes []types.BudgetNode) []types.BudgetNode {
	return tree.Flatten(nodes, func(node types.BudgetNode) []types.BudgetNode {
		return node.Children
	}, func(node *types.BudgetNode) {
		node.Children = nil
	})
}
