package authz_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestResolveMemberPermissionsSuperAdmin(t *testing.T) {
	t.Parallel()
	member := types.Member{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000000e01"), Roles: []string{permission.RoleSuperAdmin},
	}
	roles := []types.Role{
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000a101"), Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	// * expands to CompanyPermissions only — platform permissions are excluded.
	if len(perms) != len(permission.CompanyPermissions) {
		t.Fatalf("expected %d permissions (CompanyPermissions), got %d", len(permission.CompanyPermissions), len(perms))
	}
	for _, p := range perms {
		if p == permission.PlatformManage || p == permission.PlatformRead {
			t.Fatalf("super admin should NOT have platform permission %q", p)
		}
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

func TestCustomRoleBudgetManageIncludesBudgetRead(t *testing.T) {
	t.Parallel()
	perms := authz.ResolveMemberPermissions(
		types.Member{Roles: []string{"budget-manager"}},
		[]types.Role{{Name: "budget-manager", Type: "custom", Permissions: []string{"p-5"}}},
	)
	foundRead := false
	for _, p := range perms {
		if p == permission.BudgetRead {
			foundRead = true
			break
		}
	}
	if !foundRead {
		t.Fatal("expected budget:manage to include budget:read via hierarchy")
	}
}

func TestPlatformAdminRoleGetsPlatformPermissions(t *testing.T) {
	t.Parallel()
	member := types.Member{Roles: []string{permission.RolePlatformAdmin}}
	roles := []types.Role{
		{Name: permission.RolePlatformAdmin, Type: "preset"},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	found := map[string]bool{}
	for _, p := range perms {
		found[p] = true
	}
	if !found[permission.PlatformManage] {
		t.Fatal("platform admin should have platform:manage")
	}
	if !found[permission.PlatformRead] {
		t.Fatal("platform admin should have platform:read")
	}
	if !found[permission.SelfKeys] {
		t.Fatal("platform admin should have self:keys")
	}
}

func TestPlatformReadRoleGetsPlatformRead(t *testing.T) {
	t.Parallel()
	member := types.Member{Roles: []string{permission.RolePlatformRead}}
	roles := []types.Role{
		{Name: permission.RolePlatformRead, Type: "preset"},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	found := map[string]bool{}
	for _, p := range perms {
		found[p] = true
	}
	if !found[permission.PlatformRead] {
		t.Fatal("platform reader should have platform:read")
	}
	if found[permission.PlatformManage] {
		t.Fatal("platform reader should NOT have platform:manage")
	}
}

func TestCustomRoleWildcardCannotGetPlatformPermissions(t *testing.T) {
	t.Parallel()
	member := types.Member{Roles: []string{"custom-admin"}}
	roles := []types.Role{
		{Name: "custom-admin", Type: "custom", Permissions: []string{"*"}},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	for _, p := range perms {
		if p == permission.PlatformManage || p == permission.PlatformRead {
			t.Fatalf("custom role with * should NOT get platform permission %q", p)
		}
	}
	// Should still get all company permissions
	if len(perms) != len(permission.CompanyPermissions) {
		t.Fatalf("expected %d company permissions, got %d", len(permission.CompanyPermissions), len(perms))
	}
}

func TestCustomRoleDirectPlatformManageIgnored(t *testing.T) {
	t.Parallel()
	// Attempt to directly specify platform:manage in a custom role
	member := types.Member{Roles: []string{"sneaky"}}
	roles := []types.Role{
		{Name: "sneaky", Type: "custom", Permissions: []string{"platform:manage", "platform:read", "self:keys"}},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	for _, p := range perms {
		if p == permission.PlatformManage || p == permission.PlatformRead {
			t.Fatalf("custom role should NOT resolve platform permission %q even when directly specified", p)
		}
	}
	// Only self:keys should resolve (it's in CompanyPermissions)
	if len(perms) != 1 || perms[0] != permission.SelfKeys {
		t.Fatalf("expected only [self:keys], got %v", perms)
	}
}

// --- scopePermissions tests ---

func TestScopePermissions_PlatformAdminCompanySaas_Preserves(t *testing.T) {
	t.Parallel()
	perms := []string{permission.PlatformManage, permission.PlatformRead, permission.SelfKeys}
	result := authz.ScopePermissions(perms, "platform", true)
	if len(result) != 3 {
		t.Fatalf("expected 3 permissions preserved, got %d: %v", len(result), result)
	}
}

func TestScopePermissions_TrialCompanySaas_FiltersPlatform(t *testing.T) {
	t.Parallel()
	perms := []string{permission.PlatformManage, permission.PlatformRead, permission.SelfKeys, permission.OrgRead}
	result := authz.ScopePermissions(perms, "trial", true)
	for _, p := range result {
		if permission.IsPlatformPermission(p) {
			t.Fatalf("trial company should not have platform permission %q", p)
		}
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 company permissions, got %d: %v", len(result), result)
	}
}

func TestScopePermissions_StandardCompanySaas_FiltersPlatform(t *testing.T) {
	t.Parallel()
	perms := []string{permission.PlatformManage, permission.SelfKeys}
	result := authz.ScopePermissions(perms, "standard", true)
	if len(result) != 1 || result[0] != permission.SelfKeys {
		t.Fatalf("expected [self:keys], got %v", result)
	}
}

func TestScopePermissions_NonSaasMode_AlwaysFiltersPlatform(t *testing.T) {
	t.Parallel()
	// Even platform company type in non-SaaS mode should NOT get platform permissions.
	perms := []string{permission.PlatformManage, permission.PlatformRead, permission.SelfKeys}
	result := authz.ScopePermissions(perms, "platform", false)
	for _, p := range result {
		if permission.IsPlatformPermission(p) {
			t.Fatalf("non-SaaS mode should filter platform permission %q", p)
		}
	}
	if len(result) != 1 || result[0] != permission.SelfKeys {
		t.Fatalf("expected [self:keys], got %v", result)
	}
}

func TestIsPlatformPermission(t *testing.T) {
	t.Parallel()
	cases := []struct {
		perm     string
		expected bool
	}{
		{"platform:manage", true},
		{"platform:read", true},
		{"platform:anything", true},
		{"org:read", false},
		{"self:keys", false},
		{"budget:read", false},
		{"platform", false},  // no colon — not a platform permission
		{"platform:", false}, // too short after prefix check
	}
	for _, tc := range cases {
		if got := permission.IsPlatformPermission(tc.perm); got != tc.expected {
			t.Errorf("IsPlatformPermission(%q) = %v, want %v", tc.perm, got, tc.expected)
		}
	}
}

func TestDirectPermissionsMergedIntoSession(t *testing.T) {
	t.Parallel()
	member := types.Member{
		Roles:             []string{permission.RolePlatformAdmin},
		DirectPermissions: []string{permission.PlatformAdmin},
	}
	roles := []types.Role{
		{Name: permission.RolePlatformAdmin, Type: "preset"},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	found := map[string]bool{}
	for _, p := range perms {
		found[p] = true
	}
	if !found[permission.PlatformAdmin] {
		t.Fatal("expected platform:admin from DirectPermissions")
	}
	if !found[permission.PlatformManage] {
		t.Fatal("expected platform:manage from role")
	}
	if !found[permission.PlatformRead] {
		t.Fatal("expected platform:read from role")
	}
}

func TestDirectPermissionsNotInRoleSystem(t *testing.T) {
	t.Parallel()
	// Even if someone sneaks platform:admin into a role definition, it won't resolve
	// because it's not in CompanyPermissions.
	member := types.Member{Roles: []string{"sneaky"}}
	roles := []types.Role{
		{Name: "sneaky", Type: "custom", Permissions: []string{"platform:admin", "self:keys"}},
	}

	perms := authz.ResolveMemberPermissions(member, roles)
	for _, p := range perms {
		if p == permission.PlatformAdmin {
			t.Fatal("platform:admin should NOT be resolvable through role system")
		}
	}
}

func TestScopePermissions_PlatformAdminPermPreservedForPlatformCompany(t *testing.T) {
	t.Parallel()
	perms := []string{permission.PlatformAdmin, permission.PlatformManage, permission.SelfKeys}
	result := authz.ScopePermissions(perms, "platform", true)
	if len(result) != 3 {
		t.Fatalf("expected all 3 preserved for platform company, got %d: %v", len(result), result)
	}
}

func TestScopePermissions_PlatformAdminPermFilteredForTrial(t *testing.T) {
	t.Parallel()
	perms := []string{permission.PlatformAdmin, permission.PlatformManage, permission.SelfKeys}
	result := authz.ScopePermissions(perms, "trial", true)
	if len(result) != 1 || result[0] != permission.SelfKeys {
		t.Fatalf("expected only [self:keys] for trial, got %v", result)
	}
}

func TestNormalizeGrantIDsRejectsPlatformAdmin(t *testing.T) {
	t.Parallel()
	_, err := permission.NormalizeGrantIDs([]string{"platform:admin"})
	if err == nil {
		t.Fatal("expected error when trying to normalize platform:admin in grant system")
	}
}
