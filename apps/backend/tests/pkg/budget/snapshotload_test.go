package budget_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
)

func TestResolveKeyPeriodKeyUsesMemberDepartment(t *testing.T) {
	t.Parallel()
	deptA := uuid.MustParse("00000000-0000-7000-0000-00000000da01")
	memberID := uuid.MustParse("00000000-0000-7000-0000-00000000ee01")
	key := types.PlatformKey{ID: uuid.MustParse("00000000-0000-7000-0000-00000000f001"), MemberID: &memberID}
	members := []types.Member{{ID: memberID, DepartmentID: deptA}}
	deptPeriod := map[uuid.UUID]string{deptA: "2026-06"}

	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := budget.ResolveKeyPeriodKey(key, members, nil, deptPeriod, "2026-07", at)
	want := budget.SnapshotKey("2026-06", at)
	if got != want {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want %q", got, want)
	}
}

func TestResolveKeyPeriodKeyFallsBackToRoot(t *testing.T) {
	t.Parallel()
	key := types.PlatformKey{ID: uuid.MustParse("00000000-0000-7000-0000-00000000f001")}
	got := budget.ResolveKeyPeriodKey(key, nil, nil, map[uuid.UUID]string{}, "2026-06", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if got != "2026-06" {
		t.Fatalf("ResolveKeyPeriodKey() = %q, want 2026-06", got)
	}
}

func TestResolveProjectPeriodKeysUsesOwnerDepartment(t *testing.T) {
	t.Parallel()
	deptA := uuid.MustParse("00000000-0000-7000-0000-00000000da01")
	deptB := uuid.MustParse("00000000-0000-7000-0000-00000000da02")
	project := types.Project{ID: uuid.MustParse("00000000-0000-7000-0000-000000000a01"), OwnerDepartmentID: deptA}
	deptPeriod := map[uuid.UUID]string{deptA: "2026-05", deptB: "2026-06"}
	at := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := budget.ResolveProjectPeriodKeys(project, nil, deptPeriod, "2026-07", at)
	want := budget.SnapshotKey("2026-05", at)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("ResolveProjectPeriodKeys() = %v, want [%q]", got, want)
	}
}

func TestResolveProjectPeriodKeysIncludesMemberDepartments(t *testing.T) {
	t.Parallel()
	deptA := uuid.MustParse("00000000-0000-7000-0000-00000000da01")
	deptB := uuid.MustParse("00000000-0000-7000-0000-00000000da02")
	deptC := uuid.MustParse("00000000-0000-7000-0000-00000000da03")
	m1 := uuid.MustParse("00000000-0000-7000-0000-00000000ee01")
	m2 := uuid.MustParse("00000000-0000-7000-0000-00000000ee02")
	project := types.Project{
		ID:                uuid.MustParse("00000000-0000-7000-0000-000000000a01"),
		OwnerDepartmentID: deptA,
		MemberIDs:         []uuid.UUID{m1, m2},
	}
	members := []types.Member{
		{ID: m1, DepartmentID: deptB},
		{ID: m2, DepartmentID: deptC},
	}
	deptPeriod := map[uuid.UUID]string{
		deptA: "2026-05",
		deptB: "2026-06",
		deptC: "2026-07",
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
