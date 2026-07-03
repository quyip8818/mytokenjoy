package authz

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func expandRoleDefinition(role types.Role) []string {
	if role.Type == "preset" {
		if caps, ok := permission.PresetRoleCapabilities()[role.Name]; ok {
			return append([]string{}, caps...)
		}
		return nil
	}

	budgetWriteCapabilities := permission.BudgetWriteCapabilitiesFromManifest()

	caps := make(map[string]struct{})
	for _, permID := range role.Permissions {
		if permID == "*" {
			for _, p := range permission.AllPermissions {
				caps[p] = struct{}{}
			}
			continue
		}
		if mapped, ok := permission.PermissionIDMap[permID]; ok {
			caps[mapped] = struct{}{}
		} else if contains(permission.AllPermissions, permID) {
			caps[permID] = struct{}{}
		}
	}

	expanded := make([]string, 0, len(caps))
	for p := range caps {
		expanded = append(expanded, p)
	}
	for _, p := range expanded {
		if contains(budgetWriteCapabilities, p) {
			caps[permission.BudgetRead] = struct{}{}
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
	for _, writeCap := range permission.WriteCapabilitiesFromManifest() {
		if contains(permissions, writeCap) {
			return false
		}
	}
	return true
}

func HasAny(have []string, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	for _, p := range have {
		if p == "*" {
			return true
		}
	}
	for _, req := range required {
		if contains(have, req) {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
