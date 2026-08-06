package grants_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/grants"
)

func TestNormalizeGrantIDsWildcard(t *testing.T) {
	t.Parallel()
	ids, err := grants.NormalizeGrantIDs([]string{"*"})
	if err != nil {
		t.Fatal(err)
	}
	// * expands to all entries in PermissionIDMap (company permissions only, 19 total).
	if len(ids) != 19 {
		t.Fatalf("expected 19 permission ids, got %d", len(ids))
	}
}

func TestNormalizeGrantIDsCapability(t *testing.T) {
	t.Parallel()
	ids, err := grants.NormalizeGrantIDs([]string{"audit:read"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "p-16" {
		t.Fatalf("got %v want [p-16]", ids)
	}
}

func TestPresetRolePermissionIDsMember(t *testing.T) {
	t.Parallel()
	ids, err := grants.PresetRolePermissionIDs(grants.RoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 grants for member, got %v", ids)
	}
}

func TestPresetRolePermissionIDsPlatformAdmin(t *testing.T) {
	t.Parallel()
	// Platform roles contain platform:* capabilities that have no ID mapping.
	// PresetRolePermissionIDs should succeed, returning only the company-level IDs.
	ids, err := grants.PresetRolePermissionIDs(grants.RolePlatformAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 平台管理员 = ["platform:manage", "self:keys"]
	// Only self:keys has an ID mapping.
	if len(ids) != 1 {
		t.Fatalf("expected 1 company-level grant ID, got %d: %v", len(ids), ids)
	}
}

func TestPresetRolePermissionIDsPlatformRead(t *testing.T) {
	t.Parallel()
	ids, err := grants.PresetRolePermissionIDs(grants.RolePlatformRead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 平台只读 = ["platform:read", "self:keys"] → only self:keys has an ID.
	if len(ids) != 1 {
		t.Fatalf("expected 1 company-level grant ID, got %d: %v", len(ids), ids)
	}
}
