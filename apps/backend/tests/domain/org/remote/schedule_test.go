package remote_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

func TestComputeNextOrgSyncRespectsFrequency(t *testing.T) {
	t.Parallel()
	clk := clock.Fixed(time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC))
	last := time.Date(2026, 6, 20, 8, 0, 0, 0, time.UTC)
	cfg := types.SyncConfig{Enabled: true, FrequencyHours: 6}

	next := remote.ComputeNextOrgSync(cfg, &last, clk)
	want := last.Add(6 * time.Hour)
	if !next.Equal(want) {
		t.Fatalf("next = %s, want %s", next, want)
	}
}

func TestComputeNextOrgSyncDisabledReturnsNow(t *testing.T) {
	t.Parallel()
	clk := clock.Fixed(time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC))
	cfg := types.SyncConfig{Enabled: false, FrequencyHours: 24}

	next := remote.ComputeNextOrgSync(cfg, nil, clk)
	if !next.Equal(clk.Now().UTC()) {
		t.Fatalf("disabled next = %s, want now %s", next, clk.Now().UTC())
	}
}

func TestComputeNextOrgSyncAlignsStartTime(t *testing.T) {
	t.Parallel()
	clk := clock.Fixed(time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC))
	cfg := types.SyncConfig{Enabled: true, FrequencyHours: 24, StartTime: "09:00"}

	next := remote.ComputeNextOrgSync(cfg, nil, clk)
	if next.Hour() != 9 || next.Minute() != 0 {
		t.Fatalf("expected 09:00 alignment, got %s", next)
	}
	if !next.After(clk.Now()) {
		t.Fatalf("expected next in the future, got %s", next)
	}
}
