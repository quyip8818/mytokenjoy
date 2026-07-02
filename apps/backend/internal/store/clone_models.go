package store

import "github.com/tokenjoy/backend/internal/domain/types"

func cloneModels(items []types.ModelInfo) []types.ModelInfo {
	result := make([]types.ModelInfo, len(items))
	for i, model := range items {
		result[i] = types.ModelInfo{
			ID: model.ID, Provider: model.Provider, Name: model.Name,
			DisplayName: model.DisplayName, InputPrice: model.InputPrice,
			OutputPrice: model.OutputPrice, MaxContext: model.MaxContext, Enabled: model.Enabled,
			Capabilities: append([]string{}, model.Capabilities...),
		}
	}
	return result
}

func cloneRoutingRules(items []types.RoutingRule) []types.RoutingRule {
	result := make([]types.RoutingRule, len(items))
	for i, rule := range items {
		cloned := types.RoutingRule{
			ID: rule.ID, NodeID: rule.NodeID, NodeName: rule.NodeName,
			AllowedModels: append([]string{}, rule.AllowedModels...),
			Inherited:     rule.Inherited,
		}
		if rule.DefaultModel != nil {
			defaultModel := *rule.DefaultModel
			cloned.DefaultModel = &defaultModel
		}
		if rule.FallbackModel != nil {
			fallbackModel := *rule.FallbackModel
			cloned.FallbackModel = &fallbackModel
		}
		result[i] = cloned
	}
	return result
}

func cloneOperationLogs(items []types.OperationLog) []types.OperationLog {
	result := make([]types.OperationLog, len(items))
	copy(result, items)
	return result
}

func cloneUsageLedger(items []types.UsageLedgerEntry) []types.UsageLedgerEntry {
	result := make([]types.UsageLedgerEntry, len(items))
	for i, item := range items {
		result[i] = CloneUsageLedgerEntry(item)
	}
	return result
}

func CloneUsageLedgerEntry(item types.UsageLedgerEntry) types.UsageLedgerEntry {
	cloned := item
	if item.MemberID != nil {
		memberID := *item.MemberID
		cloned.MemberID = &memberID
	}
	if item.BudgetGroupID != nil {
		groupID := *item.BudgetGroupID
		cloned.BudgetGroupID = &groupID
	}
	return cloned
}

func CloneUsageLedger(items []types.UsageLedgerEntry) []types.UsageLedgerEntry {
	return cloneUsageLedger(items)
}

func CloneModels(items []types.ModelInfo) []types.ModelInfo { return cloneModels(items) }

func CloneRoutingRules(items []types.RoutingRule) []types.RoutingRule {
	return cloneRoutingRules(items)
}

func CloneOperationLogs(items []types.OperationLog) []types.OperationLog {
	return cloneOperationLogs(items)
}
