package structure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) ListRoles(ctx context.Context) ([]types.Role, error) {
	return s.d.Store.Org().Roles(ctx)
}

func (s *LocalService) CreateRole(ctx context.Context, name string, permissions []string) (types.Role, error) {
	// Validate role name
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return types.Role{}, domain.Validation("role name must not be empty")
	}
	if grants.IsPresetRole(trimmedName) {
		return types.Role{}, domain.NewDomainError(400, "role name already exists")
	}

	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}
	for _, existing := range roles {
		if existing.Name == trimmedName {
			return types.Role{}, domain.NewDomainError(400, "role name already exists")
		}
	}

	grantIDs, err := s.d.Grants.NormalizeGrantIDs(permissions)
	if err != nil {
		return types.Role{}, domain.NewDomainError(400, err.Error())
	}
	role := types.Role{
		ID:   uuid.Must(uuid.NewV7()),
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

func (s *LocalService) UpdateRole(ctx context.Context, id uuid.UUID, name string, permissions []string) (types.Role, error) {
	roles, err := s.d.Store.Org().Roles(ctx)
	if err != nil {
		return types.Role{}, err
	}
	for i := range roles {
		if roles[i].ID == id {
			if roles[i].Type == "preset" {
				return types.Role{}, domain.NewDomainError(400, "Cannot modify preset role")
			}
			grantIDs, err := s.d.Grants.NormalizeGrantIDs(permissions)
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

func (s *LocalService) DeleteRole(ctx context.Context, id uuid.UUID) error {
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

func (s *LocalService) ListPermissions(ctx context.Context) ([]types.Permission, error) {
	return s.d.Store.Org().Permissions(ctx)
}
