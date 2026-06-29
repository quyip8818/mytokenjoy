package org

import (
	"github.com/tokenjoy/backend/internal/domain/types"
)

func FlattenDepartmentTree(departments []types.Department) []types.Department {
	result := make([]types.Department, 0)
	for _, dept := range departments {
		result = append(result, dept)
		if len(dept.Children) > 0 {
			result = append(result, FlattenDepartmentTree(dept.Children)...)
		}
	}
	return result
}

func GetDeptPath(departments []types.Department, targetID string) *string {
	var walk func(nodes []types.Department, path []string) *string
	walk = func(nodes []types.Department, path []string) *string {
		for _, node := range nodes {
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
	return walk(departments, nil)
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " / " + parts[i]
	}
	return result
}

func FindDepartment(departments []types.Department, id string) *types.Department {
	for i := range departments {
		if departments[i].ID == id {
			return &departments[i]
		}
		if len(departments[i].Children) > 0 {
			if found := FindDepartment(departments[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func InsertDepartmentChild(departments []types.Department, parentID string, dept types.Department) bool {
	for i := range departments {
		if departments[i].ID == parentID {
			departments[i].Children = append(departments[i].Children, dept)
			return true
		}
		if len(departments[i].Children) > 0 && InsertDepartmentChild(departments[i].Children, parentID, dept) {
			return true
		}
	}
	return false
}

func RemoveDepartment(departments []types.Department, id string) ([]types.Department, bool) {
	filtered := make([]types.Department, 0, len(departments))
	removed := false
	for _, dept := range departments {
		if dept.ID == id {
			removed = true
			continue
		}
		cloned := dept
		if len(dept.Children) > 0 {
			var childRemoved bool
			cloned.Children, childRemoved = RemoveDepartment(dept.Children, id)
			removed = removed || childRemoved
		}
		filtered = append(filtered, cloned)
	}
	return filtered, removed
}

func UpdateDepartmentName(departments []types.Department, id, name string) bool {
	for i := range departments {
		if departments[i].ID == id {
			departments[i].Name = name
			return true
		}
		if len(departments[i].Children) > 0 && UpdateDepartmentName(departments[i].Children, id, name) {
			return true
		}
	}
	return false
}

func HasDirectChildDepartments(departments []types.Department, id string) bool {
	dept := FindDepartment(departments, id)
	if dept == nil {
		return false
	}
	return len(dept.Children) > 0
}

func HasDirectActiveMembers(members []types.Member, deptID string) bool {
	for _, member := range members {
		if member.DepartmentID == deptID && member.Status == "active" {
			return true
		}
	}
	return false
}

func RecalcMemberCounts(departments []types.Department, members []types.Member) []types.Department {
	directCounts := make(map[string]int)
	for _, member := range members {
		directCounts[member.DepartmentID]++
	}

	var walk func(nodes []types.Department) []types.Department
	walk = func(nodes []types.Department) []types.Department {
		result := make([]types.Department, len(nodes))
		for i, node := range nodes {
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
	return walk(departments)
}
