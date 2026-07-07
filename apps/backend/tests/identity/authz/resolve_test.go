package authz_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestResolveMemberPermissionsSuperAdmin(t *testing.T) {
	t.Parallel()
	member := types.Member{
		ID: "m-admin", Roles: []string{permission.RoleSuperAdmin},
	}
	roles := []types.Role{
		{ID: "role-1", Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	if len(perms) != len(permission.AllPermissions) {
		t.Fatalf("expected %d permissions, got %d", len(permission.AllPermissions), len(perms))
	}
}

func TestIsReadOnlySessionMember(t *testing.T) {
	t.Parallel()
	perms := authz.ResolveMemberPermissions(
		types.Member{Roles: []string{permission.RoleMember}},
		[]types.Role{{Name: permission.RoleMember, Type: "preset"}},
	)
	if !authz.IsReadOnlySession(perms) {
		t.Fatal("expected member session to be read-only")
	}
}

func TestIsReadOnlySessionOrgAdmin(t *testing.T) {
	t.Parallel()
	perms := authz.ResolveMemberPermissions(
		types.Member{Roles: []string{permission.RoleOrgAdmin}},
		[]types.Role{{Name: permission.RoleOrgAdmin, Type: "preset"}},
	)
	if authz.IsReadOnlySession(perms) {
		t.Fatal("expected org admin session to have write access")
	}
}

func TestCustomRoleBudgetApproverIncludesBudgetRead(t *testing.T) {
	t.Parallel()
	perms := authz.ResolveMemberPermissions(
		types.Member{Roles: []string{permission.RoleBudgetApprover}},
		[]types.Role{{Name: permission.RoleBudgetApprover, Type: "custom", Permissions: []string{"p-6"}}},
	)
	foundRead := false
	for _, p := range perms {
		if p == permission.BudgetRead {
			foundRead = true
			break
		}
	}
	if !foundRead {
		t.Fatal("expected budget approver to include budget:read")
	}
}
