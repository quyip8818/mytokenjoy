package authz

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func expandRoleDefinition(role types.Role) []string {
	if role.Type == "preset" {
		if caps, ok := permission.PresetRoleCapabilities()[role.Name]; ok {
			return append([]string{}, caps...)
		}
		return nil
	}

	caps := make(map[string]struct{})
	for _, permID := range role.Permissions {
		if permID == "*" {
			for _, p := range permission.CompanyPermissions {
				caps[p] = struct{}{}
			}
			continue
		}
		if mapped, ok := permission.PermissionIDMap[permID]; ok {
			caps[mapped] = struct{}{}
		} else if contains(permission.CompanyPermissions, permID) {
			caps[permID] = struct{}{}
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
	// Merge direct permissions (e.g. platform:admin set during bootstrap).
	for _, p := range member.DirectPermissions {
		caps[p] = struct{}{}
	}
	raw := make([]string, 0, len(caps))
	for p := range caps {
		raw = append(raw, p)
	}
	// Apply hierarchy: admin implies manage+read, manage implies read, etc.
	return permission.ExpandHierarchy(raw)
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
	return common.HasAny(have, required...)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
