package org

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func FlattenOrgNodeTree(nodes []types.OrgNode) []types.OrgNode {
	result := make([]types.OrgNode, 0)
	var walk func(items []types.OrgNode)
	walk = func(items []types.OrgNode) {
		for _, node := range items {
			flat := node
			flat.Children = nil
			result = append(result, flat)
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(nodes)
	return result
}

func FindOrgNode(nodes []types.OrgNode, id uuid.UUID) *types.OrgNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if len(nodes[i].Children) > 0 {
			if found := FindOrgNode(nodes[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func InsertOrgNodeChild(nodes []types.OrgNode, parentID uuid.UUID, child types.OrgNode) bool {
	for i := range nodes {
		if nodes[i].ID == parentID {
			nodes[i].Children = append(nodes[i].Children, child)
			return true
		}
		if len(nodes[i].Children) > 0 && InsertOrgNodeChild(nodes[i].Children, parentID, child) {
			return true
		}
	}
	return false
}

func RemoveOrgNodeByID(nodes []types.OrgNode, id uuid.UUID) []types.OrgNode {
	filtered := make([]types.OrgNode, 0, len(nodes))
	for _, node := range nodes {
		if node.ID == id {
			continue
		}
		node.Children = RemoveOrgNodeByID(node.Children, id)
		filtered = append(filtered, node)
	}
	return filtered
}

func UpdateOrgNodeName(nodes []types.OrgNode, id uuid.UUID, name string) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Name = name
			return true
		}
		if len(nodes[i].Children) > 0 && UpdateOrgNodeName(nodes[i].Children, id, name) {
			return true
		}
	}
	return false
}

func GetOrgNodePath(nodes []types.OrgNode, targetID uuid.UUID) *string {
	var walk func(items []types.OrgNode, path []string) *string
	walk = func(items []types.OrgNode, path []string) *string {
		for _, node := range items {
			current := append(path, node.Name)
			if node.ID == targetID {
				joined := joinPath(current)
				return &joined
			}
			if len(node.Children) > 0 {
				if found := walk(node.Children, current); found != nil {
					return found
				}
			}
		}
		return nil
	}
	return walk(nodes, nil)
}

func HasDirectChildOrgNodes(nodes []types.OrgNode, id uuid.UUID) bool {
	node := FindOrgNode(nodes, id)
	if node == nil {
		return false
	}
	return len(node.Children) > 0
}

func RecalcOrgNodeMemberCounts(nodes []types.OrgNode, members []types.Member) []types.OrgNode {
	directCounts := make(map[uuid.UUID]int)
	for _, member := range members {
		if member.Status == types.MemberStatusInactive {
			continue
		}
		directCounts[member.DepartmentID]++
	}

	var walk func(items []types.OrgNode) []types.OrgNode
	walk = func(items []types.OrgNode) []types.OrgNode {
		result := make([]types.OrgNode, len(items))
		for i, node := range items {
			cloned := node
			cloned.Children = walk(node.Children)
			total := directCounts[node.ID]
			for _, child := range cloned.Children {
				total += child.MemberCount
			}
			cloned.MemberCount = total
			result[i] = cloned
		}
		return result
	}
	return walk(nodes)
}

func orgNodeTreeSize(nodes []types.OrgNode) int {
	n := 0
	for _, node := range nodes {
		n++
		n += orgNodeTreeSize(node.Children)
	}
	return n
}

func departmentToOrgNode(dept types.Department) types.OrgNode {
	return types.OrgNode{
		ID:          dept.ID,
		Name:        dept.Name,
		ParentID:    dept.ParentID,
		MemberCount: dept.MemberCount,
		ExternalID:  dept.ExternalID,
		Source:      dept.Source,
		ManagerID:   dept.ManagerID,
		Children:    departmentsToOrgNodes(dept.Children),
	}
}

func departmentsToOrgNodes(departments []types.Department) []types.OrgNode {
	result := make([]types.OrgNode, len(departments))
	for i, dept := range departments {
		result[i] = departmentToOrgNode(dept)
	}
	return result
}
