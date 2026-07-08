package permission_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestNormalizeGrantIDsWildcard(t *testing.T) {
	t.Parallel()
	ids, err := permission.NormalizeGrantIDs([]string{"*"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 23 {
		t.Fatalf("expected 23 permission ids, got %d", len(ids))
	}
}

func TestNormalizeGrantIDsCapability(t *testing.T) {
	t.Parallel()
	ids, err := permission.NormalizeGrantIDs([]string{"audit:read"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "p-10" {
		t.Fatalf("got %v want [p-10]", ids)
	}
}

func TestPresetRolePermissionIDsMember(t *testing.T) {
	t.Parallel()
	ids, err := permission.PresetRolePermissionIDs(permission.RoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 grants for member, got %v", ids)
	}
}
