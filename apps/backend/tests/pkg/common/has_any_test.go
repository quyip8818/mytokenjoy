package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestHasAnyWildcard(t *testing.T) {
	t.Parallel()
	if !common.HasAny([]string{"*"}, permission.OrgAdmin) {
		t.Fatal("expected wildcard to match any permission")
	}
}

func TestHasAnyMatch(t *testing.T) {
	t.Parallel()
	if !common.HasAny([]string{permission.OrgManage}, permission.OrgManage, permission.OrgAdmin) {
		t.Fatal("expected match on org:manage")
	}
}

func TestHasAnyMiss(t *testing.T) {
	t.Parallel()
	if common.HasAny([]string{permission.SelfKeys}, permission.OrgAdmin) {
		t.Fatal("expected no match")
	}
}

func TestHasAnyEmptyRequired(t *testing.T) {
	t.Parallel()
	if !common.HasAny([]string{permission.SelfKeys}) {
		t.Fatal("empty required should return true")
	}
}
