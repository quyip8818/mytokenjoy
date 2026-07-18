package grants_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestNormalizerWildcardViaPort(t *testing.T) {
	t.Parallel()
	var normalizer grants.Normalizer = permission.NewGrantNormalizer()
	ids, err := normalizer.NormalizeGrantIDs([]string{"*"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 24 {
		t.Fatalf("expected 24 permission ids, got %d", len(ids))
	}
}

func TestNormalizerCapabilityViaPort(t *testing.T) {
	t.Parallel()
	var normalizer grants.Normalizer = permission.NewGrantNormalizer()
	ids, err := normalizer.NormalizeGrantIDs([]string{"audit:read"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "p-10" {
		t.Fatalf("got %v want [p-10]", ids)
	}
}

func TestRoleGrantIDsMemberPreset(t *testing.T) {
	t.Parallel()
	var normalizer grants.Normalizer = permission.NewGrantNormalizer()
	ids, err := normalizer.RoleGrantIDs("preset", grants.RoleMember, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 grants for member, got %v", ids)
	}
}
