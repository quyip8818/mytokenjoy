package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func TestLocalDeptID(t *testing.T) {
	t.Parallel()
	got := pkgorg.LocalDeptID(types.PlatformFeishu, "ou-abc")
	if got == uuid.Nil {
		t.Fatal("expected non-nil uuid")
	}
	// Deterministic: calling again with same input yields same result.
	got2 := pkgorg.LocalDeptID(types.PlatformFeishu, "ou-abc")
	if got != got2 {
		t.Fatalf("expected deterministic id, got %s vs %s", got, got2)
	}
	// Different input yields different result.
	got3 := pkgorg.LocalDeptID(types.PlatformFeishu, "ou-xyz")
	if got == got3 {
		t.Fatal("expected different ids for different inputs")
	}
}

func TestLocalMemberID(t *testing.T) {
	t.Parallel()
	got := pkgorg.LocalMemberID(types.PlatformFeishu, "ou-user")
	if got == uuid.Nil {
		t.Fatal("expected non-nil uuid")
	}
	// Deterministic.
	got2 := pkgorg.LocalMemberID(types.PlatformFeishu, "ou-user")
	if got != got2 {
		t.Fatalf("expected deterministic id, got %s vs %s", got, got2)
	}
}

func TestIsManualDeptSource(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	if !pkgorg.IsManualMemberSource(types.MemberSourceManual) {
		t.Fatal("expected manual member source")
	}
	if pkgorg.IsManualMemberSource(types.MemberSourceImported) {
		t.Fatal("expected imported member not manual")
	}
}
