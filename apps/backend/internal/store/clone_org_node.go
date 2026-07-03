package store

import "github.com/tokenjoy/backend/internal/domain/types"

func cloneOrgNode(node types.OrgNode) types.OrgNode {
	cloned := types.OrgNode{
		ID: node.ID, Name: node.Name, MemberCount: node.MemberCount,
		Budget: node.Budget, Consumed: node.Consumed, Period: node.Period,
		RoutingInherited: node.RoutingInherited,
	}
	if node.ParentID != nil {
		parentID := *node.ParentID
		cloned.ParentID = &parentID
	}
	if node.ExternalID != nil {
		externalID := *node.ExternalID
		cloned.ExternalID = &externalID
	}
	if node.Source != nil {
		source := *node.Source
		cloned.Source = &source
	}
	if node.ManagerID != nil {
		managerID := *node.ManagerID
		cloned.ManagerID = &managerID
	}
	if node.ReservedPool != nil {
		reserved := *node.ReservedPool
		cloned.ReservedPool = &reserved
	}
	if node.DefaultModel != nil {
		defaultModel := *node.DefaultModel
		cloned.DefaultModel = &defaultModel
	}
	if node.FallbackModel != nil {
		fallbackModel := *node.FallbackModel
		cloned.FallbackModel = &fallbackModel
	}
	if len(node.Children) > 0 {
		cloned.Children = make([]types.OrgNode, len(node.Children))
		for i, child := range node.Children {
			cloned.Children[i] = cloneOrgNode(child)
		}
	}
	return cloned
}

func cloneOrgNodes(items []types.OrgNode) []types.OrgNode {
	result := make([]types.OrgNode, len(items))
	for i, node := range items {
		result[i] = cloneOrgNode(node)
	}
	return result
}

func cloneModelAllowlist(rows []ModelAllowlistRow) []ModelAllowlistRow {
	result := make([]ModelAllowlistRow, len(rows))
	copy(result, rows)
	return result
}

func CloneOrgNodes(items []types.OrgNode) []types.OrgNode { return cloneOrgNodes(items) }

func CloneModelAllowlist(rows []ModelAllowlistRow) []ModelAllowlistRow {
	return cloneModelAllowlist(rows)
}
