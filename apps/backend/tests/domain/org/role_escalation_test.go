package org_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

func TestAddRoleMemberRejectsProtectedPresetRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// role-1 is the preset super admin role seeded by default
	members, err := svc.ListMembers(ctx, "", "", false, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(members.Items) == 0 {
		t.Fatal("expected seeded members")
	}
	memberID := members.Items[0].ID

	err = svc.AddRoleMember(ctx, "role-1", memberID)
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

	// Create a custom role
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

	err = svc.AddRoleMember(ctx, role.ID, memberID)
	if err != nil {
		t.Fatalf("expected no error adding member to custom role, got: %v", err)
	}
}
