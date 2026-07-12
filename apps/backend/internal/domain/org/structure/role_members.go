package structure

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *LocalService) ListRoleMembers(ctx context.Context, roleID string) ([]types.Member, error) {
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

func (s *LocalService) AddRoleMember(ctx context.Context, roleID, memberID string) error {
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

func (s *LocalService) RemoveRoleMember(ctx context.Context, roleID, memberID string) error {
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
	if role.Name == grants.RoleMember {
		return domain.NewDomainError(400, "Cannot remove base member role")
	}

	// Prevent removing the last super admin
	if role.Name == grants.RoleSuperAdmin {
		adminCount := 0
		for _, m := range members {
			if pkgorg.ContainsRole(m.Roles, grants.RoleSuperAdmin) {
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
