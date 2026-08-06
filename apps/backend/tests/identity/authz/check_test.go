package authz_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/identity/authz"
)

func TestHasAnyWildcard(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{"*"}, grants.OrgAdmin) {
		t.Fatal("expected wildcard to satisfy any permission")
	}
}

func TestHasAnySingleMatch(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{grants.OrgManage}, grants.OrgManage, grants.OrgAdmin) {
		t.Fatal("expected org:manage to match")
	}
}

func TestHasAnyNoMatch(t *testing.T) {
	t.Parallel()
	if authz.HasAny([]string{grants.SelfKeys}, grants.OrgAdmin) {
		t.Fatal("expected no match")
	}
}

func TestHasAnyEmptyRequired(t *testing.T) {
	t.Parallel()
	if !authz.HasAny([]string{grants.SelfKeys}) {
		t.Fatal("expected empty required to pass")
	}
}
