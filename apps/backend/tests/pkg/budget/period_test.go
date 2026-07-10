package budget_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

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
	parent := "root"
	nodes := []types.OrgNode{
		{ID: "root", Period: budget.PeriodMonthly},
		{ID: "child", ParentID: &parent, Period: "2026-06"},
	}
	got := budget.RootPeriodKey(nodes, at)
	if got != "2026-07" {
		t.Fatalf("RootPeriodKey() = %q, want 2026-07", got)
	}
	if budget.RootPeriodKey(nil, at) != "2026-07" {
		t.Fatal("empty tree should fall back to monthly at")
	}
}
