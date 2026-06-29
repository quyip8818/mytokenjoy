package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newTestOrgService(t *testing.T) org.Service {
	t.Helper()
	_, _, svc := testutil.NewOrgServiceFromStore(t)
	return svc
}

func newTestOrgServiceWithStore(t *testing.T) (org.Service, store.Store) {
	t.Helper()
	_, st, svc := testutil.NewOrgServiceFromStore(t)
	return svc, st
}

func TestDeletePresetRoleReturns400(t *testing.T) {
	svc := newTestOrgService(t)
	err := svc.DeleteRole("role-1")
	if err == nil {
		t.Fatal("expected error deleting preset role")
	}
}

func TestListMembersPagination(t *testing.T) {
	svc := newTestOrgService(t)
	page := svc.ListMembers("", "", false, 1, 20)
	if len(page.Items) != 20 {
		t.Fatalf("expected 20 items, got %d", len(page.Items))
	}
	if page.Total < 120 {
		t.Fatalf("expected total >= 120, got %d", page.Total)
	}
}

func TestRemoveBaseMemberRoleReturns400(t *testing.T) {
	svc := newTestOrgService(t)
	err := svc.RemoveRoleMember("role-3", "m-1")
	if err == nil {
		t.Fatal("expected error removing base member role")
	}
}

func TestBatchImportUnknownDepartment(t *testing.T) {
	svc := newTestOrgService(t)
	result := svc.BatchImport([]types.BatchImportRow{
		{Name: "Test", Phone: "13800000000", Email: "t@example.com", DepartmentName: "不存在部门"},
	})
	if result.Imported != 0 {
		t.Fatalf("expected 0 imported, got %d", result.Imported)
	}
	if len(result.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(result.Failures))
	}
}

func TestCreateRoleAndList(t *testing.T) {
	svc := newTestOrgService(t)
	role, err := svc.CreateRole("测试角色", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if role.Name != "测试角色" {
		t.Fatalf("unexpected role name %s", role.Name)
	}
	found := false
	for _, r := range svc.ListRoles() {
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
	svc := newTestOrgService(t)
	role, err := svc.CreateRole("附加角色", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddRoleMember(role.ID, "m-3"); err != nil {
		t.Fatal(err)
	}
	for _, m := range svc.ListMembers("", "", false, 1, 200).Items {
		if m.ID != "m-3" {
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
	svc := newTestOrgService(t)
	role, err := svc.CreateRole("可移除角色", []string{"p-2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddRoleMember(role.ID, "m-3"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveRoleMember(role.ID, "m-3"); err != nil {
		t.Fatal(err)
	}
	for _, m := range svc.ListMembers("", "", false, 1, 200).Items {
		if m.ID != "m-3" {
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
