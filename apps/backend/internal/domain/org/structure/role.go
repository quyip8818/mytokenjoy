package structure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *Local) ListRoles(ctx context.Context) ([]types.Role, error) {
	return s.d.Store.Org().Roles(ctx)
}

func (s *Local) CreateRole(ctx context.Context, name string, permissions []string) (types.Role, error) {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}

	// Validate role name
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return types.Role{}, domain.Validation("role name must not be empty")
	}
	for _, existing := range roles {
		if existing.Name == trimmedName {
			return types.Role{}, domain.NewDomainError(400, "role name already exists")
		}
	}

	grantIDs, err := permission.NormalizeGrantIDs(permissions)
	if err != nil {
		return types.Role{}, domain.NewDomainError(400, err.Error())
	}
	role := types.Role{
		ID:   fmt.Sprintf("role-%d", time.Now().UnixMilli()),
		Name: trimmedName, Type: "custom", Permissions: grantIDs, MemberCount: 0,
	}
	roles = append(roles, role)
	if err := s.d.Store.Org().SetRoles(ctx, roles); err != nil {
		return types.Role{}, fmt.Errorf("persist roles: %w", err)
	}
	if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
		return types.Role{}, err
	}
	return role, nil
}

func (s *Local) UpdateRole(ctx context.Context, id, name string, permissions []string) (types.Role, error) {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}
	for i := range roles {
		if roles[i].ID == id {
			if roles[i].Type == "preset" {
				return types.Role{}, domain.NewDomainError(400, "Cannot modify preset role")
			}
			grantIDs, err := permission.NormalizeGrantIDs(permissions)
			if err != nil {
				return types.Role{}, domain.NewDomainError(400, err.Error())
			}
			roles[i].Name = name
			roles[i].Permissions = grantIDs
			if err := s.d.Store.Org().SetRoles(ctx, roles); err != nil {
				return types.Role{}, err
			}
			if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
				return types.Role{}, err
			}
			return roles[i], nil
		}
	}
	return types.Role{}, domain.NewDomainError(404, "Not found")
}

func (s *Local) DeleteRole(ctx context.Context, id string) error {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return err
	}
	idx := -1
	for i := range roles {
		if roles[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.NewDomainError(404, "Not found")
	}
	role := roles[idx]
	if role.Type == "preset" {
		return domain.NewDomainError(400, "Cannot delete preset role")
	}

	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		for i := range members {
			filtered := make([]string, 0, len(members[i].Roles))
			for _, roleName := range members[i].Roles {
				if roleName != role.Name {
					filtered = append(filtered, roleName)
				}
			}
			members[i].Roles = filtered
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}

		roles = append(roles[:idx], roles[idx+1:]...)
		if err := st.Org().SetRoles(ctx, roles); err != nil {
			return err
		}
		return core.BumpAuthzRevision(ctx, s.d)
	})
}

func (s *Local) ListRoleMembers(ctx context.Context, roleID string) ([]types.Member, error) {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return nil, err
	}
	var role *types.Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return []types.Member{}, nil
	}

	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]types.Member, 0)
	for _, member := range members {
		for _, roleName := range member.Roles {
			if roleName == role.Name {
				result = append(result, member)
				break
			}
		}
	}
	return result, nil
}

func (s *Local) AddRoleMember(ctx context.Context, roleID, memberID string) error {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return err
	}
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return err
	}

	var role *types.Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return domain.NewDomainError(404, "Not found")
	}

	// Prevent adding members to protected preset roles via this endpoint
	if role.Type == "preset" {
		if _, protected := protectedRoles[role.Name]; protected {
			return domain.Forbidden("cannot assign protected role directly")
		}
	}

	found := false
	for i := range members {
		if members[i].ID != memberID {
			continue
		}
		found = true
		if !pkgorg.ContainsRole(members[i].Roles, role.Name) {
			members[i].Roles = append(members[i].Roles, role.Name)
			if err := s.d.Store.Org().SetMembers(ctx, members); err != nil {
				return err
			}
			return core.BumpAuthzRevision(ctx, s.d)
		}
		break
	}
	if !found {
		return domain.NewDomainError(404, "Member not found")
	}
	return nil
}

func (s *Local) RemoveRoleMember(ctx context.Context, roleID, memberID string) error {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return err
	}
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return err
	}

	var role *types.Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	var member *types.Member
	for i := range members {
		if members[i].ID == memberID {
			member = &members[i]
			break
		}
	}
	if role == nil || member == nil {
		return domain.NewDomainError(404, "Not found")
	}
	if role.Name == permission.RoleMember {
		return domain.NewDomainError(400, "Cannot remove base member role")
	}

	// Prevent removing the last super admin
	if role.Name == permission.RoleSuperAdmin {
		adminCount := 0
		for _, m := range members {
			if pkgorg.ContainsRole(m.Roles, permission.RoleSuperAdmin) {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return domain.NewDomainError(400, "Cannot remove the last super admin")
		}
	}

	filtered := make([]string, 0, len(member.Roles))
	for _, roleName := range member.Roles {
		if roleName != role.Name {
			filtered = append(filtered, roleName)
		}
	}
	for i := range members {
		if members[i].ID == memberID {
			members[i].Roles = filtered
			break
		}
	}
	if err := s.d.Store.Org().SetMembers(ctx, members); err != nil {
		return err
	}
	return core.BumpAuthzRevision(ctx, s.d)
}

func (s *Local) ListPermissions(ctx context.Context) ([]types.Permission, error) {
	return s.d.Store.Org().Permissions(ctx)
}
