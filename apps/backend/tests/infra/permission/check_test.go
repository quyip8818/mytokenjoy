package permission_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestHasAnyWildcard(t *testing.T) {
	if !permission.HasAny([]string{"*"}, permission.OrgStructure) {
		t.Fatal("expected wildcard to satisfy any permission")
	}
}

func TestHasAnySingleMatch(t *testing.T) {
	if !permission.HasAny([]string{permission.OrgMembers}, permission.OrgMembers, permission.OrgRoles) {
		t.Fatal("expected org:members to match")
	}
}

func TestHasAnyNoMatch(t *testing.T) {
	if permission.HasAny([]string{permission.SelfKeys}, permission.OrgStructure) {
		t.Fatal("expected no match")
	}
}

func TestHasAnyEmptyRequired(t *testing.T) {
	if !permission.HasAny([]string{permission.SelfKeys}) {
		t.Fatal("expected empty required to pass")
	}
}
