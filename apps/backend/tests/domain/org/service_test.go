package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestDeletePresetRoleReturns400(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	err := svc.DeleteRole(testutil.Ctx(), "role-1")
	if err == nil {
		t.Fatal("expected error deleting preset role")
	}
}

func TestListMembersPagination(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	page, err := svc.ListMembers(testutil.Ctx(), "", "", false, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 20 {
		t.Fatalf("expected 20 items, got %d", len(page.Items))
	}
	if page.Total < 35 {
		t.Fatalf("expected total >= 35, got %d", page.Total)
	}
}

func TestRemoveBaseMemberRoleReturns400(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	err := svc.RemoveRoleMember(testutil.Ctx(), "role-3", "m-1")
	if err == nil {
		t.Fatal("expected error removing base member role")
	}
}

func TestBatchImportUnknownDepartment(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	result, err := svc.BatchImport(testutil.Ctx(), []types.BatchImportRow{
		{Name: "Test", Phone: "13800000000", Email: "t@example.com", DepartmentName: "不存在部门"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 0 {
		t.Fatalf("expected 0 imported, got %d", result.Imported)
	}
	if len(result.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(result.Failures))
	}
}

func TestCreateRoleAndList(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()
	role, err := svc.CreateRole(ctx, "测试角色", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if role.Name != "测试角色" {
		t.Fatalf("unexpected role name %s", role.Name)
	}
	roles, err := svc.ListRoles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range roles {
		if r.ID == role.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created role not found in list")
	}
}

func TestAddRoleMember(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()
	role, err := svc.CreateRole(ctx, "附加角色", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddRoleMember(ctx, role.ID.String(), "m-3"); err != nil {
		t.Fatal(err)
	}
	page, err := svc.ListMembers(ctx, "", "", false, 1, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range page.Items {
		if m.ID != contract.IDMember3 {
			continue
		}
		found := false
		for _, r := range m.Roles {
			if r == "附加角色" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected m-3 to have role 附加角色, got %v", m.Roles)
		}
		return
	}
	t.Fatal("m-3 not found")
}

func TestRemoveRoleMemberSuccess(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()
	role, err := svc.CreateRole(ctx, "可移除角色", []string{"p-2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddRoleMember(ctx, role.ID.String(), "m-3"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveRoleMember(ctx, role.ID.String(), "m-3"); err != nil {
		t.Fatal(err)
	}
	page, err := svc.ListMembers(ctx, "", "", false, 1, 200)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range page.Items {
		if m.ID != contract.IDMember3 {
			continue
		}
		for _, r := range m.Roles {
			if r == "可移除角色" {
				t.Fatalf("expected role removed from m-3, still has %v", m.Roles)
			}
		}
		return
	}
	t.Fatal("m-3 not found")
}
