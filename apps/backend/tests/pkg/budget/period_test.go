package budget_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

// --- cost range resolution ---

func TestResolveCurrentMonth(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	rng, err := budget.Resolve(types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, now, types.UsageDefaultTimezone)
	if err != nil {
		t.Fatal(err)
	}
	if rng.Timezone != types.UsageDefaultTimezone {
		t.Fatalf("expected timezone %s, got %s", types.UsageDefaultTimezone, rng.Timezone)
	}
	if rng.End.Sub(rng.Start) < 28*24*time.Hour {
		t.Fatalf("unexpected current month range: %+v", rng)
	}
}

func TestPreviousRange(t *testing.T) {
	t.Parallel()
	current := budget.ResolvedRange{
		Start: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	prev := budget.PreviousRange(current)
	if !prev.End.Equal(current.Start) {
		t.Fatalf("expected previous end at current start, got %+v", prev)
	}
}

// --- snapshot key ---

func TestSnapshotKeyFixedOrgPeriod(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := budget.SnapshotKey("2026-06", at)
	if got != "2026-06" {
		t.Fatalf("SnapshotKey() = %q, want 2026-06", got)
	}
}

func TestSnapshotKeyMonthlyFallback(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := budget.SnapshotKey("", at)
	if got != "2026-07" {
		t.Fatalf("SnapshotKey() = %q, want 2026-07", got)
	}
}

func TestSnapshotKeyMonthlySpec(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := budget.SnapshotKey(budget.PeriodMonthly, at)
	if got != "2026-07" {
		t.Fatalf("SnapshotKey() = %q, want 2026-07", got)
	}
}

func TestOpenBudgetPeriodIsZero(t *testing.T) {
	t.Parallel()
	var zero budget.OpenBudgetPeriod
	if !zero.IsZero() {
		t.Fatal("zero OpenBudgetPeriod should be zero")
	}
	var occZero budget.OccurrencePeriod
	if !occZero.IsZero() {
		t.Fatal("zero OccurrencePeriod should be zero")
	}
	p := budget.OpenSnapshotKey("2026-06", clock.System())
	if p.IsZero() || p.String() != "2026-06" {
		t.Fatalf("OpenSnapshotKey(fixed) = %+v", p)
	}
}

func TestOpenSnapshotKeyUsesClock(t *testing.T) {
	t.Parallel()
	clk := clock.Fixed(time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	got := budget.OpenSnapshotKey(budget.PeriodMonthly, clk)
	if got.String() != "2026-07" {
		t.Fatalf("OpenSnapshotKey() = %q, want 2026-07", got.String())
	}
	fixed := budget.OpenSnapshotKey("2026-06", clk)
	if fixed.String() != "2026-06" {
		t.Fatalf("OpenSnapshotKey(fixed) = %q, want 2026-06", fixed.String())
	}
}

func TestOccurrenceSnapshotKey(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	got := budget.OccurrenceSnapshotKey(budget.PeriodMonthly, at)
	if got.String() != "2026-01" {
		t.Fatalf("OccurrenceSnapshotKey() = %q, want 2026-01", got.String())
	}
}

func TestRootPeriodKey(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	rootID := uuid.MustParse("00000000-0000-7000-0000-00000000ff01")
	childID := uuid.MustParse("00000000-0000-7000-0000-00000000ff02")
	parentRef := rootID
	nodes := []types.OrgNode{
		{ID: rootID, Period: budget.PeriodMonthly},
		{ID: childID, ParentID: &parentRef, Period: "2026-06"},
	}
	got := budget.RootPeriodKey(nodes, at)
	if got != "2026-07" {
		t.Fatalf("RootPeriodKey() = %q, want 2026-07", got)
	}
	if budget.RootPeriodKey(nil, at) != "2026-07" {
		t.Fatal("empty tree should fall back to monthly at")
	}
}
