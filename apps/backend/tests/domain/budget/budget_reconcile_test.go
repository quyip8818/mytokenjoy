package budget_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/budget"
)

func TestReconcileWindowStart(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	since := budget.ReconcileWindowStart(now)
	expected := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	if !since.Equal(expected) {
		t.Errorf("ReconcileWindowStart(%v) = %v, want %v", now, since, expected)
	}
}

func TestCollectPeriodKeys(t *testing.T) {
	t.Parallel()
	expected := map[budget.AxisKey]float64{
		{Kind: "platform_key", AxisID: uuid.MustParse("00000000-0000-7000-0000-000000000f01"), PeriodKey: "2026-07"}: 100,
		{Kind: "member", AxisID: uuid.MustParse("00000000-0000-7000-0000-000000000e01"), PeriodKey: "2026-07"}:       50,
		{Kind: "platform_key", AxisID: uuid.MustParse("00000000-0000-7000-0000-000000000f02"), PeriodKey: "2026-06"}: 200,
	}
	keys := budget.CollectPeriodKeys(expected)
	if len(keys) != 2 {
		t.Fatalf("expected 2 unique period keys, got %d: %v", len(keys), keys)
	}
	set := make(map[string]struct{})
	for _, k := range keys {
		set[k] = struct{}{}
	}
	if _, ok := set["2026-07"]; !ok {
		t.Error("missing 2026-07")
	}
	if _, ok := set["2026-06"]; !ok {
		t.Error("missing 2026-06")
	}
}

func TestSortedKeys(t *testing.T) {
	t.Parallel()
	m := map[uuid.UUID]struct{}{
		uuid.MustParse("00000000-0000-7000-0000-000000000c01"): {},
		uuid.MustParse("00000000-0000-7000-0000-000000000a01"): {},
		uuid.MustParse("00000000-0000-7000-0000-000000000b01"): {},
	}
	got := budget.SortedKeys(m)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	// Verify sorted by string representation
	for i := 1; i < len(got); i++ {
		if got[i-1].String() >= got[i].String() {
			t.Errorf("not sorted: got[%d]=%s >= got[%d]=%s", i-1, got[i-1], i, got[i])
		}
	}
}

func TestConsumedDrift(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expected float64
		actual   float64
		drift    bool
	}{
		{"exact match", 100.5, 100.5, false},
		{"within epsilon", 100.0, 100.0000001, false},
		{"positive drift", 100.0, 99.0, true},
		{"negative drift", 100.0, 101.0, true},
		{"zero vs zero", 0, 0, false},
		{"zero vs small", 0, 0.001, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := budget.ConsumedDrift(tc.expected, tc.actual)
			if got != tc.drift {
				t.Errorf("ConsumedDrift(%v, %v) = %v, want %v", tc.expected, tc.actual, got, tc.drift)
			}
		})
	}
}
