package budgetutil

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func FindParentNode(nodes []types.BudgetNode, childID string) *types.BudgetNode {
	var parent *types.BudgetNode
	var walk func([]types.BudgetNode) bool
	walk = func(list []types.BudgetNode) bool {
		for i := range list {
			for _, child := range list[i].Children {
				if child.ID == childID {
					parent = &list[i]
					return true
				}
			}
			if len(list[i].Children) > 0 && walk(list[i].Children) {
				return true
			}
		}
		return false
	}
	walk(nodes)
	return parent
}

func ValidateBudgetNodeUpdate(
	tree []types.BudgetNode,
	nodeID string,
	newBudget float64,
	newReservedPool float64,
) *string {
	node := FindBudgetNode(tree, nodeID)
	if node == nil {
		msg := "Node not found"
		return &msg
	}
	childrenSum := SumChildrenBudget(*node)
	if newBudget < childrenSum+newReservedPool {
		msg := fmt.Sprintf("部门预算不能低于子级预算与预留池之和（¥%.0f）", childrenSum+newReservedPool)
		return &msg
	}
	parent := FindParentNode(tree, nodeID)
	if parent != nil {
		siblingsSum := 0.0
		for _, child := range parent.Children {
			if child.ID != nodeID {
				siblingsSum += child.Budget
			}
		}
		parentReserved := 0.0
		if parent.ReservedPool != nil {
			parentReserved = *parent.ReservedPool
		}
		if siblingsSum+newReservedPool+newBudget > parent.Budget-parentReserved {
			remaining := parent.Budget - parentReserved - siblingsSum
			if remaining < 0 {
				remaining = 0
			}
			msg := fmt.Sprintf("超出上级可分配预算，当前剩余约 ¥%.0f", remaining)
			return &msg
		}
	}
	return nil
}
