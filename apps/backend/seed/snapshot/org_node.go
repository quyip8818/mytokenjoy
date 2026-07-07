package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

type orgNodeRoutingSeed struct {
	allowedModels []string
	defaultModel  *string
	fallbackModel *string
	inherited     bool
}

func orgNodeRoutingByID() map[string]orgNodeRoutingSeed {
	gpt4oMini := "gpt-4o-mini"
	deepseek := "deepseek-v3"
	claudeSonnet := "claude-sonnet-4-6"
	gpt4o := "gpt-4o"
	qwenPlus := "qwen-plus"
	return map[string]orgNodeRoutingSeed{
		"dept-1": {
			allowedModels: []string{"gpt-4o", "gpt-4o-mini", "claude-sonnet-4-6", "deepseek-v3", "qwen-plus"},
			defaultModel:  &gpt4oMini, fallbackModel: &deepseek,
		},
		"dept-2": {
			allowedModels: []string{"gpt-4o", "gpt-4o-mini", "claude-sonnet-4-6", "claude-opus-4-8", "deepseek-v3"},
			defaultModel:  &gpt4o, fallbackModel: &deepseek,
		},
		contract.IDDept3: {
			allowedModels: []string{"gpt-4o", "claude-sonnet-4-6", "deepseek-v3"},
			inherited:     true,
		},
		"dept-6": {
			allowedModels: []string{"gpt-4o-mini", "deepseek-v3", "qwen-plus"},
			defaultModel:  &gpt4oMini, fallbackModel: &qwenPlus,
		},
		"dept-4": {
			allowedModels: []string{"gpt-4o-mini", "claude-sonnet-4-6", "deepseek-v3"},
			defaultModel:  &claudeSonnet, fallbackModel: &gpt4oMini, inherited: true,
		},
		"dept-5": {
			allowedModels: []string{"gpt-4o-mini", "deepseek-v3"},
			defaultModel:  &deepseek, fallbackModel: &gpt4oMini, inherited: true,
		},
		"dept-7": {
			allowedModels: []string{"gpt-4o-mini", "qwen-plus", "deepseek-v3"},
			defaultModel:  &qwenPlus, fallbackModel: &gpt4oMini,
		},
		"dept-8": {
			allowedModels: []string{"gpt-4o-mini"},
			defaultModel:  &gpt4oMini, inherited: true,
		},
	}
}

func buildOrgNodes() []types.OrgNode {
	depts := buildDepartments()
	budgetTree := buildBudgetTree()
	routing := orgNodeRoutingByID()
	ruleByNode := make(map[string]types.RoutingRule, len(routing))
	for nodeID, cfg := range routing {
		ruleByNode[nodeID] = types.RoutingRule{
			ID: nodeID, NodeID: nodeID,
			AllowedModels: cfg.allowedModels,
			DefaultModel:  cfg.defaultModel,
			FallbackModel: cfg.fallbackModel,
			Inherited:     cfg.inherited,
		}
	}
	return mergeOrgNodeTree(depts, budgetTree, ruleByNode)
}

func mergeOrgNodeTree(
	depts []types.Department,
	budgetTree []types.BudgetNode,
	ruleByNode map[string]types.RoutingRule,
) []types.OrgNode {
	budgetByID := flattenBudgetByID(budgetTree)
	return mergeOrgNodeChildren(depts, budgetByID, ruleByNode)
}

func mergeOrgNodeChildren(
	depts []types.Department,
	budgetByID map[string]types.BudgetNode,
	ruleByNode map[string]types.RoutingRule,
) []types.OrgNode {
	nodes := make([]types.OrgNode, len(depts))
	for i, dept := range depts {
		budget := budgetByID[dept.ID]
		rule := ruleByNode[dept.ID]
		node := types.OrgNode{
			ID: dept.ID, Name: dept.Name, ParentID: dept.ParentID, MemberCount: dept.MemberCount,
			ExternalID: dept.ExternalID, Source: dept.Source, ManagerID: dept.ManagerID,
			Budget: budget.Budget, Consumed: budget.Consumed, ReservedPool: budget.ReservedPool, Period: budget.Period,
			DefaultModel: rule.DefaultModel, FallbackModel: rule.FallbackModel, RoutingInherited: rule.Inherited,
		}
		if len(dept.Children) > 0 {
			node.Children = mergeOrgNodeChildren(dept.Children, budgetByID, ruleByNode)
		}
		nodes[i] = node
	}
	return nodes
}

func flattenBudgetByID(tree []types.BudgetNode) map[string]types.BudgetNode {
	result := make(map[string]types.BudgetNode)
	var walk func(nodes []types.BudgetNode)
	walk = func(nodes []types.BudgetNode) {
		for _, node := range nodes {
			flat := node
			flat.Children = nil
			result[node.ID] = flat
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(tree)
	return result
}

func buildModelAllowlist() []store.ModelAllowlistRow {
	rows := make([]store.ModelAllowlistRow, 0)
	for nodeID, cfg := range orgNodeRoutingByID() {
		for _, modelName := range cfg.allowedModels {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerOrgNode,
				OwnerID:   nodeID,
				ModelName: modelName,
			})
		}
	}
	for _, key := range loadPlatformKeys() {
		for _, modelName := range key.ModelWhitelist {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerPlatformKey,
				OwnerID:   key.ID,
				ModelName: modelName,
			})
		}
	}
	for _, approval := range buildApprovals() {
		for _, modelName := range approval.RequestedModels {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerKeyApproval,
				OwnerID:   approval.ID,
				ModelName: modelName,
			})
		}
	}
	return rows
}
