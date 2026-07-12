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

	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := budget.ResolveKeyPeriodKey(key, members, nil, deptPeriod, "2026-07", at)
	want := budget.SnapshotKey("2026-06", at)
	if got != want {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want %q", got, want)
	}
}

func TestResolveKeyPeriodKeyFallsBackToRoot(t *testing.T) {
	t.Parallel()
	key := types.PlatformKey{ID: "plk-1"}
	got := budget.ResolveKeyPeriodKey(key, nil, nil, map[string]string{}, "2026-06", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if got != "2026-06" {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want 2026-06", got)
	}
}

func TestResolveProjectPeriodKeysUsesOwnerDepartment(t *testing.T) {
	t.Parallel()
	project := types.Project{ID: "proj-1", OwnerDepartmentID: "dept-a"}
	deptPeriod := map[string]string{"dept-a": "2026-05", "dept-b": "2026-06"}
	at := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := budget.ResolveProjectPeriodKeys(project, nil, deptPeriod, "2026-07", at)
	want := budget.SnapshotKey("2026-05", at)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("ResolveProjectPeriodKeys() = %v, want [%q]", got, want)
	}
}

func TestResolveProjectPeriodKeysIncludesMemberDepartments(t *testing.T) {
	t.Parallel()
	project := types.Project{
		ID:                "proj-1",
		OwnerDepartmentID: "dept-a",
		MemberIDs:         []string{"m-1", "m-2"},
	}
	members := []types.Member{
		{ID: "m-1", DepartmentID: "dept-b"},
		{ID: "m-2", DepartmentID: "dept-c"},
	}
	deptPeriod := map[string]string{
		"dept-a": "2026-05",
		"dept-b": "2026-06",
		"dept-c": "2026-07",
	}
	at := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := budget.ResolveProjectPeriodKeys(project, members, deptPeriod, "2026-08", at)
	want := []string{
		budget.SnapshotKey("2026-05", at),
		budget.SnapshotKey("2026-06", at),
		budget.SnapshotKey("2026-07", at),
	}
	if len(got) != len(want) {
		t.Fatalf("ResolveProjectPeriodKeys() = %v, want %v", got, want)
	}
	seen := make(map[string]struct{}, len(got))
	for _, key := range got {
		seen[key] = struct{}{}
	}
	for _, key := range want {
		if _, ok := seen[key]; !ok {
			t.Fatalf("ResolveProjectPeriodKeys() = %v, missing %q", got, key)
		}
	}
}
