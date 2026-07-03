package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func TestLocalDeptID(t *testing.T) {
	got := pkgorg.LocalDeptID(types.PlatformFeishu, "ou-abc")
	want := "dept-feishu-ou-abc"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalMemberID(t *testing.T) {
	got := pkgorg.LocalMemberID(types.PlatformFeishu, "ou-user")
	want := "m-feishu-ou-user"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestIsManualDeptSource(t *testing.T) {
	manual := types.DeptSourceManual
	imported := types.DeptSourceImported
	if !pkgorg.IsManualDeptSource(&manual) {
		t.Fatal("expected manual dept source")
	}
	if pkgorg.IsManualDeptSource(&imported) {
		t.Fatal("expected imported dept not manual")
	}
	if pkgorg.IsManualDeptSource(nil) {
		t.Fatal("expected nil source not manual")
	}
}

func TestIsManualMemberSource(t *testing.T) {
	if !pkgorg.IsManualMemberSource(types.MemberSourceManual) {
		t.Fatal("expected manual member source")
	}
	if pkgorg.IsManualMemberSource(types.MemberSourceImported) {
		t.Fatal("expected imported member not manual")
	}
}
