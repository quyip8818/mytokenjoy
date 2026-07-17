package snapshot

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

type orgNodeRoutingSeed struct {
	allowedModelIDs []uuid.UUID
	defaultModelID  *uuid.UUID
	fallbackModelID *uuid.UUID
	inherited       bool
}

func orgNodeRoutingByID() map[uuid.UUID]orgNodeRoutingSeed {
	return map[uuid.UUID]orgNodeRoutingSeed{
		contract.IDDept1: {
			allowedModelIDs: []uuid.UUID{
				contract.IDModel1, contract.IDModel2, contract.IDModel4,
				contract.IDModel5, contract.IDModel8, contract.IDModelLocalTest,
			},
			defaultModelID:  modelIDPtr("gpt-4o-mini"),
			fallbackModelID: modelIDPtr("deepseek-v3"),
		},
		contract.IDDept2: {
			allowedModelIDs: []uuid.UUID{
				contract.IDModel1, contract.IDModel2, contract.IDModel4,
				contract.IDModel3, contract.IDModel5,
			},
			defaultModelID:  modelIDPtr("gpt-4o"),
			fallbackModelID: modelIDPtr("deepseek-v3"),
		},
		contract.IDDept3: {
			allowedModelIDs: []uuid.UUID{contract.IDModel1, contract.IDModel4, contract.IDModel5, contract.IDModelLocalTest},
			inherited:       true,
		},
		contract.IDDept6: {
			allowedModelIDs: []uuid.UUID{contract.IDModel2, contract.IDModel5, contract.IDModel8},
			defaultModelID:  modelIDPtr("gpt-4o-mini"),
			fallbackModelID: modelIDPtr("qwen-plus"),
		},
		contract.IDDept4: {
			allowedModelIDs: []uuid.UUID{contract.IDModel2, contract.IDModel4, contract.IDModel5},
			defaultModelID:  modelIDPtr("claude-sonnet-4-6"),
			fallbackModelID: modelIDPtr("gpt-4o-mini"),
			inherited:       true,
		},
		contract.IDDept5: {
			allowedModelIDs: []uuid.UUID{contract.IDModel2, contract.IDModel5},
			defaultModelID:  modelIDPtr("deepseek-v3"),
			fallbackModelID: modelIDPtr("gpt-4o-mini"),
			inherited:       true,
		},
		contract.IDDept7: {
			allowedModelIDs: []uuid.UUID{contract.IDModel2, contract.IDModel8, contract.IDModel5},
			defaultModelID:  modelIDPtr("qwen-plus"),
			fallbackModelID: modelIDPtr("gpt-4o-mini"),
		},
		contract.IDDept8: {
			allowedModelIDs: []uuid.UUID{contract.IDModel2},
			defaultModelID:  modelIDPtr("gpt-4o-mini"),
			inherited:       true,
		},
	}
}

func buildOrgNodes() []types.OrgNode {
	return assembleOrgNodes(buildDepartments())
}

func assembleOrgNodes(depts []types.Department) []types.OrgNode {
	budgetTree := buildBudgetTree()
	routing := orgNodeRoutingByID()
	ruleByNode := make(map[uuid.UUID]types.RoutingRule, len(routing))
	for nodeID, cfg := range routing {
		ruleByNode[nodeID] = types.RoutingRule{
			ID:              nodeID,
			NodeID:          nodeID,
			AllowedModelIDs: append([]uuid.UUID{}, cfg.allowedModelIDs...),
			DefaultModelID:  cfg.defaultModelID,
			FallbackModelID: cfg.fallbackModelID,
			Inherited:       cfg.inherited,
		}
	}
	return mergeOrgNodeTree(depts, budgetTree, ruleByNode)
}

func mergeOrgNodeTree(
	depts []types.Department,
	budgetTree []types.BudgetNode,
	ruleByNode map[uuid.UUID]types.RoutingRule,
) []types.OrgNode {
	budgetByID := flattenBudgetByID(budgetTree)
	return mergeOrgNodeChildren(depts, budgetByID, ruleByNode)
}

func mergeOrgNodeChildren(
	depts []types.Department,
	budgetByID map[uuid.UUID]types.BudgetNode,
	ruleByNode map[uuid.UUID]types.RoutingRule,
) []types.OrgNode {
	nodes := make([]types.OrgNode, len(depts))
	for i, dept := range depts {
		budget := budgetByID[dept.ID]
		rule := ruleByNode[dept.ID]
		node := types.OrgNode{
			ID: dept.ID, Name: dept.Name, ParentID: dept.ParentID,
			ExternalID: dept.ExternalID, Source: dept.Source, ManagerID: dept.ManagerID,
			Budget: budget.Budget, ReservedPool: budget.ReservedPool, Period: budget.Period,
			DefaultModelID: rule.DefaultModelID, FallbackModelID: rule.FallbackModelID,
			RoutingInherited: rule.Inherited,
		}
		if len(dept.Children) > 0 {
			node.Children = mergeOrgNodeChildren(dept.Children, budgetByID, ruleByNode)
		}
		nodes[i] = node
	}
	return nodes
}

func flattenBudgetByID(tree []types.BudgetNode) map[uuid.UUID]types.BudgetNode {
	result := make(map[uuid.UUID]types.BudgetNode)
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
		for _, modelID := range cfg.allowedModelIDs {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerOrgNode,
				OwnerID:   nodeID,
				ModelID:   modelID,
			})
		}
	}
	for _, key := range loadPlatformKeys() {
		for _, modelID := range key.ModelWhitelist {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerPlatformKey,
				OwnerID:   key.ID,
				ModelID:   modelID,
			})
		}
	}
	for _, approval := range buildApprovals() {
		for _, modelID := range approval.RequestedModels {
			rows = append(rows, store.ModelAllowlistRow{
				OwnerType: types.AllowlistOwnerKeyApproval,
				OwnerID:   approval.ID,
				ModelID:   modelID,
			})
		}
	}
	return rows
}
