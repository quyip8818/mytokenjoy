package org

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) ListRoles(ctx context.Context) ([]types.Role, error) {
	return s.store.Org().Roles(ctx)
}

func (s *service) CreateRole(ctx context.Context, name string, permissions []string) (types.Role, error) {
	roles, err := s.store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}
	role := types.Role{
		ID:   fmt.Sprintf("role-%d", time.Now().UnixMilli()),
		Name: name, Type: "custom", Permissions: permissions, MemberCount: 0,
	}
	roles = append(roles, role)
	if err := s.store.Org().SetRoles(ctx, roles); err != nil {
		return types.Role{}, fmt.Errorf("persist roles: %w", err)
	}
	if err := s.bumpAuthzRevision(ctx); err != nil {
		return types.Role{}, err
	}
	return role, nil
}

func (s *service) UpdateRole(ctx context.Context, id, name string, permissions []string) (types.Role, error) {
	roles, err := s.store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}
	for i := range roles {
		if roles[i].ID == id {
			roles[i].Name = name
			roles[i].Permissions = permissions
			if err := s.store.Org().SetRoles(ctx, roles); err != nil {
				return types.Role{}, err
			}
			if err := s.bumpAuthzRevision(ctx); err != nil {
				return types.Role{}, err
			}
			return roles[i], nil
		}
	}
	return types.Role{}, domain.NewDomainError(404, "Not found")
}

func (s *service) DeleteRole(ctx context.Context, id string) error {
	roles, err := s.store.Org().Roles(ctx)
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

	members, err := s.store.Org().Members(ctx)
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
	if err := s.store.Org().SetMembers(ctx, members); err != nil {
		return err
	}

	roles = append(roles[:idx], roles[idx+1:]...)
	if err := s.recalcRoleMemberCounts(ctx, roles); err != nil {
		return err
	}
	if err := s.store.Org().SetRoles(ctx, roles); err != nil {
		return err
	}
	return s.bumpAuthzRevision(ctx)
}

func (s *service) ListRoleMembers(ctx context.Context, roleID string) ([]types.Member, error) {
	roles, err := s.store.Org().Roles(ctx)
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

	members, err := s.store.Org().Members(ctx)
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

func (s *service) AddRoleMember(ctx context.Context, roleID, memberID string) error {
	roles, err := s.store.Org().Roles(ctx)
	if err != nil {
		return err
	}
	members, err := s.store.Org().Members(ctx)
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

	for i := range members {
		if members[i].ID != memberID {
			continue
		}
		if !pkgorg.ContainsRole(members[i].Roles, role.Name) {
			members[i].Roles = append(members[i].Roles, role.Name)
			if err := s.recalcRoleMemberCounts(ctx, roles); err != nil {
				return err
			}
			if err := s.store.Org().SetMembers(ctx, members); err != nil {
				return err
			}
			if err := s.bumpAuthzRevision(ctx); err != nil {
				return err
			}
			return s.store.Org().SetRoles(ctx, roles)
		}
		break
	}
	return nil
}

func (s *service) RemoveRoleMember(ctx context.Context, roleID, memberID string) error {
	roles, err := s.store.Org().Roles(ctx)
	if err != nil {
		return err
	}
	members, err := s.store.Org().Members(ctx)
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
	if err := s.recalcRoleMemberCounts(ctx, roles); err != nil {
		return err
	}
	if err := s.store.Org().SetMembers(ctx, members); err != nil {
		return err
	}
	if err := s.bumpAuthzRevision(ctx); err != nil {
		return err
	}
	return s.store.Org().SetRoles(ctx, roles)
}
