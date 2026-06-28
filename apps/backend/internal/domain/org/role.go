package org

import (
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg/roleutil"
)

func (s *service) ListRoles() []Role {
	return s.store.Org().Roles()
}

func (s *service) CreateRole(name string, permissions []string) Role {
	roles := s.store.Org().Roles()
	role := Role{
		ID:   fmt.Sprintf("role-%d", time.Now().UnixMilli()),
		Name: name, Type: "custom", Permissions: permissions, MemberCount: 0,
	}
	roles = append(roles, role)
	_ = s.store.Org().SetRoles(roles)
	return role
}

func (s *service) UpdateRole(id, name string, permissions []string) (Role, error) {
	roles := s.store.Org().Roles()
	for i := range roles {
		if roles[i].ID == id {
			roles[i].Name = name
			roles[i].Permissions = permissions
			if err := s.store.Org().SetRoles(roles); err != nil {
				return Role{}, err
			}
			return roles[i], nil
		}
	}
	return Role{}, domain.NewDomainError(404, "Not found")
}

func (s *service) DeleteRole(id string) error {
	roles := s.store.Org().Roles()
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

	members := s.store.Org().Members()
	for i := range members {
		filtered := make([]string, 0, len(members[i].Roles))
		for _, roleName := range members[i].Roles {
			if roleName != role.Name {
				filtered = append(filtered, roleName)
			}
		}
		members[i].Roles = filtered
	}
	if err := s.store.Org().SetMembers(members); err != nil {
		return err
	}

	roles = append(roles[:idx], roles[idx+1:]...)
	s.recalcRoleMemberCounts(roles)
	return s.store.Org().SetRoles(roles)
}

func (s *service) ListRoleMembers(roleID string) []Member {
	roles := s.store.Org().Roles()
	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return []Member{}
	}

	members := s.store.Org().Members()
	result := make([]Member, 0)
	for _, member := range members {
		for _, roleName := range member.Roles {
			if roleName == role.Name {
				result = append(result, member)
				break
			}
		}
	}
	return result
}

func (s *service) AddRoleMember(roleID, memberID string) error {
	roles := s.store.Org().Roles()
	members := s.store.Org().Members()

	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return nil
	}

	for i := range members {
		if members[i].ID != memberID {
			continue
		}
		if !roleutil.ContainsRole(members[i].Roles, role.Name) {
			members[i].Roles = append(members[i].Roles, role.Name)
			s.recalcRoleMemberCounts(roles)
			if err := s.store.Org().SetMembers(members); err != nil {
				return err
			}
			return s.store.Org().SetRoles(roles)
		}
		break
	}
	return nil
}

func (s *service) RemoveRoleMember(roleID, memberID string) error {
	roles := s.store.Org().Roles()
	members := s.store.Org().Members()

	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	var member *Member
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
	s.recalcRoleMemberCounts(roles)
	if err := s.store.Org().SetMembers(members); err != nil {
		return err
	}
	return s.store.Org().SetRoles(roles)
}
