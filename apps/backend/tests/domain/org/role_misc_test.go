package org_test

import (
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAddRoleMemberNotFoundMember(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "TestRole", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.AddRoleMember(ctx, role.ID, "non-existent-member-id")
	if err == nil {
		t.Fatal("expected error for non-existent member, got nil")
	}
	de := asDomainError(t, err)
	if de.Status != 404 {
		t.Fatalf("expected status 404, got %d", de.Status)
	}
}

func TestDeleteRoleTransactional(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Create a role
	role, err := svc.CreateRole(ctx, "ToDelete", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	// Assign the role to a member
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) == 0 {
		t.Fatal("expected at least one seeded member")
	}
	memberID := members[0].ID

	err = svc.AddRoleMember(ctx, role.ID, memberID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the member has the role
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

	// Delete the role
	err = svc.DeleteRole(ctx, role.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify role is removed from roles list
	roles, err := st.Org().Roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range roles {
		if r.ID == role.ID {
			t.Fatal("role should have been deleted from roles list")
		}
	}

	// Verify role is removed from member
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
