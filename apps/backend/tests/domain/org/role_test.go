package org_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

// ---------------------------------------------------------------------------
// CreateRole
// ---------------------------------------------------------------------------

func TestCreateRoleRejectsDuplicateName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	_, err := svc.CreateRole(ctx, "TestRole", []string{"p-1"})
	if err != nil {
		t.Fatalf("first CreateRole failed: %v", err)
	}

	_, err = svc.CreateRole(ctx, "TestRole", []string{"p-2"})
	if err == nil {
		t.Fatal("expected error for duplicate role name")
	}
	de := asDomainError(t, err)
	if !strings.Contains(de.Message, "already exists") {
		t.Fatalf("expected 'already exists' error, got: %s", de.Message)
	}
}

func TestCreateRoleRejectsEmptyName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	_, err := svc.CreateRole(ctx, "  ", []string{"p-1"})
	if err == nil {
		t.Fatal("expected error for whitespace-only role name")
	}
	de := asDomainError(t, err)
	if !strings.Contains(de.Message, "empty") {
		t.Fatalf("expected 'empty' error, got: %s", de.Message)
	}
}

func TestCreateRoleSucceedsWithUniqueName(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "UniqueRole", []string{"p-1"})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	if role.Name != "UniqueRole" {
		t.Fatalf("expected name 'UniqueRole', got '%s'", role.Name)
	}
}

// ---------------------------------------------------------------------------
// UpdateRole
// ---------------------------------------------------------------------------

func TestUpdateRolePersistsAndBumpsAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "Editable", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	before, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateRole(ctx, role.ID.String(), "Renamed Role", []string{"p-1", "p-2"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed Role" || len(updated.Permissions) != 2 {
		t.Fatalf("unexpected role %+v", updated)
	}
	after, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision bump, before=%d after=%d", before, after)
	}
}

func TestUpdateRolePresetNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.UpdateRole(testutil.Ctx(), "missing-role", "X", []string{"p-1"})
	if err == nil {
		t.Fatal("expected error for missing role")
	}
}

func TestUpdateRoleRejectsPresetRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	_, err := svc.UpdateRole(ctx, contract.IDRole1.String(), "Hacked Admin", []string{"*"})
	if err == nil {
		t.Fatal("expected error when updating preset role")
	}
	if !strings.Contains(err.Error(), "preset") {
		t.Fatalf("expected error about preset role, got: %v", err)
	}
}

func TestUpdateRoleAllowsCustomRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "Custom Role", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateRole(ctx, role.ID.String(), "Updated Custom", []string{"p-1", "p-2"})
	if err != nil {
		t.Fatalf("expected update to succeed for custom role, got: %v", err)
	}
	if updated.Name != "Updated Custom" {
		t.Fatalf("expected name 'Updated Custom', got %q", updated.Name)
	}
	if len(updated.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(updated.Permissions))
	}
}

// ---------------------------------------------------------------------------
// AddRoleMember / RemoveRoleMember
// ---------------------------------------------------------------------------

func TestAddRoleMemberRejectsProtectedPresetRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	members, err := svc.ListMembers(ctx, "", "", false, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(members.Items) == 0 {
		t.Fatal("expected seeded members")
	}
	memberID := members.Items[0].ID

	err = svc.AddRoleMember(ctx, contract.IDRole1.String(), memberID.String())
	if err == nil {
		t.Fatal("expected error when adding member to protected preset role")
	}
	if !strings.Contains(err.Error(), "protected") {
		t.Fatalf("expected error to mention 'protected', got: %v", err)
	}
}

func TestAddRoleMemberAllowsCustomRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "CustomRole", []string{"org:read"})
	if err != nil {
		t.Fatal(err)
	}

	members, err := svc.ListMembers(ctx, "", "", false, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(members.Items) == 0 {
		t.Fatal("expected seeded members")
	}
	memberID := members.Items[0].ID

	err = svc.AddRoleMember(ctx, role.ID.String(), memberID.String())
	if err != nil {
		t.Fatalf("expected no error adding member to custom role, got: %v", err)
	}
}

func TestAddRoleMemberNotFoundMember(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "TestRole", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.AddRoleMember(ctx, role.ID.String(), "non-existent-member-id")
	if err == nil {
		t.Fatal("expected error for non-existent member, got nil")
	}
	de := asDomainError(t, err)
	if de.Status != 404 {
		t.Fatalf("expected status 404, got %d", de.Status)
	}
}

func TestRemoveRoleMemberLastSuperAdmin(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	err := svc.RemoveRoleMember(ctx, contract.IDRole1.String(), contract.IDMemberAdmin.String())
	if err == nil {
		t.Fatal("expected error when removing the last super admin")
	}
	if !strings.Contains(err.Error(), "last super admin") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRemoveRoleMemberSuperAdminAllowedWhenMultiple(t *testing.T) {
	t.Parallel()
	svc, st := newTestOrgServiceWithStore(t)
	ctx := testutil.Ctx()

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range members {
		if members[i].ID == contract.IDMember1 {
			members[i].Roles = append(members[i].Roles, "超级管理员")
			break
		}
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}

	if err := svc.RemoveRoleMember(ctx, contract.IDRole1.String(), contract.IDMemberAdmin.String()); err != nil {
		t.Fatalf("expected removal to succeed with multiple admins: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DeleteRole
// ---------------------------------------------------------------------------

func TestDeleteRoleTransactional(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "ToDelete", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) == 0 {
		t.Fatal("expected at least one seeded member")
	}
	memberID := members[0].ID

	err = svc.AddRoleMember(ctx, role.ID.String(), memberID.String())
	if err != nil {
		t.Fatal(err)
	}

	members, err = st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range members {
		if m.ID == memberID {
			for _, r := range m.Roles {
				if r == "ToDelete" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("member should have the role before deletion")
	}

	err = svc.DeleteRole(ctx, role.ID.String())
	if err != nil {
		t.Fatal(err)
	}

	roles, err := st.Org().Roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range roles {
		if r.ID == role.ID {
			t.Fatal("role should have been deleted from roles list")
		}
	}

	members, err = st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range members {
		if m.ID == memberID {
			for _, r := range m.Roles {
				if r == "ToDelete" {
					t.Fatal("role should have been removed from member")
				}
			}
		}
	}
}
