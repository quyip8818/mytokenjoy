package budget_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
)

func TestResolveKeyPeriodKeyUsesMemberDepartment(t *testing.T) {
	t.Parallel()
	const deptA = "dept-a"
	memberID := "m-1"
	key := types.PlatformKey{ID: "plk-1", MemberID: &memberID}
	members := []types.Member{{ID: memberID, DepartmentID: deptA}}
	deptPeriod := map[string]string{deptA: "2026-06"}

	got := budget.ResolveKeyPeriodKey(key, members, nil, deptPeriod, "2026-07")
	want := budget.SnapshotKey("2026-06", time.Now().UTC())
	if got != want {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want %q", got, want)
	}
}

func TestResolveKeyPeriodKeyFallsBackToRoot(t *testing.T) {
	t.Parallel()
	key := types.PlatformKey{ID: "plk-1"}
	got := budget.ResolveKeyPeriodKey(key, nil, nil, map[string]string{}, "2026-06")
	if got != "2026-06" {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want 2026-06", got)
	}
}

func TestResolveGroupPeriodKeysCollectsAllDepartments(t *testing.T) {
	t.Parallel()
	group := types.BudgetGroup{ID: "bg-1", DepartmentIDs: []string{"dept-a", "dept-b"}}
	deptPeriod := map[string]string{"dept-a": "2026-05", "dept-b": "2026-06"}
	got := budget.ResolveGroupPeriodKeys(group, deptPeriod, "2026-07")
	want := []string{
		budget.SnapshotKey("2026-05", time.Now().UTC()),
		budget.SnapshotKey("2026-06", time.Now().UTC()),
	}
	if len(got) != len(want) {
		t.Fatalf("ResolveGroupPeriodKeys() = %v, want %v", got, want)
	}
	seen := make(map[string]struct{}, len(got))
	for _, key := range got {
		seen[key] = struct{}{}
	}
	for _, key := range want {
		if _, ok := seen[key]; !ok {
			t.Fatalf("ResolveGroupPeriodKeys() = %v, missing %q", got, key)
		}
	}
}
