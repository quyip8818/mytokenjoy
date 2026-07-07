package usagequery_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store/usagequery"
)

var loc = time.UTC

func makeRow(bucket time.Time, dept, member, model string, cost float64, calls int) types.UsageBucketRow {
	return types.UsageBucketRow{
		BucketStart:  bucket,
		DepartmentID: dept,
		MemberID:     member,
		Model:        model,
		CostCNY:      cost,
		CallCount:    calls,
		InputTokens:  int64(calls * 100),
		OutputTokens: int64(calls * 50),
	}
}

func TestContainsString(t *testing.T) {
	items := []string{"a", "b", "c"}
	if !usagequery.ContainsString(items, "b") {
		t.Error("expected true for 'b'")
	}
	if usagequery.ContainsString(items, "d") {
		t.Error("expected false for 'd'")
	}
	if usagequery.ContainsString(nil, "a") {
		t.Error("expected false for nil slice")
	}
}

func TestFilterRows(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	rows := []types.UsageBucketRow{
		makeRow(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 10, 1),
		makeRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 20, 2),
		makeRow(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), "d2", "m2", "gpt-3.5", 30, 3),
		makeRow(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 40, 4),
	}

	t.Run("time range filter", func(t *testing.T) {
		result := usagequery.FilterRows(rows, start, end, "", "", nil, nil)
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
	})

	t.Run("department filter", func(t *testing.T) {
		result := usagequery.FilterRows(rows, start, end, "d1", "", nil, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("member filter", func(t *testing.T) {
		result := usagequery.FilterRows(rows, start, end, "", "m2", nil, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("scopeDeptIDs filter", func(t *testing.T) {
		result := usagequery.FilterRows(rows, start, end, "", "", []string{"d2"}, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("departmentIDs filter", func(t *testing.T) {
		result := usagequery.FilterRows(rows, start, end, "", "", nil, []string{"d1", "d2"})
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
	})
}

func TestTruncateBucket(t *testing.T) {
	ts := time.Date(2024, 3, 15, 14, 35, 22, 0, time.UTC)

	tests := []struct {
		name        string
		granularity string
		want        time.Time
	}{
		{"day", types.UsageGranularityDay, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{"hour", types.UsageGranularityHour, time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC)},
		{"minute", types.UsageGranularityMinute, time.Date(2024, 3, 15, 14, 35, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usagequery.TruncateBucket(ts, tt.granularity, loc)
			if !got.Equal(tt.want) {
				t.Errorf("TruncateBucket(%v, %q) = %v, want %v", ts, tt.granularity, got, tt.want)
			}
		})
	}
}

func TestSummaryTotals(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	rows := []types.UsageBucketRow{
		makeRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 10, 1),
		makeRow(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 20, 2),
		makeRow(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 30, 3), // excluded
	}

	totals := usagequery.SummaryTotals(rows, start, end)
	if totals.CostCNY != 30 {
		t.Errorf("CostCNY = %v, want 30", totals.CostCNY)
	}
	if totals.CallCount != 3 {
		t.Errorf("CallCount = %v, want 3", totals.CallCount)
	}
}

func TestLimitByCost(t *testing.T) {
	rows := []types.UsageAggregateRow{
		{CostCNY: 10},
		{CostCNY: 50},
		{CostCNY: 30},
		{CostCNY: 20},
	}

	t.Run("limits to top N", func(t *testing.T) {
		result := usagequery.LimitByCost(rows, 2)
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
		if result[0].CostCNY != 50 {
			t.Errorf("first should be highest cost 50, got %v", result[0].CostCNY)
		}
	})

	t.Run("zero limit returns all", func(t *testing.T) {
		result := usagequery.LimitByCost(rows, 0)
		if len(result) != 4 {
			t.Fatalf("expected 4, got %d", len(result))
		}
	})

	t.Run("limit larger than slice returns all", func(t *testing.T) {
		result := usagequery.LimitByCost(rows, 100)
		if len(result) != 4 {
			t.Fatalf("expected 4, got %d", len(result))
		}
	})
}

func TestTopModelPerDepartment(t *testing.T) {
	rows := []types.UsageBucketRow{
		makeRow(time.Now(), "d1", "m1", "gpt-4", 50, 5),
		makeRow(time.Now(), "d1", "m1", "gpt-3.5", 10, 10),
		makeRow(time.Now(), "d2", "m2", "claude", 30, 3),
		makeRow(time.Now(), "d2", "m2", "gpt-4", 20, 2),
	}

	result := usagequery.TopModelPerDepartment(rows, []string{"d1", "d2"})
	if result["d1"] != "gpt-4" {
		t.Errorf("d1 top model = %q, want 'gpt-4'", result["d1"])
	}
	if result["d2"] != "claude" {
		t.Errorf("d2 top model = %q, want 'claude'", result["d2"])
	}
}

func TestTopModelPerDepartmentEmpty(t *testing.T) {
	result := usagequery.TopModelPerDepartment(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestAggregateRows(t *testing.T) {
	rows := []types.UsageBucketRow{
		makeRow(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 10, 1),
		makeRow(time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), "d1", "m1", "gpt-4", 20, 2),
		makeRow(time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), "d2", "m2", "gpt-3.5", 30, 3),
	}

	t.Run("aggregate by day no group", func(t *testing.T) {
		result := usagequery.AggregateRows(rows, types.UsageGranularityDay, types.UsageGroupByNone, loc)
		if len(result) != 2 {
			t.Fatalf("expected 2 daily buckets, got %d", len(result))
		}
	})

	t.Run("aggregate by day group by department", func(t *testing.T) {
		result := usagequery.AggregateRows(rows, types.UsageGranularityDay, types.UsageGroupByDepartment, loc)
		if len(result) < 2 {
			t.Fatalf("expected at least 2 entries, got %d", len(result))
		}
	})

	t.Run("aggregate by model", func(t *testing.T) {
		result := usagequery.AggregateRows(rows, types.UsageGranularityDay, types.UsageGroupByModel, loc)
		if len(result) < 2 {
			t.Fatalf("expected at least 2 entries, got %d", len(result))
		}
	})
}

func TestSortSeriesPoints(t *testing.T) {
	points := []types.UsageSeriesPoint{
		{Bucket: "2024-01-02", DepartmentID: "d1"},
		{Bucket: "2024-01-01", DepartmentID: "d2"},
		{Bucket: "2024-01-01", DepartmentID: "d1"},
	}
	usagequery.SortSeriesPoints(points)
	if points[0].Bucket != "2024-01-01" || points[0].DepartmentID != "d1" {
		t.Errorf("sort failed: first = %+v", points[0])
	}
	if points[1].Bucket != "2024-01-01" || points[1].DepartmentID != "d2" {
		t.Errorf("sort failed: second = %+v", points[1])
	}
}
