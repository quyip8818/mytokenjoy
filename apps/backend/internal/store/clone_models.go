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

func cloneCallLogs(items []types.CallLog) []types.CallLog {
	result := make([]types.CallLog, len(items))
	copy(result, items)
	return result
}

func CloneModels(items []types.ModelInfo) []types.ModelInfo { return cloneModels(items) }

func CloneRoutingRules(items []types.RoutingRule) []types.RoutingRule {
	return cloneRoutingRules(items)
}

func CloneOperationLogs(items []types.OperationLog) []types.OperationLog {
	return cloneOperationLogs(items)
}

func CloneCallLogs(items []types.CallLog) []types.CallLog { return cloneCallLogs(items) }
