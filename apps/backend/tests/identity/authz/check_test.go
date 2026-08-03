package authz_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestHasAnyWildcard(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{"*"}, permission.OrgAdmin) {
		t.Fatal("expected wildcard to satisfy any permission")
	}
}

func TestHasAnySingleMatch(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{permission.OrgManage}, permission.OrgManage, permission.OrgAdmin) {
		t.Fatal("expected org:manage to match")
	}
}

func TestHasAnyNoMatch(t *testing.T) {
	t.Parallel()
	if authz.HasAny([]string{permission.SelfKeys}, permission.OrgAdmin) {
		t.Fatal("expected no match")
	}
}

func TestHasAnyEmptyRequired(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{permission.SelfKeys}) {
		t.Fatal("expected empty required to pass")
	}
}
