package store

import "github.com/tokenjoy/backend/internal/domain/types"

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

func CloneBudgetGroups(items []types.BudgetGroup) []types.BudgetGroup {
	return cloneBudgetGroups(items)
}

func CloneAlertRules(items []types.AlertRule) []types.AlertRule { return cloneAlertRules(items) }
