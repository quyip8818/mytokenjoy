package store

import "github.com/tokenjoy/backend/internal/domain/types"

func cloneBudgetNode(node types.BudgetNode) types.BudgetNode {
	cloned := types.BudgetNode{
		ID: node.ID, Name: node.Name, Budget: node.Budget,
		Consumed: node.Consumed, Period: node.Period,
	}
	if node.ParentID != nil {
		parentID := *node.ParentID
		cloned.ParentID = &parentID
	}
	if node.ReservedPool != nil {
		reserved := *node.ReservedPool
		cloned.ReservedPool = &reserved
	}
	if len(node.Children) > 0 {
		cloned.Children = make([]types.BudgetNode, len(node.Children))
		for i, child := range node.Children {
			cloned.Children[i] = cloneBudgetNode(child)
		}
	}
	return cloned
}

func cloneBudgetTree(items []types.BudgetNode) []types.BudgetNode {
	result := make([]types.BudgetNode, len(items))
	for i, node := range items {
		result[i] = cloneBudgetNode(node)
	}
	return result
}

func cloneBudgetGroups(items []types.BudgetGroup) []types.BudgetGroup {
	result := make([]types.BudgetGroup, len(items))
	for i, group := range items {
		result[i] = types.BudgetGroup{
			ID: group.ID, Name: group.Name, Budget: group.Budget, Consumed: group.Consumed,
			MemberIDs:     append([]string{}, group.MemberIDs...),
			DepartmentIDs: append([]string{}, group.DepartmentIDs...),
		}
	}
	return result
}

func cloneAlertRules(items []types.AlertRule) []types.AlertRule {
	result := make([]types.AlertRule, len(items))
	for i, rule := range items {
		result[i] = types.AlertRule{
			ID: rule.ID, NodeID: rule.NodeID, NodeName: rule.NodeName,
			Thresholds:    append([]int{}, rule.Thresholds...),
			NotifyRoleIDs: append([]string{}, rule.NotifyRoleIDs...),
			Enabled:       rule.Enabled,
		}
	}
	return result
}

func CloneBudgetTree(items []types.BudgetNode) []types.BudgetNode { return cloneBudgetTree(items) }

func CloneBudgetGroups(items []types.BudgetGroup) []types.BudgetGroup {
	return cloneBudgetGroups(items)
}

func CloneAlertRules(items []types.AlertRule) []types.AlertRule { return cloneAlertRules(items) }
