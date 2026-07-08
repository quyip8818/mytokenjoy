package org_test

import (
	"strings"
	"testing"

	"github.com/tokenjoy/backend/tests/testutil"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

func TestUpdateRoleRejectsPresetRole(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// The seeded store has preset roles (role-1 is super admin).
	_, err := svc.UpdateRole(ctx, "role-1", "Hacked Admin", []string{"*"})
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

	// Create a custom role, then update it.
	role, err := svc.CreateRole(ctx, "Custom Role", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateRole(ctx, role.ID, "Updated Custom", []string{"p-1", "p-2"})
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
