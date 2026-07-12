package authzscope_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/authzscope"
)

func TestHasAnyWildcard(t *testing.T) {
	t.Parallel()
	if !authzscope.HasAny([]string{"*"}, permission.OrgStructure) {
		t.Fatal("expected wildcard to match any permission")
	}
}

func TestHasAnyMatch(t *testing.T) {
	t.Parallel()
	if !authzscope.HasAny([]string{permission.OrgMembers}, permission.OrgMembers, permission.OrgRoles) {
		t.Fatal("expected match on org.members")
	}
}

func TestHasAnyMiss(t *testing.T) {
	t.Parallel()
	if authzscope.HasAny([]string{permission.SelfKeys}, permission.OrgStructure) {
		t.Fatal("expected no match")
	}
}

func TestHasAnyEmptyRequired(t *testing.T) {
	t.Parallel()
	if !authzscope.HasAny([]string{permission.SelfKeys}) {
		t.Fatal("empty required should return true")
	}
}
