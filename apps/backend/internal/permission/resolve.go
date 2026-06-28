package permission

import "github.com/tokenjoy/backend/internal/domain/types"

var budgetWriteCapabilities = []string{
	BudgetAllocate,
	BudgetApprove,
	BudgetPolicy,
}

var presetRoleCapabilities = map[string][]string{
	RoleSuperAdmin: append([]string{}, AllPermissions...),
	RoleOrgAdmin:   append([]string{}, AllPermissions...),
	RoleMember:     {SelfKeys, SelfApproval},
	RoleAuditor: {
		OrgRead, BudgetRead, KeysRead, ModelRead,
		AuditRead, DashboardCost, DashboardUsage, SelfApproval,
	},
	RoleAPICaller:  {APICall},
}

var writeCapabilities = []string{
	OrgDatasource,
	OrgStructure,
	OrgRoles,
	OrgMembers,
	BudgetAllocate,
	BudgetApprove,
	BudgetPolicy,
	ModelManage,
	ModelWhitelist,
	KeysAdmin,
	KeysProvider,
}

func expandRoleDefinition(role types.Role) []string {
	if role.Type == "preset" {
		if caps, ok := presetRoleCapabilities[role.Name]; ok {
			return append([]string{}, caps...)
		}
		return nil
	}

	caps := make(map[string]struct{})
	for _, permID := range role.Permissions {
		if permID == "*" {
			for _, p := range AllPermissions {
				caps[p] = struct{}{}
			}
			continue
		}
		if mapped, ok := PermissionIDMap[permID]; ok {
			caps[mapped] = struct{}{}
		} else if contains(AllPermissions, permID) {
			caps[permID] = struct{}{}
		}
	}

	expanded := make([]string, 0, len(caps))
	for p := range caps {
		expanded = append(expanded, p)
	}
	for _, p := range expanded {
		if contains(budgetWriteCapabilities, p) {
			caps[BudgetRead] = struct{}{}
			break
		}
	}

	result := make([]string, 0, len(caps))
	for p := range caps {
		result = append(result, p)
	}
	return result
}

func ResolveMemberPermissions(member types.Member, roles []types.Role) []string {
	caps := make(map[string]struct{})
	for _, roleName := range member.Roles {
		for _, role := range roles {
			if role.Name != roleName {
				continue
			}
			for _, p := range expandRoleDefinition(role) {
				caps[p] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(caps))
	for p := range caps {
		result = append(result, p)
	}
	return result
}

func IsReadOnlySession(permissions []string) bool {
	for _, p := range permissions {
		if p == "*" {
			return false
		}
	}
	for _, writeCap := range writeCapabilities {
		if contains(permissions, writeCap) {
			return false
		}
	}
	return true
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
