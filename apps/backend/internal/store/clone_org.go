package store

import "github.com/tokenjoy/backend/internal/domain/types"

func cloneImportFailures(items []types.ImportFailure) []types.ImportFailure {
	result := make([]types.ImportFailure, len(items))
	copy(result, items)
	return result
}

func cloneSyncLogs(items []types.SyncLog) []types.SyncLog {
	result := make([]types.SyncLog, len(items))
	copy(result, items)
	return result
}

func clonePermissions(items []types.Permission) []types.Permission {
	result := make([]types.Permission, len(items))
	copy(result, items)
	return result
}

func cloneRoles(items []types.Role) []types.Role {
	result := make([]types.Role, len(items))
	for i, role := range items {
		result[i] = types.Role{
			ID: role.ID, Name: role.Name, Type: role.Type,
			Permissions: append([]string{}, role.Permissions...),
			MemberCount: role.MemberCount,
		}
	}
	return result
}

func cloneMembers(items []types.Member) []types.Member {
	result := make([]types.Member, len(items))
	for i, member := range items {
		result[i] = types.Member{
			ID: member.ID, CompanyID: member.CompanyID, Name: member.Name, Phone: member.Phone, Email: member.Email,
			DepartmentID: member.DepartmentID, DepartmentName: member.DepartmentName,
			Status: member.Status, Roles: append([]string{}, member.Roles...), Source: member.Source,
		}
		if member.ExternalID != nil {
			externalID := *member.ExternalID
			result[i].ExternalID = &externalID
		}
	}
	return result
}

func cloneDepartments(items []types.Department) []types.Department {
	result := make([]types.Department, len(items))
	for i, dept := range items {
		result[i] = cloneDepartment(dept)
	}
	return result
}

func cloneDepartment(dept types.Department) types.Department {
	cloned := types.Department{ID: dept.ID, Name: dept.Name, MemberCount: dept.MemberCount}
	if dept.ParentID != nil {
		parentID := *dept.ParentID
		cloned.ParentID = &parentID
	}
	if dept.ExternalID != nil {
		externalID := *dept.ExternalID
		cloned.ExternalID = &externalID
	}
	if dept.Source != nil {
		source := *dept.Source
		cloned.Source = &source
	}
	if dept.ManagerID != nil {
		managerID := *dept.ManagerID
		cloned.ManagerID = &managerID
	}
	if len(dept.Children) > 0 {
		cloned.Children = make([]types.Department, len(dept.Children))
		for i, child := range dept.Children {
			cloned.Children[i] = cloneDepartment(child)
		}
	}
	return cloned
}

func CloneImportFailures(items []types.ImportFailure) []types.ImportFailure {
	return cloneImportFailures(items)
}

func CloneSyncLogs(items []types.SyncLog) []types.SyncLog { return cloneSyncLogs(items) }

func ClonePermissions(items []types.Permission) []types.Permission { return clonePermissions(items) }

func CloneRoles(items []types.Role) []types.Role { return cloneRoles(items) }

func CloneMembers(items []types.Member) []types.Member { return cloneMembers(items) }

func CloneDepartments(items []types.Department) []types.Department { return cloneDepartments(items) }
