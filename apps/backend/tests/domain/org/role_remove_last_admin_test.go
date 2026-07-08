package org_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestRemoveRoleMemberLastSuperAdmin(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	// role-1 is the preset super admin role; only m-admin holds it in seed data.
	err := svc.RemoveRoleMember(ctx, "role-1", contract.IDMemberAdmin)
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

	// Directly add super admin role to member1 via store (bypasses AddRoleMember protection)
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

	// Now removing one should succeed since there are 2 super admins
	if err := svc.RemoveRoleMember(ctx, "role-1", contract.IDMemberAdmin); err != nil {
		t.Fatalf("expected removal to succeed with multiple admins: %v", err)
	}
}
