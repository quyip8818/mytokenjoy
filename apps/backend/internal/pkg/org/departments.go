package org

import "github.com/tokenjoy/backend/internal/domain/types"

func FlattenDepartmentTree(departments []types.Department) []types.Department {
	return types.OrgNodesToDepartments(FlattenOrgNodeTree(departmentsToOrgNodes(departments)))
}

func GetDeptPath(departments []types.Department, targetID string) *string {
	return GetOrgNodePath(departmentsToOrgNodes(departments), targetID)
}

func FindDepartment(departments []types.Department, id string) *types.Department {
	node := FindOrgNode(departmentsToOrgNodes(departments), id)
	if node == nil {
		return nil
	}
	dept := types.OrgNodeToDepartment(*node)
	return &dept
}

func InsertDepartmentChild(departments []types.Department, parentID string, dept types.Department) bool {
	nodes := departmentsToOrgNodes(departments)
	if !InsertOrgNodeChild(nodes, parentID, departmentToOrgNode(dept)) {
		return false
	}
	syncDepartmentChildren(departments, nodes)
	return true
}

func RemoveDepartment(departments []types.Department, id string) ([]types.Department, bool) {
	nodes := departmentsToOrgNodes(departments)
	before := orgNodeTreeSize(nodes)
	updated := RemoveOrgNodeByID(nodes, id)
	return types.OrgNodesToDepartments(updated), orgNodeTreeSize(updated) != before
}

func UpdateDepartmentName(departments []types.Department, id, name string) bool {
	nodes := departmentsToOrgNodes(departments)
	if !UpdateOrgNodeName(nodes, id, name) {
		return false
	}
	syncDepartmentNames(departments, nodes)
	return true
}

func HasDirectChildDepartments(departments []types.Department, id string) bool {
	return HasDirectChildOrgNodes(departmentsToOrgNodes(departments), id)
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
	return types.OrgNodesToDepartments(RecalcOrgNodeMemberCounts(departmentsToOrgNodes(departments), members))
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

func syncDepartmentChildren(departments []types.Department, nodes []types.OrgNode) {
	for i := range departments {
		if i >= len(nodes) || departments[i].ID != nodes[i].ID {
			continue
		}
		departments[i].Children = types.OrgNodesToDepartments(nodes[i].Children)
		syncDepartmentChildren(departments[i].Children, nodes[i].Children)
	}
}

func syncDepartmentNames(departments []types.Department, nodes []types.OrgNode) {
	for i := range departments {
		if i >= len(nodes) || departments[i].ID != nodes[i].ID {
			continue
		}
		departments[i].Name = nodes[i].Name
		syncDepartmentNames(departments[i].Children, nodes[i].Children)
	}
}
