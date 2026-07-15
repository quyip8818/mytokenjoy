package budget

import (
	"testing"
	"time"
)

func TestReconcileWindowStart(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	since := reconcileWindowStart(now)
	expected := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	if !since.Equal(expected) {
		t.Errorf("reconcileWindowStart(%v) = %v, want %v", now, since, expected)
	}
}

func TestCollectPeriodKeys(t *testing.T) {
	t.Parallel()
	expected := map[AxisKey]float64{
		{Kind: "platform_key", AxisID: "pk-1", PeriodKey: "2026-07"}: 100,
		{Kind: "member", AxisID: "m-1", PeriodKey: "2026-07"}:        50,
		{Kind: "platform_key", AxisID: "pk-2", PeriodKey: "2026-06"}: 200,
	}
	keys := collectPeriodKeys(expected)
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
	m := map[string]struct{}{
		"charlie": {},
		"alpha":   {},
		"bravo":   {},
	}
	got := sortedKeys(m)
	want := []string{"alpha", "bravo", "charlie"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, v, want[i])
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
			got := ConsumedDrift(tc.expected, tc.actual)
			if got != tc.drift {
				t.Errorf("ConsumedDrift(%v, %v) = %v, want %v", tc.expected, tc.actual, got, tc.drift)
			}
		})
	}
}
